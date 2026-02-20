package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

type penaltyResponse struct {
	ID               uuid.UUID `json:"id"`
	StudentID        uuid.UUID `json:"student_id"`
	StudentFirstName string    `json:"student_first_name"`
	StudentLastName  string    `json:"student_last_name"`
	PenaltyTypeID    uuid.UUID `json:"penalty_type_id"`
	PenaltyTypeName  string    `json:"penalty_type_name"`
}

type paginatedPenaltyResponse struct {
	Page       int               `json:"page"`
	TotalCount int64             `json:"total_count"`
	Data       []penaltyResponse `json:"data"`
}

func TestPenaltyHandlerCRUDSuccess(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newPenaltyRouter(repo, cfg)
	userID := uuid.New()
	studentID := uuid.New()
	penaltyTypeID := uuid.New()

	repo.SeedStudent(repository.Student{
		ID:        studentID,
		UserID:    userID,
		FirstName: "Jean",
		LastName:  "Dupont",
	})
	repo.SeedPenaltyType(repository.PenaltyType{
		ID:     penaltyTypeID,
		UserID: userID,
		Name:   "Retard",
	})

	createReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/penalties/", map[string]any{
		"student_id":      studentID.String(),
		"penalty_type_id": penaltyTypeID.String(),
	}, userID, cfg)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)

	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createRR.Code)
	}

	created := httpx.DecodeJSONResponse[penaltyResponse](t, createRR)
	if created.ID == uuid.Nil {
		t.Fatal("expected created penalty id")
	}
	if created.StudentFirstName != "Jean" || created.StudentLastName != "Dupont" || created.PenaltyTypeName != "Retard" {
		t.Fatalf("expected enriched create payload, got %+v", created)
	}

	listReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/penalties/", userID, cfg)
	listRR := httptest.NewRecorder()
	router.ServeHTTP(listRR, listReq)

	if listRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listRR.Code)
	}

	listResp := httpx.DecodeJSONResponse[paginatedPenaltyResponse](t, listRR)
	if listResp.TotalCount != 1 || len(listResp.Data) != 1 {
		t.Fatalf("unexpected list response: %+v", listResp)
	}
	if listResp.Data[0].StudentFirstName != "Jean" || listResp.Data[0].StudentLastName != "Dupont" || listResp.Data[0].PenaltyTypeName != "Retard" {
		t.Fatalf("expected enriched list payload, got %+v", listResp.Data[0])
	}

	getReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/penalties/"+created.ID.String(), userID, cfg)
	getRR := httptest.NewRecorder()
	router.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getRR.Code)
	}

	getResp := httpx.DecodeJSONResponse[penaltyResponse](t, getRR)
	if getResp.ID != created.ID {
		t.Fatalf("expected penalty id %s, got %s", created.ID, getResp.ID)
	}
	if getResp.StudentFirstName != "Jean" || getResp.StudentLastName != "Dupont" || getResp.PenaltyTypeName != "Retard" {
		t.Fatalf("expected enriched get payload, got %+v", getResp)
	}

	listByStudentReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+studentID.String()+"/penalties", userID, cfg)
	listByStudentRR := httptest.NewRecorder()
	router.ServeHTTP(listByStudentRR, listByStudentReq)

	if listByStudentRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listByStudentRR.Code)
	}

	listByStudentResp := httpx.DecodeJSONResponse[paginatedPenaltyResponse](t, listByStudentRR)
	if listByStudentResp.TotalCount != 1 || len(listByStudentResp.Data) != 1 {
		t.Fatalf("unexpected list by student response: %+v", listByStudentResp)
	}
	if listByStudentResp.Data[0].StudentFirstName != "Jean" || listByStudentResp.Data[0].StudentLastName != "Dupont" || listByStudentResp.Data[0].PenaltyTypeName != "Retard" {
		t.Fatalf("expected enriched list-by-student payload, got %+v", listByStudentResp.Data[0])
	}

	deleteReq := handlertest.NewAuthorizedRequest(t, http.MethodDelete, "/v1/penalties/"+created.ID.String(), userID, cfg)
	deleteRR := httptest.NewRecorder()
	router.ServeHTTP(deleteRR, deleteReq)

	if deleteRR.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, deleteRR.Code)
	}

	getDeletedReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/penalties/"+created.ID.String(), userID, cfg)
	getDeletedRR := httptest.NewRecorder()
	router.ServeHTTP(getDeletedRR, getDeletedReq)

	if getDeletedRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, getDeletedRR.Code)
	}

	getDeletedResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, getDeletedRR)
	if getDeletedResp.Error != api.ErrPenaltyNotFound.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrPenaltyNotFound.Error(), getDeletedResp.Error)
	}
}

func TestPenaltyHandlerValidationsAndDecodeErrors(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newPenaltyRouter(repo, cfg)
	userID := uuid.New()

	t.Run("create_validations", func(t *testing.T) {
		req := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/penalties/", map[string]any{}, userID, cfg)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrValidationFailed.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrValidationFailed.Error(), resp.Error)
		}

		shared.AssertHasErrorDetail(t, resp.ErrorDetails, "student_id", api.KeyValidationFieldRequired)
		shared.AssertHasErrorDetail(t, resp.ErrorDetails, "penalty_type_id", api.KeyValidationFieldRequired)
	})

	t.Run("decode_unknown_and_malformed", func(t *testing.T) {
		unknownReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/penalties/", map[string]any{
			"student_id":      uuid.New().String(),
			"penalty_type_id": uuid.New().String(),
			"unknown":         "x",
		}, userID, cfg)
		unknownRR := httptest.NewRecorder()
		router.ServeHTTP(unknownRR, unknownReq)

		if unknownRR.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, unknownRR.Code)
		}

		unknownResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, unknownRR)
		if unknownResp.Error != api.ErrInvalidRequestBody.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrInvalidRequestBody.Error(), unknownResp.Error)
		}
		shared.AssertHasErrorDetail(t, unknownResp.ErrorDetails, "unknown", api.KeyValidationUnknownField)

		malformedReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/penalties/", map[string]any{
			"student_id":      123,
			"penalty_type_id": uuid.New().String(),
		}, userID, cfg)
		malformedRR := httptest.NewRecorder()
		router.ServeHTTP(malformedRR, malformedReq)

		if malformedRR.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, malformedRR.Code)
		}

		malformedResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, malformedRR)
		if malformedResp.Error != api.ErrMalformedParameter.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrMalformedParameter.Error(), malformedResp.Error)
		}
		shared.AssertHasErrorDetail(t, malformedResp.ErrorDetails, "student_id", "validation_malformed_parameter:expected_string")
	})

	t.Run("malformed_ids", func(t *testing.T) {
		getReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/penalties/not-a-uuid", userID, cfg)
		getRR := httptest.NewRecorder()
		router.ServeHTTP(getRR, getReq)
		if getRR.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, getRR.Code)
		}

		deleteReq := handlertest.NewAuthorizedRequest(t, http.MethodDelete, "/v1/penalties/not-a-uuid", userID, cfg)
		deleteRR := httptest.NewRecorder()
		router.ServeHTTP(deleteRR, deleteReq)
		if deleteRR.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, deleteRR.Code)
		}

		listByStudentReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/not-a-uuid/penalties", userID, cfg)
		listByStudentRR := httptest.NewRecorder()
		router.ServeHTTP(listByStudentRR, listByStudentReq)
		if listByStudentRR.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, listByStudentRR.Code)
		}
	})
}

func TestPenaltyHandlerBusinessAndInternalErrors(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newPenaltyRouter(repo, cfg)
	userID := uuid.New()
	studentID := uuid.New()
	penaltyTypeID := uuid.New()

	repo.SeedStudent(repository.Student{
		ID:        studentID,
		UserID:    userID,
		FirstName: "Jean",
		LastName:  "Dupont",
	})
	repo.SeedPenaltyType(repository.PenaltyType{
		ID:     penaltyTypeID,
		UserID: userID,
		Name:   "Retard",
	})

	createMissingStudentReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/penalties/", map[string]any{
		"student_id":      uuid.New().String(),
		"penalty_type_id": penaltyTypeID.String(),
	}, userID, cfg)
	createMissingStudentRR := httptest.NewRecorder()
	router.ServeHTTP(createMissingStudentRR, createMissingStudentReq)

	if createMissingStudentRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, createMissingStudentRR.Code)
	}

	createMissingStudentResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, createMissingStudentRR)
	if createMissingStudentResp.Error != api.ErrStudentNotFound.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrStudentNotFound.Error(), createMissingStudentResp.Error)
	}

	createMissingTypeReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/penalties/", map[string]any{
		"student_id":      studentID.String(),
		"penalty_type_id": uuid.New().String(),
	}, userID, cfg)
	createMissingTypeRR := httptest.NewRecorder()
	router.ServeHTTP(createMissingTypeRR, createMissingTypeReq)

	if createMissingTypeRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, createMissingTypeRR.Code)
	}

	createMissingTypeResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, createMissingTypeRR)
	if createMissingTypeResp.Error != api.ErrPenaltyTypeNotFound.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrPenaltyTypeNotFound.Error(), createMissingTypeResp.Error)
	}

	listByMissingStudentReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+uuid.New().String()+"/penalties", userID, cfg)
	listByMissingStudentRR := httptest.NewRecorder()
	router.ServeHTTP(listByMissingStudentRR, listByMissingStudentReq)

	if listByMissingStudentRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, listByMissingStudentRR.Code)
	}

	listByMissingStudentResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, listByMissingStudentRR)
	if listByMissingStudentResp.Error != api.ErrStudentNotFound.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrStudentNotFound.Error(), listByMissingStudentResp.Error)
	}

	repo.SetError(inmemory.OpCreatePenalty, errors.New("database unavailable"))
	createInternalReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/penalties/", map[string]any{
		"student_id":      studentID.String(),
		"penalty_type_id": penaltyTypeID.String(),
	}, userID, cfg)
	createInternalRR := httptest.NewRecorder()
	router.ServeHTTP(createInternalRR, createInternalReq)

	if createInternalRR.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, createInternalRR.Code)
	}

	createInternalResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, createInternalRR)
	if createInternalResp.Error != api.ErrInternalError.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrInternalError.Error(), createInternalResp.Error)
	}

	repo.ClearError(inmemory.OpCreatePenalty)
	repo.SetError(inmemory.OpListPenaltiesByUser, errors.New("database unavailable"))

	listReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/penalties/", userID, cfg)
	listRR := httptest.NewRecorder()
	router.ServeHTTP(listRR, listReq)

	if listRR.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, listRR.Code)
	}

	listResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, listRR)
	if listResp.Error != api.ErrInternalError.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrInternalError.Error(), listResp.Error)
	}
}

func TestPenaltyHandlerCreateTriggersPunishmentFromRule(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newPenaltyRouter(repo, cfg)
	userID := uuid.New()
	studentID := uuid.New()
	penaltyTypeID := uuid.New()
	punishmentTypeID := uuid.New()
	ruleID := uuid.New()

	repo.SeedStudent(repository.Student{
		ID:        studentID,
		UserID:    userID,
		FirstName: "Jean",
		LastName:  "Dupont",
	})
	repo.SeedPenaltyType(repository.PenaltyType{
		ID:     penaltyTypeID,
		UserID: userID,
		Name:   "Retard",
	})
	repo.SeedPunishmentType(repository.PunishmentType{
		ID:     punishmentTypeID,
		UserID: userID,
		Name:   "Heure de colle",
	})
	repo.SeedRule(repository.Rule{
		ID:                        ruleID,
		UserID:                    userID,
		Name:                      "2 retards => colle",
		ResultingPunishmentTypeID: punishmentTypeID,
		PenaltyTypeID:             penaltyTypeID,
		Threshold:                 2,
		DueAtAfterDays:            3,
		Mode:                      "at",
		IsActive:                  true,
	})

	firstPenaltyReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/penalties/", map[string]any{
		"student_id":      studentID.String(),
		"penalty_type_id": penaltyTypeID.String(),
	}, userID, cfg)
	firstPenaltyRR := httptest.NewRecorder()
	router.ServeHTTP(firstPenaltyRR, firstPenaltyReq)

	if firstPenaltyRR.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, firstPenaltyRR.Code)
	}

	secondPenaltyReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/penalties/", map[string]any{
		"student_id":      studentID.String(),
		"penalty_type_id": penaltyTypeID.String(),
	}, userID, cfg)
	secondPenaltyRR := httptest.NewRecorder()
	router.ServeHTTP(secondPenaltyRR, secondPenaltyReq)

	if secondPenaltyRR.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, secondPenaltyRR.Code)
	}

	listPunishmentsReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/punishments/?state=pending", userID, cfg)
	listPunishmentsRR := httptest.NewRecorder()
	router.ServeHTTP(listPunishmentsRR, listPunishmentsReq)

	if listPunishmentsRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listPunishmentsRR.Code)
	}

	listPunishmentsResp := httpx.DecodeJSONResponse[paginatedPunishmentResponse](t, listPunishmentsRR)
	if listPunishmentsResp.TotalCount != 1 || len(listPunishmentsResp.Data) != 1 {
		t.Fatalf("expected one punishment created by rule, got %+v", listPunishmentsResp)
	}

	createdPunishment := listPunishmentsResp.Data[0]
	if createdPunishment.StudentID != studentID {
		t.Fatalf("expected student id %s, got %s", studentID, createdPunishment.StudentID)
	}
	if createdPunishment.PunishmentTypeID != punishmentTypeID {
		t.Fatalf("expected punishment type id %s, got %s", punishmentTypeID, createdPunishment.PunishmentTypeID)
	}
	if createdPunishment.TriggeringRuleID == nil {
		t.Fatal("expected triggering_rule_id to be set")
	}
	if *createdPunishment.TriggeringRuleID != ruleID {
		t.Fatalf("expected triggering_rule_id=%s, got %s", ruleID, *createdPunishment.TriggeringRuleID)
	}
	if createdPunishment.TriggeringRuleName == nil {
		t.Fatal("expected triggering_rule_name to be set")
	}
	if *createdPunishment.TriggeringRuleName != "2 retards => colle" {
		t.Fatalf("expected triggering_rule_name=%q, got %q", "2 retards => colle", *createdPunishment.TriggeringRuleName)
	}
	if !createdPunishment.Automated {
		t.Fatalf("expected automated=true for punishment created by rule, got %+v", createdPunishment)
	}
	if createdPunishment.StudentFirstName != "Jean" || createdPunishment.StudentLastName != "Dupont" {
		t.Fatalf("expected enriched student names, got %+v", createdPunishment)
	}
	if createdPunishment.PunishmentTypeName != "Heure de colle" {
		t.Fatalf("expected punishment_type_name=%q, got %q", "Heure de colle", createdPunishment.PunishmentTypeName)
	}

	expectedDueAt := time.Now().UTC().Add(3 * 24 * time.Hour)
	delta := createdPunishment.DueAt.Sub(expectedDueAt)
	if delta < 0 {
		delta = -delta
	}
	if delta > 2*time.Minute {
		t.Fatalf("expected due_at around %s, got %s", expectedDueAt, createdPunishment.DueAt)
	}
}

func newPenaltyRouter(repo *inmemory.Repository, cfg config.JWTConfig) http.Handler {
	penaltySvc := service.NewPenaltyService(repo)
	penaltyHandler := handler.NewPenaltyHandler(penaltySvc)

	punishmentSvc := service.NewPunishmentService(repo)
	punishmentHandler := handler.NewPunishmentHandler(punishmentSvc)

	r := chi.NewRouter()
	r.Use(platformauth.AuthMiddleware(cfg.AccessSecret, cfg.Issuer, cfg.Audience))

	r.Route("/v1/penalties", func(r chi.Router) {
		r.Post("/", penaltyHandler.CreatePenalty)
		r.Get("/", penaltyHandler.ListPenalties)
		r.Get("/{id}", penaltyHandler.GetPenalty)
		r.Delete("/{id}", penaltyHandler.DeletePenalty)
	})

	r.Route("/v1/students", func(r chi.Router) {
		r.Get("/{id}/penalties", penaltyHandler.ListPenaltiesByStudent)
	})

	// Only list endpoint is needed to assert rule-triggered punishments from penalties.
	r.Route("/v1/punishments", func(r chi.Router) {
		r.Get("/", punishmentHandler.ListPunishments)
	})

	return r
}
