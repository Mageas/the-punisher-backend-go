//go:build integration

package integration

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	. "github.com/mageas/the-punisher-backend/internal/service"
)

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

	current, err := svc.GetCurrentUser(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetCurrentUser returned error: %v", err)
	}
	if current.ID != created.ID {
		t.Fatalf("expected same user id, got %s vs %s", current.ID, created.ID)
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
