package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/config"
)

type fakeUserService struct {
	confirmErr   error
	confirmToken string
	resendErr    error
	resendEmail  string
}

func (f *fakeUserService) CreateUser(_ context.Context, _ dto.RequestUserDto) (*dto.ReturnUserDto, error) {
	return nil, nil
}

func (f *fakeUserService) ConfirmEmail(_ context.Context, token string) error {
	f.confirmToken = token
	return f.confirmErr
}

func (f *fakeUserService) ResendEmailConfirmation(_ context.Context, email string) error {
	f.resendEmail = email
	return f.resendErr
}

func (f *fakeUserService) GetCurrentUser(_ context.Context, _ uuid.UUID) (*dto.ReturnUserDto, error) {
	return nil, nil
}

func TestUserHandler_ConfirmEmailMissingToken(t *testing.T) {
	h := NewUserHandler(&fakeUserService{}, config.Config{})
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/confirm-email", nil)
	rr := httptest.NewRecorder()

	h.ConfirmEmail(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestUserHandler_ConfirmEmailSuccess(t *testing.T) {
	fakeSvc := &fakeUserService{}
	h := NewUserHandler(fakeSvc, config.Config{})
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/confirm-email?token=test-token", nil)
	rr := httptest.NewRecorder()

	h.ConfirmEmail(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if fakeSvc.confirmToken != "test-token" {
		t.Fatalf("expected token to be passed to service, got %q", fakeSvc.confirmToken)
	}

	var payload dto.ConfirmEmailResponseDto
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if payload.Status != "email_confirmed" {
		t.Fatalf("unexpected status payload: %+v", payload)
	}
}

func TestUserHandler_ConfirmEmailServiceError(t *testing.T) {
	h := NewUserHandler(&fakeUserService{confirmErr: api.ErrEmailConfirmationTokenExpired}, config.Config{})
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/confirm-email?token=test-token", nil)
	rr := httptest.NewRecorder()

	h.ConfirmEmail(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestUserHandler_ResendConfirmEmailSuccess(t *testing.T) {
	fakeSvc := &fakeUserService{}
	h := NewUserHandler(fakeSvc, config.Config{})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/confirm-email/resend", strings.NewReader(`{"email":"teacher@school.test"}`))
	rr := httptest.NewRecorder()

	h.ResendConfirmEmail(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if fakeSvc.resendEmail != "teacher@school.test" {
		t.Fatalf("expected email to be passed to service, got %q", fakeSvc.resendEmail)
	}
}

func TestUserHandler_ResendConfirmEmailValidationError(t *testing.T) {
	h := NewUserHandler(&fakeUserService{}, config.Config{})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/confirm-email/resend", strings.NewReader(`{"email":"bad"}`))
	rr := httptest.NewRecorder()

	h.ResendConfirmEmail(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}
