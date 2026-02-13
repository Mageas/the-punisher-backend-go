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

type BonusHandler struct {
	service service.BonusService
}

func NewBonusHandler(service service.BonusService) *BonusHandler {
	return &BonusHandler{service: service}
}

func (h *BonusHandler) CreateBonus(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	var req dto.RequestBonusDto
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

	bonusTypeID, err := uuid.Parse(req.BonusTypeID)
	if err != nil {
		web.WriteError(w, http.StatusBadRequest, api.ErrInvalidRequestBody, nil)
		return
	}

	bonus, err := h.service.CreateBonus(r.Context(), userID, studentID, bonusTypeID, req.Points)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusCreated, bonus, nil)
}

func (h *BonusHandler) ListBonuses(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	limit, offset, page := web.ParsePagination(r)

	bonuses, totalCount, err := h.service.ListBonuses(r.Context(), userID, limit, offset)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(bonuses, totalCount, page)
	web.WriteJSON(w, http.StatusOK, response, nil)
}

func (h *BonusHandler) GetBonus(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	bonusID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		web.WriteError(w, http.StatusBadRequest, api.ErrMalformedParameter, nil)
		return
	}

	bonus, err := h.service.GetBonus(r.Context(), userID, bonusID)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, bonus, nil)
}

func (h *BonusHandler) UseBonus(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	bonusID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		web.WriteError(w, http.StatusBadRequest, api.ErrMalformedParameter, nil)
		return
	}

	bonus, err := h.service.UseBonus(r.Context(), userID, bonusID)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, bonus, nil)
}

func (h *BonusHandler) DeleteBonus(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	bonusID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		web.WriteError(w, http.StatusBadRequest, api.ErrMalformedParameter, nil)
		return
	}

	if err := h.service.DeleteBonus(r.Context(), userID, bonusID); err != nil {
		web.WriteFromError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BonusHandler) ListBonusesByStudent(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	studentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		web.WriteError(w, http.StatusBadRequest, api.ErrMalformedParameter, nil)
		return
	}

	limit, offset, page := web.ParsePagination(r)

	bonuses, totalCount, err := h.service.ListBonusesByStudent(r.Context(), userID, studentID, limit, offset)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(bonuses, totalCount, page)
	web.WriteJSON(w, http.StatusOK, response, nil)
}
