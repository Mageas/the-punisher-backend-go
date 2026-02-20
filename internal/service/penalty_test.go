package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mageas/the-punisher-backend/internal/repository"
	"github.com/mageas/the-punisher-backend/internal/testutil/inmemory"
)

type nonTransactionalPenaltyRepo struct {
	repository.Querier
}

func TestPenaltyServiceCreatePenalty(t *testing.T) {
	t.Parallel()

	t.Run("creates_penalty_and_rule_based_punishment_in_transaction", func(t *testing.T) {
		t.Parallel()

		repo := inmemory.NewRepository()
		svc := NewPenaltyService(repo)

		userID := uuid.New()
		studentID := uuid.New()
		penaltyTypeID := uuid.New()
		punishmentTypeID := uuid.New()
		ruleID := uuid.New()

		repo.SeedStudent(repository.Student{ID: studentID, UserID: userID, FirstName: "Alice", LastName: "Martin"})
		repo.SeedPenaltyType(repository.PenaltyType{ID: penaltyTypeID, UserID: userID, Name: "Retard"})
		repo.SeedPunishmentType(repository.PunishmentType{ID: punishmentTypeID, UserID: userID, Name: "Retenue"})
		repo.SeedRule(repository.Rule{
			ID:                        ruleID,
			UserID:                    userID,
			Name:                      "1 retard => retenue",
			ResultingPunishmentTypeID: punishmentTypeID,
			PenaltyTypeID:             penaltyTypeID,
			Threshold:                 1,
			Mode:                      "at",
			DueAtAfterDays:            3,
			IsActive:                  true,
		})

		start := time.Now().UTC()
		_, err := svc.CreatePenalty(context.Background(), userID, studentID, penaltyTypeID)
		end := time.Now().UTC()
		if err != nil {
			t.Fatalf("expected no error, got=%v", err)
		}

		penaltiesCount, err := repo.CountPenaltiesByStudentAndType(context.Background(), repository.CountPenaltiesByStudentAndTypeParams{
			StudentID:     studentID,
			UserID:        userID,
			PenaltyTypeID: penaltyTypeID,
		})
		if err != nil {
			t.Fatalf("expected no error counting penalties, got=%v", err)
		}
		if penaltiesCount != 1 {
			t.Fatalf("expected penalties count=1, got=%d", penaltiesCount)
		}

		punishments, err := repo.ListPunishmentsByStudent(context.Background(), repository.ListPunishmentsByStudentParams{
			StudentID:   studentID,
			UserID:      userID,
			Resolved:    pgtype.Bool{},
			QueryLimit:  10,
			QueryOffset: 0,
		})
		if err != nil {
			t.Fatalf("expected no error listing punishments, got=%v", err)
		}
		if len(punishments) != 1 {
			t.Fatalf("expected one automated punishment, got=%d", len(punishments))
		}

		punishment := punishments[0]
		if !punishment.Automated {
			t.Fatal("expected punishment to be automated")
		}
		if !punishment.TriggeringRuleID.Valid {
			t.Fatal("expected triggering_rule_id to be set")
		}
		if uuid.UUID(punishment.TriggeringRuleID.Bytes) != ruleID {
			t.Fatalf("expected triggering_rule_id=%s, got=%s", ruleID, uuid.UUID(punishment.TriggeringRuleID.Bytes))
		}

		dueAt := punishment.DueAt.UTC()
		minExpected := start.Add(3 * 24 * time.Hour)
		maxExpected := end.Add(3 * 24 * time.Hour).Add(2 * time.Second)
		if dueAt.Before(minExpected) || dueAt.After(maxExpected) {
			t.Fatalf("expected due_at between %s and %s, got=%s", minExpected, maxExpected, dueAt)
		}
	})

	t.Run("rolls_back_transaction_when_rule_based_punishment_creation_fails", func(t *testing.T) {
		t.Parallel()

		repo := inmemory.NewRepository()
		svc := NewPenaltyService(repo)

		userID := uuid.New()
		studentID := uuid.New()
		penaltyTypeID := uuid.New()
		punishmentTypeID := uuid.New()

		repo.SeedStudent(repository.Student{ID: studentID, UserID: userID, FirstName: "Alice", LastName: "Martin"})
		repo.SeedPenaltyType(repository.PenaltyType{ID: penaltyTypeID, UserID: userID, Name: "Retard"})
		repo.SeedPunishmentType(repository.PunishmentType{ID: punishmentTypeID, UserID: userID, Name: "Retenue"})
		repo.SeedRule(repository.Rule{
			ID:                        uuid.New(),
			UserID:                    userID,
			Name:                      "1 retard => retenue",
			ResultingPunishmentTypeID: punishmentTypeID,
			PenaltyTypeID:             penaltyTypeID,
			Threshold:                 1,
			Mode:                      "at",
			DueAtAfterDays:            1,
			IsActive:                  true,
		})
		repo.SetError(inmemory.OpCreatePunishmentFromRule, errors.New("insert failure"))

		_, err := svc.CreatePenalty(context.Background(), userID, studentID, penaltyTypeID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to create punishment from rule: insert failure") {
			t.Fatalf("expected wrapped punishment error, got=%v", err)
		}

		penaltiesCount, err := repo.CountPenaltiesByStudentAndType(context.Background(), repository.CountPenaltiesByStudentAndTypeParams{
			StudentID:     studentID,
			UserID:        userID,
			PenaltyTypeID: penaltyTypeID,
		})
		if err != nil {
			t.Fatalf("expected no error counting penalties, got=%v", err)
		}
		if penaltiesCount != 0 {
			t.Fatalf("expected rollback to keep penalties count=0, got=%d", penaltiesCount)
		}

		punishmentsCount, err := repo.CountPunishmentsByStudent(context.Background(), repository.CountPunishmentsByStudentParams{
			StudentID: studentID,
			UserID:    userID,
			Resolved:  pgtype.Bool{},
		})
		if err != nil {
			t.Fatalf("expected no error counting punishments, got=%v", err)
		}
		if punishmentsCount != 0 {
			t.Fatalf("expected rollback to keep punishments count=0, got=%d", punishmentsCount)
		}
	})

	t.Run("returns_explicit_error_when_repository_is_not_transactional", func(t *testing.T) {
		t.Parallel()

		repo := nonTransactionalPenaltyRepo{Querier: inmemory.NewRepository()}
		svc := NewPenaltyService(repo)

		_, err := svc.CreatePenalty(context.Background(), uuid.New(), uuid.New(), uuid.New())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "does not support transactions") {
			t.Fatalf("expected transaction capability error, got=%v", err)
		}
	})
}

func TestShouldTriggerRule(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		mode      string
		threshold int32
		count     int64
		want      bool
	}{
		{
			name:      "at_triggers_when_count_equals_threshold",
			mode:      "at",
			threshold: 3,
			count:     3,
			want:      true,
		},
		{
			name:      "at_does_not_trigger_when_count_is_greater",
			mode:      "at",
			threshold: 3,
			count:     4,
			want:      false,
		},
		{
			name:      "every_does_not_trigger_at_zero",
			mode:      "every",
			threshold: 3,
			count:     0,
			want:      false,
		},
		{
			name:      "every_triggers_on_multiple",
			mode:      "every",
			threshold: 3,
			count:     6,
			want:      true,
		},
		{
			name:      "every_does_not_trigger_on_non_multiple",
			mode:      "every",
			threshold: 3,
			count:     5,
			want:      false,
		},
		{
			name:      "after_triggers_when_count_is_greater",
			mode:      "after",
			threshold: 3,
			count:     4,
			want:      true,
		},
		{
			name:      "after_does_not_trigger_when_count_equals_threshold",
			mode:      "after",
			threshold: 3,
			count:     3,
			want:      false,
		},
		{
			name:      "invalid_threshold_never_triggers",
			mode:      "every",
			threshold: 0,
			count:     10,
			want:      false,
		},
		{
			name:      "invalid_mode_never_triggers",
			mode:      "unknown",
			threshold: 3,
			count:     3,
			want:      false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := shouldTriggerRule(tt.mode, tt.threshold, tt.count)
			if got != tt.want {
				t.Fatalf("shouldTriggerRule(%q, %d, %d) = %v, want %v", tt.mode, tt.threshold, tt.count, got, tt.want)
			}
		})
	}
}
