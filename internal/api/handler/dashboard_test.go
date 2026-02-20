package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/api/handler"
	platformauth "github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/config"
	"github.com/mageas/the-punisher-backend/internal/repository"
	"github.com/mageas/the-punisher-backend/internal/service"
	"github.com/mageas/the-punisher-backend/internal/testutil/handlertest"
	"github.com/mageas/the-punisher-backend/internal/testutil/httpx"
	"github.com/mageas/the-punisher-backend/internal/testutil/inmemory"
	shared "github.com/mageas/the-punisher-backend/internal/testutil/shared"
)

type dashboardKpisResponse struct {
	StudentCount           int64   `json:"student_count"`
	AvailableBonusPoints   float64 `json:"available_bonus_points"`
	UnusedBonusCount       int64   `json:"unused_bonus_count"`
	PenaltyCount           int64   `json:"penalty_count"`
	PendingPunishmentCount int64   `json:"pending_punishment_count"`
}

type dashboardPenaltyResponse struct {
	StudentID uuid.UUID `json:"student_id"`
}

type dashboardBonusResponse struct {
	StudentID uuid.UUID `json:"student_id"`
}

type dashboardPunishmentResponse struct {
	StudentID          uuid.UUID  `json:"student_id"`
	TriggeringRuleID   *uuid.UUID `json:"triggering_rule_id"`
	TriggeringRuleName *string    `json:"triggering_rule_name"`
	Automated          bool       `json:"automated"`
}

type dashboardResponse struct {
	Kpis               dashboardKpisResponse         `json:"kpis"`
	RecentPenalties    []dashboardPenaltyResponse    `json:"recent_penalties"`
	RecentBonuses      []dashboardBonusResponse      `json:"recent_bonuses"`
	PendingPunishments []dashboardPunishmentResponse `json:"pending_punishments"`
}

func TestDashboardHandlerSuccess(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newDashboardRouter(repo, cfg)

	userID := uuid.New()
	studentInClassroomID := uuid.New()
	studentOutsideClassroomID := uuid.New()
	classroomID := uuid.New()
	bonusTypeID := uuid.New()
	penaltyTypeID := uuid.New()
	punishmentTypeID := uuid.New()
	ruleID := uuid.New()

	base := time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC)

	repo.SeedStudent(repository.Student{ID: studentInClassroomID, UserID: userID, FirstName: "Jean", LastName: "Dupont", CreatedAt: base.Add(-3 * time.Hour), UpdatedAt: base.Add(-3 * time.Hour)})
	repo.SeedStudent(repository.Student{ID: studentOutsideClassroomID, UserID: userID, FirstName: "Emma", LastName: "Martin", CreatedAt: base.Add(-2 * time.Hour), UpdatedAt: base.Add(-2 * time.Hour)})

	repo.SeedClassroom(repository.Classroom{ID: classroomID, UserID: userID, Name: "CM1 A", CreatedAt: base.Add(-4 * time.Hour), UpdatedAt: base.Add(-4 * time.Hour)})
	if _, err := repo.AddStudentToClassroom(t.Context(), repository.AddStudentToClassroomParams{StudentID: studentInClassroomID, ClassroomID: classroomID, UserID: userID}); err != nil {
		t.Fatalf("failed to seed student/classroom relation: %v", err)
	}

	repo.SeedBonusType(repository.BonusType{ID: bonusTypeID, UserID: userID, Name: "Participation"})
	repo.SeedPenaltyType(repository.PenaltyType{ID: penaltyTypeID, UserID: userID, Name: "Bavardage"})
	repo.SeedPunishmentType(repository.PunishmentType{ID: punishmentTypeID, UserID: userID, Name: "Retenue"})
	repo.SeedRule(repository.Rule{ID: ruleID, UserID: userID, Name: "3 bavardages => retenue", PenaltyTypeID: penaltyTypeID, ResultingPunishmentTypeID: punishmentTypeID, Mode: "every", Threshold: 3, IsActive: true, DueAtAfterDays: 7})

	repo.SeedBonus(repository.Bonus{ID: uuid.New(), UserID: userID, StudentID: studentInClassroomID, BonusTypeID: bonusTypeID, Points: 2, CreatedAt: base.Add(1 * time.Hour)})
	repo.SeedBonus(repository.Bonus{ID: uuid.New(), UserID: userID, StudentID: studentOutsideClassroomID, BonusTypeID: bonusTypeID, Points: 3, CreatedAt: base.Add(2 * time.Hour)})
	repo.SeedBonus(repository.Bonus{ID: uuid.New(), UserID: userID, StudentID: studentInClassroomID, BonusTypeID: bonusTypeID, Points: 4, CreatedAt: base.Add(3 * time.Hour), UsedAt: pgtype.Timestamptz{Time: base.Add(4 * time.Hour), Valid: true}})

	repo.SeedPenalty(repository.Penalty{ID: uuid.New(), UserID: userID, StudentID: studentInClassroomID, PenaltyTypeID: penaltyTypeID, CreatedAt: base.Add(1 * time.Hour)})
	repo.SeedPenalty(repository.Penalty{ID: uuid.New(), UserID: userID, StudentID: studentOutsideClassroomID, PenaltyTypeID: penaltyTypeID, CreatedAt: base.Add(2 * time.Hour)})

	repo.SeedPunishment(repository.Punishment{ID: uuid.New(), UserID: userID, StudentID: studentInClassroomID, PunishmentTypeID: punishmentTypeID, TriggeringRuleID: pgtype.UUID{Bytes: ruleID, Valid: true}, Automated: true, CreatedAt: base.Add(1 * time.Hour), DueAt: base.Add(24 * time.Hour)})
	repo.SeedPunishment(repository.Punishment{ID: uuid.New(), UserID: userID, StudentID: studentOutsideClassroomID, PunishmentTypeID: punishmentTypeID, CreatedAt: base.Add(2 * time.Hour), DueAt: base.Add(24 * time.Hour)})
	repo.SeedPunishment(repository.Punishment{ID: uuid.New(), UserID: userID, StudentID: studentInClassroomID, PunishmentTypeID: punishmentTypeID, CreatedAt: base.Add(3 * time.Hour), DueAt: base.Add(24 * time.Hour), ResolvedAt: pgtype.Timestamptz{Time: base.Add(5 * time.Hour), Valid: true}})

	t.Run("without_classroom_filter", func(t *testing.T) {
		req := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/dashboard/", userID, cfg)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[dashboardResponse](t, rr)
		if resp.Kpis.StudentCount != 2 {
			t.Fatalf("expected student_count %d, got %d", 2, resp.Kpis.StudentCount)
		}
		if resp.Kpis.AvailableBonusPoints != 5 {
			t.Fatalf("expected available_bonus_points %.1f, got %.1f", 5.0, resp.Kpis.AvailableBonusPoints)
		}
		if resp.Kpis.UnusedBonusCount != 2 || resp.Kpis.PenaltyCount != 2 || resp.Kpis.PendingPunishmentCount != 2 {
			t.Fatalf("unexpected kpis: %+v", resp.Kpis)
		}
		if len(resp.RecentPenalties) != 2 || len(resp.RecentBonuses) != 3 || len(resp.PendingPunishments) != 2 {
			t.Fatalf("unexpected dashboard lists: %+v", resp)
		}
		if resp.RecentPenalties[0].StudentID != studentOutsideClassroomID {
			t.Fatalf("expected penalties sorted by created_at desc, got %+v", resp.RecentPenalties)
		}
		if resp.RecentBonuses[0].StudentID != studentInClassroomID {
			t.Fatalf("expected bonuses sorted by created_at desc, got %+v", resp.RecentBonuses)
		}

		autoPunishmentCount := 0
		for _, pendingPunishment := range resp.PendingPunishments {
			if pendingPunishment.TriggeringRuleID != nil && pendingPunishment.TriggeringRuleName != nil && *pendingPunishment.TriggeringRuleName == "3 bavardages => retenue" {
				if !pendingPunishment.Automated {
					t.Fatalf("expected automated=true when triggering rule is present, got %+v", pendingPunishment)
				}
				autoPunishmentCount++
				continue
			}
			if pendingPunishment.Automated {
				t.Fatalf("expected automated=false when no triggering rule is present, got %+v", pendingPunishment)
			}
		}
		if autoPunishmentCount != 1 {
			t.Fatalf("expected one pending automated punishment, got %+v", resp.PendingPunishments)
		}
	})

	t.Run("with_classroom_filter", func(t *testing.T) {
		req := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/dashboard/?classroom_id="+classroomID.String(), userID, cfg)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[dashboardResponse](t, rr)
		if resp.Kpis.StudentCount != 1 {
			t.Fatalf("expected student_count %d, got %d", 1, resp.Kpis.StudentCount)
		}
		if resp.Kpis.AvailableBonusPoints != 2 {
			t.Fatalf("expected available_bonus_points %.1f, got %.1f", 2.0, resp.Kpis.AvailableBonusPoints)
		}
		if resp.Kpis.UnusedBonusCount != 1 || resp.Kpis.PenaltyCount != 1 || resp.Kpis.PendingPunishmentCount != 1 {
			t.Fatalf("unexpected filtered kpis: %+v", resp.Kpis)
		}
		if len(resp.RecentPenalties) != 1 || len(resp.RecentBonuses) != 2 || len(resp.PendingPunishments) != 1 {
			t.Fatalf("unexpected filtered dashboard lists: %+v", resp)
		}
		if resp.RecentPenalties[0].StudentID != studentInClassroomID || resp.RecentBonuses[0].StudentID != studentInClassroomID || resp.PendingPunishments[0].StudentID != studentInClassroomID {
			t.Fatalf("expected classroom-filtered lists to contain only classroom students, got %+v", resp)
		}
	})
}

func TestDashboardHandlerErrors(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newDashboardRouter(repo, cfg)
	userID := uuid.New()

	t.Run("malformed_classroom_id", func(t *testing.T) {
		req := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/dashboard/?classroom_id=not-a-uuid", userID, cfg)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrMalformedParameter.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrMalformedParameter.Error(), resp.Error)
		}
		shared.AssertHasErrorDetail(t, resp.ErrorDetails, "classroom_id", "validation_malformed_parameter:expected_uuid")
	})

	t.Run("classroom_not_found", func(t *testing.T) {
		req := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/dashboard/?classroom_id="+uuid.New().String(), userID, cfg)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrClassroomNotFound.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrClassroomNotFound.Error(), resp.Error)
		}
	})

	t.Run("internal_error", func(t *testing.T) {
		repo.SetError(inmemory.OpGetDashboardKpis, errors.New("database unavailable"))

		req := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/dashboard/", userID, cfg)
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

func newDashboardRouter(repo *inmemory.Repository, cfg config.JWTConfig) http.Handler {
	dashboardSvc := service.NewDashboardService(repo)
	dashboardHandler := handler.NewDashboardHandler(dashboardSvc)

	r := chi.NewRouter()
	r.Use(platformauth.AuthMiddleware(cfg.AccessSecret))
	r.Route("/v1/dashboard", func(r chi.Router) {
		r.Get("/", dashboardHandler.GetDashboard)
	})

	return r
}
