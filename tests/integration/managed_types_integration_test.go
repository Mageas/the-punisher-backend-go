//go:build integration

package integration

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	. "github.com/mageas/the-punisher-backend/internal/service"
)

func TestBonusTypeService_CRUD_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	userSvc := NewUserService(repo)
	user, err := userSvc.CreateUser(ctx, dto.RequestUserDto{
		Email:     "types-bonus@example.com",
		FirstName: "Bonus",
		LastName:  "Type",
		Password:  "VeryStrongPassword123!",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	svc := NewBonusTypeService(repo)
	created, err := svc.CreateBonusType(ctx, user.ID, dto.RequestBonusTypeDto{Name: "Participation"})
	if err != nil {
		t.Fatalf("CreateBonusType returned error: %v", err)
	}
	if _, err := svc.CreateBonusType(ctx, user.ID, dto.RequestBonusTypeDto{Name: "Aide"}); err != nil {
		t.Fatalf("CreateBonusType (extra fixture) returned error: %v", err)
	}

	got, err := svc.GetBonusType(ctx, user.ID, created.ID)
	if err != nil {
		t.Fatalf("GetBonusType returned error: %v", err)
	}
	if got.Name != "Participation" {
		t.Fatalf("unexpected name: %s", got.Name)
	}

	all, total, err := svc.ListBonusTypes(ctx, user.ID, nil, 20, 0)
	if err != nil {
		t.Fatalf("ListBonusTypes returned error: %v", err)
	}
	if total != 2 || len(all) != 2 {
		t.Fatalf("expected two rows, got total=%d len=%d", total, len(all))
	}

	search := "parti"
	filtered, filteredTotal, err := svc.ListBonusTypes(ctx, user.ID, &search, 20, 0)
	if err != nil {
		t.Fatalf("ListBonusTypes (with search) returned error: %v", err)
	}
	if filteredTotal != 1 || len(filtered) != 1 {
		t.Fatalf("expected one filtered row, got total=%d len=%d", filteredTotal, len(filtered))
	}
	if filtered[0].ID != created.ID {
		t.Fatalf("unexpected filtered bonus type id: %s", filtered[0].ID)
	}

	newName := "Assiduite"
	updated, err := svc.UpdateBonusType(ctx, user.ID, created.ID, dto.UpdateBonusTypeDto{Name: &newName})
	if err != nil {
		t.Fatalf("UpdateBonusType returned error: %v", err)
	}
	if updated.Name != newName {
		t.Fatalf("unexpected updated name: %s", updated.Name)
	}

	if err := svc.DeleteBonusType(ctx, user.ID, created.ID); err != nil {
		t.Fatalf("DeleteBonusType returned error: %v", err)
	}

	_, err = svc.GetBonusType(ctx, user.ID, created.ID)
	if !errors.Is(err, api.ErrBonusTypeNotFound) {
		t.Fatalf("expected ErrBonusTypeNotFound, got %v", err)
	}
}

func TestPenaltyTypeService_CRUD_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	userSvc := NewUserService(repo)
	user, err := userSvc.CreateUser(ctx, dto.RequestUserDto{
		Email:     "types-penalty@example.com",
		FirstName: "Penalty",
		LastName:  "Type",
		Password:  "VeryStrongPassword123!",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	svc := NewPenaltyTypeService(repo)
	created, err := svc.CreatePenaltyType(ctx, user.ID, dto.RequestPenaltyTypeDto{Name: "Retard"})
	if err != nil {
		t.Fatalf("CreatePenaltyType returned error: %v", err)
	}
	if _, err := svc.CreatePenaltyType(ctx, user.ID, dto.RequestPenaltyTypeDto{Name: "Bavardage"}); err != nil {
		t.Fatalf("CreatePenaltyType (extra fixture) returned error: %v", err)
	}

	got, err := svc.GetPenaltyType(ctx, user.ID, created.ID)
	if err != nil {
		t.Fatalf("GetPenaltyType returned error: %v", err)
	}
	if got.Name != "Retard" {
		t.Fatalf("unexpected name: %s", got.Name)
	}

	all, total, err := svc.ListPenaltyTypes(ctx, user.ID, nil, 20, 0)
	if err != nil {
		t.Fatalf("ListPenaltyTypes returned error: %v", err)
	}
	if total != 2 || len(all) != 2 {
		t.Fatalf("expected two rows, got total=%d len=%d", total, len(all))
	}

	search := "reta"
	filtered, filteredTotal, err := svc.ListPenaltyTypes(ctx, user.ID, &search, 20, 0)
	if err != nil {
		t.Fatalf("ListPenaltyTypes (with search) returned error: %v", err)
	}
	if filteredTotal != 1 || len(filtered) != 1 {
		t.Fatalf("expected one filtered row, got total=%d len=%d", filteredTotal, len(filtered))
	}
	if filtered[0].ID != created.ID {
		t.Fatalf("unexpected filtered penalty type id: %s", filtered[0].ID)
	}

	newName := "Bavardage"
	updated, err := svc.UpdatePenaltyType(ctx, user.ID, created.ID, dto.UpdatePenaltyTypeDto{Name: &newName})
	if err != nil {
		t.Fatalf("UpdatePenaltyType returned error: %v", err)
	}
	if updated.Name != newName {
		t.Fatalf("unexpected updated name: %s", updated.Name)
	}

	if err := svc.DeletePenaltyType(ctx, user.ID, created.ID); err != nil {
		t.Fatalf("DeletePenaltyType returned error: %v", err)
	}

	_, err = svc.GetPenaltyType(ctx, user.ID, created.ID)
	if !errors.Is(err, api.ErrPenaltyTypeNotFound) {
		t.Fatalf("expected ErrPenaltyTypeNotFound, got %v", err)
	}
}

func TestPunishmentTypeService_CRUD_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	userSvc := NewUserService(repo)
	user, err := userSvc.CreateUser(ctx, dto.RequestUserDto{
		Email:     "types-punishment@example.com",
		FirstName: "Punishment",
		LastName:  "Type",
		Password:  "VeryStrongPassword123!",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	svc := NewPunishmentTypeService(repo)
	created, err := svc.CreatePunishmentType(ctx, user.ID, dto.RequestPunishmentTypeDto{Name: "Copie"})
	if err != nil {
		t.Fatalf("CreatePunishmentType returned error: %v", err)
	}
	if _, err := svc.CreatePunishmentType(ctx, user.ID, dto.RequestPunishmentTypeDto{Name: "Heure de colle"}); err != nil {
		t.Fatalf("CreatePunishmentType (extra fixture) returned error: %v", err)
	}

	got, err := svc.GetPunishmentType(ctx, user.ID, created.ID)
	if err != nil {
		t.Fatalf("GetPunishmentType returned error: %v", err)
	}
	if got.Name != "Copie" {
		t.Fatalf("unexpected name: %s", got.Name)
	}

	all, total, err := svc.ListPunishmentTypes(ctx, user.ID, nil, 20, 0)
	if err != nil {
		t.Fatalf("ListPunishmentTypes returned error: %v", err)
	}
	if total != 2 || len(all) != 2 {
		t.Fatalf("expected two rows, got total=%d len=%d", total, len(all))
	}

	search := "copi"
	filtered, filteredTotal, err := svc.ListPunishmentTypes(ctx, user.ID, &search, 20, 0)
	if err != nil {
		t.Fatalf("ListPunishmentTypes (with search) returned error: %v", err)
	}
	if filteredTotal != 1 || len(filtered) != 1 {
		t.Fatalf("expected one filtered row, got total=%d len=%d", filteredTotal, len(filtered))
	}
	if filtered[0].ID != created.ID {
		t.Fatalf("unexpected filtered punishment type id: %s", filtered[0].ID)
	}

	newName := "Heure de colle"
	updated, err := svc.UpdatePunishmentType(ctx, user.ID, created.ID, dto.UpdatePunishmentTypeDto{Name: &newName})
	if err != nil {
		t.Fatalf("UpdatePunishmentType returned error: %v", err)
	}
	if updated.Name != newName {
		t.Fatalf("unexpected updated name: %s", updated.Name)
	}

	if err := svc.DeletePunishmentType(ctx, user.ID, created.ID); err != nil {
		t.Fatalf("DeletePunishmentType returned error: %v", err)
	}

	_, err = svc.GetPunishmentType(ctx, user.ID, created.ID)
	if !errors.Is(err, api.ErrPunishmentTypeNotFound) {
		t.Fatalf("expected ErrPunishmentTypeNotFound, got %v", err)
	}
}

func TestManagedTypeServices_NotFoundBranches_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	missingID := uuid.New()
	userID := uuid.New()

	bonusTypeSvc := NewBonusTypeService(repo)
	if err := bonusTypeSvc.DeleteBonusType(ctx, userID, missingID); !errors.Is(err, api.ErrBonusTypeNotFound) {
		t.Fatalf("expected ErrBonusTypeNotFound, got %v", err)
	}

	penaltyTypeSvc := NewPenaltyTypeService(repo)
	if err := penaltyTypeSvc.DeletePenaltyType(ctx, userID, missingID); !errors.Is(err, api.ErrPenaltyTypeNotFound) {
		t.Fatalf("expected ErrPenaltyTypeNotFound, got %v", err)
	}

	punishmentTypeSvc := NewPunishmentTypeService(repo)
	if err := punishmentTypeSvc.DeletePunishmentType(ctx, userID, missingID); !errors.Is(err, api.ErrPunishmentTypeNotFound) {
		t.Fatalf("expected ErrPunishmentTypeNotFound, got %v", err)
	}
}
