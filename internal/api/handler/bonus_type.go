package handler

import (
	"net/http"

	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/validator"
	"github.com/mageas/the-punisher-backend/internal/platform/web"
	"github.com/mageas/the-punisher-backend/internal/service"
)

type BonusTypeHandler struct {
	service service.BonusTypeService
}

func NewBonusTypeHandler(service service.BonusTypeService) *BonusTypeHandler {
	return &BonusTypeHandler{
		service: service,
	}
}

func (h *BonusTypeHandler) CreateBonusType(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	var req dto.RequestBonusTypeDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	res, err := h.service.CreateBonusType(r.Context(), userID, req)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusCreated, res, nil)
}

func (h *BonusTypeHandler) ListBonusTypes(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	limit, offset, page := web.ParsePagination(r)
	search := web.ParseSearchQueryParam(r, "search")

	items, totalCount, err := h.service.ListBonusTypes(r.Context(), userID, search, limit, offset)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(items, totalCount, page, int(limit))
	web.WriteJSON(w, http.StatusOK, response, nil)
}

func (h *BonusTypeHandler) GetBonusType(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	bonusTypeID, ok := parsePathUUID(w, r, "bonus_type_id")
	if !ok {
		return
	}

	res, err := h.service.GetBonusType(r.Context(), userID, bonusTypeID)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, res, nil)
}

func (h *BonusTypeHandler) UpdateBonusType(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	bonusTypeID, ok := parsePathUUID(w, r, "bonus_type_id")
	if !ok {
		return
	}

	var req dto.UpdateBonusTypeDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	res, err := h.service.UpdateBonusType(r.Context(), userID, bonusTypeID, req)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, res, nil)
}

func (h *BonusTypeHandler) DeleteBonusType(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	bonusTypeID, ok := parsePathUUID(w, r, "bonus_type_id")
	if !ok {
		return
	}

	if err := h.service.DeleteBonusType(r.Context(), userID, bonusTypeID); err != nil {
		web.WriteFromError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
