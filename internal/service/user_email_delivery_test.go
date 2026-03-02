package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type spyConfirmationMailer struct {
	sendCalls int
	sendErr   error
}

func (m *spyConfirmationMailer) SendConfirmationEmail(
	_ context.Context,
	_ string,
	_ string,
	_ string,
	_ time.Duration,
) error {
	m.sendCalls++
	return m.sendErr
}

type fakeTransactionalUserRepo struct {
	repository.Querier
	commitErr error

	userEmailExists bool
	createUserRow   repository.CreateUserRow

	verificationRow repository.GetUserEmailVerificationStateByEmailRow
	verificationErr error
}

func (f *fakeTransactionalUserRepo) WithinTransaction(_ context.Context, fn func(repository.Querier) error) error {
	if err := fn(f); err != nil {
		return err
	}
	return f.commitErr
}

func (f *fakeTransactionalUserRepo) UserEmailExists(_ context.Context, _ string) (bool, error) {
	return f.userEmailExists, nil
}

func (f *fakeTransactionalUserRepo) CreateUser(_ context.Context, _ repository.CreateUserParams) (repository.CreateUserRow, error) {
	return f.createUserRow, nil
}

func (f *fakeTransactionalUserRepo) CreateEmailConfirmationToken(_ context.Context, arg repository.CreateEmailConfirmationTokenParams) (repository.EmailConfirmationToken, error) {
	return repository.EmailConfirmationToken{
		ID:        uuid.New(),
		UserID:    arg.UserID,
		TokenHash: arg.TokenHash,
		ExpiresAt: arg.ExpiresAt,
		CreatedAt: time.Now(),
	}, nil
}

func (f *fakeTransactionalUserRepo) GetUserEmailVerificationStateByEmail(_ context.Context, _ string) (repository.GetUserEmailVerificationStateByEmailRow, error) {
	if f.verificationErr != nil {
		return repository.GetUserEmailVerificationStateByEmailRow{}, f.verificationErr
	}
	return f.verificationRow, nil
}

func (f *fakeTransactionalUserRepo) InvalidateEmailConfirmationTokensByUserID(_ context.Context, _ uuid.UUID) (int64, error) {
	return 1, nil
}

func TestUserService_CreateUserDoesNotSendConfirmationEmailWhenCommitFails(t *testing.T) {
	commitErr := errors.New("failed to commit transaction")
	repo := &fakeTransactionalUserRepo{
		commitErr: commitErr,
		createUserRow: repository.CreateUserRow{
			ID:        uuid.New(),
			Email:     "john.doe@example.com",
			FirstName: "John",
			LastName:  "Doe",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	mailer := &spyConfirmationMailer{}

	svc := NewUserServiceWithEmailConfirmation(repo, testEmailConfirmationConfig(), mailer)

	_, err := svc.CreateUser(context.Background(), dto.RequestUserDto{
		Email:     "john.doe@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Password:  "password123",
	})
	if !errors.Is(err, commitErr) {
		t.Fatalf("expected commit error, got %v", err)
	}
	if mailer.sendCalls != 0 {
		t.Fatalf("expected no email to be sent when commit fails, got %d calls", mailer.sendCalls)
	}
}

func TestUserService_ResendEmailConfirmationDoesNotSendEmailWhenCommitFails(t *testing.T) {
	commitErr := errors.New("failed to commit transaction")
	repo := &fakeTransactionalUserRepo{
		commitErr: commitErr,
		verificationRow: repository.GetUserEmailVerificationStateByEmailRow{
			ID:        uuid.New(),
			Email:     "john.doe@example.com",
			FirstName: "John",
		},
	}
	mailer := &spyConfirmationMailer{}

	svc := NewUserServiceWithEmailConfirmation(repo, testEmailConfirmationConfig(), mailer)

	err := svc.ResendEmailConfirmation(context.Background(), "john.doe@example.com")
	if !errors.Is(err, commitErr) {
		t.Fatalf("expected commit error, got %v", err)
	}
	if mailer.sendCalls != 0 {
		t.Fatalf("expected no email to be sent when commit fails, got %d calls", mailer.sendCalls)
	}
}
