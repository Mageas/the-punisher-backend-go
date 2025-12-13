package handler

import (
	"encoding/json"
	"net/http"

	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/service"
	"github.com/mageas/the-punisher-backend/internal/utils"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(service service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequestDto
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"}, nil)
		return
	}

	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials or user doesn't exist"}, nil)
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp, nil)
}
