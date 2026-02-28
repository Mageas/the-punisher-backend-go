package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	platformjwt "github.com/mageas/the-punisher-backend/internal/platform/jwt"
)

func decodeAuthError(t *testing.T, rr *httptest.ResponseRecorder) api.ErrorResponse {
	t.Helper()

	var body api.ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode auth error response: %v", err)
	}
	return body
}

func TestAuthMiddlewareRejectsMissingHeader(t *testing.T) {
	h := AuthMiddleware("secret", "issuer", "aud")(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatalf("next handler should not be called")
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddlewareRejectsMalformedBearerHeader(t *testing.T) {
	h := AuthMiddleware("secret", "issuer", "aud")(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatalf("next handler should not be called")
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Token abc")
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddlewareInjectsUserID(t *testing.T) {
	userID := uuid.New()
	token, err := platformjwt.Generate(platformjwt.Config{
		Secret:     "secret",
		Expiration: time.Minute,
		Issuer:     "issuer",
		Audience:   "aud",
	}, userID.String())
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	var got uuid.UUID
	h := AuthMiddleware("secret", "issuer", "aud")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v, ok := UserIDFromContext(r.Context())
		if !ok {
			t.Fatalf("expected user id in context")
		}
		got = v
		w.WriteHeader(http.StatusNoContent)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	if got != userID {
		t.Fatalf("unexpected user id: %s", got)
	}
}

func TestAuthMiddlewareRejectsExpiredToken(t *testing.T) {
	token, err := platformjwt.Generate(platformjwt.Config{
		Secret:     "secret",
		Expiration: -1 * time.Second,
		Issuer:     "issuer",
		Audience:   "aud",
	}, uuid.NewString())
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	h := AuthMiddleware("secret", "issuer", "aud")(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatalf("next handler should not be called")
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	body := decodeAuthError(t, rr)
	if body.Error != api.ErrJWTExpired.Message {
		t.Fatalf("expected jwt_expired, got %s", body.Error)
	}
}

func TestAuthMiddlewareRejectsNonUUIDSubject(t *testing.T) {
	token, err := platformjwt.Generate(platformjwt.Config{
		Secret:     "secret",
		Expiration: time.Minute,
		Issuer:     "issuer",
		Audience:   "aud",
	}, "not-a-uuid")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	h := AuthMiddleware("secret", "issuer", "aud")(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatalf("next handler should not be called")
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}

	body := decodeAuthError(t, rr)
	if body.Error != api.ErrUnauthorized.Message {
		t.Fatalf("unexpected error: %s", body.Error)
	}
}

func TestMustUserIDFromContext(t *testing.T) {
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), userIDKey, userID)

	if got := MustUserIDFromContext(ctx); got != userID {
		t.Fatalf("unexpected user id: %s", got)
	}
}

func TestMustUserIDFromContextPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic")
		}
	}()

	_ = MustUserIDFromContext(context.Background())
}
