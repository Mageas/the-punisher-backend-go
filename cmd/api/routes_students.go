package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/mageas/the-punisher-backend/internal/api/handler"
)

func (app *application) mountStudentRoutes(
	r chi.Router,
	studentHandler *handler.StudentHandler,
	classroomHandler *handler.ClassroomHandler,
	bonusHandler *handler.BonusHandler,
	penaltyHandler *handler.PenaltyHandler,
	punishmentHandler *handler.PunishmentHandler,
) {
	r.Route("/students", func(r chi.Router) {
		r.Post("/", studentHandler.CreateStudent)
		r.Get("/", studentHandler.ListStudents)
		r.Delete("/", studentHandler.DeleteAllStudents)
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
}
