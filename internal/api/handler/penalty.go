package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
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

	studentID, err := uuid.Parse(req.StudentID)
	if err != nil {
		web.WriteError(w, http.StatusBadRequest, api.ErrInvalidRequestBody, nil)
		return
	}

	penaltyTypeID, err := uuid.Parse(req.PenaltyTypeID)
	if err != nil {
		web.WriteError(w, http.StatusBadRequest, api.ErrInvalidRequestBody, nil)
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

	penaltyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		web.WriteError(w, http.StatusBadRequest, api.ErrMalformedParameter, nil)
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

	penaltyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		web.WriteError(w, http.StatusBadRequest, api.ErrMalformedParameter, nil)
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

	studentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		web.WriteError(w, http.StatusBadRequest, api.ErrMalformedParameter, nil)
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
