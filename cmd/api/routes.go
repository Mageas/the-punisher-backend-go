package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/mageas/the-punisher-backend/internal/api/handler"
	"github.com/mageas/the-punisher-backend/internal/platform/auth"
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
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

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

	r.Route("/v1/students", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(app.config.JWT.AccessSecret))
		r.Post("/", studentHandler.CreateStudent)
		r.Get("/", studentHandler.ListStudents)
		r.Get("/{id}", studentHandler.GetStudent)
		r.Get("/{id}/kpis", studentHandler.GetStudentKpis)
		r.Get("/{id}/history", studentHandler.GetStudentHistory)
		r.Put("/{id}", studentHandler.UpdateStudent)
		r.Delete("/{id}", studentHandler.DeleteStudent)
		r.Get("/{id}/classrooms", classroomHandler.ListClassroomsByStudent)
		r.Get("/{id}/bonuses", bonusHandler.ListBonusesByStudent)
		r.Get("/{id}/penalties", penaltyHandler.ListPenaltiesByStudent)
		r.Get("/{id}/punishments", punishmentHandler.ListPunishmentsByStudent)
	})

	r.Route("/v1/classrooms", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(app.config.JWT.AccessSecret))
		r.Post("/", classroomHandler.CreateClassroom)
		r.Get("/", classroomHandler.ListClassrooms)
		r.Get("/{id}", classroomHandler.GetClassroom)
		r.Put("/{id}", classroomHandler.UpdateClassroom)
		r.Delete("/{id}", classroomHandler.DeleteClassroom)
		r.Post("/{id}/students", classroomHandler.AddStudentToClassroom)
		r.Delete("/{id}/students/{studentId}", classroomHandler.RemoveStudentFromClassroom)
		r.Get("/{id}/students", classroomHandler.ListStudentsByClassroom)
	})

	bonusTypeService := service.NewBonusTypeService(repo)
	bonusTypeHandler := handler.NewBonusTypeHandler(bonusTypeService)

	penaltyTypeService := service.NewPenaltyTypeService(repo)
	penaltyTypeHandler := handler.NewPenaltyTypeHandler(penaltyTypeService)

	punishmentTypeService := service.NewPunishmentTypeService(repo)
	punishmentTypeHandler := handler.NewPunishmentTypeHandler(punishmentTypeService)

	r.Route("/v1/bonus-types", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(app.config.JWT.AccessSecret))
		r.Post("/", bonusTypeHandler.CreateBonusType)
		r.Get("/", bonusTypeHandler.ListBonusTypes)
		r.Get("/{id}", bonusTypeHandler.GetBonusType)
		r.Put("/{id}", bonusTypeHandler.UpdateBonusType)
		r.Delete("/{id}", bonusTypeHandler.DeleteBonusType)
	})

	r.Route("/v1/penalty-types", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(app.config.JWT.AccessSecret))
		r.Post("/", penaltyTypeHandler.CreatePenaltyType)
		r.Get("/", penaltyTypeHandler.ListPenaltyTypes)
		r.Get("/{id}", penaltyTypeHandler.GetPenaltyType)
		r.Put("/{id}", penaltyTypeHandler.UpdatePenaltyType)
		r.Delete("/{id}", penaltyTypeHandler.DeletePenaltyType)
	})

	r.Route("/v1/punishment-types", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(app.config.JWT.AccessSecret))
		r.Post("/", punishmentTypeHandler.CreatePunishmentType)
		r.Get("/", punishmentTypeHandler.ListPunishmentTypes)
		r.Get("/{id}", punishmentTypeHandler.GetPunishmentType)
		r.Put("/{id}", punishmentTypeHandler.UpdatePunishmentType)
		r.Delete("/{id}", punishmentTypeHandler.DeletePunishmentType)
	})

	r.Route("/v1/bonuses", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(app.config.JWT.AccessSecret))
		r.Post("/", bonusHandler.CreateBonus)
		r.Get("/", bonusHandler.ListBonuses)
		r.Get("/{id}", bonusHandler.GetBonus)
		r.Post("/{id}/use", bonusHandler.UseBonus)
		r.Delete("/{id}", bonusHandler.DeleteBonus)
	})

	r.Route("/v1/penalties", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(app.config.JWT.AccessSecret))
		r.Post("/", penaltyHandler.CreatePenalty)
		r.Get("/", penaltyHandler.ListPenalties)
		r.Get("/{id}", penaltyHandler.GetPenalty)
		r.Delete("/{id}", penaltyHandler.DeletePenalty)
	})

	r.Route("/v1/punishments", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(app.config.JWT.AccessSecret))
		r.Post("/", punishmentHandler.CreatePunishment)
		r.Get("/", punishmentHandler.ListPunishments)
		r.Get("/{id}", punishmentHandler.GetPunishment)
		r.Post("/{id}/resolve", punishmentHandler.ResolvePunishment)
		r.Delete("/{id}", punishmentHandler.DeletePunishment)
	})

	r.Route("/v1/rules", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(app.config.JWT.AccessSecret))
		r.Post("/", ruleHandler.CreateRule)
		r.Get("/", ruleHandler.ListRules)
		r.Get("/{id}", ruleHandler.GetRule)
		r.Put("/{id}", ruleHandler.UpdateRule)
		r.Delete("/{id}", ruleHandler.DeleteRule)
	})

	r.Route("/v1/dashboard", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(app.config.JWT.AccessSecret))
		r.Get("/", dashboardHandler.GetDashboard)
	})

	return r
}
