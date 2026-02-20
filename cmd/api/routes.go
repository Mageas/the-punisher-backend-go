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
		AllowedOrigins:   app.config.CORS.AllowedOrigins,
		AllowedMethods:   app.config.CORS.AllowedMethods,
		AllowedHeaders:   app.config.CORS.AllowedHeaders,
		ExposedHeaders:   app.config.CORS.ExposedHeaders,
		AllowCredentials: app.config.CORS.AllowCredentials,
		MaxAge:           app.config.CORS.MaxAge, // Maximum value not ignored by any of major browsers
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

	r.Route("/v1/user", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(app.config.JWT.AccessSecret, app.config.JWT.Issuer, app.config.JWT.Audience))
		r.Get("/me", userHandler.GetMe)
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
		r.Use(auth.AuthMiddleware(app.config.JWT.AccessSecret, app.config.JWT.Issuer, app.config.JWT.Audience))
		r.Post("/", studentHandler.CreateStudent)
		r.Get("/", studentHandler.ListStudents)
		r.Get("/{student_id}", studentHandler.GetStudent)
		r.Get("/{student_id}/kpis", studentHandler.GetStudentKpis)
		r.Get("/{student_id}/history", studentHandler.GetStudentHistory)
		r.Put("/{student_id}", studentHandler.UpdateStudent)
		r.Delete("/{student_id}", studentHandler.DeleteStudent)
		r.Get("/{student_id}/classrooms", classroomHandler.ListClassroomsByStudent)
		r.Get("/{student_id}/bonuses", bonusHandler.ListBonusesByStudent)
		r.Get("/{student_id}/penalties", penaltyHandler.ListPenaltiesByStudent)
		r.Get("/{student_id}/punishments", punishmentHandler.ListPunishmentsByStudent)
	})

	r.Route("/v1/classrooms", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(app.config.JWT.AccessSecret, app.config.JWT.Issuer, app.config.JWT.Audience))
		r.Post("/", classroomHandler.CreateClassroom)
		r.Get("/", classroomHandler.ListClassrooms)
		r.Get("/{classroom_id}", classroomHandler.GetClassroom)
		r.Put("/{classroom_id}", classroomHandler.UpdateClassroom)
		r.Delete("/{classroom_id}", classroomHandler.DeleteClassroom)
		r.Post("/{classroom_id}/students", classroomHandler.AddStudentToClassroom)
		r.Delete("/{classroom_id}/students/{student_id}", classroomHandler.RemoveStudentFromClassroom)
		r.Get("/{classroom_id}/students", classroomHandler.ListStudentsByClassroom)
	})

	bonusTypeService := service.NewBonusTypeService(repo)
	bonusTypeHandler := handler.NewBonusTypeHandler(bonusTypeService)

	penaltyTypeService := service.NewPenaltyTypeService(repo)
	penaltyTypeHandler := handler.NewPenaltyTypeHandler(penaltyTypeService)

	punishmentTypeService := service.NewPunishmentTypeService(repo)
	punishmentTypeHandler := handler.NewPunishmentTypeHandler(punishmentTypeService)

	r.Route("/v1/bonus-types", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(app.config.JWT.AccessSecret, app.config.JWT.Issuer, app.config.JWT.Audience))
		r.Post("/", bonusTypeHandler.CreateBonusType)
		r.Get("/", bonusTypeHandler.ListBonusTypes)
		r.Get("/{bonus_type_id}", bonusTypeHandler.GetBonusType)
		r.Put("/{bonus_type_id}", bonusTypeHandler.UpdateBonusType)
		r.Delete("/{bonus_type_id}", bonusTypeHandler.DeleteBonusType)
	})

	r.Route("/v1/penalty-types", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(app.config.JWT.AccessSecret, app.config.JWT.Issuer, app.config.JWT.Audience))
		r.Post("/", penaltyTypeHandler.CreatePenaltyType)
		r.Get("/", penaltyTypeHandler.ListPenaltyTypes)
		r.Get("/{penalty_type_id}", penaltyTypeHandler.GetPenaltyType)
		r.Put("/{penalty_type_id}", penaltyTypeHandler.UpdatePenaltyType)
		r.Delete("/{penalty_type_id}", penaltyTypeHandler.DeletePenaltyType)
	})

	r.Route("/v1/punishment-types", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(app.config.JWT.AccessSecret, app.config.JWT.Issuer, app.config.JWT.Audience))
		r.Post("/", punishmentTypeHandler.CreatePunishmentType)
		r.Get("/", punishmentTypeHandler.ListPunishmentTypes)
		r.Get("/{punishment_type_id}", punishmentTypeHandler.GetPunishmentType)
		r.Put("/{punishment_type_id}", punishmentTypeHandler.UpdatePunishmentType)
		r.Delete("/{punishment_type_id}", punishmentTypeHandler.DeletePunishmentType)
	})

	r.Route("/v1/bonuses", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(app.config.JWT.AccessSecret, app.config.JWT.Issuer, app.config.JWT.Audience))
		r.Post("/", bonusHandler.CreateBonus)
		r.Get("/", bonusHandler.ListBonuses)
		r.Get("/{bonus_id}", bonusHandler.GetBonus)
		r.Post("/{bonus_id}/use", bonusHandler.UseBonus)
		r.Delete("/{bonus_id}", bonusHandler.DeleteBonus)
	})

	r.Route("/v1/penalties", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(app.config.JWT.AccessSecret, app.config.JWT.Issuer, app.config.JWT.Audience))
		r.Post("/", penaltyHandler.CreatePenalty)
		r.Get("/", penaltyHandler.ListPenalties)
		r.Get("/{penalty_id}", penaltyHandler.GetPenalty)
		r.Delete("/{penalty_id}", penaltyHandler.DeletePenalty)
	})

	r.Route("/v1/punishments", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(app.config.JWT.AccessSecret, app.config.JWT.Issuer, app.config.JWT.Audience))
		r.Post("/", punishmentHandler.CreatePunishment)
		r.Get("/", punishmentHandler.ListPunishments)
		r.Get("/{punishment_id}", punishmentHandler.GetPunishment)
		r.Post("/{punishment_id}/resolve", punishmentHandler.ResolvePunishment)
		r.Delete("/{punishment_id}", punishmentHandler.DeletePunishment)
	})

	r.Route("/v1/rules", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(app.config.JWT.AccessSecret, app.config.JWT.Issuer, app.config.JWT.Audience))
		r.Post("/", ruleHandler.CreateRule)
		r.Get("/", ruleHandler.ListRules)
		r.Get("/{rule_id}", ruleHandler.GetRule)
		r.Put("/{rule_id}", ruleHandler.UpdateRule)
		r.Delete("/{rule_id}", ruleHandler.DeleteRule)
	})

	r.Route("/v1/dashboard", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(app.config.JWT.AccessSecret, app.config.JWT.Issuer, app.config.JWT.Audience))
		r.Get("/", dashboardHandler.GetDashboard)
	})

	return r
}
