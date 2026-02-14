package handlertest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/platform/config"
	"github.com/mageas/the-punisher-backend/internal/platform/jwt"
	"github.com/mageas/the-punisher-backend/internal/testutil/httpx"
)

func NewAuthorizedJSONRequest(t *testing.T, method, target string, payload any, userID uuid.UUID, cfg config.JWTConfig) *http.Request {
	t.Helper()

	req := httpx.NewJSONRequest(t, method, target, payload)
	req.Header.Set("Authorization", "Bearer "+MustAccessToken(t, userID, cfg))
	return req
}

func NewAuthorizedRequest(t *testing.T, method, target string, userID uuid.UUID, cfg config.JWTConfig) *http.Request {
	t.Helper()

	req := httptest.NewRequest(method, target, nil)
	req.Header.Set("Authorization", "Bearer "+MustAccessToken(t, userID, cfg))
	return req
}

func MustAccessToken(t *testing.T, userID uuid.UUID, cfg config.JWTConfig) string {
	t.Helper()

	token, err := jwt.Generate(jwt.Config{
		Secret:     cfg.AccessSecret,
		Expiration: cfg.AccessExpiration,
		Issuer:     cfg.Issuer,
		Audience:   cfg.Audience,
	}, userID.String())
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	return token
}
