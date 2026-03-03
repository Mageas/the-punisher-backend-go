package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	platformauth "github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/config"
	platformjwt "github.com/mageas/the-punisher-backend/internal/platform/jwt"
)

type fakeAuthService struct {
	changePasswordErr  error
	changePasswordReq  dto.ChangePasswordRequestDto
	changePasswordUser uuid.UUID
	changeCalled       bool
}

func (f *fakeAuthService) Login(_ context.Context, _ dto.LoginRequestDto) (*dto.LoginResponseDto, error) {
	return nil, nil
}

func (f *fakeAuthService) Refresh(_ context.Context, _ string) (*dto.RefreshResponseDto, error) {
	return nil, nil
}

func (f *fakeAuthService) Logout(_ context.Context, _ string) error {
	return nil
}

func (f *fakeAuthService) LogoutAll(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (f *fakeAuthService) ChangePassword(_ context.Context, userID uuid.UUID, req dto.ChangePasswordRequestDto) error {
	f.changeCalled = true
	f.changePasswordUser = userID
	f.changePasswordReq = req
	return f.changePasswordErr
}

func authBearerToken(t *testing.T, userID uuid.UUID) string {
	t.Helper()

	token, err := platformjwt.Generate(platformjwt.Config{
		Secret:     "access-secret",
		Expiration: 15 * time.Minute,
		Issuer:     "issuer",
		Audience:   "audience",
	}, userID.String())
	if err != nil {
		t.Fatalf("failed to create bearer token: %v", err)
	}

	return token
}

func TestAuthHandler_ChangePasswordSuccess(t *testing.T) {
	fakeSvc := &fakeAuthService{}
	h := NewAuthHandler(fakeSvc, config.JWTConfig{RefreshCookieSecure: true}, "/v1/auth")

	userID := uuid.New()
	body := `{"current_password":"CurrentPass1!","new_password":"NewSecurePass1!","confirm_password":"NewSecurePass1!"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/change-password", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+authBearerToken(t, userID))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := platformauth.AuthMiddleware("access-secret", "issuer", "audience")(http.HandlerFunc(h.ChangePassword))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if !fakeSvc.changeCalled {
		t.Fatalf("expected ChangePassword to be called")
	}
	if fakeSvc.changePasswordUser != userID {
		t.Fatalf("expected user id %s, got %s", userID, fakeSvc.changePasswordUser)
	}
	if fakeSvc.changePasswordReq.NewPassword != "NewSecurePass1!" {
		t.Fatalf("unexpected request payload: %+v", fakeSvc.changePasswordReq)
	}

	var payload dto.ChangePasswordResponseDto
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload.Status != "password_changed" {
		t.Fatalf("unexpected response payload: %+v", payload)
	}

	cookies := rr.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected one cookie, got %d", len(cookies))
	}
	if cookies[0].Name != refreshTokenName || cookies[0].MaxAge != -1 {
		t.Fatalf("expected cleared refresh cookie, got %+v", cookies[0])
	}
}

func TestAuthHandler_ChangePasswordValidationError(t *testing.T) {
	fakeSvc := &fakeAuthService{}
	h := NewAuthHandler(fakeSvc, config.JWTConfig{}, "/v1/auth")

	userID := uuid.New()
	body := `{"current_password":"short","new_password":"short","confirm_password":"mismatch"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/change-password", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+authBearerToken(t, userID))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := platformauth.AuthMiddleware("access-secret", "issuer", "audience")(http.HandlerFunc(h.ChangePassword))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
	if fakeSvc.changeCalled {
		t.Fatalf("expected ChangePassword not to be called")
	}

	var payload api.ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload.Error != api.ErrValidationFailed.Message {
		t.Fatalf("expected validation_failed, got %s", payload.Error)
	}
}

func TestAuthHandler_ChangePasswordServiceError(t *testing.T) {
	fakeSvc := &fakeAuthService{changePasswordErr: api.ErrInvalidCurrentPassword}
	h := NewAuthHandler(fakeSvc, config.JWTConfig{}, "/v1/auth")

	userID := uuid.New()
	body := `{"current_password":"CurrentPass1!","new_password":"NewSecurePass1!","confirm_password":"NewSecurePass1!"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/change-password", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+authBearerToken(t, userID))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := platformauth.AuthMiddleware("access-secret", "issuer", "audience")(http.HandlerFunc(h.ChangePassword))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
	if !fakeSvc.changeCalled {
		t.Fatalf("expected ChangePassword to be called")
	}

	var payload api.ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload.Error != api.ErrInvalidCurrentPassword.Message {
		t.Fatalf("expected invalid_current_password, got %s", payload.Error)
	}
}
