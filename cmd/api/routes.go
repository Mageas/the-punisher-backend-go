package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/api/handler"
	"github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/mailer"
	"github.com/mageas/the-punisher-backend/internal/platform/web"
	"github.com/mageas/the-punisher-backend/internal/repository"
	"github.com/mageas/the-punisher-backend/internal/service"
)

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   app.config.CORS.AllowedOrigins,
		AllowedMethods:   app.config.CORS.AllowedMethods,
		AllowedHeaders:   app.config.CORS.AllowedHeaders,
		ExposedHeaders:   app.config.CORS.ExposedHeaders,
		AllowCredentials: app.config.CORS.AllowCredentials,
		MaxAge:           app.config.CORS.MaxAge, // Maximum value not ignored by any of major browsers
	}))

	r.Use(middleware.Timeout(60 * time.Second))

	// Custom 404 Handler
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		web.WriteAPIError(w, api.ErrNotFound, nil)
	})

	healthService := service.NewHealthService(&app.config, app.db)
	healthHandler := handler.NewHealthHandler(healthService)
	r.Get("/v1/health", healthHandler.GetHealth)

	repo := repository.New(app.db)

	smtpMailer := mailer.NewSMTPMailer(app.config.SMTP)
	userService := service.NewUserServiceWithEmailConfirmation(repo, app.config.EmailConfirm, smtpMailer)
	userHandler := handler.NewUserHandler(userService, app.config)

	authService := service.NewAuthService(repo, app.config.JWT)
	authHandler := handler.NewAuthHandler(authService, app.config.JWT, "/v1/auth")
	authMiddleware := auth.AuthMiddleware(app.config.JWT.AccessSecret, app.config.JWT.Issuer, app.config.JWT.Audience)

	app.mountAuthRoutes(r, userHandler, authHandler, authMiddleware)

	studentService := service.NewStudentService(repo)
	studentHandler := handler.NewStudentHandler(studentService)

	classroomService := service.NewClassroomService(repo)
	classroomHandler := handler.NewClassroomHandler(classroomService)

	bonusService := service.NewBonusService(repo)
	bonusHandler := handler.NewBonusHandler(bonusService)

	penaltyService := service.NewPenaltyService(repo)
	penaltyHandler := handler.NewPenaltyHandler(penaltyService)

	punishmentService := service.NewPunishmentService(repo)
	punishmentHandler := handler.NewPunishmentHandler(punishmentService)

	ruleService := service.NewRuleService(repo)
	ruleHandler := handler.NewRuleHandler(ruleService)

	dashboardService := service.NewDashboardService(repo)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)

	bonusTypeService := service.NewBonusTypeService(repo)
	bonusTypeHandler := handler.NewBonusTypeHandler(bonusTypeService)

	penaltyTypeService := service.NewPenaltyTypeService(repo)
	penaltyTypeHandler := handler.NewPenaltyTypeHandler(penaltyTypeService)

	punishmentTypeService := service.NewPunishmentTypeService(repo)
	punishmentTypeHandler := handler.NewPunishmentTypeHandler(punishmentTypeService)

	r.Route("/v1", func(r chi.Router) {
		r.Use(authMiddleware)

		app.mountUserRoutes(r, userHandler)
		app.mountStudentRoutes(r, studentHandler, classroomHandler, bonusHandler, penaltyHandler, punishmentHandler)
		app.mountClassroomRoutes(r, classroomHandler)
		app.mountBonusRoutes(r, bonusHandler, bonusTypeHandler)
		app.mountPenaltyRoutes(r, penaltyHandler, penaltyTypeHandler)
		app.mountPunishmentRoutes(r, punishmentHandler, punishmentTypeHandler)
		app.mountRuleRoutes(r, ruleHandler)
		app.mountDashboardRoutes(r, dashboardHandler)
	})

	return r
}
