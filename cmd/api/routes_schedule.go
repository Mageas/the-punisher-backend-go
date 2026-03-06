package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/mageas/the-punisher-backend/internal/api/handler"
)

func (app *application) mountScheduleRoutes(r chi.Router, scheduleHandler *handler.ScheduleHandler) {
	r.Route("/schedule", func(r chi.Router) {
		r.Route("/slots", func(r chi.Router) {
			r.Post("/", scheduleHandler.CreateScheduleSlot)
			r.Get("/", scheduleHandler.ListScheduleSlots)
			r.Get("/{schedule_slot_id}", scheduleHandler.GetScheduleSlot)
			r.Put("/{schedule_slot_id}", scheduleHandler.UpdateScheduleSlot)
			r.Delete("/{schedule_slot_id}", scheduleHandler.DeleteScheduleSlot)
		})

		r.Route("/exceptions", func(r chi.Router) {
			r.Post("/", scheduleHandler.CreateScheduleException)
			r.Get("/", scheduleHandler.ListScheduleExceptions)
			r.Get("/{schedule_exception_id}", scheduleHandler.GetScheduleException)
			r.Put("/{schedule_exception_id}", scheduleHandler.UpdateScheduleException)
			r.Delete("/{schedule_exception_id}", scheduleHandler.DeleteScheduleException)
		})
	})
}
