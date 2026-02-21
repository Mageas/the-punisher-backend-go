package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/jwt"
)

func TestAuthMiddleware(t *testing.T) {
	secret := "test-secret-key-very-long-and-secure-enough-for-hs256"
	issuer := "test-issuer"
	audience := "test-audience"

	// Setup a simple handler that returns 200 OK and checks if userID is present
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			t.Error("expected userID in context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if userID == uuid.Nil {
			t.Error("expected non-nil userID")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := auth.AuthMiddleware(secret, issuer, audience)
	handler := middleware(nextHandler)

	t.Run("valid_token", func(t *testing.T) {
		userID := uuid.New()
		token, err := jwt.Generate(jwt.Config{
			Secret:     secret,
			Expiration: 1 * time.Hour,
			Issuer:     issuer,
			Audience:   audience,
		}, userID.String())
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
		}
	})

	t.Run("missing_header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
		}
	})

	t.Run("malformed_header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "InvalidToken")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
		}
	})

	t.Run("invalid_signature", func(t *testing.T) {
		userID := uuid.New()
		token, err := jwt.Generate(jwt.Config{
			Secret:     "wrong-secret",
			Expiration: 1 * time.Hour,
			Issuer:     issuer,
			Audience:   audience,
		}, userID.String())
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
		}
	})

	t.Run("expired_token", func(t *testing.T) {
		userID := uuid.New()
		// Generate an expired token manually by setting a negative expiration time
		token, err := jwt.Generate(jwt.Config{
			Secret:     secret,
			Expiration: -1 * time.Hour,
			Issuer:     issuer,
			Audience:   audience,
		}, userID.String())
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
		}
	})
}
