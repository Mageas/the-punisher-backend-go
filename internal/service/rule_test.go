package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
	"github.com/mageas/the-punisher-backend/internal/testutil/inmemory"
)

func TestRuleServiceCreateRule(t *testing.T) {
	t.Parallel()

	t.Run("uses_default_is_active_true_when_omitted", func(t *testing.T) {
		t.Parallel()

		repo := inmemory.NewRepository()
		svc := NewRuleService(repo)

		userID := uuid.New()
		penaltyTypeID := uuid.New()
		punishmentTypeID := uuid.New()

		repo.SeedPenaltyType(repository.PenaltyType{ID: penaltyTypeID, UserID: userID, Name: "Retard"})
		repo.SeedPunishmentType(repository.PunishmentType{ID: punishmentTypeID, UserID: userID, Name: "Retenue"})

		rule, err := svc.CreateRule(context.Background(), userID, dto.RequestRuleDto{
			Name:                      "2 retards => retenue",
			ResultingPunishmentTypeID: punishmentTypeID.String(),
			PenaltyTypeID:             penaltyTypeID.String(),
			Threshold:                 2,
			DueAtAfterDays:            3,
			Mode:                      "at",
		})
		if err != nil {
			t.Fatalf("expected no error, got=%v", err)
		}
		if !rule.IsActive {
			t.Fatal("expected is_active=true by default")
		}
	})

	t.Run("returns_invalid_request_body_when_punishment_type_id_is_invalid_uuid", func(t *testing.T) {
		t.Parallel()

		repo := inmemory.NewRepository()
		svc := NewRuleService(repo)

		_, err := svc.CreateRule(context.Background(), uuid.New(), dto.RequestRuleDto{
			Name:                      "Rule test",
			ResultingPunishmentTypeID: "not-a-uuid",
			PenaltyTypeID:             uuid.New().String(),
			Threshold:                 1,
			DueAtAfterDays:            0,
			Mode:                      "at",
		})
		if err != api.ErrInvalidRequestBody {
			t.Fatalf("expected err=%v, got=%v", api.ErrInvalidRequestBody, err)
		}
	})
}

func TestRuleServiceUpdateRule(t *testing.T) {
	t.Parallel()

	repo := inmemory.NewRepository()
	svc := NewRuleService(repo)

	invalidID := "not-a-uuid"
	_, err := svc.UpdateRule(context.Background(), uuid.New(), uuid.New(), dto.UpdateRuleDto{
		ResultingPunishmentTypeID: &invalidID,
	})
	if err != api.ErrInvalidRequestBody {
		t.Fatalf("expected err=%v, got=%v", api.ErrInvalidRequestBody, err)
	}
}
