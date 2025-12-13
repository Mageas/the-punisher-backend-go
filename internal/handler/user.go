package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"

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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"}, nil)
		return
	}

	user, err := h.service.CreateUser(r.Context(), req)
	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": validationErrors.Error()}, nil)
			return
		}

		if errors.Is(err, service.ErrEmailAlreadyExists) {
			utils.WriteError(w, http.StatusConflict, err.Error())
			return
		}

		utils.ServerError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, user, nil)
}
