package handler_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/api/handler"
	"github.com/mageas/the-punisher-backend/internal/dto"
	platformauth "github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/config"
	"github.com/mageas/the-punisher-backend/internal/platform/hash"
	"github.com/mageas/the-punisher-backend/internal/repository"
	"github.com/mageas/the-punisher-backend/internal/service"
	"github.com/mageas/the-punisher-backend/internal/testutil/handlertest"
	"github.com/mageas/the-punisher-backend/internal/testutil/httpx"
	"github.com/mageas/the-punisher-backend/internal/testutil/inmemory"
	shared "github.com/mageas/the-punisher-backend/internal/testutil/shared"
)

func TestUserHandlerCreateUserSuccess(t *testing.T) {
	repo := inmemory.NewRepository()
	userHandler := newUserHandler(repo, true)

	req := httpx.NewJSONRequest(t, http.MethodPost, "/v1/auth/register", map[string]any{
		"email":      "Teacher@Example.com",
		"first_name": "Jean",
		"last_name":  "Dupont",
		"password":   "password123",
	})
	rr := httptest.NewRecorder()
	userHandler.CreateUser(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	resp := httpx.DecodeJSONResponse[dto.ReturnUserDto](t, rr)
	if resp.ID.String() == "" {
		t.Fatal("expected created user id")
	}
	if resp.Email != "teacher@example.com" {
		t.Fatalf("expected normalized email %q, got %q", "teacher@example.com", resp.Email)
	}
	if resp.FirstName != "Jean" || resp.LastName != "Dupont" {
		t.Fatalf("unexpected user payload: %+v", resp)
	}
}

func TestUserHandlerCreateUserRegisterNotAllowed(t *testing.T) {
	repo := inmemory.NewRepository()
	userHandler := newUserHandler(repo, false)

	req := httpx.NewJSONRequest(t, http.MethodPost, "/v1/auth/register", map[string]any{
		"email":      "teacher@example.com",
		"first_name": "Jean",
		"last_name":  "Dupont",
		"password":   "password123",
	})
	rr := httptest.NewRecorder()
	userHandler.CreateUser(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
	if resp.Error != api.ErrRegisterNotAllowed.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrRegisterNotAllowed.Error(), resp.Error)
	}
}

func TestUserHandlerCreateUserDecodeErrors(t *testing.T) {
	repo := inmemory.NewRepository()
	userHandler := newUserHandler(repo, true)

	t.Run("unknown_field", func(t *testing.T) {
		req := httpx.NewJSONRequest(t, http.MethodPost, "/v1/auth/register", map[string]any{
			"email":      "teacher@example.com",
			"first_name": "Jean",
			"last_name":  "Dupont",
			"password":   "password123",
			"unknown":    "value",
		})
		rr := httptest.NewRecorder()

		userHandler.CreateUser(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrInvalidRequestBody.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrInvalidRequestBody.Error(), resp.Error)
		}

		shared.AssertHasErrorDetail(t, resp.ErrorDetails, "unknown", api.KeyValidationUnknownField)
	})

	t.Run("malformed_parameter_type", func(t *testing.T) {
		req := httpx.NewJSONRequest(t, http.MethodPost, "/v1/auth/register", map[string]any{
			"email":      "teacher@example.com",
			"first_name": 123,
			"last_name":  "Dupont",
			"password":   "password123",
		})
		rr := httptest.NewRecorder()

		userHandler.CreateUser(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrMalformedParameter.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrMalformedParameter.Error(), resp.Error)
		}

		shared.AssertHasErrorDetail(t, resp.ErrorDetails, "first_name", "validation_malformed_parameter:expected_string")
	})

	t.Run("invalid_json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBufferString("{"))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		userHandler.CreateUser(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrInvalidRequestBody.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrInvalidRequestBody.Error(), resp.Error)
		}
	})
}

func TestUserHandlerCreateUserDTOValidations(t *testing.T) {
	repo := inmemory.NewRepository()
	userHandler := newUserHandler(repo, true)

	tests := []struct {
		name          string
		payload       map[string]any
		expectedField string
		expectedError string
	}{
		{
			name: "email_required",
			payload: map[string]any{
				"first_name": "Jean",
				"last_name":  "Dupont",
				"password":   "password123",
			},
			expectedField: "email",
			expectedError: api.KeyValidationFieldRequired,
		},
		{
			name: "email_invalid",
			payload: map[string]any{
				"email":      "not-an-email",
				"first_name": "Jean",
				"last_name":  "Dupont",
				"password":   "password123",
			},
			expectedField: "email",
			expectedError: api.KeyValidationInvalidEmail,
		},
		{
			name: "first_name_min_length",
			payload: map[string]any{
				"email":      "teacher@example.com",
				"first_name": "J",
				"last_name":  "Dupont",
				"password":   "password123",
			},
			expectedField: "first_name",
			expectedError: "validation_min_length:2",
		},
		{
			name: "password_min_length",
			payload: map[string]any{
				"email":      "teacher@example.com",
				"first_name": "Jean",
				"last_name":  "Dupont",
				"password":   "short",
			},
			expectedField: "password",
			expectedError: "validation_min_length:8",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httpx.NewJSONRequest(t, http.MethodPost, "/v1/auth/register", tc.payload)
			rr := httptest.NewRecorder()

			userHandler.CreateUser(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
			}

			resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
			if resp.Error != api.ErrValidationFailed.Error() {
				t.Fatalf("expected error %q, got %q", api.ErrValidationFailed.Error(), resp.Error)
			}

			shared.AssertHasErrorDetail(t, resp.ErrorDetails, tc.expectedField, tc.expectedError)
		})
	}
}

func TestUserHandlerCreateUserBusinessAndInternalErrors(t *testing.T) {
	t.Run("email_already_exists", func(t *testing.T) {
		repo := inmemory.NewRepository()
		userHandler := newUserHandler(repo, true)

		passwordHash, err := hash.HashPassword("password123")
		if err != nil {
			t.Fatalf("failed to hash password: %v", err)
		}

		repo.SeedUser(repository.User{
			Email:        "teacher@example.com",
			FirstName:    "Jean",
			LastName:     "Dupont",
			PasswordHash: passwordHash,
		})

		req := httpx.NewJSONRequest(t, http.MethodPost, "/v1/auth/register", map[string]any{
			"email":      "teacher@example.com",
			"first_name": "Jean",
			"last_name":  "Dupont",
			"password":   "password123",
		})
		rr := httptest.NewRecorder()

		userHandler.CreateUser(rr, req)

		if rr.Code != http.StatusConflict {
			t.Fatalf("expected status %d, got %d", http.StatusConflict, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrEmailAlreadyExists.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrEmailAlreadyExists.Error(), resp.Error)
		}

		shared.AssertHasErrorDetail(t, resp.ErrorDetails, "email", api.KeyValidationEmailAlreadyExists)
	})

	t.Run("email_exists_repository_failure", func(t *testing.T) {
		repo := inmemory.NewRepository()
		userHandler := newUserHandler(repo, true)
		repo.SetError(inmemory.OpUserEmailExists, errors.New("database unavailable"))

		req := httpx.NewJSONRequest(t, http.MethodPost, "/v1/auth/register", map[string]any{
			"email":      "teacher@example.com",
			"first_name": "Jean",
			"last_name":  "Dupont",
			"password":   "password123",
		})
		rr := httptest.NewRecorder()

		userHandler.CreateUser(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrInternalError.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrInternalError.Error(), resp.Error)
		}
	})

	t.Run("create_repository_failure", func(t *testing.T) {
		repo := inmemory.NewRepository()
		userHandler := newUserHandler(repo, true)
		repo.SetError(inmemory.OpCreateUser, errors.New("database unavailable"))

		req := httpx.NewJSONRequest(t, http.MethodPost, "/v1/auth/register", map[string]any{
			"email":      "teacher@example.com",
			"first_name": "Jean",
			"last_name":  "Dupont",
			"password":   "password123",
		})
		rr := httptest.NewRecorder()

		userHandler.CreateUser(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrInternalError.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrInternalError.Error(), resp.Error)
		}
	})
}

func TestUserHandlerGetMe(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := inmemory.NewRepository()
		cfg := shared.TestJWTConfig()
		router := newUserRouter(repo, cfg, true)
		userID := uuid.New()

		repo.SeedUser(repository.User{
			ID:        userID,
			Email:     "teacher@example.com",
			FirstName: "Jean",
			LastName:  "Dupont",
		})

		req := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/user/me/", userID, cfg)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[dto.ReturnUserDto](t, rr)
		if resp.ID != userID {
			t.Fatalf("expected user id %s, got %s", userID, resp.ID)
		}
		if resp.Email != "teacher@example.com" || resp.FirstName != "Jean" || resp.LastName != "Dupont" {
			t.Fatalf("unexpected /user/me payload: %+v", resp)
		}
	})

	t.Run("unauthorized_without_token", func(t *testing.T) {
		repo := inmemory.NewRepository()
		cfg := shared.TestJWTConfig()
		router := newUserRouter(repo, cfg, true)

		req := httptest.NewRequest(http.MethodGet, "/v1/user/me/", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
		}
	})

	t.Run("returns_unauthorized_when_user_not_found", func(t *testing.T) {
		repo := inmemory.NewRepository()
		cfg := shared.TestJWTConfig()
		router := newUserRouter(repo, cfg, true)
		userID := uuid.New()

		req := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/user/me/", userID, cfg)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrUnauthorized.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrUnauthorized.Error(), resp.Error)
		}
	})

	t.Run("internal_error", func(t *testing.T) {
		repo := inmemory.NewRepository()
		cfg := shared.TestJWTConfig()
		router := newUserRouter(repo, cfg, true)
		userID := uuid.New()

		repo.SeedUser(repository.User{
			ID:        userID,
			Email:     "teacher@example.com",
			FirstName: "Jean",
			LastName:  "Dupont",
		})
		repo.SetGetUserByIDError(errors.New("database unavailable"))

		req := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/user/me/", userID, cfg)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrInternalError.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrInternalError.Error(), resp.Error)
		}
	})
}

func newUserHandler(repo *inmemory.Repository, allowRegister bool) *handler.UserHandler {
	return handler.NewUserHandler(
		service.NewUserService(repo),
		config.Config{AllowRegister: allowRegister},
	)
}

func newUserRouter(repo *inmemory.Repository, jwtCfg config.JWTConfig, allowRegister bool) http.Handler {
	userHandler := handler.NewUserHandler(
		service.NewUserService(repo),
		config.Config{AllowRegister: allowRegister},
	)

	r := chi.NewRouter()
	r.Route("/v1/auth", func(r chi.Router) {
		r.Post("/register", userHandler.CreateUser)
	})

	r.Route("/v1/user/me", func(r chi.Router) {
		r.Use(platformauth.AuthMiddleware(jwtCfg.AccessSecret))
		r.Get("/", userHandler.GetMe)
	})

	return r
}
