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

	studentID, ok := parseBodyUUID(w, req.StudentID, "student_id")
	if !ok {
		return
	}

	bonusTypeID, ok := parseBodyUUID(w, req.BonusTypeID, "bonus_type_id")
	if !ok {
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
	stateValue, hasState, err := web.ParseEnumQueryParam(r, "state", []string{
		string(service.BonusStateUsed),
		string(service.BonusStateUnused),
	})
	if err != nil {
		expected := web.EnumExpected([]string{string(service.BonusStateUsed), string(service.BonusStateUnused)})
		details := []api.ErrorDetail{
			{
				Field: "state",
				Error: fmt.Sprintf(api.KeyValidationMalformedParameter, expected),
			},
		}
		web.WriteAPIError(w, api.ErrMalformedParameter, details)
		return
	}

	var state *service.BonusState
	if hasState {
		parsedState := service.BonusState(stateValue)
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

	bonusTypeID, details, err := web.ParseOptionalUUIDQueryParam(r, "bonus_type_id")
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

	filters := service.ListBonusesFilters{
		StudentID:   studentID,
		ClassroomID: classroomID,
		BonusTypeID: bonusTypeID,
		State:       state,
		CreatedFrom: createdFrom,
		CreatedTo:   createdTo,
		Limit:       limit,
		Offset:      offset,
	}

	bonuses, totalCount, err := h.service.ListBonuses(r.Context(), userID, filters)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(bonuses, totalCount, page, int(limit))
	web.WriteJSON(w, http.StatusOK, response, nil)
}

func (h *BonusHandler) GetBonus(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	bonusID, ok := parsePathUUID(w, r, "bonus_id")
	if !ok {
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

	bonusID, ok := parsePathUUID(w, r, "bonus_id")
	if !ok {
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

	bonusID, ok := parsePathUUID(w, r, "bonus_id")
	if !ok {
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

	studentID, ok := parsePathUUID(w, r, "student_id")
	if !ok {
		return
	}

	limit, offset, page := web.ParsePagination(r)

	used, details, err := web.ParseEnumQueryParamToBool(r, "state", "used", "unused")
	if err != nil {
		web.WriteAPIError(w, api.ErrMalformedParameter, details)
		return
	}

	bonuses, totalCount, err := h.service.ListBonusesByStudent(r.Context(), userID, studentID, used, limit, offset)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(bonuses, totalCount, page, int(limit))
	web.WriteJSON(w, http.StatusOK, response, nil)
}
