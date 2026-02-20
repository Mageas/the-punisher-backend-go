package integration

import (
	"context"
	"testing"
	"time"

	"github.com/mageas/the-punisher-backend/internal/repository"
	"github.com/mageas/the-punisher-backend/internal/service"
)

func TestBusinessFlow_PenaltyRulePunishment(t *testing.T) {
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

	// Setup Types
	penaltyType, err := repo.CreatePenaltyType(ctx, repository.CreatePenaltyTypeParams{
		UserID: user.ID,
		Name:   "Chattering",
	})
	if err != nil {
		t.Fatalf("Failed to create penalty type: %v", err)
	}

	punishmentType, err := repo.CreatePunishmentType(ctx, repository.CreatePunishmentTypeParams{
		UserID: user.ID,
		Name:   "Detention",
	})
	if err != nil {
		t.Fatalf("Failed to create punishment type: %v", err)
	}

	// Setup Rule: 3 Chattering -> Detention
	_, err = repo.CreateRule(ctx, repository.CreateRuleParams{
		UserID:                     user.ID,
		Name:                       "3 Chatterings = Detention",
		ResultingPunishmentTypeID:  punishmentType.ID,
		PenaltyTypeID:              penaltyType.ID,
		Threshold:                  3,
		Mode:                       "at", // Use "at" mode to trigger exactly at 3
		IsActive:                   true,
		DueAtAfterDays:             7,
	})
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	// Apply 2 Penalties via Repo (bypassing service rule check for speed/simplicity, or use service)
	// Using service is safer to ensure consistency, but repo is faster for setup.
	for i := 0; i < 2; i++ {
		_, err := repo.CreatePenalty(ctx, repository.CreatePenaltyParams{
			UserID:        user.ID,
			StudentID:     student.ID,
			PenaltyTypeID: penaltyType.ID,
		})
		if err != nil {
			t.Fatalf("Failed to create penalty %d: %v", i, err)
		}
	}

	// Check NO punishment yet
	punishments, err := repo.ListPunishmentsByStudent(ctx, repository.ListPunishmentsByStudentParams{
		StudentID:   student.ID,
		UserID:      user.ID,
		QueryLimit:  10,
		QueryOffset: 0,
	})
	if err != nil {
		t.Fatalf("Failed to list punishments: %v", err)
	}
	if len(punishments) != 0 {
		t.Errorf("Expected 0 punishments, got %d", len(punishments))
	}

	// Apply 3rd Penalty (The Trigger) via Service
	penaltySvc := service.NewPenaltyService(repo)
	_, err = penaltySvc.CreatePenalty(ctx, user.ID, student.ID, penaltyType.ID)
	if err != nil {
		t.Fatalf("Failed to create 3rd penalty via service: %v", err)
	}

	// Check Punishment Created
	punishments, err = repo.ListPunishmentsByStudent(ctx, repository.ListPunishmentsByStudentParams{
		StudentID:   student.ID,
		UserID:      user.ID,
		QueryLimit:  10,
		QueryOffset: 0,
	})
	if err != nil {
		t.Fatalf("Failed to list punishments: %v", err)
	}
	if len(punishments) != 1 {
		t.Fatalf("Expected 1 punishment, got %d", len(punishments))
	}
	// Verify details
	if punishments[0].PunishmentTypeID != punishmentType.ID {
		t.Errorf("Expected punishment type %s, got %s", punishmentType.ID, punishments[0].PunishmentTypeID)
	}
	if !punishments[0].Automated {
		t.Errorf("Expected automated punishment")
	}
	if punishments[0].TriggeringRuleID == nil {
		t.Errorf("Expected triggering rule ID")
	}
	
	// Verify Due Date is roughly 7 days from now
	expectedDue := time.Now().Add(7 * 24 * time.Hour)
	if punishments[0].DueAt.Before(expectedDue.Add(-1*time.Hour)) || punishments[0].DueAt.After(expectedDue.Add(1*time.Hour)) {
		t.Errorf("Expected due date around %v, got %v", expectedDue, punishments[0].DueAt)
	}
}
