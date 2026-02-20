package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/repository"
	"github.com/mageas/the-punisher-backend/internal/testutil/inmemory"
)

func TestPunishmentServiceResolvePunishment(t *testing.T) {
	t.Parallel()

	t.Run("returns_not_found_when_punishment_does_not_exist", func(t *testing.T) {
		t.Parallel()

		repo := inmemory.NewRepository()
		svc := NewPunishmentService(repo)

		_, err := svc.ResolvePunishment(context.Background(), uuid.New(), uuid.New())
		if err != api.ErrPunishmentNotFound {
			t.Fatalf("expected err=%v, got=%v", api.ErrPunishmentNotFound, err)
		}
	})

	t.Run("returns_already_resolved_when_punishment_is_already_resolved", func(t *testing.T) {
		t.Parallel()

		repo := inmemory.NewRepository()
		svc := NewPunishmentService(repo)

		userID := uuid.New()
		punishmentID := uuid.New()
		now := time.Now()
		repo.SeedPunishment(repository.Punishment{
			ID:               punishmentID,
			UserID:           userID,
			StudentID:        uuid.New(),
			PunishmentTypeID: uuid.New(),
			DueAt:            time.Now().Add(24 * time.Hour),
			ResolvedAt:       &now,
		})

		_, err := svc.ResolvePunishment(context.Background(), userID, punishmentID)
		if err != api.ErrPunishmentAlreadyResolved {
			t.Fatalf("expected err=%v, got=%v", api.ErrPunishmentAlreadyResolved, err)
		}
	})

	t.Run("returns_internal_error_when_lookup_fails_after_no_rows", func(t *testing.T) {
		t.Parallel()

		repo := inmemory.NewRepository()
		svc := NewPunishmentService(repo)

		repo.SetError(inmemory.OpResolvePunishment, pgx.ErrNoRows)
		repo.SetError(inmemory.OpGetPunishmentByUser, errors.New("db down"))

		_, err := svc.ResolvePunishment(context.Background(), uuid.New(), uuid.New())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to get punishment: db down") {
			t.Fatalf("expected wrapped get error, got=%v", err)
		}
	})
}
