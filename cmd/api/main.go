package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mageas/the-punisher-backend/internal/api/handler"
	"github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/config"
	"github.com/mageas/the-punisher-backend/internal/repository"
	"github.com/mageas/the-punisher-backend/internal/service"
)

func main() {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	conn, err := pgxpool.New(ctx, cfg.DB.DSN)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if err := conn.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	logger.Info("Connected to database")

	app := &application{
		config: *cfg,
		db:     conn,
	}

	if err := app.run(app.mount()); err != nil {
		log.Fatal(err)
	}
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	healthService := service.NewHealthService(&app.config, app.db)
	healthHandler := handler.NewHealthHandler(healthService)
	r.Get("/v1/health", healthHandler.GetHealth)

	repo := repository.New(app.db)

	userService := service.NewUserService(repo)
	userHandler := handler.NewUserHandler(userService, app.config)

	authService := service.NewAuthService(repo, app.config.JWT)
	authHandler := handler.NewAuthHandler(authService, app.config.JWT, "/v1/auth/refresh")

	r.Route("/v1/auth", func(r chi.Router) {
		r.Post("/register", userHandler.CreateUser)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
	})

	studentService := service.NewStudentService(repo)
	studentHandler := handler.NewStudentHandler(studentService)

	r.Route("/v1/students", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(app.config.JWT.AccessSecret))
		r.Post("/", studentHandler.CreateStudent)
		r.Get("/", studentHandler.ListStudents)
		r.Get("/{id}", studentHandler.GetStudent)
		r.Put("/{id}", studentHandler.UpdateStudent)
		r.Delete("/{id}", studentHandler.DeleteStudent)
	})

	return r
}

func (app *application) run(h http.Handler) error {
	srv := &http.Server{
		Addr:              app.config.Addr,
		Handler:           h,
		WriteTimeout:      30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 3 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	shutdown := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		s := <-quit

		slog.Info("Shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		shutdown <- srv.Shutdown(ctx)
	}()

	slog.Info("Starting server", "address", app.config.Addr)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdown
	if err != nil {
		return err
	}

	slog.Info("Server stopped gracefully")
	return nil
}

type application struct {
	config config.Config
	db     *pgxpool.Pool
}
