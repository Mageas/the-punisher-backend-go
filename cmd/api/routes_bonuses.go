package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/mageas/the-punisher-backend/internal/api/handler"
)

func (app *application) mountBonusRoutes(r chi.Router, bonusHandler *handler.BonusHandler, bonusTypeHandler *handler.BonusTypeHandler) {
	r.Route("/bonus-types", func(r chi.Router) {
		r.Post("/", bonusTypeHandler.CreateBonusType)
		r.Get("/", bonusTypeHandler.ListBonusTypes)
		r.Get("/{bonus_type_id}", bonusTypeHandler.GetBonusType)
		r.Put("/{bonus_type_id}", bonusTypeHandler.UpdateBonusType)
		r.Delete("/{bonus_type_id}", bonusTypeHandler.DeleteBonusType)
	})

	r.Route("/bonuses", func(r chi.Router) {
		r.Post("/", bonusHandler.CreateBonus)
		r.Get("/", bonusHandler.ListBonuses)
		r.Get("/{bonus_id}", bonusHandler.GetBonus)
		r.Put("/{bonus_id}", bonusHandler.UpdateBonus)
		r.Post("/{bonus_id}/use", bonusHandler.UseBonus)
		r.Delete("/{bonus_id}", bonusHandler.DeleteBonus)
	})
}
