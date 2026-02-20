package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/mageas/the-punisher-backend/internal/api/handler"
)

func (app *application) mountDashboardRoutes(r chi.Router, dashboardHandler *handler.DashboardHandler) {
	r.Route("/dashboard", func(r chi.Router) {
		r.Get("/", dashboardHandler.GetDashboard)
	})
}
