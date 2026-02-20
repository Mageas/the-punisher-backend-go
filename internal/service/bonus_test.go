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

func TestBonusServiceUseBonus(t *testing.T) {
	t.Parallel()

	t.Run("returns_not_found_when_bonus_does_not_exist", func(t *testing.T) {
		t.Parallel()

		repo := inmemory.NewRepository()
		svc := NewBonusService(repo)

		_, err := svc.UseBonus(context.Background(), uuid.New(), uuid.New())
		if err != api.ErrBonusNotFound {
			t.Fatalf("expected err=%v, got=%v", api.ErrBonusNotFound, err)
		}
	})

	t.Run("returns_already_used_when_bonus_is_already_consumed", func(t *testing.T) {
		t.Parallel()

		repo := inmemory.NewRepository()
		svc := NewBonusService(repo)

		userID := uuid.New()
		bonusID := uuid.New()
		now := time.Now()
		repo.SeedBonus(repository.Bonus{
			ID:          bonusID,
			UserID:      userID,
			StudentID:   uuid.New(),
			BonusTypeID: uuid.New(),
			Points:      1,
			UsedAt:      &now,
		})

		_, err := svc.UseBonus(context.Background(), userID, bonusID)
		if err != api.ErrBonusAlreadyUsed {
			t.Fatalf("expected err=%v, got=%v", api.ErrBonusAlreadyUsed, err)
		}
	})

	t.Run("returns_internal_error_when_lookup_fails_after_no_rows", func(t *testing.T) {
		t.Parallel()

		repo := inmemory.NewRepository()
		svc := NewBonusService(repo)

		repo.SetError(inmemory.OpUseBonus, pgx.ErrNoRows)
		repo.SetError(inmemory.OpGetBonusByUser, errors.New("db down"))

		_, err := svc.UseBonus(context.Background(), uuid.New(), uuid.New())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to get bonus: db down") {
			t.Fatalf("expected wrapped get error, got=%v", err)
		}
	})
}
