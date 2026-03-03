package handler

import (
	"context"
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
	"github.com/mageas/the-punisher-backend/internal/platform/jwt"
)

const (
	testAccessSecret = "test-access-secret"
	testJWTIssuer    = "test-issuer"
	testJWTAudience  = "test-audience"
)

type fakeAuthService struct {
	changePasswordErr    error
	changePasswordCalled bool
	changePasswordUserID uuid.UUID
	changePasswordReq    dto.ChangePasswordRequestDto
}

func (f *fakeAuthService) Login(context.Context, dto.LoginRequestDto) (*dto.LoginResponseDto, error) {
	return nil, nil
}

func (f *fakeAuthService) Refresh(context.Context, string) (*dto.RefreshResponseDto, error) {
	return nil, nil
}

func (f *fakeAuthService) Logout(context.Context, string) error {
	return nil
}

func (f *fakeAuthService) LogoutAll(context.Context, uuid.UUID) error {
	return nil
}

func (f *fakeAuthService) ChangePassword(_ context.Context, userID uuid.UUID, req dto.ChangePasswordRequestDto) error {
	f.changePasswordCalled = true
	f.changePasswordUserID = userID
	f.changePasswordReq = req
	return f.changePasswordErr
}

func TestAuthHandler_ChangePasswordServiceErrorDoesNotClearRefreshCookie(t *testing.T) {
	userID := uuid.New()
	fakeSvc := &fakeAuthService{
		changePasswordErr: api.ErrCurrentPasswordInvalid,
	}
	h := NewAuthHandler(fakeSvc, config.JWTConfig{}, "/v1/auth")

	req := authenticatedChangePasswordRequest(t, userID)
	rr := httptest.NewRecorder()

	withAuthMiddleware(h).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
	if !fakeSvc.changePasswordCalled {
		t.Fatalf("expected ChangePassword to be called")
	}
	if fakeSvc.changePasswordUserID != userID {
		t.Fatalf("expected userID %s, got %s", userID, fakeSvc.changePasswordUserID)
	}
	if len(rr.Result().Cookies()) != 0 {
		t.Fatalf("expected no Set-Cookie header on service error, got %d cookies", len(rr.Result().Cookies()))
	}
}

func TestAuthHandler_ChangePasswordSuccessClearsRefreshCookie(t *testing.T) {
	userID := uuid.New()
	fakeSvc := &fakeAuthService{}
	h := NewAuthHandler(fakeSvc, config.JWTConfig{}, "/v1/auth")

	req := authenticatedChangePasswordRequest(t, userID)
	rr := httptest.NewRecorder()

	withAuthMiddleware(h).ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
	if !fakeSvc.changePasswordCalled {
		t.Fatalf("expected ChangePassword to be called")
	}
	cookies := rr.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected one cookie to be cleared on success, got %d", len(cookies))
	}
	if cookies[0].Name != refreshTokenName || cookies[0].MaxAge != -1 {
		t.Fatalf("expected refresh_token cookie to be cleared, got %+v", cookies[0])
	}
}

func withAuthMiddleware(h *AuthHandler) http.Handler {
	authMiddleware := platformauth.AuthMiddleware(testAccessSecret, testJWTIssuer, testJWTAudience)
	return authMiddleware(http.HandlerFunc(h.ChangePassword))
}

func authenticatedChangePasswordRequest(t *testing.T, userID uuid.UUID) *http.Request {
	t.Helper()

	accessToken, err := jwt.Generate(jwt.Config{
		Secret:     testAccessSecret,
		Expiration: time.Hour,
		Issuer:     testJWTIssuer,
		Audience:   testJWTAudience,
	}, userID.String())
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/change-password", strings.NewReader(`{
		"current_password":"VeryStrongPassword123!",
		"new_password":"EvenStrongerPassword456!",
		"confirm_password":"EvenStrongerPassword456!"
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.AddCookie(&http.Cookie{
		Name:  refreshTokenName,
		Value: "refresh-token-value",
	})

	return req
}
