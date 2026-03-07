//go:build integration

package integration

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

func TestRepository_CreateBonusRejectsCrossUserReferences(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	otherUser := mustCreateUserRecord(t, repo, ctx)
	student := mustCreateStudentRecord(t, repo, ctx, user.ID)
	otherBonusType := mustCreateBonusTypeRecord(t, repo, ctx, otherUser.ID)

	_, err := repo.CreateBonus(ctx, repository.CreateBonusParams{
		UserID:      user.ID,
		StudentID:   student.ID,
		BonusTypeID: otherBonusType.ID,
		Points:      2,
	})
	if !isForeignKeyViolation(err) {
		t.Fatalf("expected foreign key violation, got %v", err)
	}
}

func TestRepository_CreateRuleRejectsCrossUserReferences(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	otherUser := mustCreateUserRecord(t, repo, ctx)
	penaltyType := mustCreatePenaltyTypeRecord(t, repo, ctx, user.ID)
	otherPunishmentType := mustCreatePunishmentTypeRecord(t, repo, ctx, otherUser.ID)

	_, err := repo.CreateRule(ctx, repository.CreateRuleParams{
		UserID:                    user.ID,
		Name:                      uniqueValue("rule-cross-user"),
		ResultingPunishmentTypeID: otherPunishmentType.ID,
		PenaltyTypeID:             penaltyType.ID,
		Threshold:                 1,
		Mode:                      "at",
		IsActive:                  true,
		DueAtAfterDays:            ptr(int32(1)),
		DueAtMode:                 "days",
	})
	if !isForeignKeyViolation(err) {
		t.Fatalf("expected foreign key violation, got %v", err)
	}
}

func TestRepository_AddStudentToClassroomScopesByUser(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	otherUser := mustCreateUserRecord(t, repo, ctx)
	student := mustCreateStudentRecord(t, repo, ctx, user.ID)
	otherClassroom := mustCreateClassroomRecord(t, repo, ctx, otherUser.ID)

	rowsAffected, err := repo.AddStudentToClassroom(ctx, repository.AddStudentToClassroomParams{
		UserID:      user.ID,
		StudentID:   student.ID,
		ClassroomID: otherClassroom.ID,
	})
	if err != nil {
		t.Fatalf("AddStudentToClassroom returned error: %v", err)
	}
	if rowsAffected != 0 {
		t.Fatalf("expected 0 rows affected, got %d", rowsAffected)
	}

	total, err := repo.CountClassroomsByStudent(ctx, repository.CountClassroomsByStudentParams{
		UserID:    user.ID,
		StudentID: student.ID,
	})
	if err != nil {
		t.Fatalf("CountClassroomsByStudent returned error: %v", err)
	}
	if total != 0 {
		t.Fatalf("expected no classroom relation for user-scoped query, got %d", total)
	}
}

func TestRepository_CreateScheduleSlotClassroomRelationScopesByUser(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	otherUser := mustCreateUserRecord(t, repo, ctx)
	slot, err := repo.CreateScheduleSlot(ctx, repository.CreateScheduleSlotParams{
		UserID:      user.ID,
		Weekday:     1,
		StartTime:   "08:00",
		EndTime:     "09:00",
		WeekPattern: "every_week",
	})
	if err != nil {
		t.Fatalf("CreateScheduleSlot returned error: %v", err)
	}
	otherClassroom := mustCreateClassroomRecord(t, repo, ctx, otherUser.ID)

	rowsAffected, err := repo.CreateScheduleSlotClassroomRelation(ctx, repository.CreateScheduleSlotClassroomRelationParams{
		UserID:         user.ID,
		ScheduleSlotID: slot.ID,
		ClassroomID:    otherClassroom.ID,
	})
	if err != nil {
		t.Fatalf("CreateScheduleSlotClassroomRelation returned error: %v", err)
	}
	if rowsAffected != 0 {
		t.Fatalf("expected 0 rows affected, got %d", rowsAffected)
	}

	refs, err := repo.ListScheduleSlotClassroomRefsBySlotIDs(ctx, repository.ListScheduleSlotClassroomRefsBySlotIDsParams{
		UserID:          user.ID,
		ScheduleSlotIds: []uuid.UUID{slot.ID},
	})
	if err != nil {
		t.Fatalf("ListScheduleSlotClassroomRefsBySlotIDs returned error: %v", err)
	}
	if len(refs) != 0 {
		t.Fatalf("expected no schedule slot classroom relation, got %d", len(refs))
	}
}

func TestRepository_DeleteRuleNullsTriggeringRuleOnPunishments(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	student := mustCreateStudentRecord(t, repo, ctx, user.ID)
	penaltyType := mustCreatePenaltyTypeRecord(t, repo, ctx, user.ID)
	punishmentType := mustCreatePunishmentTypeRecord(t, repo, ctx, user.ID)
	rule := mustCreateRuleRecord(t, repo, ctx, user.ID, penaltyType.ID, punishmentType.ID, "at", 1, 2, true)

	punishment, err := repo.CreatePunishmentFromRule(ctx, repository.CreatePunishmentFromRuleParams{
		UserID:           user.ID,
		StudentID:        student.ID,
		PunishmentTypeID: punishmentType.ID,
		TriggeringRuleID: &rule.ID,
		Automated:        true,
		DueAt:            time.Now().UTC().Add(2 * time.Hour),
	})
	if err != nil {
		t.Fatalf("CreatePunishmentFromRule returned error: %v", err)
	}
	if punishment.TriggeringRuleID == nil {
		t.Fatalf("expected punishment to reference a rule before deletion")
	}

	rowsAffected, err := repo.DeleteRuleByUser(ctx, repository.DeleteRuleByUserParams{
		ID:     rule.ID,
		UserID: user.ID,
	})
	if err != nil {
		t.Fatalf("DeleteRuleByUser returned error: %v", err)
	}
	if rowsAffected != 1 {
		t.Fatalf("expected 1 deleted rule, got %d", rowsAffected)
	}

	got, err := repo.GetPunishmentByUser(ctx, repository.GetPunishmentByUserParams{
		ID:     punishment.ID,
		UserID: user.ID,
	})
	if err != nil {
		t.Fatalf("GetPunishmentByUser returned error: %v", err)
	}
	if got.TriggeringRuleID != nil {
		t.Fatalf("expected triggering_rule_id to be NULL after rule deletion, got %s", *got.TriggeringRuleID)
	}
}
