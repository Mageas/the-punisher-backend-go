//go:build integration

package integration

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/config"
	. "github.com/mageas/the-punisher-backend/internal/service"
)

type capturedConfirmationEmail struct {
	ToEmail         string
	FirstName       string
	ConfirmationURL string
	ExpiresIn       time.Duration
}

type fakeConfirmationMailer struct {
	emails []capturedConfirmationEmail
	err    error
}

func (f *fakeConfirmationMailer) SendConfirmationEmail(
	_ context.Context,
	toEmail string,
	firstName string,
	confirmationURL string,
	expiresIn time.Duration,
) error {
	if f.err != nil {
		return f.err
	}

	f.emails = append(f.emails, capturedConfirmationEmail{
		ToEmail:         toEmail,
		FirstName:       firstName,
		ConfirmationURL: confirmationURL,
		ExpiresIn:       expiresIn,
	})
	return nil
}

func integrationEmailConfirmationConfig() config.EmailConfirmationConfig {
	return config.EmailConfirmationConfig{
		Secret:     "email-confirm-secret",
		Expiration: 24 * time.Hour,
		BaseURL:    "http://localhost:8080/v1/auth/confirm-email",
		Issuer:     "test-issuer",
		Audience:   "test-audience",
	}
}

func tokenFromConfirmationURL(t *testing.T, confirmationURL string) string {
	t.Helper()

	parsed, err := url.Parse(confirmationURL)
	if err != nil {
		t.Fatalf("failed to parse confirmation URL: %v", err)
	}

	token := strings.TrimSpace(parsed.Query().Get("token"))
	if token == "" {
		t.Fatalf("missing token in confirmation URL: %s", confirmationURL)
	}

	return token
}

func TestUserService_CreateUserAndGetCurrentUser_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	svc := NewUserService(repo)

	created, err := svc.CreateUser(ctx, dto.RequestUserDto{
		Email:     "TeSt.User@example.com",
		FirstName: "Test",
		LastName:  "User",
		Password:  "VeryStrongPassword123",
	})
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}
	if created.ID == uuid.Nil {
		t.Fatalf("expected non-nil user id")
	}
	if created.Email != "test.user@example.com" {
		t.Fatalf("expected lowercased email, got %s", created.Email)
	}
	if created.Timezone != testUserTimezone {
		t.Fatalf("expected created user timezone %s, got %s", testUserTimezone, created.Timezone)
	}

	storedUser, err := repo.GetUserByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetUserByID returned error: %v", err)
	}
	if storedUser.Timezone != "Europe/Paris" {
		t.Fatalf("expected default timezone Europe/Paris, got %s", storedUser.Timezone)
	}

	current, err := svc.GetCurrentUser(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetCurrentUser returned error: %v", err)
	}
	if current.ID != created.ID {
		t.Fatalf("expected same user id, got %s vs %s", current.ID, created.ID)
	}
	if current.Timezone != testUserTimezone {
		t.Fatalf("expected current user timezone %s, got %s", testUserTimezone, current.Timezone)
	}
}

func TestUserService_CreateUserDuplicateEmail_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	svc := NewUserService(repo)

	_, err := svc.CreateUser(ctx, dto.RequestUserDto{
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Password:  "VeryStrongPassword123",
	})
	if err != nil {
		t.Fatalf("CreateUser first call returned error: %v", err)
	}

	_, err = svc.CreateUser(ctx, dto.RequestUserDto{
		Email:     "TEST@example.com",
		FirstName: "Jane",
		LastName:  "Doe",
		Password:  "VeryStrongPassword123",
	})
	if !errors.Is(err, api.ErrEmailAlreadyExists) {
		t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestUserService_GetCurrentUserNotFound_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	svc := NewUserService(repo)

	_, err := svc.GetCurrentUser(ctx, uuid.New())
	if !errors.Is(err, api.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestUserService_CreateUserAndConfirmEmailFlow_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	mailer := &fakeConfirmationMailer{}
	svc := NewUserServiceWithEmailConfirmation(repo, integrationEmailConfirmationConfig(), mailer)

	created, err := svc.CreateUser(ctx, dto.RequestUserDto{
		Email:     "confirm.user@example.com",
		FirstName: "Confirm",
		LastName:  "User",
		Password:  "VeryStrongPassword123",
	})
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	if len(mailer.emails) != 1 {
		t.Fatalf("expected one confirmation email, got %d", len(mailer.emails))
	}

	confirmationToken := tokenFromConfirmationURL(t, mailer.emails[0].ConfirmationURL)
	if err := svc.ConfirmEmail(ctx, confirmationToken); err != nil {
		t.Fatalf("ConfirmEmail returned error: %v", err)
	}

	verificationState, err := repo.GetUserEmailVerificationStateByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get user verification state: %v", err)
	}
	if verificationState.EmailVerifiedAt == nil {
		t.Fatalf("expected email to be verified")
	}

	err = svc.ConfirmEmail(ctx, confirmationToken)
	if !errors.Is(err, api.ErrEmailConfirmationTokenAlreadyUsed) {
		t.Fatalf("expected ErrEmailConfirmationTokenAlreadyUsed, got %v", err)
	}
}

func TestUserService_ConfirmEmailInvalidToken_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	svc := NewUserServiceWithEmailConfirmation(repo, integrationEmailConfirmationConfig(), &fakeConfirmationMailer{})

	err := svc.ConfirmEmail(ctx, "not-a-jwt-token")
	if !errors.Is(err, api.ErrEmailConfirmationTokenInvalid) {
		t.Fatalf("expected ErrEmailConfirmationTokenInvalid, got %v", err)
	}
}

func TestUserService_ConfirmEmailExpiredToken_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	expiredConfig := integrationEmailConfirmationConfig()
	expiredConfig.Expiration = 1 * time.Millisecond

	mailer := &fakeConfirmationMailer{}
	svc := NewUserServiceWithEmailConfirmation(repo, expiredConfig, mailer)

	_, err := svc.CreateUser(ctx, dto.RequestUserDto{
		Email:     "expired.confirm.user@example.com",
		FirstName: "Expired",
		LastName:  "Confirm",
		Password:  "VeryStrongPassword123",
	})
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	if len(mailer.emails) != 1 {
		t.Fatalf("expected one confirmation email, got %d", len(mailer.emails))
	}

	confirmationToken := tokenFromConfirmationURL(t, mailer.emails[0].ConfirmationURL)
	time.Sleep(25 * time.Millisecond)
	err = svc.ConfirmEmail(ctx, confirmationToken)
	if !errors.Is(err, api.ErrEmailConfirmationTokenExpired) {
		t.Fatalf("expected ErrEmailConfirmationTokenExpired, got %v", err)
	}
}

func TestUserService_ResendEmailConfirmationFlow_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	mailer := &fakeConfirmationMailer{}
	svc := NewUserServiceWithEmailConfirmation(repo, integrationEmailConfirmationConfig(), mailer)

	_, err := svc.CreateUser(ctx, dto.RequestUserDto{
		Email:     "resend.confirm.user@example.com",
		FirstName: "Resend",
		LastName:  "User",
		Password:  "VeryStrongPassword123",
	})
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	if len(mailer.emails) != 1 {
		t.Fatalf("expected one confirmation email after register, got %d", len(mailer.emails))
	}

	firstToken := tokenFromConfirmationURL(t, mailer.emails[0].ConfirmationURL)

	if err := svc.ResendEmailConfirmation(ctx, "RESEND.CONFIRM.USER@EXAMPLE.COM"); err != nil {
		t.Fatalf("ResendEmailConfirmation returned error: %v", err)
	}

	if len(mailer.emails) != 2 {
		t.Fatalf("expected second confirmation email after resend, got %d", len(mailer.emails))
	}

	secondToken := tokenFromConfirmationURL(t, mailer.emails[1].ConfirmationURL)
	if secondToken == firstToken {
		t.Fatalf("expected a newly generated confirmation token")
	}

	err = svc.ConfirmEmail(ctx, firstToken)
	if !errors.Is(err, api.ErrEmailConfirmationTokenAlreadyUsed) {
		t.Fatalf("expected old token to be invalidated, got %v", err)
	}

	if err := svc.ConfirmEmail(ctx, secondToken); err != nil {
		t.Fatalf("ConfirmEmail with resent token returned error: %v", err)
	}
}

func TestUserService_ResendEmailConfirmationUnknownOrAlreadyVerified_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	mailer := &fakeConfirmationMailer{}
	svc := NewUserServiceWithEmailConfirmation(repo, integrationEmailConfirmationConfig(), mailer)

	if err := svc.ResendEmailConfirmation(ctx, "missing.user@example.com"); err != nil {
		t.Fatalf("ResendEmailConfirmation on unknown user returned error: %v", err)
	}
	if len(mailer.emails) != 0 {
		t.Fatalf("expected no confirmation email for unknown user, got %d", len(mailer.emails))
	}

	_, err := svc.CreateUser(ctx, dto.RequestUserDto{
		Email:     "verified.resend.user@example.com",
		FirstName: "Verified",
		LastName:  "Resend",
		Password:  "VeryStrongPassword123",
	})
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	if len(mailer.emails) != 1 {
		t.Fatalf("expected one confirmation email after register, got %d", len(mailer.emails))
	}

	confirmationToken := tokenFromConfirmationURL(t, mailer.emails[0].ConfirmationURL)
	if err := svc.ConfirmEmail(ctx, confirmationToken); err != nil {
		t.Fatalf("ConfirmEmail returned error: %v", err)
	}

	if err := svc.ResendEmailConfirmation(ctx, "verified.resend.user@example.com"); err != nil {
		t.Fatalf("ResendEmailConfirmation on verified user returned error: %v", err)
	}
	if len(mailer.emails) != 1 {
		t.Fatalf("expected no additional email for already verified user, got %d", len(mailer.emails))
	}
}
