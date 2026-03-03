package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mageas/the-punisher-backend/internal/api/handler"
)

func (app *application) mountAuthRoutes(
	r chi.Router,
	userHandler *handler.UserHandler,
	authHandler *handler.AuthHandler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/v1/auth", func(r chi.Router) {
		r.Get("/register/status", userHandler.GetRegisterStatus)
		r.Post("/register", userHandler.CreateUser)
		r.Get("/confirm-email", userHandler.ConfirmEmail)
		r.Post("/confirm-email/resend", userHandler.ResendConfirmEmail)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
		r.Post("/logout", authHandler.Logout)
	})

	r.With(authMiddleware).Post("/v1/auth/change-password", authHandler.ChangePassword)
	r.With(authMiddleware).Delete("/v1/auth/refresh-tokens", authHandler.LogoutAll)
}
