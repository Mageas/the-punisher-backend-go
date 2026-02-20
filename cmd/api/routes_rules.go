package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/mageas/the-punisher-backend/internal/api/handler"
)

func (app *application) mountRuleRoutes(r chi.Router, ruleHandler *handler.RuleHandler) {
	r.Route("/rules", func(r chi.Router) {
		r.Post("/", ruleHandler.CreateRule)
		r.Get("/", ruleHandler.ListRules)
		r.Get("/{rule_id}", ruleHandler.GetRule)
		r.Put("/{rule_id}", ruleHandler.UpdateRule)
		r.Delete("/{rule_id}", ruleHandler.DeleteRule)
	})
}
