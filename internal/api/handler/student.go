package handler

import (
	"net/http"

	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/validator"
	"github.com/mageas/the-punisher-backend/internal/platform/web"
	"github.com/mageas/the-punisher-backend/internal/service"
)

type StudentHandler struct {
	service service.StudentService
}

func NewStudentHandler(service service.StudentService) *StudentHandler {
	return &StudentHandler{service: service}
}

func (h *StudentHandler) CreateStudent(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	var req dto.RequestStudentDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	student, err := h.service.CreateStudent(r.Context(), userID, req)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusCreated, student, nil)
}

func (h *StudentHandler) GetStudent(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	studentID, ok := parsePathUUID(w, r, "student_id", "student_id", "id")
	if !ok {
		return
	}

	student, err := h.service.GetStudent(r.Context(), userID, studentID)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, student, nil)
}

func (h *StudentHandler) GetStudentKpis(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	studentID, ok := parsePathUUID(w, r, "student_id", "student_id", "id")
	if !ok {
		return
	}

	kpis, err := h.service.GetStudentKpis(r.Context(), userID, studentID)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, kpis, nil)
}

func (h *StudentHandler) GetStudentHistory(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	studentID, ok := parsePathUUID(w, r, "student_id", "student_id", "id")
	if !ok {
		return
	}

	limit, offset, _ := web.ParsePagination(r)

	if r.URL.Query().Has("history_page") {
		w.Header().Set("Warning", `299 - "query parameter 'history_page' is deprecated; use 'page'"`)
	}

	history, err := h.service.ListStudentHistory(r.Context(), userID, studentID, limit, offset)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, history, nil)
}

func (h *StudentHandler) ListStudents(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	limit, offset, page := web.ParsePagination(r)
	search := web.ParseSearchQueryParam(r, "search")

	students, totalCount, err := h.service.ListStudents(r.Context(), userID, search, limit, offset)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(students, totalCount, page)
	web.WriteJSON(w, http.StatusOK, response, nil)
}

func (h *StudentHandler) UpdateStudent(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	studentID, ok := parsePathUUID(w, r, "student_id", "student_id", "id")
	if !ok {
		return
	}

	var req dto.UpdateStudentDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	student, err := h.service.UpdateStudent(r.Context(), userID, studentID, req)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, student, nil)
}

func (h *StudentHandler) DeleteStudent(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	studentID, ok := parsePathUUID(w, r, "student_id", "student_id", "id")
	if !ok {
		return
	}

	if err := h.service.DeleteStudent(r.Context(), userID, studentID); err != nil {
		web.WriteFromError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
