package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/mageas/the-punisher-backend/internal/api/handler"
)

func (app *application) mountUserRoutes(r chi.Router, userHandler *handler.UserHandler) {
	r.Route("/user", func(r chi.Router) {
		r.Get("/me", userHandler.GetMe)
	})
}
