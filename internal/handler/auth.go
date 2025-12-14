package handler

import (
	"net/http"

	"github.com/mageas/the-punisher-backend/internal/apierr"
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
	if err := utils.DecodeJSON(w, r, &req); err != nil {
		utils.WriteJSONDecodeError(w, err)
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.WriteValidationError(w, err)
		return
	}

	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, apierr.ErrInvalidCredentialsOrUserDoesntExist, nil)
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp, nil)
}
