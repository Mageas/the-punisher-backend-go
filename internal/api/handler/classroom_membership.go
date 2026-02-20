package handler

import (
	"net/http"

	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/validator"
	"github.com/mageas/the-punisher-backend/internal/platform/web"
)

func (h *ClassroomHandler) AddStudentToClassroom(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	classroomID, ok := parsePathUUID(w, r, "classroom_id")
	if !ok {
		return
	}

	var req dto.StudentClassroomRequestDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	studentID, ok := parseBodyUUID(w, req.StudentID, "student_id")
	if !ok {
		return
	}

	if err := h.service.AddStudentToClassroom(r.Context(), userID, classroomID, studentID); err != nil {
		web.WriteFromError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ClassroomHandler) RemoveStudentFromClassroom(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	classroomID, ok := parsePathUUID(w, r, "classroom_id")
	if !ok {
		return
	}

	studentID, ok := parsePathUUID(w, r, "student_id")
	if !ok {
		return
	}

	if err := h.service.RemoveStudentFromClassroom(r.Context(), userID, classroomID, studentID); err != nil {
		web.WriteFromError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ClassroomHandler) ListStudentsByClassroom(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	classroomID, ok := parsePathUUID(w, r, "classroom_id")
	if !ok {
		return
	}

	limit, offset, page := web.ParsePagination(r)

	students, totalCount, err := h.service.ListStudentsByClassroom(r.Context(), userID, classroomID, limit, offset)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(students, totalCount, page)
	web.WriteJSON(w, http.StatusOK, response, nil)
}

func (h *ClassroomHandler) ListClassroomsByStudent(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	studentID, ok := parsePathUUID(w, r, "student_id")
	if !ok {
		return
	}

	limit, offset, page := web.ParsePagination(r)

	classrooms, totalCount, err := h.service.ListClassroomsByStudent(r.Context(), userID, studentID, limit, offset)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(classrooms, totalCount, page)
	web.WriteJSON(w, http.StatusOK, response, nil)
}
