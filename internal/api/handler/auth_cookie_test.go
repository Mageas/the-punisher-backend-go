package handler

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mageas/the-punisher-backend/internal/platform/config"
)

func TestSetRefreshTokenCookie(t *testing.T) {
	h := &AuthHandler{
		cfg: config.JWTConfig{
			RefreshExpiration:   24 * time.Hour,
			RefreshCookieSecure: true,
		},
		refreshTokenPath: "/auth/refresh",
	}

	rr := httptest.NewRecorder()
	h.setRefreshTokenCookie(rr, "token-value")

	cookies := rr.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected one cookie, got %d", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != refreshTokenName || cookie.Value != "token-value" {
		t.Fatalf("unexpected cookie identity: %+v", cookie)
	}
	if !cookie.HttpOnly || !cookie.Secure {
		t.Fatalf("expected httponly+secure cookie")
	}
	if cookie.Path != "/auth/refresh" {
		t.Fatalf("unexpected path: %s", cookie.Path)
	}
	if cookie.MaxAge <= 0 {
		t.Fatalf("expected positive max age, got %d", cookie.MaxAge)
	}
}

func TestClearRefreshTokenCookie(t *testing.T) {
	h := &AuthHandler{
		cfg: config.JWTConfig{
			RefreshCookieSecure: true,
		},
		refreshTokenPath: "/auth/refresh",
	}

	rr := httptest.NewRecorder()
	h.clearRefreshTokenCookie(rr)

	cookies := rr.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected one cookie, got %d", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != refreshTokenName {
		t.Fatalf("unexpected cookie name: %s", cookie.Name)
	}
	if cookie.Value != "" {
		t.Fatalf("expected empty value")
	}
	if cookie.MaxAge != -1 {
		t.Fatalf("expected MaxAge -1, got %d", cookie.MaxAge)
	}
}
