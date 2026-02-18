package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/repository"
	"github.com/mageas/the-punisher-backend/internal/testutil/handlertest"
	"github.com/mageas/the-punisher-backend/internal/testutil/httpx"
	"github.com/mageas/the-punisher-backend/internal/testutil/inmemory"
	shared "github.com/mageas/the-punisher-backend/internal/testutil/shared"
)

type studentProfileStudentResponse struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
}

type studentProfileKpisResponse struct {
	AvailableBonusPoints   float64 `json:"available_bonus_points"`
	ActiveBonusCount       int64   `json:"active_bonus_count"`
	TotalPenaltyCount      int64   `json:"total_penalty_count"`
	PendingPunishmentCount int64   `json:"pending_punishment_count"`
}

type studentProfilePendingPunishmentResponse struct {
	ID                 uuid.UUID  `json:"id"`
	PunishmentTypeID   uuid.UUID  `json:"punishment_type_id"`
	PunishmentTypeName string     `json:"punishment_type_name"`
	TriggeringRuleID   *uuid.UUID `json:"triggering_rule_id"`
	TriggeringRuleName *string    `json:"triggering_rule_name"`
}

type studentProfileAvailableBonusResponse struct {
	ID            uuid.UUID `json:"id"`
	BonusTypeID   uuid.UUID `json:"bonus_type_id"`
	BonusTypeName string    `json:"bonus_type_name"`
	Points        float64   `json:"points"`
}

type studentProfileHistoryItemResponse struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type studentProfileResponse struct {
	Student            studentProfileStudentResponse             `json:"student"`
	Classrooms         []studentClassroomResponse                `json:"classrooms"`
	Kpis               studentProfileKpisResponse                `json:"kpis"`
	PendingPunishments []studentProfilePendingPunishmentResponse `json:"pending_punishments"`
	AvailableBonuses   []studentProfileAvailableBonusResponse    `json:"available_bonuses"`
	History            []studentProfileHistoryItemResponse       `json:"history"`
}

func TestStudentProfileHandlerSuccess(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newStudentRouter(repo, cfg)

	userID := uuid.New()
	studentID := uuid.New()
	classroomID := uuid.New()
	bonusTypeID := uuid.New()
	penaltyTypeID := uuid.New()
	punishmentTypeID := uuid.New()
	ruleID := uuid.New()

	base := time.Date(2026, 2, 2, 10, 0, 0, 0, time.UTC)

	repo.SeedStudent(repository.Student{
		ID:        studentID,
		UserID:    userID,
		FirstName: "Lucas",
		LastName:  "Dubois",
		CreatedAt: base,
		UpdatedAt: base,
	})

	repo.SeedClassroom(repository.Classroom{ID: classroomID, UserID: userID, Name: "6eme A", CreatedAt: base.Add(-time.Hour), UpdatedAt: base.Add(-time.Hour)})
	if _, err := repo.AddStudentToClassroom(t.Context(), repository.AddStudentToClassroomParams{StudentID: studentID, ClassroomID: classroomID, UserID: userID}); err != nil {
		t.Fatalf("failed to seed student/classroom relation: %v", err)
	}

	repo.SeedBonusType(repository.BonusType{ID: bonusTypeID, UserID: userID, Name: "Participation"})
	repo.SeedPenaltyType(repository.PenaltyType{ID: penaltyTypeID, UserID: userID, Name: "Bavardage"})
	repo.SeedPunishmentType(repository.PunishmentType{ID: punishmentTypeID, UserID: userID, Name: "Retenue"})
	repo.SeedRule(repository.Rule{ID: ruleID, UserID: userID, Name: "3 bavardages => retenue", PenaltyTypeID: penaltyTypeID, ResultingPunishmentTypeID: punishmentTypeID, Mode: "every", Threshold: 3, IsActive: true, DueAtAfterDays: 7})

	repo.SeedBonus(repository.Bonus{ID: uuid.New(), UserID: userID, StudentID: studentID, BonusTypeID: bonusTypeID, Points: 2, CreatedAt: base.Add(1 * time.Hour)})
	repo.SeedBonus(repository.Bonus{ID: uuid.New(), UserID: userID, StudentID: studentID, BonusTypeID: bonusTypeID, Points: 1, CreatedAt: base.Add(2 * time.Hour), UsedAt: pgtype.Timestamptz{Time: base.Add(3 * time.Hour), Valid: true}})

	repo.SeedPenalty(repository.Penalty{ID: uuid.New(), UserID: userID, StudentID: studentID, PenaltyTypeID: penaltyTypeID, CreatedAt: base.Add(4 * time.Hour)})

	repo.SeedPunishment(repository.Punishment{ID: uuid.New(), UserID: userID, StudentID: studentID, PunishmentTypeID: punishmentTypeID, TriggeringRuleID: pgtype.UUID{Bytes: ruleID, Valid: true}, CreatedAt: base.Add(5 * time.Hour), DueAt: base.Add(24 * time.Hour)})
	repo.SeedPunishment(repository.Punishment{ID: uuid.New(), UserID: userID, StudentID: studentID, PunishmentTypeID: punishmentTypeID, CreatedAt: base.Add(6 * time.Hour), DueAt: base.Add(24 * time.Hour), ResolvedAt: pgtype.Timestamptz{Time: base.Add(8 * time.Hour), Valid: true}})

	req := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+studentID.String()+"/profile", userID, cfg)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	resp := httpx.DecodeJSONResponse[studentProfileResponse](t, rr)
	if resp.Student.ID != studentID || resp.Student.FirstName != "Lucas" || resp.Student.LastName != "Dubois" {
		t.Fatalf("unexpected student payload: %+v", resp.Student)
	}
	if len(resp.Classrooms) != 1 || resp.Classrooms[0].ID != classroomID || resp.Classrooms[0].Name != "6eme A" {
		t.Fatalf("unexpected classrooms payload: %+v", resp.Classrooms)
	}
	if resp.Kpis.AvailableBonusPoints != 2 || resp.Kpis.ActiveBonusCount != 1 || resp.Kpis.TotalPenaltyCount != 1 || resp.Kpis.PendingPunishmentCount != 1 {
		t.Fatalf("unexpected kpis: %+v", resp.Kpis)
	}
	if len(resp.PendingPunishments) != 1 {
		t.Fatalf("expected one pending punishment, got %+v", resp.PendingPunishments)
	}
	if resp.PendingPunishments[0].PunishmentTypeID != punishmentTypeID || resp.PendingPunishments[0].PunishmentTypeName != "Retenue" {
		t.Fatalf("unexpected pending punishment payload: %+v", resp.PendingPunishments[0])
	}
	if resp.PendingPunishments[0].TriggeringRuleID == nil || *resp.PendingPunishments[0].TriggeringRuleID != ruleID {
		t.Fatalf("expected triggering_rule_id %s, got %+v", ruleID, resp.PendingPunishments[0].TriggeringRuleID)
	}
	if resp.PendingPunishments[0].TriggeringRuleName == nil || *resp.PendingPunishments[0].TriggeringRuleName != "3 bavardages => retenue" {
		t.Fatalf("unexpected triggering_rule_name: %+v", resp.PendingPunishments[0].TriggeringRuleName)
	}
	if len(resp.AvailableBonuses) != 1 || resp.AvailableBonuses[0].BonusTypeID != bonusTypeID || resp.AvailableBonuses[0].BonusTypeName != "Participation" || resp.AvailableBonuses[0].Points != 2 {
		t.Fatalf("unexpected available_bonuses payload: %+v", resp.AvailableBonuses)
	}

	if len(resp.History) != 5 {
		t.Fatalf("expected 5 history items, got %d (%+v)", len(resp.History), resp.History)
	}
	if resp.History[0].Type != "punishment" {
		t.Fatalf("expected first history item to be latest punishment, got %+v", resp.History[0])
	}

	typesCount := map[string]int{}
	for _, item := range resp.History {
		typesCount[item.Type]++
	}
	if typesCount["bonus"] != 2 || typesCount["penalty"] != 1 || typesCount["punishment"] != 2 {
		t.Fatalf("unexpected history type distribution: %+v", typesCount)
	}
}

func TestStudentProfileHandlerHistoryPagination(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newStudentRouter(repo, cfg)
	userID := uuid.New()
	studentID := uuid.New()

	repo.SeedStudent(inmemoryStudent(studentID, userID, "Jean", "Dupont"))
	repo.SeedPenaltyType(repository.PenaltyType{ID: uuid.New(), UserID: userID, Name: "Retard"})

	req := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+studentID.String()+"/profile?history_page=2", userID, cfg)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	resp := httpx.DecodeJSONResponse[studentProfileResponse](t, rr)
	if len(resp.History) != 0 {
		t.Fatalf("expected empty second history page, got %+v", resp.History)
	}

	reqInvalidPage := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+studentID.String()+"/profile?history_page=bad", userID, cfg)
	rrInvalidPage := httptest.NewRecorder()
	router.ServeHTTP(rrInvalidPage, reqInvalidPage)

	if rrInvalidPage.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rrInvalidPage.Code)
	}
}

func TestStudentProfileHandlerErrors(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newStudentRouter(repo, cfg)
	userID := uuid.New()

	t.Run("malformed_student_id", func(t *testing.T) {
		req := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/not-a-uuid/profile", userID, cfg)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrMalformedParameter.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrMalformedParameter.Error(), resp.Error)
		}
	})

	t.Run("student_not_found", func(t *testing.T) {
		req := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+uuid.New().String()+"/profile", userID, cfg)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrStudentNotFound.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrStudentNotFound.Error(), resp.Error)
		}
	})

	t.Run("internal_error", func(t *testing.T) {
		studentID := uuid.New()
		repo.SeedStudent(inmemoryStudent(studentID, userID, "Jean", "Dupont"))
		repo.SetError(inmemory.OpGetStudentProfileKpis, errors.New("database unavailable"))

		req := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+studentID.String()+"/profile", userID, cfg)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrInternalError.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrInternalError.Error(), resp.Error)
		}
	})
}
