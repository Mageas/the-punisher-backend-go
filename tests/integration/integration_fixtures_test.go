//go:build integration

package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

const testUserTimezone = "Europe/Paris"

func ptr[T any](v T) *T {
	return &v
}

func assertTimeEqualToPostgresPrecision(t *testing.T, field string, got, expected time.Time) {
	t.Helper()

	gotNormalized := got.UTC().Truncate(time.Microsecond)
	expectedNormalized := expected.UTC().Truncate(time.Microsecond)
	if !gotNormalized.Equal(expectedNormalized) {
		t.Fatalf("expected %s %s, got %s", field, expectedNormalized, gotNormalized)
	}
}

func uniqueValue(prefix string) string {
	compact := strings.ReplaceAll(uuid.NewString(), "-", "")
	return fmt.Sprintf("%s-%s", prefix, compact[:10])
}

func mustLoadLocation(t *testing.T, timezone string) *time.Location {
	t.Helper()

	location, err := time.LoadLocation(timezone)
	if err != nil {
		t.Fatalf("failed to load timezone %s: %v", timezone, err)
	}

	return location
}

func filterDateInTimezone(t *testing.T, value time.Time, timezone string) time.Time {
	t.Helper()

	location := mustLoadLocation(t, timezone)
	localValue := value.In(location)
	return time.Date(localValue.Year(), localValue.Month(), localValue.Day(), 0, 0, 0, 0, time.UTC)
}

func startOfDayInTimezone(t *testing.T, value time.Time, timezone string) time.Time {
	t.Helper()

	location := mustLoadLocation(t, timezone)
	localValue := value.In(location)
	return time.Date(localValue.Year(), localValue.Month(), localValue.Day(), 0, 0, 0, 0, location)
}

func mustCreateUserRecord(t *testing.T, repo repository.Querier, ctx context.Context) repository.CreateUserRow {
	t.Helper()

	row, err := repo.CreateUser(ctx, repository.CreateUserParams{
		Email:        uniqueValue("user") + "@example.com",
		FirstName:    "John",
		LastName:     "Doe",
		PasswordHash: "$2a$10$Q6TSf7MmtQ6bfmI7T4B5du5n2QHnQjPdr6Vv2Wj3Z3J0v6YxS8GgK", // bcrypt for test password
	})
	if err != nil {
		t.Fatalf("failed to create user fixture: %v", err)
	}

	return row
}

func mustCreateStudentRecord(t *testing.T, repo repository.Querier, ctx context.Context, userID uuid.UUID) repository.CreateStudentRow {
	t.Helper()

	row, err := repo.CreateStudent(ctx, repository.CreateStudentParams{
		UserID:    userID,
		FirstName: uniqueValue("first"),
		LastName:  strings.ToUpper(uniqueValue("last")),
	})
	if err != nil {
		t.Fatalf("failed to create student fixture: %v", err)
	}

	return row
}

func mustCreateClassroomRecord(t *testing.T, repo repository.Querier, ctx context.Context, userID uuid.UUID) repository.CreateClassroomRow {
	t.Helper()

	year := "2025"
	teacher := "Mme Test"
	row, err := repo.CreateClassroom(ctx, repository.CreateClassroomParams{
		UserID:      userID,
		Name:        uniqueValue("class"),
		Year:        &year,
		MainTeacher: &teacher,
	})
	if err != nil {
		t.Fatalf("failed to create classroom fixture: %v", err)
	}

	return row
}

func mustCreateBonusTypeRecord(t *testing.T, repo repository.Querier, ctx context.Context, userID uuid.UUID) repository.BonusType {
	t.Helper()

	row, err := repo.CreateBonusType(ctx, repository.CreateBonusTypeParams{
		UserID: userID,
		Name:   uniqueValue("bonus-type"),
	})
	if err != nil {
		t.Fatalf("failed to create bonus type fixture: %v", err)
	}

	return row
}

func mustCreatePenaltyTypeRecord(t *testing.T, repo repository.Querier, ctx context.Context, userID uuid.UUID) repository.PenaltyType {
	t.Helper()

	row, err := repo.CreatePenaltyType(ctx, repository.CreatePenaltyTypeParams{
		UserID: userID,
		Name:   uniqueValue("penalty-type"),
	})
	if err != nil {
		t.Fatalf("failed to create penalty type fixture: %v", err)
	}

	return row
}

func mustCreatePunishmentTypeRecord(t *testing.T, repo repository.Querier, ctx context.Context, userID uuid.UUID) repository.PunishmentType {
	t.Helper()

	row, err := repo.CreatePunishmentType(ctx, repository.CreatePunishmentTypeParams{
		UserID: userID,
		Name:   uniqueValue("punishment-type"),
	})
	if err != nil {
		t.Fatalf("failed to create punishment type fixture: %v", err)
	}

	return row
}

func mustCreateBonusRecord(t *testing.T, repo repository.Querier, ctx context.Context, userID, studentID, bonusTypeID uuid.UUID, points float64) repository.CreateBonusRow {
	t.Helper()

	row, err := repo.CreateBonus(ctx, repository.CreateBonusParams{
		UserID:      userID,
		StudentID:   studentID,
		BonusTypeID: bonusTypeID,
		Points:      points,
	})
	if err != nil {
		t.Fatalf("failed to create bonus fixture: %v", err)
	}

	return row
}

func mustCreatePenaltyRecord(t *testing.T, repo repository.Querier, ctx context.Context, userID, studentID, penaltyTypeID uuid.UUID) repository.CreatePenaltyRow {
	t.Helper()

	row, err := repo.CreatePenalty(ctx, repository.CreatePenaltyParams{
		UserID:        userID,
		StudentID:     studentID,
		PenaltyTypeID: penaltyTypeID,
	})
	if err != nil {
		t.Fatalf("failed to create penalty fixture: %v", err)
	}

	return row
}

func mustCreatePunishmentRecord(t *testing.T, repo repository.Querier, ctx context.Context, userID, studentID, punishmentTypeID uuid.UUID, dueAt time.Time) repository.CreatePunishmentRow {
	t.Helper()

	row, err := repo.CreatePunishment(ctx, repository.CreatePunishmentParams{
		UserID:           userID,
		StudentID:        studentID,
		PunishmentTypeID: punishmentTypeID,
		DueAt:            dueAt,
	})
	if err != nil {
		t.Fatalf("failed to create punishment fixture: %v", err)
	}

	return row
}

func mustCreateRuleRecord(
	t *testing.T,
	repo repository.Querier,
	ctx context.Context,
	userID, penaltyTypeID, punishmentTypeID uuid.UUID,
	mode string,
	threshold int32,
	dueAtAfterDays int32,
	isActive bool,
) repository.CreateRuleRow {
	t.Helper()

	row, err := repo.CreateRule(ctx, repository.CreateRuleParams{
		UserID:                    userID,
		Name:                      uniqueValue("rule"),
		ResultingPunishmentTypeID: punishmentTypeID,
		PenaltyTypeID:             penaltyTypeID,
		Threshold:                 threshold,
		Mode:                      mode,
		IsActive:                  isActive,
		DueAtAfterDays:            &dueAtAfterDays,
		DueAtMode:                 "days",
	})
	if err != nil {
		t.Fatalf("failed to create rule fixture: %v", err)
	}

	return row
}

func mustCreateNextLessonsRuleRecord(
	t *testing.T,
	repo repository.Querier,
	ctx context.Context,
	userID, penaltyTypeID, punishmentTypeID uuid.UUID,
	mode string,
	threshold int32,
	dueAtAfterLessons int32,
	isActive bool,
) repository.CreateRuleRow {
	t.Helper()

	row, err := repo.CreateRule(ctx, repository.CreateRuleParams{
		UserID:                    userID,
		Name:                      uniqueValue("rule-next-lessons"),
		ResultingPunishmentTypeID: punishmentTypeID,
		PenaltyTypeID:             penaltyTypeID,
		Threshold:                 threshold,
		Mode:                      mode,
		IsActive:                  isActive,
		DueAtAfterDays:            nil,
		DueAtMode:                 "next_lessons",
		DueAtAfterLessons:         &dueAtAfterLessons,
	})
	if err != nil {
		t.Fatalf("failed to create next_lessons rule fixture: %v", err)
	}

	return row
}
