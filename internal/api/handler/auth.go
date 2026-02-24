package handler

import (
	"net/http"
	"time"

	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/config"
	"github.com/mageas/the-punisher-backend/internal/platform/validator"
	"github.com/mageas/the-punisher-backend/internal/platform/web"
	"github.com/mageas/the-punisher-backend/internal/service"
)

var (
	refreshTokenName = "refresh_token"
)

type AuthHandler struct {
	service          service.AuthService
	cfg              config.JWTConfig
	refreshTokenPath string
}

func NewAuthHandler(service service.AuthService, cfg config.JWTConfig, refreshTokenPath string) *AuthHandler {
	return &AuthHandler{
		service:          service,
		cfg:              cfg,
		refreshTokenPath: refreshTokenPath,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequestDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	req.RemoteAddr = r.RemoteAddr

	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	h.setRefreshTokenCookie(w, resp.RefreshToken)

	// TODO: retrieve 'X-Auth-Mode' header, if 'body' is set, return the refresh token in the response

	web.WriteJSON(w, http.StatusOK, resp, nil)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshTokenName)
	if err != nil {
		web.WriteAPIError(w, api.ErrUnauthorized, nil)
		return
	}

	resp, err := h.service.Refresh(r.Context(), cookie.Value)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	h.setRefreshTokenCookie(w, resp.RefreshToken)

	web.WriteJSON(w, http.StatusOK, resp, nil)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshTokenName)
	if err == nil {
		if err := h.service.Logout(r.Context(), cookie.Value); err != nil {
			h.clearRefreshTokenCookie(w)
			web.WriteFromError(w, err)
			return
		}
	}

	h.clearRefreshTokenCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) setRefreshTokenCookie(w http.ResponseWriter, token string) {
	cookieDuration := h.cfg.RefreshExpiration

	http.SetCookie(w, &http.Cookie{
		Name:  refreshTokenName,
		Value: token,

		Path:     h.refreshTokenPath,
		HttpOnly: true,
		Secure:   h.cfg.RefreshCookieSecure,
		SameSite: http.SameSiteStrictMode,

		Expires: time.Now().Add(cookieDuration),
		MaxAge:  int(cookieDuration.Seconds()),
	})
}

func (h *AuthHandler) clearRefreshTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenName,
		Value:    "",
		Path:     h.refreshTokenPath,
		HttpOnly: true,
		Secure:   h.cfg.RefreshCookieSecure,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}
