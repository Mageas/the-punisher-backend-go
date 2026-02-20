package handler_test

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api/handler"
	platformauth "github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/config"
	"github.com/mageas/the-punisher-backend/internal/service"
	"github.com/mageas/the-punisher-backend/internal/testutil/inmemory"
)

type classroomResponse struct {
	ID                uuid.UUID                       `json:"id"`
	Name              string                          `json:"name"`
	Year              *string                         `json:"year"`
	MainTeacher       *string                         `json:"main_teacher"`
	StudentCount      int64                           `json:"student_count"`
	StudentsPreview   []classroomStudentPreviewRecord `json:"students_preview"`
	TotalBonusPoints  float64                         `json:"total_bonus_points"`
	TotalPenaltyCount int64                           `json:"total_penalty_count"`
}

type classroomStudentPreviewRecord struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
}

type paginatedClassroomResponse struct {
	Page       int                 `json:"page"`
	TotalCount int64               `json:"total_count"`
	Data       []classroomResponse `json:"data"`
}

func newClassroomRouter(repo *inmemory.Repository, cfg config.JWTConfig) http.Handler {
	classroomSvc := service.NewClassroomService(repo)
	classroomHandler := handler.NewClassroomHandler(classroomSvc)

	r := chi.NewRouter()
	r.Use(platformauth.AuthMiddleware(cfg.AccessSecret, cfg.Issuer, cfg.Audience))

	r.Route("/v1/classrooms", func(r chi.Router) {
		r.Post("/", classroomHandler.CreateClassroom)
		r.Get("/", classroomHandler.ListClassrooms)
		r.Get("/{classroom_id}", classroomHandler.GetClassroom)
		r.Put("/{classroom_id}", classroomHandler.UpdateClassroom)
		r.Delete("/{classroom_id}", classroomHandler.DeleteClassroom)
		r.Post("/{classroom_id}/students", classroomHandler.AddStudentToClassroom)
		r.Delete("/{classroom_id}/students/{student_id}", classroomHandler.RemoveStudentFromClassroom)
		r.Get("/{classroom_id}/students", classroomHandler.ListStudentsByClassroom)
	})

	r.Route("/v1/students", func(r chi.Router) {
		r.Get("/{student_id}/classrooms", classroomHandler.ListClassroomsByStudent)
	})

	return r
}
