package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/web"
	"github.com/mageas/the-punisher-backend/internal/service"
)

type DashboardHandler struct {
	service service.DashboardService
}

func NewDashboardHandler(service service.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	var classroomID *uuid.UUID
	rawClassroomID := strings.TrimSpace(r.URL.Query().Get("classroom_id"))
	if rawClassroomID != "" {
		parsedClassroomID, err := uuid.Parse(rawClassroomID)
		if err != nil {
			web.WriteError(w, http.StatusBadRequest, api.ErrMalformedParameter, []api.ErrorDetail{
				{Field: "classroom_id", Error: fmt.Sprintf(api.KeyValidationMalformedParameter, "uuid")},
			})
			return
		}
		classroomID = &parsedClassroomID
	}

	dashboard, err := h.service.GetDashboard(r.Context(), userID, classroomID)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, dashboard, nil)
}
