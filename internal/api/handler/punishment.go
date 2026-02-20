package handler

import (
	"net/http"

	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/validator"
	"github.com/mageas/the-punisher-backend/internal/platform/web"
	"github.com/mageas/the-punisher-backend/internal/service"
)

type PunishmentHandler struct {
	service service.PunishmentService
}

func NewPunishmentHandler(service service.PunishmentService) *PunishmentHandler {
	return &PunishmentHandler{service: service}
}

func (h *PunishmentHandler) CreatePunishment(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	var req dto.RequestPunishmentDto
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

	punishmentTypeID, ok := parseBodyUUID(w, req.PunishmentTypeID, "punishment_type_id")
	if !ok {
		return
	}

	dueAt, ok := parseBodyRFC3339(w, req.DueAt, "due_at")
	if !ok {
		return
	}

	punishment, err := h.service.CreatePunishment(r.Context(), userID, studentID, punishmentTypeID, dueAt)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusCreated, punishment, nil)
}

func (h *PunishmentHandler) ListPunishments(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	limit, offset, page := web.ParsePagination(r)
	search := web.ParseSearchQueryParam(r, "search")

	resolved, details, err := web.ParseEnumQueryParamToBool(r, "state", "resolved", "pending")
	if err != nil {
		web.WriteError(w, http.StatusBadRequest, api.ErrMalformedParameter, details)
		return
	}

	punishments, totalCount, err := h.service.ListPunishments(r.Context(), userID, resolved, search, limit, offset)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(punishments, totalCount, page)
	web.WriteJSON(w, http.StatusOK, response, nil)
}

func (h *PunishmentHandler) GetPunishment(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	punishmentID, ok := parsePathUUID(w, r, "punishment_id", "punishment_id", "id")
	if !ok {
		return
	}

	punishment, err := h.service.GetPunishment(r.Context(), userID, punishmentID)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, punishment, nil)
}

func (h *PunishmentHandler) ResolvePunishment(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	punishmentID, ok := parsePathUUID(w, r, "punishment_id", "punishment_id", "id")
	if !ok {
		return
	}

	punishment, err := h.service.ResolvePunishment(r.Context(), userID, punishmentID)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, punishment, nil)
}

func (h *PunishmentHandler) DeletePunishment(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	punishmentID, ok := parsePathUUID(w, r, "punishment_id", "punishment_id", "id")
	if !ok {
		return
	}

	if err := h.service.DeletePunishment(r.Context(), userID, punishmentID); err != nil {
		web.WriteFromError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PunishmentHandler) ListPunishmentsByStudent(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	studentID, ok := parsePathUUID(w, r, "student_id", "student_id", "id")
	if !ok {
		return
	}

	limit, offset, page := web.ParsePagination(r)

	resolved, details, err := web.ParseEnumQueryParamToBool(r, "state", "resolved", "pending")
	if err != nil {
		web.WriteError(w, http.StatusBadRequest, api.ErrMalformedParameter, details)
		return
	}

	punishments, totalCount, err := h.service.ListPunishmentsByStudent(r.Context(), userID, studentID, resolved, limit, offset)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(punishments, totalCount, page)
	web.WriteJSON(w, http.StatusOK, response, nil)
}
