package handler

import (
	"net/http"

	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/validator"
	"github.com/mageas/the-punisher-backend/internal/platform/web"
	"github.com/mageas/the-punisher-backend/internal/service"
)

type PenaltyHandler struct {
	service service.PenaltyService
}

func NewPenaltyHandler(service service.PenaltyService) *PenaltyHandler {
	return &PenaltyHandler{service: service}
}

func (h *PenaltyHandler) CreatePenalty(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	var req dto.RequestPenaltyDto
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

	penaltyTypeID, ok := parseBodyUUID(w, req.PenaltyTypeID, "penalty_type_id")
	if !ok {
		return
	}

	penalty, err := h.service.CreatePenalty(r.Context(), userID, studentID, penaltyTypeID)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusCreated, penalty, nil)
}

func (h *PenaltyHandler) ListPenalties(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	limit, offset, page := web.ParsePagination(r)

	penalties, totalCount, err := h.service.ListPenalties(r.Context(), userID, limit, offset)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(penalties, totalCount, page)
	web.WriteJSON(w, http.StatusOK, response, nil)
}

func (h *PenaltyHandler) GetPenalty(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	penaltyID, ok := parsePathUUID(w, r, "penalty_id", "penalty_id", "id")
	if !ok {
		return
	}

	penalty, err := h.service.GetPenalty(r.Context(), userID, penaltyID)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, penalty, nil)
}

func (h *PenaltyHandler) DeletePenalty(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	penaltyID, ok := parsePathUUID(w, r, "penalty_id", "penalty_id", "id")
	if !ok {
		return
	}

	if err := h.service.DeletePenalty(r.Context(), userID, penaltyID); err != nil {
		web.WriteFromError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PenaltyHandler) ListPenaltiesByStudent(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	studentID, ok := parsePathUUID(w, r, "student_id", "student_id", "id")
	if !ok {
		return
	}

	limit, offset, page := web.ParsePagination(r)

	penalties, totalCount, err := h.service.ListPenaltiesByStudent(r.Context(), userID, studentID, limit, offset)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(penalties, totalCount, page)
	web.WriteJSON(w, http.StatusOK, response, nil)
}
