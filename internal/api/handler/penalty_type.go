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

type PenaltyTypeHandler struct {
	service service.PenaltyTypeService
}

func NewPenaltyTypeHandler(service service.PenaltyTypeService) *PenaltyTypeHandler {
	return &PenaltyTypeHandler{
		service: service,
	}
}

func (h *PenaltyTypeHandler) CreatePenaltyType(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	var req dto.RequestPenaltyTypeDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	res, err := h.service.CreatePenaltyType(r.Context(), userID, req)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusCreated, res, nil)
}

func (h *PenaltyTypeHandler) ListPenaltyTypes(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	limit, offset, page := web.ParsePagination(r)

	penaltyTypes, totalCount, err := h.service.ListPenaltyTypes(r.Context(), userID, limit, offset)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(penaltyTypes, totalCount, page)
	web.WriteJSON(w, http.StatusOK, response, nil)
}

func (h *PenaltyTypeHandler) GetPenaltyType(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	penaltyTypeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		web.WriteError(w, http.StatusBadRequest, api.ErrMalformedParameter, nil)
		return
	}

	res, err := h.service.GetPenaltyType(r.Context(), userID, penaltyTypeID)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, res, nil)
}

func (h *PenaltyTypeHandler) UpdatePenaltyType(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	penaltyTypeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		web.WriteError(w, http.StatusBadRequest, api.ErrMalformedParameter, nil)
		return
	}

	var req dto.UpdatePenaltyTypeDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	penaltyType, err := h.service.UpdatePenaltyType(r.Context(), userID, penaltyTypeID, req)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, penaltyType, nil)
}

func (h *PenaltyTypeHandler) DeletePenaltyType(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	penaltyTypeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		web.WriteError(w, http.StatusBadRequest, api.ErrMalformedParameter, nil)
		return
	}

	if err := h.service.DeletePenaltyType(r.Context(), userID, penaltyTypeID); err != nil {
		web.WriteFromError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
