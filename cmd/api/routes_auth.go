package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/mageas/the-punisher-backend/internal/api/handler"
)

func (app *application) mountAuthRoutes(r chi.Router, userHandler *handler.UserHandler, authHandler *handler.AuthHandler) {
	r.Route("/v1/auth", func(r chi.Router) {
		r.Post("/register", userHandler.CreateUser)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
	})
}
