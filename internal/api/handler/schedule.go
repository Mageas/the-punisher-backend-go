package handler

import (
	"net/http"

	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/validator"
	"github.com/mageas/the-punisher-backend/internal/platform/web"
	"github.com/mageas/the-punisher-backend/internal/service"
)

type ScheduleHandler struct {
	service service.ScheduleService
}

func NewScheduleHandler(service service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{service: service}
}

func (h *ScheduleHandler) CreateScheduleSlot(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	var req dto.RequestScheduleSlotDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	slot, err := h.service.CreateScheduleSlot(r.Context(), userID, req)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusCreated, slot, nil)
}

func (h *ScheduleHandler) GetScheduleSlot(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	scheduleSlotID, ok := parsePathUUID(w, r, "schedule_slot_id")
	if !ok {
		return
	}

	slot, err := h.service.GetScheduleSlot(r.Context(), userID, scheduleSlotID)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, slot, nil)
}

func (h *ScheduleHandler) ListScheduleSlots(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	slots, err := h.service.ListScheduleSlots(r.Context(), userID)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, slots, nil)
}

func (h *ScheduleHandler) UpdateScheduleSlot(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	scheduleSlotID, ok := parsePathUUID(w, r, "schedule_slot_id")
	if !ok {
		return
	}

	var req dto.UpdateScheduleSlotDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	slot, err := h.service.UpdateScheduleSlot(r.Context(), userID, scheduleSlotID, req)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, slot, nil)
}

func (h *ScheduleHandler) DeleteScheduleSlot(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	scheduleSlotID, ok := parsePathUUID(w, r, "schedule_slot_id")
	if !ok {
		return
	}

	if err := h.service.DeleteScheduleSlot(r.Context(), userID, scheduleSlotID); err != nil {
		web.WriteFromError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ScheduleHandler) CreateScheduleException(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	var req dto.RequestScheduleExceptionDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	exception, err := h.service.CreateScheduleException(r.Context(), userID, req)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusCreated, exception, nil)
}

func (h *ScheduleHandler) GetScheduleException(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	scheduleExceptionID, ok := parsePathUUID(w, r, "schedule_exception_id")
	if !ok {
		return
	}

	exception, err := h.service.GetScheduleException(r.Context(), userID, scheduleExceptionID)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, exception, nil)
}

func (h *ScheduleHandler) ListScheduleExceptions(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	exceptions, err := h.service.ListScheduleExceptions(r.Context(), userID)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, exceptions, nil)
}

func (h *ScheduleHandler) UpdateScheduleException(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	scheduleExceptionID, ok := parsePathUUID(w, r, "schedule_exception_id")
	if !ok {
		return
	}

	var req dto.UpdateScheduleExceptionDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	exception, err := h.service.UpdateScheduleException(r.Context(), userID, scheduleExceptionID, req)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, exception, nil)
}

func (h *ScheduleHandler) DeleteScheduleException(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	scheduleExceptionID, ok := parsePathUUID(w, r, "schedule_exception_id")
	if !ok {
		return
	}

	if err := h.service.DeleteScheduleException(r.Context(), userID, scheduleExceptionID); err != nil {
		web.WriteFromError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ScheduleHandler) ListNextLessons(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	classroomID, ok := parsePathUUID(w, r, "classroom_id")
	if !ok {
		return
	}

	nextLessons, err := h.service.ListNextLessons(r.Context(), userID, classroomID)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, nextLessons, nil)
}
