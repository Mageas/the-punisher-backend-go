package handler

import (
	"errors"
	"net/http"

	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/validator"
	"github.com/mageas/the-punisher-backend/internal/platform/web"
	"github.com/mageas/the-punisher-backend/internal/service"
)

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req dto.RequestUserDto
	if err := web.DecodeJSON(w, r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	user, err := h.service.CreateUser(r.Context(), req)
	if err != nil {
		if errors.Is(err, api.ErrEmailAlreadyExists) {
			web.WriteConflictError(w, "email", api.KeyValidationEmailAlreadyExists)
			return
		}

		web.WriteServerError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusCreated, user, nil)
}
