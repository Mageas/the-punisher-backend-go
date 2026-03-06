package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/mageas/the-punisher-backend/internal/api/handler"
)

func (app *application) mountPunishmentRoutes(r chi.Router, punishmentHandler *handler.PunishmentHandler, punishmentTypeHandler *handler.PunishmentTypeHandler) {
	r.Route("/punishment-types", func(r chi.Router) {
		r.Post("/", punishmentTypeHandler.CreatePunishmentType)
		r.Get("/", punishmentTypeHandler.ListPunishmentTypes)
		r.Get("/{punishment_type_id}", punishmentTypeHandler.GetPunishmentType)
		r.Put("/{punishment_type_id}", punishmentTypeHandler.UpdatePunishmentType)
		r.Delete("/{punishment_type_id}", punishmentTypeHandler.DeletePunishmentType)
	})

	r.Route("/punishments", func(r chi.Router) {
		r.Post("/", punishmentHandler.CreatePunishment)
		r.Get("/", punishmentHandler.ListPunishments)
		r.Get("/{punishment_id}", punishmentHandler.GetPunishment)
		r.Put("/{punishment_id}", punishmentHandler.UpdatePunishment)
		r.Post("/{punishment_id}/resolve", punishmentHandler.ResolvePunishment)
		r.Delete("/{punishment_id}", punishmentHandler.DeletePunishment)
	})
}
