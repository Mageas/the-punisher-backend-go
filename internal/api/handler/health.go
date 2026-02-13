package handler

import (
	"net/http"

	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/web"
	"github.com/mageas/the-punisher-backend/internal/service"
)

type HealthHandler struct {
	service service.HealthChecker
}

func NewHealthHandler(service service.HealthChecker) *HealthHandler {
	return &HealthHandler{service: service}
}

func (h *HealthHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	data := h.service.Check()

	statusCode := http.StatusOK

	if data.Status == dto.StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	if err := web.WriteJSON(w, statusCode, data, nil); err != nil {
		web.WriteServerError(w, err)
	}
}
