package handler

import (
	"net/http"

	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/validator"
	"github.com/mageas/the-punisher-backend/internal/platform/web"
	"github.com/mageas/the-punisher-backend/internal/service"
)

type PunishmentTypeHandler struct {
	service service.PunishmentTypeService
}

func NewPunishmentTypeHandler(service service.PunishmentTypeService) *PunishmentTypeHandler {
	return &PunishmentTypeHandler{
		service: service,
	}
}

func (h *PunishmentTypeHandler) CreatePunishmentType(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	var req dto.RequestPunishmentTypeDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	res, err := h.service.CreatePunishmentType(r.Context(), userID, req)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusCreated, res, nil)
}

func (h *PunishmentTypeHandler) ListPunishmentTypes(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	limit, offset, page := web.ParsePagination(r)

	items, totalCount, err := h.service.ListPunishmentTypes(r.Context(), userID, limit, offset)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(items, totalCount, page)
	web.WriteJSON(w, http.StatusOK, response, nil)
}

func (h *PunishmentTypeHandler) GetPunishmentType(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	punishmentTypeID, ok := parsePathUUID(w, r, "punishment_type_id")
	if !ok {
		return
	}

	res, err := h.service.GetPunishmentType(r.Context(), userID, punishmentTypeID)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, res, nil)
}

func (h *PunishmentTypeHandler) UpdatePunishmentType(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	punishmentTypeID, ok := parsePathUUID(w, r, "punishment_type_id")
	if !ok {
		return
	}

	var req dto.UpdatePunishmentTypeDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	res, err := h.service.UpdatePunishmentType(r.Context(), userID, punishmentTypeID, req)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, res, nil)
}

func (h *PunishmentTypeHandler) DeletePunishmentType(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	punishmentTypeID, ok := parsePathUUID(w, r, "punishment_type_id")
	if !ok {
		return
	}

	if err := h.service.DeletePunishmentType(r.Context(), userID, punishmentTypeID); err != nil {
		web.WriteFromError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
