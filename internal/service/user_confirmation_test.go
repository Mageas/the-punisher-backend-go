package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/platform/config"
	platformjwt "github.com/mageas/the-punisher-backend/internal/platform/jwt"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type noopConfirmationMailer struct{}

func (noopConfirmationMailer) SendConfirmationEmail(
	_ context.Context,
	_ string,
	_ string,
	_ string,
	_ time.Duration,
) error {
	return nil
}

type fakeConfirmationRepo struct {
	repository.Querier
	token             repository.EmailConfirmationToken
	tokenErr          error
	verificationState repository.GetUserEmailVerificationStateByIDRow
	verificationErr   error
	verifyRows        int64
	verifyErr         error
	markRows          int64
	markErr           error
}

func (f *fakeConfirmationRepo) WithinTransaction(_ context.Context, fn func(repository.Querier) error) error {
	return fn(f)
}

func (f *fakeConfirmationRepo) GetEmailConfirmationTokenByHash(_ context.Context, _ string) (repository.EmailConfirmationToken, error) {
	if f.tokenErr != nil {
		return repository.EmailConfirmationToken{}, f.tokenErr
	}
	return f.token, nil
}

func (f *fakeConfirmationRepo) GetUserEmailVerificationStateByID(_ context.Context, _ uuid.UUID) (repository.GetUserEmailVerificationStateByIDRow, error) {
	if f.verificationErr != nil {
		return repository.GetUserEmailVerificationStateByIDRow{}, f.verificationErr
	}
	return f.verificationState, nil
}

func (f *fakeConfirmationRepo) VerifyUserEmailByID(_ context.Context, _ uuid.UUID) (int64, error) {
	if f.verifyErr != nil {
		return 0, f.verifyErr
	}
	return f.verifyRows, nil
}

func (f *fakeConfirmationRepo) MarkEmailConfirmationTokenUsedByID(_ context.Context, _ uuid.UUID) (int64, error) {
	if f.markErr != nil {
		return 0, f.markErr
	}
	return f.markRows, nil
}

func testEmailConfirmationConfig() config.EmailConfirmationConfig {
	return config.EmailConfirmationConfig{
		Secret:     "email-confirm-secret",
		Expiration: 24 * time.Hour,
		BaseURL:    "http://localhost:8080/v1/auth/confirm-email",
		Issuer:     "issuer",
		Audience:   "audience",
	}
}

func generateTestConfirmationToken(t *testing.T, cfg config.EmailConfirmationConfig, userID uuid.UUID) string {
	t.Helper()

	token, err := platformjwt.Generate(platformjwt.Config{
		Secret:     cfg.Secret,
		Expiration: cfg.Expiration,
		Issuer:     cfg.Issuer,
		Audience:   cfg.Audience,
	}, userID.String())
	if err != nil {
		t.Fatalf("failed to generate confirmation token: %v", err)
	}

	return token
}

func TestUserService_ConfirmEmailUserNotFound(t *testing.T) {
	cfg := testEmailConfirmationConfig()
	userID := uuid.New()
	token := generateTestConfirmationToken(t, cfg, userID)

	repo := &fakeConfirmationRepo{
		token: repository.EmailConfirmationToken{
			ID:        uuid.New(),
			UserID:    userID,
			TokenHash: "unused-in-fake",
			ExpiresAt: time.Now().Add(time.Hour),
		},
		verificationErr: repository.ErrNoRows,
	}

	svc := NewUserServiceWithEmailConfirmation(repo, cfg, noopConfirmationMailer{})

	err := svc.ConfirmEmail(context.Background(), token)
	if !errors.Is(err, api.ErrEmailConfirmationUserNotFound) {
		t.Fatalf("expected ErrEmailConfirmationUserNotFound, got %v", err)
	}
}

func TestUserService_ConfirmEmailAlreadyVerified(t *testing.T) {
	cfg := testEmailConfirmationConfig()
	userID := uuid.New()
	token := generateTestConfirmationToken(t, cfg, userID)
	alreadyVerifiedAt := time.Now().Add(-time.Hour)

	repo := &fakeConfirmationRepo{
		token: repository.EmailConfirmationToken{
			ID:        uuid.New(),
			UserID:    userID,
			TokenHash: "unused-in-fake",
			ExpiresAt: time.Now().Add(time.Hour),
		},
		verificationState: repository.GetUserEmailVerificationStateByIDRow{
			ID:              userID,
			EmailVerifiedAt: &alreadyVerifiedAt,
		},
	}

	svc := NewUserServiceWithEmailConfirmation(repo, cfg, noopConfirmationMailer{})

	err := svc.ConfirmEmail(context.Background(), token)
	if !errors.Is(err, api.ErrEmailAlreadyVerified) {
		t.Fatalf("expected ErrEmailAlreadyVerified, got %v", err)
	}
}
