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

	created, err := svc.CreateBonus(ctx, user.ID, student.ID, bonusType.ID, 4, nil, nil)
	if err != nil {
		t.Fatalf("CreateBonus returned error: %v", err)
	}
	if created.ID == uuid.Nil || created.UsedAt != nil {
		t.Fatalf("unexpected created bonus: %+v", created)
	}
	if created.EvaluationLabel != "" {
		t.Fatalf("expected empty evaluation_label by default, got %q", created.EvaluationLabel)
	}
	if created.OccurredAt.IsZero() {
		t.Fatalf("expected occurred_at to be set")
	}
	if created.OccurredAt.Sub(created.CreatedAt) > 2*time.Second || created.CreatedAt.Sub(created.OccurredAt) > 2*time.Second {
		t.Fatalf("expected occurred_at to fallback close to created_at")
	}

	occurredAt := time.Now().UTC().Add(-48 * time.Hour)
	label := "Participation oral"
	backdated, err := svc.CreateBonus(ctx, user.ID, student.ID, bonusType.ID, 1, &occurredAt, &label)
	if err != nil {
		t.Fatalf("CreateBonus(backdated) returned error: %v", err)
	}
	assertTimeEqualToPostgresPrecision(t, "backdated occurred_at", backdated.OccurredAt, occurredAt)
	if backdated.EvaluationLabel != label {
		t.Fatalf("unexpected evaluation_label on backdated bonus: %+v", backdated.EvaluationLabel)
	}
	if !backdated.OccurredAt.Before(backdated.CreatedAt) {
		t.Fatalf("expected backdated occurred_at before created_at")
	}

	got, err := svc.GetBonus(ctx, user.ID, created.ID)
	if err != nil {
		t.Fatalf("GetBonus returned error: %v", err)
	}
	if got.ID != created.ID {
		t.Fatalf("expected same bonus id")
	}

	bonuses, total, err := svc.ListBonuses(ctx, user.ID, ListBonusesFilters{Limit: 20, Offset: 0})
	if err != nil {
		t.Fatalf("ListBonuses returned error: %v", err)
	}
	if total != 2 || len(bonuses) != 2 {
		t.Fatalf("expected two bonuses, got total=%d len=%d", total, len(bonuses))
	}

	updatedOccurredAt := time.Now().UTC().Add(-24 * time.Hour)
	updatedLabel := "Nouveau libelle"
	updated, err := svc.UpdateBonus(ctx, user.ID, created.ID, &updatedOccurredAt, &updatedLabel)
	if err != nil {
		t.Fatalf("UpdateBonus returned error: %v", err)
	}
	assertTimeEqualToPostgresPrecision(t, "updated occurred_at", updated.OccurredAt, updatedOccurredAt)
	if updated.EvaluationLabel != updatedLabel {
		t.Fatalf("unexpected updated evaluation_label: %+v", updated.EvaluationLabel)
	}

	emptyLabel := ""
	cleared, err := svc.UpdateBonus(ctx, user.ID, created.ID, nil, &emptyLabel)
	if err != nil {
		t.Fatalf("UpdateBonus(clear label) returned error: %v", err)
	}
	if cleared.EvaluationLabel != "" {
		t.Fatalf("expected evaluation_label to be cleared to empty string, got %+v", cleared.EvaluationLabel)
	}

	byStudent, totalByStudent, err := svc.ListBonusesByStudent(ctx, user.ID, student.ID, nil, 20, 0)
	if err != nil {
		t.Fatalf("ListBonusesByStudent returned error: %v", err)
	}
	if totalByStudent != 2 || len(byStudent) != 2 {
		t.Fatalf("expected two student bonuses, got total=%d len=%d", totalByStudent, len(byStudent))
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

	_, err := svc.CreateBonus(ctx, user.ID, uuid.New(), uuid.New(), 3, nil, nil)
	if !errors.Is(err, api.ErrStudentNotFound) {
		t.Fatalf("expected ErrStudentNotFound, got %v", err)
	}

	student := mustCreateStudentRecord(t, repo, ctx, user.ID)
	_, err = svc.CreateBonus(ctx, user.ID, student.ID, uuid.New(), 3, nil, nil)
	if !errors.Is(err, api.ErrBonusTypeNotFound) {
		t.Fatalf("expected ErrBonusTypeNotFound, got %v", err)
	}

	_, err = svc.UpdateBonus(ctx, user.ID, uuid.New(), nil, nil)
	if !errors.Is(err, api.ErrBonusNotFound) {
		t.Fatalf("expected ErrBonusNotFound on update, got %v", err)
	}
}

func TestBonusService_ListBonusesFilters_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	studentInClass := mustCreateStudentRecord(t, repo, ctx, user.ID)
	studentOutClass := mustCreateStudentRecord(t, repo, ctx, user.ID)
	classroom := mustCreateClassroomRecord(t, repo, ctx, user.ID)
	if _, err := repo.AddStudentToClassroom(ctx, repository.AddStudentToClassroomParams{
		StudentID:   studentInClass.ID,
		ClassroomID: classroom.ID,
		UserID:      user.ID,
	}); err != nil {
		t.Fatalf("failed to add student to classroom: %v", err)
	}

	bonusTypeA := mustCreateBonusTypeRecord(t, repo, ctx, user.ID)
	bonusTypeB := mustCreateBonusTypeRecord(t, repo, ctx, user.ID)
	svc := NewBonusService(repo)

	bonusInClass, err := svc.CreateBonus(ctx, user.ID, studentInClass.ID, bonusTypeA.ID, 2, nil, nil)
	if err != nil {
		t.Fatalf("CreateBonus(in class) returned error: %v", err)
	}
	bonusOutClass, err := svc.CreateBonus(ctx, user.ID, studentOutClass.ID, bonusTypeB.ID, 3, nil, nil)
	if err != nil {
		t.Fatalf("CreateBonus(out class) returned error: %v", err)
	}
	if _, err := svc.UseBonus(ctx, user.ID, bonusOutClass.ID); err != nil {
		t.Fatalf("UseBonus(out class) returned error: %v", err)
	}

	unused := BonusStateUnused
	filtered, total, err := svc.ListBonuses(ctx, user.ID, ListBonusesFilters{
		State:  &unused,
		Limit:  20,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListBonuses(state=unused) returned error: %v", err)
	}
	if total != 1 || len(filtered) != 1 || filtered[0].ID != bonusInClass.ID {
		t.Fatalf("unexpected state=unused result: total=%d len=%d data=%+v", total, len(filtered), filtered)
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	studentInClassID := studentInClass.ID
	classroomID := classroom.ID
	bonusTypeAID := bonusTypeA.ID
	filtered, total, err = svc.ListBonuses(ctx, user.ID, ListBonusesFilters{
		StudentID:   &studentInClassID,
		ClassroomID: &classroomID,
		BonusTypeID: &bonusTypeAID,
		CreatedFrom: &today,
		CreatedTo:   &today,
		Limit:       20,
		Offset:      0,
	})
	if err != nil {
		t.Fatalf("ListBonuses(combined filters) returned error: %v", err)
	}
	if total != 1 || len(filtered) != 1 || filtered[0].ID != bonusInClass.ID {
		t.Fatalf("unexpected combined filters result: total=%d len=%d data=%+v", total, len(filtered), filtered)
	}

	used := BonusStateUsed
	filtered, total, err = svc.ListBonuses(ctx, user.ID, ListBonusesFilters{
		State:       &used,
		ClassroomID: &classroomID,
		Limit:       20,
		Offset:      0,
	})
	if err != nil {
		t.Fatalf("ListBonuses(state=used,classroom) returned error: %v", err)
	}
	if total != 0 || len(filtered) != 0 {
		t.Fatalf("expected no used bonus in classroom, got total=%d len=%d", total, len(filtered))
	}
}

func TestBonusService_ListBonuses_UsesOccurredAtForFilterAndSort_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	student := mustCreateStudentRecord(t, repo, ctx, user.ID)
	bonusType := mustCreateBonusTypeRecord(t, repo, ctx, user.ID)
	svc := NewBonusService(repo)

	recentOccurred := time.Now().UTC()
	backdatedOccurred := recentOccurred.AddDate(0, 0, -4)

	recentBonus, err := svc.CreateBonus(ctx, user.ID, student.ID, bonusType.ID, 1, &recentOccurred, nil)
	if err != nil {
		t.Fatalf("CreateBonus(recent) returned error: %v", err)
	}
	backdatedBonus, err := svc.CreateBonus(ctx, user.ID, student.ID, bonusType.ID, 1, &backdatedOccurred, nil)
	if err != nil {
		t.Fatalf("CreateBonus(backdated) returned error: %v", err)
	}

	studentID := student.ID
	all, total, err := svc.ListBonuses(ctx, user.ID, ListBonusesFilters{
		StudentID: &studentID,
		Limit:     20,
		Offset:    0,
	})
	if err != nil {
		t.Fatalf("ListBonuses returned error: %v", err)
	}
	if total != 2 || len(all) != 2 {
		t.Fatalf("expected two bonuses, got total=%d len=%d", total, len(all))
	}
	if all[0].ID != recentBonus.ID {
		t.Fatalf("expected recent occurred_at bonus first, got %s (backdated=%s)", all[0].ID, backdatedBonus.ID)
	}

	today := recentOccurred.Truncate(24 * time.Hour)
	filtered, total, err := svc.ListBonuses(ctx, user.ID, ListBonusesFilters{
		StudentID:   &studentID,
		CreatedFrom: &today,
		CreatedTo:   &today,
		Limit:       20,
		Offset:      0,
	})
	if err != nil {
		t.Fatalf("ListBonuses(created range) returned error: %v", err)
	}
	if total != 1 || len(filtered) != 1 || filtered[0].ID != recentBonus.ID {
		t.Fatalf("expected only recent bonus in created range filter, got total=%d len=%d data=%+v", total, len(filtered), filtered)
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
	occurredAt := time.Now().UTC().Add(-24 * time.Hour)
	label := "Heure de retenue"
	created, err := svc.CreatePunishment(ctx, user.ID, student.ID, punishmentType.ID, dueAt, &occurredAt, &label)
	if err != nil {
		t.Fatalf("CreatePunishment returned error: %v", err)
	}
	assertTimeEqualToPostgresPrecision(t, "occurred_at", created.OccurredAt, occurredAt)
	if created.EvaluationLabel != label {
		t.Fatalf("unexpected evaluation_label on create: %+v", created.EvaluationLabel)
	}

	got, err := svc.GetPunishment(ctx, user.ID, created.ID)
	if err != nil {
		t.Fatalf("GetPunishment returned error: %v", err)
	}
	if got.ID != created.ID || got.ResolvedAt != nil {
		t.Fatalf("unexpected punishment: %+v", got)
	}

	all, total, err := svc.ListPunishments(ctx, user.ID, ListPunishmentsFilters{Limit: 20, Offset: 0})
	if err != nil {
		t.Fatalf("ListPunishments returned error: %v", err)
	}
	if total != 1 || len(all) != 1 {
		t.Fatalf("expected one punishment, got total=%d len=%d", total, len(all))
	}

	updatedOccurredAt := time.Now().UTC().Add(-12 * time.Hour)
	updatedLabel := "Label mis a jour"
	updated, err := svc.UpdatePunishment(ctx, user.ID, created.ID, &updatedOccurredAt, &updatedLabel)
	if err != nil {
		t.Fatalf("UpdatePunishment returned error: %v", err)
	}
	assertTimeEqualToPostgresPrecision(t, "updated occurred_at", updated.OccurredAt, updatedOccurredAt)
	if updated.EvaluationLabel != updatedLabel {
		t.Fatalf("unexpected updated label: %+v", updated.EvaluationLabel)
	}

	emptyLabel := ""
	cleared, err := svc.UpdatePunishment(ctx, user.ID, created.ID, nil, &emptyLabel)
	if err != nil {
		t.Fatalf("UpdatePunishment(clear label) returned error: %v", err)
	}
	if cleared.EvaluationLabel != "" {
		t.Fatalf("expected label to be cleared to empty string")
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

	_, err := svc.CreatePunishment(ctx, user.ID, uuid.New(), uuid.New(), time.Now().UTC(), nil, nil)
	if !errors.Is(err, api.ErrStudentNotFound) {
		t.Fatalf("expected ErrStudentNotFound, got %v", err)
	}

	student := mustCreateStudentRecord(t, repo, ctx, user.ID)
	_, err = svc.CreatePunishment(ctx, user.ID, student.ID, uuid.New(), time.Now().UTC(), nil, nil)
	if !errors.Is(err, api.ErrPunishmentTypeNotFound) {
		t.Fatalf("expected ErrPunishmentTypeNotFound, got %v", err)
	}

	_, err = svc.ResolvePunishment(ctx, user.ID, uuid.New())
	if !errors.Is(err, api.ErrPunishmentNotFound) {
		t.Fatalf("expected ErrPunishmentNotFound for missing resolve, got %v", err)
	}

	_, err = svc.UpdatePunishment(ctx, user.ID, uuid.New(), nil, nil)
	if !errors.Is(err, api.ErrPunishmentNotFound) {
		t.Fatalf("expected ErrPunishmentNotFound for missing update, got %v", err)
	}
}

func TestPunishmentService_ListPunishmentsFilters_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	studentInClass := mustCreateStudentRecord(t, repo, ctx, user.ID)
	studentOutClass := mustCreateStudentRecord(t, repo, ctx, user.ID)
	classroom := mustCreateClassroomRecord(t, repo, ctx, user.ID)
	if _, err := repo.AddStudentToClassroom(ctx, repository.AddStudentToClassroomParams{
		StudentID:   studentInClass.ID,
		ClassroomID: classroom.ID,
		UserID:      user.ID,
	}); err != nil {
		t.Fatalf("failed to add student to classroom: %v", err)
	}

	punishmentTypeA := mustCreatePunishmentTypeRecord(t, repo, ctx, user.ID)
	punishmentTypeB := mustCreatePunishmentTypeRecord(t, repo, ctx, user.ID)
	svc := NewPunishmentService(repo)

	duePast := time.Now().UTC().Add(-24 * time.Hour)
	dueFuture := time.Now().UTC().Add(24 * time.Hour)

	manualOverdue, err := svc.CreatePunishment(ctx, user.ID, studentInClass.ID, punishmentTypeA.ID, duePast, nil, nil)
	if err != nil {
		t.Fatalf("CreatePunishment(manual overdue) returned error: %v", err)
	}

	resolvedPunishment, err := svc.CreatePunishment(ctx, user.ID, studentOutClass.ID, punishmentTypeB.ID, dueFuture, nil, nil)
	if err != nil {
		t.Fatalf("CreatePunishment(resolved candidate) returned error: %v", err)
	}
	if _, err := svc.ResolvePunishment(ctx, user.ID, resolvedPunishment.ID); err != nil {
		t.Fatalf("ResolvePunishment returned error: %v", err)
	}

	automatedOverdue, err := repo.CreatePunishmentFromRule(ctx, repository.CreatePunishmentFromRuleParams{
		UserID:           user.ID,
		StudentID:        studentInClass.ID,
		PunishmentTypeID: punishmentTypeA.ID,
		TriggeringRuleID: nil,
		Automated:        true,
		DueAt:            duePast,
	})
	if err != nil {
		t.Fatalf("CreatePunishmentFromRule returned error: %v", err)
	}

	overdue := true
	filtered, total, err := svc.ListPunishments(ctx, user.ID, ListPunishmentsFilters{
		Overdue: &overdue,
		Limit:   20,
		Offset:  0,
	})
	if err != nil {
		t.Fatalf("ListPunishments(overdue=true) returned error: %v", err)
	}
	if total != 2 || len(filtered) != 2 {
		t.Fatalf("expected two overdue punishments, got total=%d len=%d", total, len(filtered))
	}

	automated := true
	filtered, total, err = svc.ListPunishments(ctx, user.ID, ListPunishmentsFilters{
		Overdue:   &overdue,
		Automated: &automated,
		Limit:     20,
		Offset:    0,
	})
	if err != nil {
		t.Fatalf("ListPunishments(overdue=true,automated=true) returned error: %v", err)
	}
	if total != 1 || len(filtered) != 1 || filtered[0].ID != automatedOverdue.ID {
		t.Fatalf("unexpected overdue+automated result: total=%d len=%d data=%+v", total, len(filtered), filtered)
	}

	resolved := PunishmentStateResolved
	filtered, total, err = svc.ListPunishments(ctx, user.ID, ListPunishmentsFilters{
		State:  &resolved,
		Limit:  20,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListPunishments(state=resolved) returned error: %v", err)
	}
	if total != 1 || len(filtered) != 1 || filtered[0].ID != resolvedPunishment.ID {
		t.Fatalf("unexpected state=resolved result: total=%d len=%d data=%+v", total, len(filtered), filtered)
	}

	pending := PunishmentStatePending
	studentInClassID := studentInClass.ID
	classroomID := classroom.ID
	punishmentTypeAID := punishmentTypeA.ID
	manual := false
	dueTo := time.Now().UTC().Truncate(24 * time.Hour)
	filtered, total, err = svc.ListPunishments(ctx, user.ID, ListPunishmentsFilters{
		State:            &pending,
		StudentID:        &studentInClassID,
		ClassroomID:      &classroomID,
		PunishmentTypeID: &punishmentTypeAID,
		Automated:        &manual,
		DueTo:            &dueTo,
		Limit:            20,
		Offset:           0,
	})
	if err != nil {
		t.Fatalf("ListPunishments(combined filters) returned error: %v", err)
	}
	if total != 1 || len(filtered) != 1 || filtered[0].ID != manualOverdue.ID {
		t.Fatalf("unexpected combined filters result: total=%d len=%d data=%+v", total, len(filtered), filtered)
	}
}

func TestPunishmentService_ListPunishments_UsesOccurredAtForFilterAndSort_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	student := mustCreateStudentRecord(t, repo, ctx, user.ID)
	punishmentType := mustCreatePunishmentTypeRecord(t, repo, ctx, user.ID)
	svc := NewPunishmentService(repo)

	recentOccurred := time.Now().UTC()
	backdatedOccurred := recentOccurred.AddDate(0, 0, -4)
	dueAt := time.Now().UTC().Add(24 * time.Hour)

	recentPunishment, err := svc.CreatePunishment(ctx, user.ID, student.ID, punishmentType.ID, dueAt, &recentOccurred, nil)
	if err != nil {
		t.Fatalf("CreatePunishment(recent) returned error: %v", err)
	}
	backdatedPunishment, err := svc.CreatePunishment(ctx, user.ID, student.ID, punishmentType.ID, dueAt, &backdatedOccurred, nil)
	if err != nil {
		t.Fatalf("CreatePunishment(backdated) returned error: %v", err)
	}

	studentID := student.ID
	all, total, err := svc.ListPunishments(ctx, user.ID, ListPunishmentsFilters{
		StudentID: &studentID,
		Limit:     20,
		Offset:    0,
	})
	if err != nil {
		t.Fatalf("ListPunishments returned error: %v", err)
	}
	if total != 2 || len(all) != 2 {
		t.Fatalf("expected two punishments, got total=%d len=%d", total, len(all))
	}
	if all[0].ID != recentPunishment.ID {
		t.Fatalf("expected recent occurred_at punishment first, got %s (backdated=%s)", all[0].ID, backdatedPunishment.ID)
	}

	today := recentOccurred.Truncate(24 * time.Hour)
	filtered, total, err := svc.ListPunishments(ctx, user.ID, ListPunishmentsFilters{
		StudentID:   &studentID,
		CreatedFrom: &today,
		CreatedTo:   &today,
		Limit:       20,
		Offset:      0,
	})
	if err != nil {
		t.Fatalf("ListPunishments(created range) returned error: %v", err)
	}
	if total != 1 || len(filtered) != 1 || filtered[0].ID != recentPunishment.ID {
		t.Fatalf("expected only recent punishment in created range filter, got total=%d len=%d data=%+v", total, len(filtered), filtered)
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

	occurredAt := time.Now().UTC().Add(-24 * time.Hour)
	label := "Retard constate"
	created, err := penaltySvc.CreatePenalty(ctx, user.ID, student.ID, penaltyType.ID, &occurredAt, &label)
	if err != nil {
		t.Fatalf("CreatePenalty returned error: %v", err)
	}
	assertTimeEqualToPostgresPrecision(t, "occurred_at", created.OccurredAt, occurredAt)
	if created.EvaluationLabel != label {
		t.Fatalf("unexpected evaluation_label on create: %+v", created.EvaluationLabel)
	}

	got, err := penaltySvc.GetPenalty(ctx, user.ID, created.ID)
	if err != nil {
		t.Fatalf("GetPenalty returned error: %v", err)
	}
	if got.ID != created.ID {
		t.Fatalf("unexpected penalty id")
	}

	all, total, err := penaltySvc.ListPenalties(ctx, user.ID, ListPenaltiesFilters{Limit: 20, Offset: 0})
	if err != nil {
		t.Fatalf("ListPenalties returned error: %v", err)
	}
	if total != 1 || len(all) != 1 {
		t.Fatalf("expected one penalty, got total=%d len=%d", total, len(all))
	}

	updatedOccurredAt := time.Now().UTC().Add(-12 * time.Hour)
	updatedLabel := "Retard corrige"
	updated, err := penaltySvc.UpdatePenalty(ctx, user.ID, created.ID, &updatedOccurredAt, &updatedLabel)
	if err != nil {
		t.Fatalf("UpdatePenalty returned error: %v", err)
	}
	assertTimeEqualToPostgresPrecision(t, "updated occurred_at", updated.OccurredAt, updatedOccurredAt)
	if updated.EvaluationLabel != updatedLabel {
		t.Fatalf("unexpected label on update: %+v", updated.EvaluationLabel)
	}

	emptyLabel := ""
	cleared, err := penaltySvc.UpdatePenalty(ctx, user.ID, created.ID, nil, &emptyLabel)
	if err != nil {
		t.Fatalf("UpdatePenalty(clear label) returned error: %v", err)
	}
	if cleared.EvaluationLabel != "" {
		t.Fatalf("expected label to be cleared to empty string")
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

	_, err := penaltySvc.CreatePenalty(ctx, user.ID, uuid.New(), uuid.New(), nil, nil)
	if !errors.Is(err, api.ErrStudentNotFound) {
		t.Fatalf("expected ErrStudentNotFound, got %v", err)
	}

	student := mustCreateStudentRecord(t, repo, ctx, user.ID)
	_, err = penaltySvc.CreatePenalty(ctx, user.ID, student.ID, uuid.New(), nil, nil)
	if !errors.Is(err, api.ErrPenaltyTypeNotFound) {
		t.Fatalf("expected ErrPenaltyTypeNotFound, got %v", err)
	}

	_, _, err = penaltySvc.ListPenaltiesByStudent(ctx, user.ID, uuid.New(), 20, 0)
	if !errors.Is(err, api.ErrStudentNotFound) {
		t.Fatalf("expected ErrStudentNotFound on ListPenaltiesByStudent, got %v", err)
	}

	_, err = penaltySvc.UpdatePenalty(ctx, user.ID, uuid.New(), nil, nil)
	if !errors.Is(err, api.ErrPenaltyNotFound) {
		t.Fatalf("expected ErrPenaltyNotFound on update, got %v", err)
	}
}

func TestPenaltyService_ListPenaltiesFilters_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	studentInClass := mustCreateStudentRecord(t, repo, ctx, user.ID)
	studentOutClass := mustCreateStudentRecord(t, repo, ctx, user.ID)
	classroom := mustCreateClassroomRecord(t, repo, ctx, user.ID)
	if _, err := repo.AddStudentToClassroom(ctx, repository.AddStudentToClassroomParams{
		StudentID:   studentInClass.ID,
		ClassroomID: classroom.ID,
		UserID:      user.ID,
	}); err != nil {
		t.Fatalf("failed to add student to classroom: %v", err)
	}

	penaltyTypeA := mustCreatePenaltyTypeRecord(t, repo, ctx, user.ID)
	penaltyTypeB := mustCreatePenaltyTypeRecord(t, repo, ctx, user.ID)
	svc := NewPenaltyService(repo)

	penaltyInClass, err := svc.CreatePenalty(ctx, user.ID, studentInClass.ID, penaltyTypeA.ID, nil, nil)
	if err != nil {
		t.Fatalf("CreatePenalty(in class) returned error: %v", err)
	}
	if _, err := svc.CreatePenalty(ctx, user.ID, studentOutClass.ID, penaltyTypeB.ID, nil, nil); err != nil {
		t.Fatalf("CreatePenalty(out class) returned error: %v", err)
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	studentInClassID := studentInClass.ID
	classroomID := classroom.ID
	penaltyTypeAID := penaltyTypeA.ID

	filtered, total, err := svc.ListPenalties(ctx, user.ID, ListPenaltiesFilters{
		StudentID:     &studentInClassID,
		ClassroomID:   &classroomID,
		PenaltyTypeID: &penaltyTypeAID,
		CreatedFrom:   &today,
		CreatedTo:     &today,
		Limit:         20,
		Offset:        0,
	})
	if err != nil {
		t.Fatalf("ListPenalties(combined filters) returned error: %v", err)
	}
	if total != 1 || len(filtered) != 1 || filtered[0].ID != penaltyInClass.ID {
		t.Fatalf("unexpected combined filters result: total=%d len=%d data=%+v", total, len(filtered), filtered)
	}

	penaltyTypeBID := penaltyTypeB.ID
	filtered, total, err = svc.ListPenalties(ctx, user.ID, ListPenaltiesFilters{
		ClassroomID:   &classroomID,
		PenaltyTypeID: &penaltyTypeBID,
		Limit:         20,
		Offset:        0,
	})
	if err != nil {
		t.Fatalf("ListPenalties(classroom+type mismatch) returned error: %v", err)
	}
	if total != 0 || len(filtered) != 0 {
		t.Fatalf("expected no penalties for classroom+type mismatch, got total=%d len=%d", total, len(filtered))
	}
}

func TestPenaltyService_ListPenalties_UsesOccurredAtForFilterAndSort_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	student := mustCreateStudentRecord(t, repo, ctx, user.ID)
	penaltyType := mustCreatePenaltyTypeRecord(t, repo, ctx, user.ID)
	svc := NewPenaltyService(repo)

	recentOccurred := time.Now().UTC()
	backdatedOccurred := recentOccurred.AddDate(0, 0, -4)

	recentPenalty, err := svc.CreatePenalty(ctx, user.ID, student.ID, penaltyType.ID, &recentOccurred, nil)
	if err != nil {
		t.Fatalf("CreatePenalty(recent) returned error: %v", err)
	}
	backdatedPenalty, err := svc.CreatePenalty(ctx, user.ID, student.ID, penaltyType.ID, &backdatedOccurred, nil)
	if err != nil {
		t.Fatalf("CreatePenalty(backdated) returned error: %v", err)
	}

	studentID := student.ID
	all, total, err := svc.ListPenalties(ctx, user.ID, ListPenaltiesFilters{
		StudentID: &studentID,
		Limit:     20,
		Offset:    0,
	})
	if err != nil {
		t.Fatalf("ListPenalties returned error: %v", err)
	}
	if total != 2 || len(all) != 2 {
		t.Fatalf("expected two penalties, got total=%d len=%d", total, len(all))
	}
	if all[0].ID != recentPenalty.ID {
		t.Fatalf("expected recent occurred_at penalty first, got %s (backdated=%s)", all[0].ID, backdatedPenalty.ID)
	}

	today := recentOccurred.Truncate(24 * time.Hour)
	filtered, total, err := svc.ListPenalties(ctx, user.ID, ListPenaltiesFilters{
		StudentID:   &studentID,
		CreatedFrom: &today,
		CreatedTo:   &today,
		Limit:       20,
		Offset:      0,
	})
	if err != nil {
		t.Fatalf("ListPenalties(created range) returned error: %v", err)
	}
	if total != 1 || len(filtered) != 1 || filtered[0].ID != recentPenalty.ID {
		t.Fatalf("expected only recent penalty in created range filter, got total=%d len=%d data=%+v", total, len(filtered), filtered)
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

func TestDashboardService_RecentListsUseOccurredAtOrder_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	student := mustCreateStudentRecord(t, repo, ctx, user.ID)
	bonusType := mustCreateBonusTypeRecord(t, repo, ctx, user.ID)
	penaltyType := mustCreatePenaltyTypeRecord(t, repo, ctx, user.ID)
	punishmentType := mustCreatePunishmentTypeRecord(t, repo, ctx, user.ID)

	recentOccurred := time.Now().UTC()
	backdatedOccurred := recentOccurred.AddDate(0, 0, -3)

	recentBonus, err := repo.CreateBonus(ctx, repository.CreateBonusParams{
		UserID:      user.ID,
		StudentID:   student.ID,
		BonusTypeID: bonusType.ID,
		Points:      2,
		OccurredAt:  &recentOccurred,
	})
	if err != nil {
		t.Fatalf("failed to create recent bonus: %v", err)
	}
	backdatedBonus, err := repo.CreateBonus(ctx, repository.CreateBonusParams{
		UserID:      user.ID,
		StudentID:   student.ID,
		BonusTypeID: bonusType.ID,
		Points:      1,
		OccurredAt:  &backdatedOccurred,
	})
	if err != nil {
		t.Fatalf("failed to create backdated bonus: %v", err)
	}

	recentPenalty, err := repo.CreatePenalty(ctx, repository.CreatePenaltyParams{
		UserID:        user.ID,
		StudentID:     student.ID,
		PenaltyTypeID: penaltyType.ID,
		OccurredAt:    &recentOccurred,
	})
	if err != nil {
		t.Fatalf("failed to create recent penalty: %v", err)
	}
	backdatedPenalty, err := repo.CreatePenalty(ctx, repository.CreatePenaltyParams{
		UserID:        user.ID,
		StudentID:     student.ID,
		PenaltyTypeID: penaltyType.ID,
		OccurredAt:    &backdatedOccurred,
	})
	if err != nil {
		t.Fatalf("failed to create backdated penalty: %v", err)
	}

	dueAt := time.Now().UTC().Add(24 * time.Hour)
	recentPunishment, err := repo.CreatePunishment(ctx, repository.CreatePunishmentParams{
		UserID:           user.ID,
		StudentID:        student.ID,
		PunishmentTypeID: punishmentType.ID,
		DueAt:            dueAt,
		OccurredAt:       &recentOccurred,
	})
	if err != nil {
		t.Fatalf("failed to create recent punishment: %v", err)
	}
	backdatedPunishment, err := repo.CreatePunishment(ctx, repository.CreatePunishmentParams{
		UserID:           user.ID,
		StudentID:        student.ID,
		PunishmentTypeID: punishmentType.ID,
		DueAt:            dueAt,
		OccurredAt:       &backdatedOccurred,
	})
	if err != nil {
		t.Fatalf("failed to create backdated punishment: %v", err)
	}

	svc := NewDashboardService(repo)
	dashboard, err := svc.GetDashboard(ctx, user.ID, nil)
	if err != nil {
		t.Fatalf("GetDashboard returned error: %v", err)
	}

	if len(dashboard.RecentBonuses) < 2 || len(dashboard.RecentPenalties) < 2 || len(dashboard.PendingPunishments) < 2 {
		t.Fatalf("expected at least two items in each dashboard list")
	}

	if dashboard.RecentBonuses[0].ID != recentBonus.ID || dashboard.RecentBonuses[1].ID != backdatedBonus.ID {
		t.Fatalf("unexpected bonus order by occurred_at: got %s,%s expected %s,%s", dashboard.RecentBonuses[0].ID, dashboard.RecentBonuses[1].ID, recentBonus.ID, backdatedBonus.ID)
	}
	if dashboard.RecentPenalties[0].ID != recentPenalty.ID || dashboard.RecentPenalties[1].ID != backdatedPenalty.ID {
		t.Fatalf("unexpected penalty order by occurred_at: got %s,%s expected %s,%s", dashboard.RecentPenalties[0].ID, dashboard.RecentPenalties[1].ID, recentPenalty.ID, backdatedPenalty.ID)
	}
	if dashboard.PendingPunishments[0].ID != recentPunishment.ID || dashboard.PendingPunishments[1].ID != backdatedPunishment.ID {
		t.Fatalf("unexpected punishment order by occurred_at: got %s,%s expected %s,%s", dashboard.PendingPunishments[0].ID, dashboard.PendingPunishments[1].ID, recentPunishment.ID, backdatedPunishment.ID)
	}
}
