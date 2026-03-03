package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/config"
	platformjwt "github.com/mageas/the-punisher-backend/internal/platform/jwt"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type noopPasswordResetMailer struct{}

func (noopPasswordResetMailer) SendPasswordResetEmail(
	_ context.Context,
	_ string,
	_ string,
	_ string,
	_ time.Duration,
) error {
	return nil
}

type fakePasswordResetRepo struct {
	repository.Querier
	token         repository.PasswordResetToken
	tokenErr      error
	updatedRows   int64
	updateErr     error
	deletedRows   int64
	deleteErr     error
	markedRows    int64
	markErr       error
	invalidateErr error
}

func (f *fakePasswordResetRepo) WithinTransaction(_ context.Context, fn func(repository.Querier) error) error {
	return fn(f)
}

func (f *fakePasswordResetRepo) GetPasswordResetTokenByHash(_ context.Context, _ string) (repository.PasswordResetToken, error) {
	if f.tokenErr != nil {
		return repository.PasswordResetToken{}, f.tokenErr
	}
	return f.token, nil
}

func (f *fakePasswordResetRepo) UpdateUserPasswordByID(_ context.Context, _ repository.UpdateUserPasswordByIDParams) (int64, error) {
	if f.updateErr != nil {
		return 0, f.updateErr
	}
	return f.updatedRows, nil
}

func (f *fakePasswordResetRepo) DeleteRefreshTokensByUserId(_ context.Context, _ uuid.UUID) (int64, error) {
	if f.deleteErr != nil {
		return 0, f.deleteErr
	}
	return f.deletedRows, nil
}

func (f *fakePasswordResetRepo) MarkPasswordResetTokenUsedByID(_ context.Context, _ uuid.UUID) (int64, error) {
	if f.markErr != nil {
		return 0, f.markErr
	}
	return f.markedRows, nil
}

func (f *fakePasswordResetRepo) InvalidatePasswordResetTokensByUserID(_ context.Context, _ uuid.UUID) (int64, error) {
	if f.invalidateErr != nil {
		return 0, f.invalidateErr
	}
	return 1, nil
}

func testPasswordResetConfig() config.PasswordResetConfig {
	return config.PasswordResetConfig{
		Secret:     "password-reset-secret",
		Expiration: time.Hour,
		BaseURL:    "http://localhost:3000/reset-password",
		Issuer:     "issuer",
		Audience:   "audience",
	}
}

func generateTestPasswordResetToken(t *testing.T, cfg config.PasswordResetConfig, userID uuid.UUID) string {
	t.Helper()

	token, err := platformjwt.Generate(platformjwt.Config{
		Secret:     cfg.Secret,
		Expiration: cfg.Expiration,
		Issuer:     cfg.Issuer,
		Audience:   cfg.Audience,
	}, userID.String())
	if err != nil {
		t.Fatalf("failed to generate reset token: %v", err)
	}

	return token
}

func TestAuthService_ResetPasswordMissingToken(t *testing.T) {
	svc := NewAuthServiceWithPasswordReset(&fakePasswordResetRepo{}, config.JWTConfig{}, testPasswordResetConfig(), noopPasswordResetMailer{})

	err := svc.ResetPassword(context.Background(), dto.ResetPasswordRequestDto{
		Token:           "",
		NewPassword:     "NewSecurePass1!",
		ConfirmPassword: "NewSecurePass1!",
	})
	if !errors.Is(err, api.ErrPasswordResetTokenMissing) {
		t.Fatalf("expected ErrPasswordResetTokenMissing, got %v", err)
	}
}

func TestAuthService_ResetPasswordMismatchConfirmation(t *testing.T) {
	svc := NewAuthServiceWithPasswordReset(&fakePasswordResetRepo{}, config.JWTConfig{}, testPasswordResetConfig(), noopPasswordResetMailer{})

	err := svc.ResetPassword(context.Background(), dto.ResetPasswordRequestDto{
		Token:           "will-not-be-used",
		NewPassword:     "NewSecurePass1!",
		ConfirmPassword: "DifferentSecurePass1!",
	})

	var apiErr *api.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %v", err)
	}
	if apiErr.Message != api.ErrValidationFailed.Message {
		t.Fatalf("expected validation_failed, got %s", apiErr.Message)
	}
	if len(apiErr.Details) != 1 {
		t.Fatalf("expected one validation detail, got %+v", apiErr.Details)
	}
	if apiErr.Details[0].Field != "confirm_password" || apiErr.Details[0].Error != api.KeyValidationPasswordConfirmationMismatch {
		t.Fatalf("unexpected validation detail: %+v", apiErr.Details[0])
	}
}

func TestAuthService_ResetPasswordUserNotFound(t *testing.T) {
	cfg := testPasswordResetConfig()
	userID := uuid.New()
	token := generateTestPasswordResetToken(t, cfg, userID)

	repo := &fakePasswordResetRepo{
		token: repository.PasswordResetToken{
			ID:        uuid.New(),
			UserID:    userID,
			TokenHash: "unused-in-fake",
			ExpiresAt: time.Now().Add(time.Hour),
		},
		updatedRows: 0,
	}

	svc := NewAuthServiceWithPasswordReset(repo, config.JWTConfig{}, cfg, noopPasswordResetMailer{})

	err := svc.ResetPassword(context.Background(), dto.ResetPasswordRequestDto{
		Token:           token,
		NewPassword:     "NewSecurePass1!",
		ConfirmPassword: "NewSecurePass1!",
	})
	if !errors.Is(err, api.ErrPasswordResetUserNotFound) {
		t.Fatalf("expected ErrPasswordResetUserNotFound, got %v", err)
	}
}
