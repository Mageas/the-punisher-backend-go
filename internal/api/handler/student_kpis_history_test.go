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

type studentKpisResponse struct {
	AvailableBonusPoints   float64 `json:"available_bonus_points"`
	ActiveBonusCount       int64   `json:"active_bonus_count"`
	TotalPenaltyCount      int64   `json:"total_penalty_count"`
	PendingPunishmentCount int64   `json:"pending_punishment_count"`
}

type studentHistoryItemResponse struct {
	Type               string     `json:"type"`
	ID                 uuid.UUID  `json:"id"`
	PenaltyTypeID      *uuid.UUID `json:"penalty_type_id,omitempty"`
	PenaltyTypeName    *string    `json:"penalty_type_name,omitempty"`
	BonusTypeID        *uuid.UUID `json:"bonus_type_id,omitempty"`
	BonusTypeName      *string    `json:"bonus_type_name,omitempty"`
	Points             *float64   `json:"points,omitempty"`
	UsedAt             *time.Time `json:"used_at,omitempty"`
	PunishmentTypeID   *uuid.UUID `json:"punishment_type_id,omitempty"`
	PunishmentTypeName *string    `json:"punishment_type_name,omitempty"`
	TriggeringRuleID   *uuid.UUID `json:"triggering_rule_id,omitempty"`
	TriggeringRuleName *string    `json:"triggering_rule_name,omitempty"`
	DueAt              *time.Time `json:"due_at,omitempty"`
	ResolvedAt         *time.Time `json:"resolved_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

func TestStudentKpisHandlerSuccess(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newStudentRouter(repo, cfg)

	userID := uuid.New()
	studentID := uuid.New()
	bonusTypeID := uuid.New()
	penaltyTypeID := uuid.New()
	punishmentTypeID := uuid.New()
	base := time.Date(2026, 2, 2, 10, 0, 0, 0, time.UTC)

	repo.SeedStudent(inmemoryStudent(studentID, userID, "Lucas", "Dubois"))
	repo.SeedBonusType(repository.BonusType{ID: bonusTypeID, UserID: userID, Name: "Participation"})
	repo.SeedPenaltyType(repository.PenaltyType{ID: penaltyTypeID, UserID: userID, Name: "Bavardage"})
	repo.SeedPunishmentType(repository.PunishmentType{ID: punishmentTypeID, UserID: userID, Name: "Retenue"})

	repo.SeedBonus(repository.Bonus{ID: uuid.New(), UserID: userID, StudentID: studentID, BonusTypeID: bonusTypeID, Points: 2, CreatedAt: base.Add(1 * time.Hour)})
	repo.SeedBonus(repository.Bonus{ID: uuid.New(), UserID: userID, StudentID: studentID, BonusTypeID: bonusTypeID, Points: 1, CreatedAt: base.Add(2 * time.Hour), UsedAt: pgtype.Timestamptz{Time: base.Add(3 * time.Hour), Valid: true}})
	repo.SeedPenalty(repository.Penalty{ID: uuid.New(), UserID: userID, StudentID: studentID, PenaltyTypeID: penaltyTypeID, CreatedAt: base.Add(4 * time.Hour)})
	repo.SeedPunishment(repository.Punishment{ID: uuid.New(), UserID: userID, StudentID: studentID, PunishmentTypeID: punishmentTypeID, CreatedAt: base.Add(5 * time.Hour), DueAt: base.Add(24 * time.Hour)})
	repo.SeedPunishment(repository.Punishment{ID: uuid.New(), UserID: userID, StudentID: studentID, PunishmentTypeID: punishmentTypeID, CreatedAt: base.Add(6 * time.Hour), DueAt: base.Add(24 * time.Hour), ResolvedAt: pgtype.Timestamptz{Time: base.Add(8 * time.Hour), Valid: true}})

	req := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+studentID.String()+"/kpis", userID, cfg)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	resp := httpx.DecodeJSONResponse[studentKpisResponse](t, rr)
	if resp.AvailableBonusPoints != 2 || resp.ActiveBonusCount != 1 || resp.TotalPenaltyCount != 1 || resp.PendingPunishmentCount != 1 {
		t.Fatalf("unexpected kpis payload: %+v", resp)
	}
}

func TestStudentHistoryHandlerSuccess(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newStudentRouter(repo, cfg)

	userID := uuid.New()
	studentID := uuid.New()
	bonusTypeID := uuid.New()
	penaltyTypeID := uuid.New()
	punishmentTypeID := uuid.New()
	ruleID := uuid.New()
	base := time.Date(2026, 2, 2, 10, 0, 0, 0, time.UTC)

	repo.SeedStudent(inmemoryStudent(studentID, userID, "Lucas", "Dubois"))
	repo.SeedBonusType(repository.BonusType{ID: bonusTypeID, UserID: userID, Name: "Participation"})
	repo.SeedPenaltyType(repository.PenaltyType{ID: penaltyTypeID, UserID: userID, Name: "Bavardage"})
	repo.SeedPunishmentType(repository.PunishmentType{ID: punishmentTypeID, UserID: userID, Name: "Retenue"})
	repo.SeedRule(repository.Rule{ID: ruleID, UserID: userID, Name: "3 bavardages => retenue", PenaltyTypeID: penaltyTypeID, ResultingPunishmentTypeID: punishmentTypeID, Mode: "every", Threshold: 3, IsActive: true, DueAtAfterDays: 7})

	repo.SeedBonus(repository.Bonus{ID: uuid.New(), UserID: userID, StudentID: studentID, BonusTypeID: bonusTypeID, Points: 2, CreatedAt: base.Add(1 * time.Hour)})
	repo.SeedBonus(repository.Bonus{ID: uuid.New(), UserID: userID, StudentID: studentID, BonusTypeID: bonusTypeID, Points: 1, CreatedAt: base.Add(2 * time.Hour), UsedAt: pgtype.Timestamptz{Time: base.Add(3 * time.Hour), Valid: true}})
	repo.SeedPenalty(repository.Penalty{ID: uuid.New(), UserID: userID, StudentID: studentID, PenaltyTypeID: penaltyTypeID, CreatedAt: base.Add(4 * time.Hour)})
	repo.SeedPunishment(repository.Punishment{ID: uuid.New(), UserID: userID, StudentID: studentID, PunishmentTypeID: punishmentTypeID, TriggeringRuleID: pgtype.UUID{Bytes: ruleID, Valid: true}, CreatedAt: base.Add(5 * time.Hour), DueAt: base.Add(24 * time.Hour)})
	repo.SeedPunishment(repository.Punishment{ID: uuid.New(), UserID: userID, StudentID: studentID, PunishmentTypeID: punishmentTypeID, CreatedAt: base.Add(6 * time.Hour), DueAt: base.Add(24 * time.Hour), ResolvedAt: pgtype.Timestamptz{Time: base.Add(8 * time.Hour), Valid: true}})

	req := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+studentID.String()+"/history", userID, cfg)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	resp := httpx.DecodeJSONResponse[[]studentHistoryItemResponse](t, rr)
	if len(resp) != 5 {
		t.Fatalf("expected 5 history items, got %d (%+v)", len(resp), resp)
	}
	if resp[0].Type != "punishment" {
		t.Fatalf("expected first history item to be latest punishment, got %+v", resp[0])
	}
	if resp[0].PunishmentTypeID == nil || resp[0].PunishmentTypeName == nil || resp[0].DueAt == nil {
		t.Fatalf("expected punishment fields on first item, got %+v", resp[0])
	}

	typesCount := map[string]int{}
	for _, item := range resp {
		typesCount[item.Type]++
	}
	if typesCount["bonus"] != 2 || typesCount["penalty"] != 1 || typesCount["punishment"] != 2 {
		t.Fatalf("unexpected history type distribution: %+v", typesCount)
	}
}

func TestStudentHistoryHandlerPagination(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newStudentRouter(repo, cfg)
	userID := uuid.New()
	studentID := uuid.New()

	repo.SeedStudent(inmemoryStudent(studentID, userID, "Jean", "Dupont"))

	req := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+studentID.String()+"/history?page=2", userID, cfg)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	resp := httpx.DecodeJSONResponse[[]studentHistoryItemResponse](t, rr)
	if len(resp) != 0 {
		t.Fatalf("expected empty second history page, got %+v", resp)
	}

	reqLegacyPage := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+studentID.String()+"/history?history_page=2", userID, cfg)
	rrLegacyPage := httptest.NewRecorder()
	router.ServeHTTP(rrLegacyPage, reqLegacyPage)

	if rrLegacyPage.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rrLegacyPage.Code)
	}
}

func TestStudentKpisAndHistoryHandlersErrors(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newStudentRouter(repo, cfg)
	userID := uuid.New()

	t.Run("malformed_student_id", func(t *testing.T) {
		reqKpis := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/not-a-uuid/kpis", userID, cfg)
		rrKpis := httptest.NewRecorder()
		router.ServeHTTP(rrKpis, reqKpis)

		if rrKpis.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rrKpis.Code)
		}

		reqHistory := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/not-a-uuid/history", userID, cfg)
		rrHistory := httptest.NewRecorder()
		router.ServeHTTP(rrHistory, reqHistory)

		if rrHistory.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rrHistory.Code)
		}
	})

	t.Run("student_not_found", func(t *testing.T) {
		reqKpis := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+uuid.New().String()+"/kpis", userID, cfg)
		rrKpis := httptest.NewRecorder()
		router.ServeHTTP(rrKpis, reqKpis)

		if rrKpis.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, rrKpis.Code)
		}

		respKpis := httpx.DecodeJSONResponse[api.ErrorResponse](t, rrKpis)
		if respKpis.Error != api.ErrStudentNotFound.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrStudentNotFound.Error(), respKpis.Error)
		}

		reqHistory := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+uuid.New().String()+"/history", userID, cfg)
		rrHistory := httptest.NewRecorder()
		router.ServeHTTP(rrHistory, reqHistory)

		if rrHistory.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, rrHistory.Code)
		}
	})

	t.Run("internal_error", func(t *testing.T) {
		studentID := uuid.New()
		repo.SeedStudent(inmemoryStudent(studentID, userID, "Jean", "Dupont"))

		repo.SetError(inmemory.OpGetStudentKpis, errors.New("database unavailable"))
		reqKpis := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+studentID.String()+"/kpis", userID, cfg)
		rrKpis := httptest.NewRecorder()
		router.ServeHTTP(rrKpis, reqKpis)

		if rrKpis.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rrKpis.Code)
		}

		respKpis := httpx.DecodeJSONResponse[api.ErrorResponse](t, rrKpis)
		if respKpis.Error != api.ErrInternalError.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrInternalError.Error(), respKpis.Error)
		}

		repo.ClearError(inmemory.OpGetStudentKpis)
		repo.SetError(inmemory.OpListStudentHistory, errors.New("database unavailable"))
		reqHistory := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+studentID.String()+"/history", userID, cfg)
		rrHistory := httptest.NewRecorder()
		router.ServeHTTP(rrHistory, reqHistory)

		if rrHistory.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rrHistory.Code)
		}

		respHistory := httpx.DecodeJSONResponse[api.ErrorResponse](t, rrHistory)
		if respHistory.Error != api.ErrInternalError.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrInternalError.Error(), respHistory.Error)
		}
	})
}
