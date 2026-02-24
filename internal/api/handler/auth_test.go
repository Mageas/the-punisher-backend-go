package handler_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/api/handler"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/config"
	"github.com/mageas/the-punisher-backend/internal/platform/hash"
	"github.com/mageas/the-punisher-backend/internal/platform/jwt"
	"github.com/mageas/the-punisher-backend/internal/repository"
	"github.com/mageas/the-punisher-backend/internal/service"
	"github.com/mageas/the-punisher-backend/internal/testutil/httpx"
	"github.com/mageas/the-punisher-backend/internal/testutil/inmemory"
)

const (
	refreshTokenCookieName = "refresh_token"
	refreshTokenCookiePath = "/v1/auth"
)

func TestAuthHandlerLoginSuccess(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := testJWTConfig()
	authHandler := handler.NewAuthHandler(service.NewAuthService(repo, cfg), cfg, refreshTokenCookiePath)

	userID := uuid.New()
	password := "password123"
	hashedPassword, err := hash.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	repo.SeedAuthUser(userID, "teacher@example.com", hashedPassword)

	req := httpx.NewJSONRequest(t, http.MethodPost, "/v1/auth/login", map[string]any{
		"email":    "teacher@example.com",
		"password": password,
	})
	req.RemoteAddr = "203.0.113.1:12345"

	rr := httptest.NewRecorder()
	authHandler.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	resp := httpx.DecodeJSONResponse[dto.LoginResponseDto](t, rr)
	if resp.AccessToken == "" {
		t.Fatal("expected access_token to be present in response body")
	}
	if resp.RefreshToken != "" {
		t.Fatal("expected refresh_token to be omitted from response body")
	}

	cookie := httpx.MustCookie(t, rr, refreshTokenCookieName)
	if cookie.Value == "" {
		t.Fatal("expected refresh token cookie value to be present")
	}
	if cookie.Path != refreshTokenCookiePath {
		t.Fatalf("expected cookie path %q, got %q", refreshTokenCookiePath, cookie.Path)
	}
	if !cookie.HttpOnly {
		t.Fatal("expected cookie to be httpOnly")
	}
	if cookie.Secure {
		t.Fatal("expected cookie secure flag to be disabled in non-production test config")
	}
	if cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("expected SameSiteStrictMode, got %v", cookie.SameSite)
	}
	if cookie.MaxAge <= 0 {
		t.Fatalf("expected positive MaxAge, got %d", cookie.MaxAge)
	}

	refreshTokenHash := hash.HashToken(cookie.Value, cfg.RefreshSecret)
	storedToken, ok := repo.StoredRefreshToken(userID, refreshTokenHash)
	if !ok {
		t.Fatal("expected refresh token to be persisted in in-memory repository")
	}
	if storedToken.ClientIp != req.RemoteAddr {
		t.Fatalf("expected persisted client ip %q, got %q", req.RemoteAddr, storedToken.ClientIp)
	}
}

func TestAuthHandlerLoginSecureCookieEnabled(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := testJWTConfig()
	cfg.RefreshCookieSecure = true
	authHandler := handler.NewAuthHandler(service.NewAuthService(repo, cfg), cfg, refreshTokenCookiePath)

	userID := uuid.New()
	password := "password123"
	hashedPassword, err := hash.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	repo.SeedAuthUser(userID, "teacher@example.com", hashedPassword)

	req := httpx.NewJSONRequest(t, http.MethodPost, "/v1/auth/login", map[string]any{
		"email":    "teacher@example.com",
		"password": password,
	})

	rr := httptest.NewRecorder()
	authHandler.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	cookie := httpx.MustCookie(t, rr, refreshTokenCookieName)
	if !cookie.Secure {
		t.Fatal("expected cookie secure flag to be enabled")
	}
}

func TestAuthHandlerLoginDecodeErrors(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := testJWTConfig()
	authHandler := handler.NewAuthHandler(service.NewAuthService(repo, cfg), cfg, refreshTokenCookiePath)

	t.Run("unknown_field", func(t *testing.T) {
		req := httpx.NewJSONRequest(t, http.MethodPost, "/v1/auth/login", map[string]any{
			"email":    "teacher@example.com",
			"password": "password123",
			"unknown":  "value",
		})

		rr := httptest.NewRecorder()
		authHandler.Login(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrInvalidRequestBody.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrInvalidRequestBody.Error(), resp.Error)
		}

		assertHasErrorDetail(t, resp.ErrorDetails, "unknown", api.KeyValidationUnknownField)
	})

	t.Run("malformed_parameter_type", func(t *testing.T) {
		req := httpx.NewJSONRequest(t, http.MethodPost, "/v1/auth/login", map[string]any{
			"email":    "teacher@example.com",
			"password": 123,
		})

		rr := httptest.NewRecorder()
		authHandler.Login(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrInvalidRequestBody.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrInvalidRequestBody.Error(), resp.Error)
		}

		assertHasErrorDetail(t, resp.ErrorDetails, "password", "validation_malformed_parameter:expected_string")
	})

	t.Run("invalid_json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBufferString("{"))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		authHandler.Login(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrInvalidRequestBody.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrInvalidRequestBody.Error(), resp.Error)
		}
	})
}

func TestAuthHandlerLoginDTOValidations(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := testJWTConfig()
	authHandler := handler.NewAuthHandler(service.NewAuthService(repo, cfg), cfg, refreshTokenCookiePath)

	tests := []struct {
		name          string
		payload       map[string]any
		expectedField string
		expectedError string
	}{
		{
			name: "email_required",
			payload: map[string]any{
				"password": "password123",
			},
			expectedField: "email",
			expectedError: api.KeyValidationFieldRequired,
		},
		{
			name: "email_format_invalid",
			payload: map[string]any{
				"email":    "not-an-email",
				"password": "password123",
			},
			expectedField: "email",
			expectedError: api.KeyValidationInvalidEmail,
		},
		{
			name: "password_required",
			payload: map[string]any{
				"email": "teacher@example.com",
			},
			expectedField: "password",
			expectedError: api.KeyValidationFieldRequired,
		},
		{
			name: "password_min_length",
			payload: map[string]any{
				"email":    "teacher@example.com",
				"password": "short",
			},
			expectedField: "password",
			expectedError: "validation_min_length:8",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httpx.NewJSONRequest(t, http.MethodPost, "/v1/auth/login", tc.payload)
			rr := httptest.NewRecorder()

			authHandler.Login(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
			}

			resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
			if resp.Error != api.ErrValidationFailed.Error() {
				t.Fatalf("expected error %q, got %q", api.ErrValidationFailed.Error(), resp.Error)
			}

			assertHasErrorDetail(t, resp.ErrorDetails, tc.expectedField, tc.expectedError)
		})
	}
}

func TestAuthHandlerLoginBusinessAndInternalErrors(t *testing.T) {
	t.Run("invalid_credentials", func(t *testing.T) {
		repo := inmemory.NewRepository()
		cfg := testJWTConfig()
		authHandler := handler.NewAuthHandler(service.NewAuthService(repo, cfg), cfg, refreshTokenCookiePath)

		req := httpx.NewJSONRequest(t, http.MethodPost, "/v1/auth/login", map[string]any{
			"email":    "missing@example.com",
			"password": "password123",
		})

		rr := httptest.NewRecorder()
		authHandler.Login(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrInvalidCredentialsOrUserDoesntExist.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrInvalidCredentialsOrUserDoesntExist.Error(), resp.Error)
		}
	})

	t.Run("repository_read_failure", func(t *testing.T) {
		repo := inmemory.NewRepository()
		cfg := testJWTConfig()
		authHandler := handler.NewAuthHandler(service.NewAuthService(repo, cfg), cfg, refreshTokenCookiePath)
		repo.SetGetUserCredentialsError(errors.New("database unavailable"))

		req := httpx.NewJSONRequest(t, http.MethodPost, "/v1/auth/login", map[string]any{
			"email":    "teacher@example.com",
			"password": "password123",
		})

		rr := httptest.NewRecorder()
		authHandler.Login(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrInternalError.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrInternalError.Error(), resp.Error)
		}
	})

	t.Run("repository_write_failure", func(t *testing.T) {
		repo := inmemory.NewRepository()
		cfg := testJWTConfig()
		authHandler := handler.NewAuthHandler(service.NewAuthService(repo, cfg), cfg, refreshTokenCookiePath)

		userID := uuid.New()
		password := "password123"
		hashedPassword, err := hash.HashPassword(password)
		if err != nil {
			t.Fatalf("failed to hash password: %v", err)
		}

		repo.SeedAuthUser(userID, "teacher@example.com", hashedPassword)
		repo.SetCreateRefreshTokenError(errors.New("database unavailable"))

		req := httpx.NewJSONRequest(t, http.MethodPost, "/v1/auth/login", map[string]any{
			"email":    "teacher@example.com",
			"password": password,
		})

		rr := httptest.NewRecorder()
		authHandler.Login(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrInternalError.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrInternalError.Error(), resp.Error)
		}
	})
}

func TestAuthHandlerRefresh(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := inmemory.NewRepository()
		cfg := testJWTConfig()
		authHandler := handler.NewAuthHandler(service.NewAuthService(repo, cfg), cfg, refreshTokenCookiePath)

		userID := uuid.New()
		refreshToken, err := jwt.Generate(jwt.Config{
			Secret:     cfg.RefreshSecret,
			Expiration: cfg.RefreshExpiration,
			Issuer:     cfg.Issuer,
			Audience:   cfg.Audience,
		}, userID.String())
		if err != nil {
			t.Fatalf("failed to generate refresh token: %v", err)
		}

		repo.SeedRefreshToken(repository.RefreshToken{
			UserID:    userID,
			Token:     hash.HashToken(refreshToken, cfg.RefreshSecret),
			ExpiresAt: time.Now().Add(cfg.RefreshExpiration),
		})

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: refreshToken})
		rr := httptest.NewRecorder()

		authHandler.Refresh(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[dto.RefreshResponseDto](t, rr)
		if resp.AccessToken == "" {
			t.Fatal("expected access_token to be present in response body")
		}

		newCookie := httpx.MustCookie(t, rr, refreshTokenCookieName)
		if newCookie.Value == "" {
			t.Fatal("expected rotated refresh token cookie value to be present")
		}
		if newCookie.Value == refreshToken {
			t.Fatal("expected rotated refresh token to differ from previous one")
		}

		oldToken, ok := repo.StoredRefreshToken(userID, hash.HashToken(refreshToken, cfg.RefreshSecret))
		if !ok {
			t.Fatal("expected previous refresh token to still exist in storage as revoked")
		}
		if oldToken.RevokedAt == nil {
			t.Fatal("expected previous refresh token to be revoked after refresh")
		}

		rotatedTokenHash := hash.HashToken(newCookie.Value, cfg.RefreshSecret)
		rotatedToken, ok := repo.StoredRefreshToken(userID, rotatedTokenHash)
		if !ok {
			t.Fatal("expected rotated refresh token to be persisted in storage")
		}
		if rotatedToken.RevokedAt != nil {
			t.Fatal("expected rotated refresh token to be active")
		}
	})

	t.Run("missing_cookie", func(t *testing.T) {
		repo := inmemory.NewRepository()
		cfg := testJWTConfig()
		authHandler := handler.NewAuthHandler(service.NewAuthService(repo, cfg), cfg, refreshTokenCookiePath)

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", nil)
		rr := httptest.NewRecorder()

		authHandler.Refresh(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrUnauthorized.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrUnauthorized.Error(), resp.Error)
		}
	})

	t.Run("invalid_jwt_refresh_token", func(t *testing.T) {
		repo := inmemory.NewRepository()
		cfg := testJWTConfig()
		authHandler := handler.NewAuthHandler(service.NewAuthService(repo, cfg), cfg, refreshTokenCookiePath)

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "not-a-jwt"})
		rr := httptest.NewRecorder()

		authHandler.Refresh(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrJWTInvalidToken.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrJWTInvalidToken.Error(), resp.Error)
		}
	})

	t.Run("refresh_token_not_found", func(t *testing.T) {
		repo := inmemory.NewRepository()
		cfg := testJWTConfig()
		authHandler := handler.NewAuthHandler(service.NewAuthService(repo, cfg), cfg, refreshTokenCookiePath)

		userID := uuid.New()
		refreshToken, err := jwt.Generate(jwt.Config{
			Secret:     cfg.RefreshSecret,
			Expiration: cfg.RefreshExpiration,
			Issuer:     cfg.Issuer,
			Audience:   cfg.Audience,
		}, userID.String())
		if err != nil {
			t.Fatalf("failed to generate refresh token: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: refreshToken})
		rr := httptest.NewRecorder()

		authHandler.Refresh(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrUnauthorized.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrUnauthorized.Error(), resp.Error)
		}
	})

	t.Run("repository_failure", func(t *testing.T) {
		repo := inmemory.NewRepository()
		cfg := testJWTConfig()
		authHandler := handler.NewAuthHandler(service.NewAuthService(repo, cfg), cfg, refreshTokenCookiePath)
		repo.SetGetRefreshTokenError(errors.New("database unavailable"))

		userID := uuid.New()
		refreshToken, err := jwt.Generate(jwt.Config{
			Secret:     cfg.RefreshSecret,
			Expiration: cfg.RefreshExpiration,
			Issuer:     cfg.Issuer,
			Audience:   cfg.Audience,
		}, userID.String())
		if err != nil {
			t.Fatalf("failed to generate refresh token: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: refreshToken})
		rr := httptest.NewRecorder()

		authHandler.Refresh(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrInternalError.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrInternalError.Error(), resp.Error)
		}
	})

	t.Run("rotation_rollback_when_create_new_refresh_token_fails", func(t *testing.T) {
		repo := inmemory.NewRepository()
		cfg := testJWTConfig()
		authHandler := handler.NewAuthHandler(service.NewAuthService(repo, cfg), cfg, refreshTokenCookiePath)

		userID := uuid.New()
		refreshToken, err := jwt.Generate(jwt.Config{
			Secret:     cfg.RefreshSecret,
			Expiration: cfg.RefreshExpiration,
			Issuer:     cfg.Issuer,
			Audience:   cfg.Audience,
		}, userID.String())
		if err != nil {
			t.Fatalf("failed to generate refresh token: %v", err)
		}

		oldTokenHash := hash.HashToken(refreshToken, cfg.RefreshSecret)
		repo.SeedRefreshToken(repository.RefreshToken{
			UserID:    userID,
			Token:     oldTokenHash,
			ExpiresAt: time.Now().Add(cfg.RefreshExpiration),
		})
		repo.SetCreateRefreshTokenError(errors.New("database unavailable"))

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: refreshToken})
		rr := httptest.NewRecorder()

		authHandler.Refresh(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrInternalError.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrInternalError.Error(), resp.Error)
		}

		oldToken, ok := repo.StoredRefreshToken(userID, oldTokenHash)
		if !ok {
			t.Fatal("expected original refresh token to still be persisted")
		}
		if oldToken.RevokedAt != nil {
			t.Fatal("expected original refresh token to remain active after rollback")
		}
	})
}

func TestAuthHandlerLogout(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := inmemory.NewRepository()
		cfg := testJWTConfig()
		authHandler := handler.NewAuthHandler(service.NewAuthService(repo, cfg), cfg, refreshTokenCookiePath)

		userID := uuid.New()
		refreshToken, err := jwt.Generate(jwt.Config{
			Secret:     cfg.RefreshSecret,
			Expiration: cfg.RefreshExpiration,
			Issuer:     cfg.Issuer,
			Audience:   cfg.Audience,
		}, userID.String())
		if err != nil {
			t.Fatalf("failed to generate refresh token: %v", err)
		}

		refreshTokenHash := hash.HashToken(refreshToken, cfg.RefreshSecret)
		repo.SeedRefreshToken(repository.RefreshToken{
			UserID:    userID,
			Token:     refreshTokenHash,
			ExpiresAt: time.Now().Add(cfg.RefreshExpiration),
		})

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", nil)
		req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: refreshToken})
		rr := httptest.NewRecorder()

		authHandler.Logout(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
		}

		storedToken, ok := repo.StoredRefreshToken(userID, refreshTokenHash)
		if !ok {
			t.Fatal("expected refresh token to still exist in storage as revoked")
		}
		if storedToken.RevokedAt == nil {
			t.Fatal("expected refresh token to be revoked")
		}

		cookie := httpx.MustCookie(t, rr, refreshTokenCookieName)
		if cookie.Path != refreshTokenCookiePath {
			t.Fatalf("expected cookie path %q, got %q", refreshTokenCookiePath, cookie.Path)
		}
		if cookie.MaxAge >= 0 {
			t.Fatalf("expected negative MaxAge for expired cookie, got %d", cookie.MaxAge)
		}
		if cookie.Value != "" {
			t.Fatal("expected expired cookie value to be empty")
		}
	})

	t.Run("missing_cookie", func(t *testing.T) {
		repo := inmemory.NewRepository()
		cfg := testJWTConfig()
		authHandler := handler.NewAuthHandler(service.NewAuthService(repo, cfg), cfg, refreshTokenCookiePath)

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", nil)
		rr := httptest.NewRecorder()

		authHandler.Logout(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
		}

		cookie := httpx.MustCookie(t, rr, refreshTokenCookieName)
		if cookie.Value != "" {
			t.Fatal("expected expired cookie value to be empty")
		}
	})

	t.Run("repository_failure", func(t *testing.T) {
		repo := inmemory.NewRepository()
		cfg := testJWTConfig()
		authHandler := handler.NewAuthHandler(service.NewAuthService(repo, cfg), cfg, refreshTokenCookiePath)
		repo.SetRevokeRefreshTokenError(errors.New("database unavailable"))

		userID := uuid.New()
		refreshToken, err := jwt.Generate(jwt.Config{
			Secret:     cfg.RefreshSecret,
			Expiration: cfg.RefreshExpiration,
			Issuer:     cfg.Issuer,
			Audience:   cfg.Audience,
		}, userID.String())
		if err != nil {
			t.Fatalf("failed to generate refresh token: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", nil)
		req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: refreshToken})
		rr := httptest.NewRecorder()

		authHandler.Logout(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrInternalError.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrInternalError.Error(), resp.Error)
		}

		cookie := httpx.MustCookie(t, rr, refreshTokenCookieName)
		if cookie.Value != "" {
			t.Fatal("expected expired cookie value to be empty")
		}
	})
}

func testJWTConfig() config.JWTConfig {
	return config.JWTConfig{
		AccessSecret:        "test-access-secret",
		AccessExpiration:    15 * time.Minute,
		RefreshSecret:       "test-refresh-secret",
		RefreshExpiration:   7 * 24 * time.Hour,
		RefreshCookieSecure: false,
		Issuer:              "the-punisher-tests",
		Audience:            "the-punisher-tests",
	}
}

func assertHasErrorDetail(t *testing.T, details []api.ErrorDetail, field, errKey string) {
	t.Helper()

	for _, detail := range details {
		if detail.Field == field && detail.Error == errKey {
			return
		}
	}

	t.Fatalf("expected error detail field=%q error=%q, got %+v", field, errKey, details)
}
