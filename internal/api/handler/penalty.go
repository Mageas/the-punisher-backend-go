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

	occurredAt, ok := parseOptionalBodyRFC3339(w, req.OccurredAt, "occurred_at")
	if !ok {
		return
	}

	classroomID, ok := parseOptionalBodyUUID(w, req.ClassroomID, "classroom_id")
	if !ok {
		return
	}

	penalty, err := h.service.CreatePenalty(r.Context(), userID, studentID, penaltyTypeID, classroomID, occurredAt, req.EvaluationLabel)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusCreated, penalty, nil)
}

func (h *PenaltyHandler) ListPenalties(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	limit, offset, page := web.ParsePagination(r)

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

	penaltyTypeID, details, err := web.ParseOptionalUUIDQueryParam(r, "penalty_type_id")
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

	filters := service.ListPenaltiesFilters{
		StudentID:     studentID,
		ClassroomID:   classroomID,
		PenaltyTypeID: penaltyTypeID,
		CreatedFrom:   createdFrom,
		CreatedTo:     createdTo,
		Limit:         limit,
		Offset:        offset,
	}

	penalties, totalCount, err := h.service.ListPenalties(r.Context(), userID, filters)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(penalties, totalCount, page, int(limit))
	web.WriteJSON(w, http.StatusOK, response, nil)
}

func (h *PenaltyHandler) GetPenalty(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	penaltyID, ok := parsePathUUID(w, r, "penalty_id")
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

func (h *PenaltyHandler) UpdatePenalty(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	penaltyID, ok := parsePathUUID(w, r, "penalty_id")
	if !ok {
		return
	}

	var req dto.UpdatePenaltyDto
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

	penalty, err := h.service.UpdatePenalty(
		r.Context(),
		userID,
		penaltyID,
		occurredAt,
		req.EvaluationLabel,
	)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, penalty, nil)
}

func (h *PenaltyHandler) DeletePenalty(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	penaltyID, ok := parsePathUUID(w, r, "penalty_id")
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

	studentID, ok := parsePathUUID(w, r, "student_id")
	if !ok {
		return
	}

	limit, offset, page := web.ParsePagination(r)

	penalties, totalCount, err := h.service.ListPenaltiesByStudent(r.Context(), userID, studentID, limit, offset)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(penalties, totalCount, page, int(limit))
	web.WriteJSON(w, http.StatusOK, response, nil)
}
