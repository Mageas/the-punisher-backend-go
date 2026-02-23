package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/mageas/the-punisher-backend/internal/api/handler"
)

func (app *application) mountClassroomRoutes(r chi.Router, classroomHandler *handler.ClassroomHandler) {
	r.Route("/classrooms", func(r chi.Router) {
		r.Post("/", classroomHandler.CreateClassroom)
		r.Get("/", classroomHandler.ListClassrooms)
		r.Get("/{classroom_id}", classroomHandler.GetClassroom)
		r.Get("/{classroom_id}/kpis", classroomHandler.GetClassroomKpis)
		r.Put("/{classroom_id}", classroomHandler.UpdateClassroom)
		r.Delete("/{classroom_id}", classroomHandler.DeleteClassroom)
		r.Post("/{classroom_id}/students", classroomHandler.AddStudentToClassroom)
		r.Delete("/{classroom_id}/students/{student_id}", classroomHandler.RemoveStudentFromClassroom)
		r.Get("/{classroom_id}/students", classroomHandler.ListStudentsByClassroom)
	})
}
