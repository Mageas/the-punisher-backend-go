package handler

import (
	"errors"
	"net/http"

	"github.com/mageas/the-punisher-backend/internal/domain_errors"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/service"
	"github.com/mageas/the-punisher-backend/internal/utils"
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
	if err := utils.DecodeJSON(w, r, &req); err != nil {
		utils.WriteJSONDecodeError(w, err)
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.WriteValidationError(w, err)
		return
	}

	user, err := h.service.CreateUser(r.Context(), req)
	if err != nil {
		if errors.Is(err, domain_errors.ErrEmailAlreadyExists) {
			utils.WriteError(w, http.StatusConflict, domain_errors.ErrEmailAlreadyExists.WithKey("testtt"), nil)
			return
		}

		utils.WriteServerError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, user, nil)
}
