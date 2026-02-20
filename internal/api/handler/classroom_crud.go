package handler

import (
	"net/http"

	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/validator"
	"github.com/mageas/the-punisher-backend/internal/platform/web"
)

func (h *ClassroomHandler) CreateClassroom(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	var req dto.RequestClassroomDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	classroom, err := h.service.CreateClassroom(r.Context(), userID, req)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusCreated, classroom, nil)
}

func (h *ClassroomHandler) GetClassroom(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	classroomID, ok := parsePathUUID(w, r, "classroom_id", "classroom_id", "id")
	if !ok {
		return
	}

	classroom, err := h.service.GetClassroom(r.Context(), userID, classroomID)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, classroom, nil)
}

func (h *ClassroomHandler) ListClassrooms(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	limit, offset, page := web.ParsePagination(r)

	classrooms, totalCount, err := h.service.ListClassrooms(r.Context(), userID, limit, offset)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(classrooms, totalCount, page)
	web.WriteJSON(w, http.StatusOK, response, nil)
}

func (h *ClassroomHandler) UpdateClassroom(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	classroomID, ok := parsePathUUID(w, r, "classroom_id", "classroom_id", "id")
	if !ok {
		return
	}

	var req dto.UpdateClassroomDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	classroom, err := h.service.UpdateClassroom(r.Context(), userID, classroomID, req)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, classroom, nil)
}

func (h *ClassroomHandler) DeleteClassroom(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	classroomID, ok := parsePathUUID(w, r, "classroom_id", "classroom_id", "id")
	if !ok {
		return
	}

	if err := h.service.DeleteClassroom(r.Context(), userID, classroomID); err != nil {
		web.WriteFromError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
