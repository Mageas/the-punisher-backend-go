package handler

import (
	"fmt"
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

	occurredAt, ok := parseOptionalBodyRFC3339(w, req.OccurredAt, "occurred_at")
	if !ok {
		return
	}

	punishment, err := h.service.CreatePunishment(r.Context(), userID, studentID, punishmentTypeID, dueAt, occurredAt, req.EvaluationLabel)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusCreated, punishment, nil)
}

func (h *PunishmentHandler) ListPunishments(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	limit, offset, page := web.ParsePagination(r)
	stateValue, hasState, err := web.ParseEnumQueryParam(r, "state", []string{
		string(service.PunishmentStatePending),
		string(service.PunishmentStateResolved),
	})
	if err != nil {
		expected := web.EnumExpected([]string{string(service.PunishmentStatePending), string(service.PunishmentStateResolved)})
		details := []api.ErrorDetail{
			{
				Field: "state",
				Error: fmt.Sprintf(api.KeyValidationMalformedParameter, expected),
			},
		}
		web.WriteAPIError(w, api.ErrMalformedParameter, details)
		return
	}

	var state *service.PunishmentState
	if hasState {
		parsedState := service.PunishmentState(stateValue)
		state = &parsedState
	}

	studentID, details, err := web.ParseOptionalUUIDQueryParam(r, "student_id")
	if err != nil {
		web.WriteAPIError(w, api.ErrMalformedParameter, details)
		return
	}

	classroomID, details, err := web.ParseOptionalUUIDQueryParam(r, "classroom_id")
	if err != nil {
		web.WriteAPIError(w, api.ErrMalformedParameter, details)
		return
	}

	punishmentTypeID, details, err := web.ParseOptionalUUIDQueryParam(r, "punishment_type_id")
	if err != nil {
		web.WriteAPIError(w, api.ErrMalformedParameter, details)
		return
	}

	automated, details, err := web.ParseOptionalBoolQueryParam(r, "automated")
	if err != nil {
		web.WriteAPIError(w, api.ErrMalformedParameter, details)
		return
	}

	overdue, details, err := web.ParseOptionalBoolQueryParam(r, "overdue")
	if err != nil {
		web.WriteAPIError(w, api.ErrMalformedParameter, details)
		return
	}

	createdFrom, details, err := web.ParseOptionalDateQueryParam(r, "created_from")
	if err != nil {
		web.WriteAPIError(w, api.ErrMalformedParameter, details)
		return
	}

	createdTo, details, err := web.ParseOptionalDateQueryParam(r, "created_to")
	if err != nil {
		web.WriteAPIError(w, api.ErrMalformedParameter, details)
		return
	}

	details, err = web.ValidateDateRange(createdFrom, createdTo, "created_from", "created_to")
	if err != nil {
		web.WriteAPIError(w, api.ErrMalformedParameter, details)
		return
	}

	dueFrom, details, err := web.ParseOptionalDateQueryParam(r, "due_from")
	if err != nil {
		web.WriteAPIError(w, api.ErrMalformedParameter, details)
		return
	}

	dueTo, details, err := web.ParseOptionalDateQueryParam(r, "due_to")
	if err != nil {
		web.WriteAPIError(w, api.ErrMalformedParameter, details)
		return
	}

	details, err = web.ValidateDateRange(dueFrom, dueTo, "due_from", "due_to")
	if err != nil {
		web.WriteAPIError(w, api.ErrMalformedParameter, details)
		return
	}

	filters := service.ListPunishmentsFilters{
		StudentID:        studentID,
		ClassroomID:      classroomID,
		PunishmentTypeID: punishmentTypeID,
		State:            state,
		Automated:        automated,
		Overdue:          overdue,
		CreatedFrom:      createdFrom,
		CreatedTo:        createdTo,
		DueFrom:          dueFrom,
		DueTo:            dueTo,
		Limit:            limit,
		Offset:           offset,
	}

	punishments, totalCount, err := h.service.ListPunishments(r.Context(), userID, filters)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(punishments, totalCount, page, int(limit))
	web.WriteJSON(w, http.StatusOK, response, nil)
}

func (h *PunishmentHandler) GetPunishment(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	punishmentID, ok := parsePathUUID(w, r, "punishment_id")
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

func (h *PunishmentHandler) UpdatePunishment(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	punishmentID, ok := parsePathUUID(w, r, "punishment_id")
	if !ok {
		return
	}

	var req dto.UpdatePunishmentDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	occurredAt, ok := parseOptionalBodyRFC3339(w, req.OccurredAt, "occurred_at")
	if !ok {
		return
	}

	punishment, err := h.service.UpdatePunishment(
		r.Context(),
		userID,
		punishmentID,
		occurredAt,
		req.EvaluationLabel.Set,
		req.EvaluationLabel.Value,
	)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, punishment, nil)
}

func (h *PunishmentHandler) ResolvePunishment(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	punishmentID, ok := parsePathUUID(w, r, "punishment_id")
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

	punishmentID, ok := parsePathUUID(w, r, "punishment_id")
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

	studentID, ok := parsePathUUID(w, r, "student_id")
	if !ok {
		return
	}

	limit, offset, page := web.ParsePagination(r)

	resolved, details, err := web.ParseEnumQueryParamToBool(r, "state", "resolved", "pending")
	if err != nil {
		web.WriteAPIError(w, api.ErrMalformedParameter, details)
		return
	}

	punishments, totalCount, err := h.service.ListPunishmentsByStudent(r.Context(), userID, studentID, resolved, limit, offset)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(punishments, totalCount, page, int(limit))
	web.WriteJSON(w, http.StatusOK, response, nil)
}
