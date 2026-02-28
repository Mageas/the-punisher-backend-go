//go:build integration

package integration

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
	. "github.com/mageas/the-punisher-backend/internal/service"
)

func TestBonusService_CRUDAndUse_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	student := mustCreateStudentRecord(t, repo, ctx, user.ID)
	bonusType := mustCreateBonusTypeRecord(t, repo, ctx, user.ID)
	svc := NewBonusService(repo)

	created, err := svc.CreateBonus(ctx, user.ID, student.ID, bonusType.ID, 4)
	if err != nil {
		t.Fatalf("CreateBonus returned error: %v", err)
	}
	if created.ID == uuid.Nil || created.UsedAt != nil {
		t.Fatalf("unexpected created bonus: %+v", created)
	}

	got, err := svc.GetBonus(ctx, user.ID, created.ID)
	if err != nil {
		t.Fatalf("GetBonus returned error: %v", err)
	}
	if got.ID != created.ID {
		t.Fatalf("expected same bonus id")
	}

	bonuses, total, err := svc.ListBonuses(ctx, user.ID, nil, nil, 20, 0)
	if err != nil {
		t.Fatalf("ListBonuses returned error: %v", err)
	}
	if total != 1 || len(bonuses) != 1 {
		t.Fatalf("expected one bonus, got total=%d len=%d", total, len(bonuses))
	}

	byStudent, totalByStudent, err := svc.ListBonusesByStudent(ctx, user.ID, student.ID, nil, 20, 0)
	if err != nil {
		t.Fatalf("ListBonusesByStudent returned error: %v", err)
	}
	if totalByStudent != 1 || len(byStudent) != 1 {
		t.Fatalf("expected one student bonus, got total=%d len=%d", totalByStudent, len(byStudent))
	}

	used, err := svc.UseBonus(ctx, user.ID, created.ID)
	if err != nil {
		t.Fatalf("UseBonus returned error: %v", err)
	}
	if used.UsedAt == nil {
		t.Fatalf("expected used_at to be set")
	}

	_, err = svc.UseBonus(ctx, user.ID, created.ID)
	if !errors.Is(err, api.ErrBonusAlreadyUsed) {
		t.Fatalf("expected ErrBonusAlreadyUsed, got %v", err)
	}

	if err := svc.DeleteBonus(ctx, user.ID, created.ID); err != nil {
		t.Fatalf("DeleteBonus returned error: %v", err)
	}
	if err := svc.DeleteBonus(ctx, user.ID, created.ID); !errors.Is(err, api.ErrBonusNotFound) {
		t.Fatalf("expected ErrBonusNotFound, got %v", err)
	}

	_, err = svc.UseBonus(ctx, user.ID, uuid.New())
	if !errors.Is(err, api.ErrBonusNotFound) {
		t.Fatalf("expected ErrBonusNotFound for missing bonus, got %v", err)
	}
}

func TestBonusService_NotFoundPrerequisites_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	svc := NewBonusService(repo)

	_, err := svc.CreateBonus(ctx, user.ID, uuid.New(), uuid.New(), 3)
	if !errors.Is(err, api.ErrStudentNotFound) {
		t.Fatalf("expected ErrStudentNotFound, got %v", err)
	}

	student := mustCreateStudentRecord(t, repo, ctx, user.ID)
	_, err = svc.CreateBonus(ctx, user.ID, student.ID, uuid.New(), 3)
	if !errors.Is(err, api.ErrBonusTypeNotFound) {
		t.Fatalf("expected ErrBonusTypeNotFound, got %v", err)
	}
}

func TestPunishmentService_CRUDAndResolve_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	student := mustCreateStudentRecord(t, repo, ctx, user.ID)
	punishmentType := mustCreatePunishmentTypeRecord(t, repo, ctx, user.ID)
	svc := NewPunishmentService(repo)

	dueAt := time.Now().UTC().Add(48 * time.Hour)
	created, err := svc.CreatePunishment(ctx, user.ID, student.ID, punishmentType.ID, dueAt)
	if err != nil {
		t.Fatalf("CreatePunishment returned error: %v", err)
	}

	got, err := svc.GetPunishment(ctx, user.ID, created.ID)
	if err != nil {
		t.Fatalf("GetPunishment returned error: %v", err)
	}
	if got.ID != created.ID || got.ResolvedAt != nil {
		t.Fatalf("unexpected punishment: %+v", got)
	}

	all, total, err := svc.ListPunishments(ctx, user.ID, nil, nil, 20, 0)
	if err != nil {
		t.Fatalf("ListPunishments returned error: %v", err)
	}
	if total != 1 || len(all) != 1 {
		t.Fatalf("expected one punishment, got total=%d len=%d", total, len(all))
	}

	byStudent, totalByStudent, err := svc.ListPunishmentsByStudent(ctx, user.ID, student.ID, nil, 20, 0)
	if err != nil {
		t.Fatalf("ListPunishmentsByStudent returned error: %v", err)
	}
	if totalByStudent != 1 || len(byStudent) != 1 {
		t.Fatalf("expected one student punishment, got total=%d len=%d", totalByStudent, len(byStudent))
	}

	resolved, err := svc.ResolvePunishment(ctx, user.ID, created.ID)
	if err != nil {
		t.Fatalf("ResolvePunishment returned error: %v", err)
	}
	if resolved.ResolvedAt == nil {
		t.Fatalf("expected resolved_at to be set")
	}

	_, err = svc.ResolvePunishment(ctx, user.ID, created.ID)
	if !errors.Is(err, api.ErrPunishmentAlreadyResolved) {
		t.Fatalf("expected ErrPunishmentAlreadyResolved, got %v", err)
	}

	if err := svc.DeletePunishment(ctx, user.ID, created.ID); err != nil {
		t.Fatalf("DeletePunishment returned error: %v", err)
	}
	if err := svc.DeletePunishment(ctx, user.ID, created.ID); !errors.Is(err, api.ErrPunishmentNotFound) {
		t.Fatalf("expected ErrPunishmentNotFound, got %v", err)
	}
}

func TestPunishmentService_NotFoundPrerequisites_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	svc := NewPunishmentService(repo)

	_, err := svc.CreatePunishment(ctx, user.ID, uuid.New(), uuid.New(), time.Now().UTC())
	if !errors.Is(err, api.ErrStudentNotFound) {
		t.Fatalf("expected ErrStudentNotFound, got %v", err)
	}

	student := mustCreateStudentRecord(t, repo, ctx, user.ID)
	_, err = svc.CreatePunishment(ctx, user.ID, student.ID, uuid.New(), time.Now().UTC())
	if !errors.Is(err, api.ErrPunishmentTypeNotFound) {
		t.Fatalf("expected ErrPunishmentTypeNotFound, got %v", err)
	}

	_, err = svc.ResolvePunishment(ctx, user.ID, uuid.New())
	if !errors.Is(err, api.ErrPunishmentNotFound) {
		t.Fatalf("expected ErrPunishmentNotFound for missing resolve, got %v", err)
	}
}

func TestRuleService_CRUDAndValidation_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	penaltyType := mustCreatePenaltyTypeRecord(t, repo, ctx, user.ID)
	punishmentType := mustCreatePunishmentTypeRecord(t, repo, ctx, user.ID)
	otherPenaltyType := mustCreatePenaltyTypeRecord(t, repo, ctx, user.ID)
	otherPunishmentType := mustCreatePunishmentTypeRecord(t, repo, ctx, user.ID)

	svc := NewRuleService(repo)

	created, err := svc.CreateRule(ctx, user.ID, dto.RequestRuleDto{
		Name:                      "Rule 1",
		ResultingPunishmentTypeID: punishmentType.ID.String(),
		PenaltyTypeID:             penaltyType.ID.String(),
		Threshold:                 2,
		DueAtAfterDays:            1,
		Mode:                      "every",
		IsActive:                  nil,
	})
	if err != nil {
		t.Fatalf("CreateRule returned error: %v", err)
	}
	if !created.IsActive {
		t.Fatalf("expected default is_active=true")
	}

	got, err := svc.GetRule(ctx, user.ID, created.ID)
	if err != nil {
		t.Fatalf("GetRule returned error: %v", err)
	}
	if got.ID != created.ID {
		t.Fatalf("expected same rule id")
	}

	all, total, err := svc.ListRules(ctx, user.ID, 20, 0)
	if err != nil {
		t.Fatalf("ListRules returned error: %v", err)
	}
	if total != 1 || len(all) != 1 {
		t.Fatalf("expected one rule, got total=%d len=%d", total, len(all))
	}

	newName := "Rule updated"
	newMode := "after"
	newThreshold := int32(3)
	newDueDays := int32(4)
	newActive := false
	newPenaltyID := otherPenaltyType.ID.String()
	newPunishmentID := otherPunishmentType.ID.String()

	updated, err := svc.UpdateRule(ctx, user.ID, created.ID, dto.UpdateRuleDto{
		Name:                      &newName,
		Mode:                      &newMode,
		Threshold:                 &newThreshold,
		DueAtAfterDays:            &newDueDays,
		IsActive:                  &newActive,
		PenaltyTypeID:             &newPenaltyID,
		ResultingPunishmentTypeID: &newPunishmentID,
	})
	if err != nil {
		t.Fatalf("UpdateRule returned error: %v", err)
	}
	if updated.Name != newName || updated.Mode != newMode || updated.IsActive != newActive {
		t.Fatalf("unexpected updated rule: %+v", updated)
	}

	if err := svc.DeleteRule(ctx, user.ID, created.ID); err != nil {
		t.Fatalf("DeleteRule returned error: %v", err)
	}
	if err := svc.DeleteRule(ctx, user.ID, created.ID); !errors.Is(err, api.ErrRuleNotFound) {
		t.Fatalf("expected ErrRuleNotFound, got %v", err)
	}

	_, err = svc.CreateRule(ctx, user.ID, dto.RequestRuleDto{
		Name:                      "invalid",
		ResultingPunishmentTypeID: "not-a-uuid",
		PenaltyTypeID:             penaltyType.ID.String(),
		Threshold:                 1,
		DueAtAfterDays:            0,
		Mode:                      "at",
	})
	if !errors.Is(err, api.ErrInvalidRequestBody) {
		t.Fatalf("expected ErrInvalidRequestBody on invalid uuid, got %v", err)
	}

	_, err = svc.UpdateRule(ctx, user.ID, uuid.New(), dto.UpdateRuleDto{PenaltyTypeID: ptr("bad-uuid")})
	if !errors.Is(err, api.ErrInvalidRequestBody) {
		t.Fatalf("expected ErrInvalidRequestBody on update invalid uuid, got %v", err)
	}
}

func TestPenaltyService_CRUDAndRuleTrigger_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	student := mustCreateStudentRecord(t, repo, ctx, user.ID)
	penaltyType := mustCreatePenaltyTypeRecord(t, repo, ctx, user.ID)
	punishmentType := mustCreatePunishmentTypeRecord(t, repo, ctx, user.ID)
	rule := mustCreateRuleRecord(t, repo, ctx, user.ID, penaltyType.ID, punishmentType.ID, "at", 1, 2, true)

	penaltySvc := NewPenaltyService(repo)
	punishmentSvc := NewPunishmentService(repo)

	created, err := penaltySvc.CreatePenalty(ctx, user.ID, student.ID, penaltyType.ID)
	if err != nil {
		t.Fatalf("CreatePenalty returned error: %v", err)
	}

	got, err := penaltySvc.GetPenalty(ctx, user.ID, created.ID)
	if err != nil {
		t.Fatalf("GetPenalty returned error: %v", err)
	}
	if got.ID != created.ID {
		t.Fatalf("unexpected penalty id")
	}

	all, total, err := penaltySvc.ListPenalties(ctx, user.ID, 20, 0)
	if err != nil {
		t.Fatalf("ListPenalties returned error: %v", err)
	}
	if total != 1 || len(all) != 1 {
		t.Fatalf("expected one penalty, got total=%d len=%d", total, len(all))
	}

	byStudent, totalByStudent, err := penaltySvc.ListPenaltiesByStudent(ctx, user.ID, student.ID, 20, 0)
	if err != nil {
		t.Fatalf("ListPenaltiesByStudent returned error: %v", err)
	}
	if totalByStudent != 1 || len(byStudent) != 1 {
		t.Fatalf("expected one student penalty, got total=%d len=%d", totalByStudent, len(byStudent))
	}

	punishments, punishmentTotal, err := punishmentSvc.ListPunishmentsByStudent(ctx, user.ID, student.ID, nil, 20, 0)
	if err != nil {
		t.Fatalf("ListPunishmentsByStudent returned error: %v", err)
	}
	if punishmentTotal != 1 || len(punishments) != 1 {
		t.Fatalf("expected one rule-triggered punishment, got total=%d len=%d", punishmentTotal, len(punishments))
	}
	if !punishments[0].Automated {
		t.Fatalf("expected rule-triggered punishment to be automated")
	}
	if punishments[0].TriggeringRuleID == nil || *punishments[0].TriggeringRuleID != rule.ID {
		t.Fatalf("expected triggering rule id %s, got %+v", rule.ID, punishments[0].TriggeringRuleID)
	}

	if err := penaltySvc.DeletePenalty(ctx, user.ID, created.ID); err != nil {
		t.Fatalf("DeletePenalty returned error: %v", err)
	}
	if err := penaltySvc.DeletePenalty(ctx, user.ID, created.ID); !errors.Is(err, api.ErrPenaltyNotFound) {
		t.Fatalf("expected ErrPenaltyNotFound, got %v", err)
	}
}

func TestPenaltyService_NotFoundPrerequisites_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	penaltySvc := NewPenaltyService(repo)

	_, err := penaltySvc.CreatePenalty(ctx, user.ID, uuid.New(), uuid.New())
	if !errors.Is(err, api.ErrStudentNotFound) {
		t.Fatalf("expected ErrStudentNotFound, got %v", err)
	}

	student := mustCreateStudentRecord(t, repo, ctx, user.ID)
	_, err = penaltySvc.CreatePenalty(ctx, user.ID, student.ID, uuid.New())
	if !errors.Is(err, api.ErrPenaltyTypeNotFound) {
		t.Fatalf("expected ErrPenaltyTypeNotFound, got %v", err)
	}

	_, _, err = penaltySvc.ListPenaltiesByStudent(ctx, user.ID, uuid.New(), 20, 0)
	if !errors.Is(err, api.ErrStudentNotFound) {
		t.Fatalf("expected ErrStudentNotFound on ListPenaltiesByStudent, got %v", err)
	}
}

func TestDashboardService_GetDashboard_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	studentInClass := mustCreateStudentRecord(t, repo, ctx, user.ID)
	studentOutClass := mustCreateStudentRecord(t, repo, ctx, user.ID)
	classroom := mustCreateClassroomRecord(t, repo, ctx, user.ID)

	if _, err := repo.AddStudentToClassroom(ctx, repository.AddStudentToClassroomParams{StudentID: studentInClass.ID, ClassroomID: classroom.ID, UserID: user.ID}); err != nil {
		t.Fatalf("failed to link student to classroom: %v", err)
	}

	bonusType := mustCreateBonusTypeRecord(t, repo, ctx, user.ID)
	penaltyType := mustCreatePenaltyTypeRecord(t, repo, ctx, user.ID)
	punishmentType := mustCreatePunishmentTypeRecord(t, repo, ctx, user.ID)

	_ = mustCreateBonusRecord(t, repo, ctx, user.ID, studentInClass.ID, bonusType.ID, 2)
	_ = mustCreateBonusRecord(t, repo, ctx, user.ID, studentOutClass.ID, bonusType.ID, 1)
	_ = mustCreatePenaltyRecord(t, repo, ctx, user.ID, studentInClass.ID, penaltyType.ID)
	_ = mustCreatePunishmentRecord(t, repo, ctx, user.ID, studentInClass.ID, punishmentType.ID, time.Now().UTC().Add(-1*time.Hour))

	svc := NewDashboardService(repo)

	dashboardAll, err := svc.GetDashboard(ctx, user.ID, nil)
	if err != nil {
		t.Fatalf("GetDashboard(all) returned error: %v", err)
	}
	if dashboardAll.Kpis.StudentCount != 2 {
		t.Fatalf("expected student_count=2, got %d", dashboardAll.Kpis.StudentCount)
	}
	if len(dashboardAll.RecentBonuses) != 2 {
		t.Fatalf("expected two recent bonuses, got %d", len(dashboardAll.RecentBonuses))
	}

	dashboardClassroom, err := svc.GetDashboard(ctx, user.ID, &classroom.ID)
	if err != nil {
		t.Fatalf("GetDashboard(classroom) returned error: %v", err)
	}
	if dashboardClassroom.Kpis.StudentCount != 1 {
		t.Fatalf("expected classroom-filtered student_count=1, got %d", dashboardClassroom.Kpis.StudentCount)
	}
	if len(dashboardClassroom.RecentBonuses) != 1 {
		t.Fatalf("expected one classroom-filtered bonus, got %d", len(dashboardClassroom.RecentBonuses))
	}

	missingID := uuid.New()
	_, err = svc.GetDashboard(ctx, user.ID, &missingID)
	if !errors.Is(err, api.ErrClassroomNotFound) {
		t.Fatalf("expected ErrClassroomNotFound, got %v", err)
	}
}
