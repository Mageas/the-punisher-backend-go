package handler

import (
	"net/http"
	"strings"

	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/config"
	"github.com/mageas/the-punisher-backend/internal/platform/validator"
	"github.com/mageas/the-punisher-backend/internal/platform/web"
	"github.com/mageas/the-punisher-backend/internal/service"
)

type UserHandler struct {
	service service.UserService
	cfg     config.Config
}

func NewUserHandler(service service.UserService, cfg config.Config) *UserHandler {
	return &UserHandler{
		service: service,
		cfg:     cfg,
	}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.AllowRegister {
		web.WriteAPIError(w, api.ErrRegisterNotAllowed, nil)
		return
	}

	var req dto.RequestUserDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	user, err := h.service.CreateUser(r.Context(), req)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusCreated, user, nil)
}

func (h *UserHandler) GetRegisterStatus(w http.ResponseWriter, r *http.Request) {
	resp := dto.RegisterStatusResponseDto{
		RegisterAllowed: h.cfg.AllowRegister,
	}

	web.WriteJSON(w, http.StatusOK, resp, nil)
}

func (h *UserHandler) ConfirmEmail(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		web.WriteAPIError(w, api.ErrEmailConfirmationTokenMissing, nil)
		return
	}

	if err := h.service.ConfirmEmail(r.Context(), token); err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, dto.ConfirmEmailResponseDto{Status: "email_confirmed"}, nil)
}

func (h *UserHandler) ResendConfirmEmail(w http.ResponseWriter, r *http.Request) {
	var req dto.ResendConfirmEmailRequestDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	if err := h.service.ResendEmailConfirmation(r.Context(), req.Email); err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, dto.ResendConfirmEmailResponseDto{Status: "confirmation_email_sent_if_needed"}, nil)
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	user, err := h.service.GetCurrentUser(r.Context(), userID)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, user, nil)
}
