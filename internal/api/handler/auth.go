package handler

import (
	"net/http"

	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/validator"
	"github.com/mageas/the-punisher-backend/internal/platform/web"
	"github.com/mageas/the-punisher-backend/internal/service"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(service service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequestDto
	if err := web.DecodeJSON(w, r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		web.WriteError(w, http.StatusUnauthorized, api.ErrInvalidCredentialsOrUserDoesntExist, nil)
		return
	}

	web.WriteJSON(w, http.StatusOK, resp, nil)
}
