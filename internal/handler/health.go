package handler

import (
	"net/http"

	"github.com/mageas/the-punisher-backend/internal/service"
	"github.com/mageas/the-punisher-backend/internal/utils"
)

type HealthHandler struct {
	service *service.HealthService
}

func NewHealthHandler(service *service.HealthService) *HealthHandler {
	return &HealthHandler{service: service}
}

func (h *HealthHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	data := h.service.Check()

	statusCode := http.StatusOK

	if data.Status == service.StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	if err := utils.WriteJSON(w, statusCode, data, nil); err != nil {
		utils.WriteServerError(w, err)
	}
}
