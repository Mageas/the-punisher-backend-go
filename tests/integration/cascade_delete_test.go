package integration

import (
	"context"
	"testing"

	"github.com/mageas/the-punisher-backend/internal/repository"
)

func TestCascadeDelete_PenaltyType(t *testing.T) {
	if testDB == nil {
		t.Skip("Database not initialized")
	}
	truncateTables(t)

	ctx := context.Background()
	repo := repository.New(testDB)

	// Setup User
	user, err := repo.CreateUser(ctx, repository.CreateUserParams{
		Email:        "teacher@example.com",
		FirstName:    "Teacher",
		LastName:     "One",
		PasswordHash: "hash",
	})
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Setup Student
	student, err := repo.CreateStudent(ctx, repository.CreateStudentParams{
		UserID:    user.ID,
		FirstName: "Bad",
		LastName:  "Student",
	})
	if err != nil {
		t.Fatalf("Failed to create student: %v", err)
	}

	// Setup Penalty Type
	penaltyType, err := repo.CreatePenaltyType(ctx, repository.CreatePenaltyTypeParams{
		UserID: user.ID,
		Name:   "Chattering",
	})
	if err != nil {
		t.Fatalf("Failed to create penalty type: %v", err)
	}

	// Create a Penalty
	penalty, err := repo.CreatePenalty(ctx, repository.CreatePenaltyParams{
		UserID:        user.ID,
		StudentID:     student.ID,
		PenaltyTypeID: penaltyType.ID,
	})
	if err != nil {
		t.Fatalf("Failed to create penalty: %v", err)
	}

	// Delete Penalty Type
	rowsAffected, err := repo.DeletePenaltyTypeByUser(ctx, repository.DeletePenaltyTypeByUserParams{
		ID:     penaltyType.ID,
		UserID: user.ID,
	})
	if err != nil {
		t.Fatalf("Failed to delete penalty type: %v", err)
	}
	if rowsAffected != 1 {
		t.Fatalf("Expected 1 row affected, got %d", rowsAffected)
	}

	// Verify Penalty is gone (Cascade)
	_, err = repo.GetPenaltyByUser(ctx, repository.GetPenaltyByUserParams{
		ID:     penalty.ID,
		UserID: user.ID,
	})
	if err == nil {
		t.Errorf("Expected penalty to be deleted (cascade), but it still exists")
	}
}

func TestCascadeDelete_PunishmentType(t *testing.T) {
	if testDB == nil {
		t.Skip("Database not initialized")
	}
	truncateTables(t)

	ctx := context.Background()
	repo := repository.New(testDB)

	// Setup User
	user, _ := repo.CreateUser(ctx, repository.CreateUserParams{
		Email:        "teacher2@example.com",
		FirstName:    "Teacher",
		LastName:     "Two",
		PasswordHash: "hash",
	})

	// Setup Student
	student, _ := repo.CreateStudent(ctx, repository.CreateStudentParams{
		UserID:    user.ID,
		FirstName: "Bad",
		LastName:  "Student",
	})

	// Setup Punishment Type
	punishmentType, err := repo.CreatePunishmentType(ctx, repository.CreatePunishmentTypeParams{
		UserID: user.ID,
		Name:   "Detention",
	})
	if err != nil {
		t.Fatalf("Failed to create punishment type: %v", err)
	}

	// Create a Punishment
	punishment, err := repo.CreatePunishment(ctx, repository.CreatePunishmentParams{
		UserID:           user.ID,
		StudentID:        student.ID,
		PunishmentTypeID: punishmentType.ID,
	})
	if err != nil {
		t.Fatalf("Failed to create punishment: %v", err)
	}

	// Delete Punishment Type
	rowsAffected, err := repo.DeletePunishmentTypeByUser(ctx, repository.DeletePunishmentTypeByUserParams{
		ID:     punishmentType.ID,
		UserID: user.ID,
	})
	if err != nil {
		t.Fatalf("Failed to delete punishment type: %v", err)
	}
	if rowsAffected != 1 {
		t.Fatalf("Expected 1 row affected, got %d", rowsAffected)
	}

	// Verify Punishment is gone (Cascade)
	_, err = repo.GetPunishmentByUser(ctx, repository.GetPunishmentByUserParams{
		ID:     punishment.ID,
		UserID: user.ID,
	})
	if err == nil {
		t.Errorf("Expected punishment to be deleted (cascade), but it still exists")
	}
}

func TestCascadeDelete_PunishmentType_InRule(t *testing.T) {
	if testDB == nil {
		t.Skip("Database not initialized")
	}
	truncateTables(t)

	ctx := context.Background()
	repo := repository.New(testDB)

	user, _ := repo.CreateUser(ctx, repository.CreateUserParams{
		Email:        "teacher3@example.com",
		FirstName:    "Teacher",
		LastName:     "Three",
		PasswordHash: "hash",
	})

	penaltyType, _ := repo.CreatePenaltyType(ctx, repository.CreatePenaltyTypeParams{
		UserID: user.ID,
		Name:   "P1",
	})

	punishmentType, _ := repo.CreatePunishmentType(ctx, repository.CreatePunishmentTypeParams{
		UserID: user.ID,
		Name:   "T1",
	})

	// Create a Rule using this punishment type
	rule, err := repo.CreateRule(ctx, repository.CreateRuleParams{
		UserID:                    user.ID,
		Name:                      "Rule 1",
		ResultingPunishmentTypeID: punishmentType.ID,
		PenaltyTypeID:             penaltyType.ID,
		Threshold:                 3,
		Mode:                      "at",
	})
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	// Delete Punishment Type
	_, err = repo.DeletePunishmentTypeByUser(ctx, repository.DeletePunishmentTypeByUserParams{
		ID:     punishmentType.ID,
		UserID: user.ID,
	})
	if err != nil {
		t.Fatalf("Failed to delete punishment type: %v", err)
	}

	// Verify Rule is gone (Cascade - from migration 000013)
	_, err = repo.GetRuleByUser(ctx, repository.GetRuleByUserParams{
		ID:     rule.ID,
		UserID: user.ID,
	})
	if err == nil {
		t.Errorf("Expected rule to be deleted (cascade), but it still exists")
	}
}
