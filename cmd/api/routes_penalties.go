package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/mageas/the-punisher-backend/internal/api/handler"
)

func (app *application) mountPenaltyRoutes(r chi.Router, penaltyHandler *handler.PenaltyHandler, penaltyTypeHandler *handler.PenaltyTypeHandler) {
	r.Route("/penalty-types", func(r chi.Router) {
		r.Post("/", penaltyTypeHandler.CreatePenaltyType)
		r.Get("/", penaltyTypeHandler.ListPenaltyTypes)
		r.Get("/{penalty_type_id}", penaltyTypeHandler.GetPenaltyType)
		r.Put("/{penalty_type_id}", penaltyTypeHandler.UpdatePenaltyType)
		r.Delete("/{penalty_type_id}", penaltyTypeHandler.DeletePenaltyType)
	})

	r.Route("/penalties", func(r chi.Router) {
		r.Post("/", penaltyHandler.CreatePenalty)
		r.Get("/", penaltyHandler.ListPenalties)
		r.Get("/{penalty_id}", penaltyHandler.GetPenalty)
		r.Put("/{penalty_id}", penaltyHandler.UpdatePenalty)
		r.Delete("/{penalty_id}", penaltyHandler.DeletePenalty)
	})
}
