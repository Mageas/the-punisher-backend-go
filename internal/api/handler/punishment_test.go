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

type punishmentResponse struct {
	ID                 uuid.UUID  `json:"id"`
	StudentID          uuid.UUID  `json:"student_id"`
	StudentFirstName   string     `json:"student_first_name"`
	StudentLastName    string     `json:"student_last_name"`
	PunishmentTypeID   uuid.UUID  `json:"punishment_type_id"`
	PunishmentTypeName string     `json:"punishment_type_name"`
	TriggeringRuleID   *uuid.UUID `json:"triggering_rule_id"`
	TriggeringRuleName *string    `json:"triggering_rule_name"`
	Automated          bool       `json:"automated"`
	DueAt              time.Time  `json:"due_at"`
	ResolvedAt         *time.Time `json:"resolved_at"`
}

type paginatedPunishmentResponse struct {
	Page       int                  `json:"page"`
	TotalCount int64                `json:"total_count"`
	Data       []punishmentResponse `json:"data"`
}

func TestPunishmentHandlerCRUDSuccess(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newPunishmentRouter(repo, cfg)
	userID := uuid.New()
	studentID := uuid.New()
	punishmentTypeID := uuid.New()

	repo.SeedStudent(repository.Student{
		ID:        studentID,
		UserID:    userID,
		FirstName: "Jean",
		LastName:  "Dupont",
	})
	repo.SeedPunishmentType(repository.PunishmentType{
		ID:     punishmentTypeID,
		UserID: userID,
		Name:   "Heure de colle",
	})

	dueAt := time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339)
	createReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/punishments/", map[string]any{
		"student_id":         studentID.String(),
		"punishment_type_id": punishmentTypeID.String(),
		"due_at":             dueAt,
	}, userID, cfg)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)

	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createRR.Code)
	}

	created := httpx.DecodeJSONResponse[punishmentResponse](t, createRR)
	if created.ID == uuid.Nil {
		t.Fatal("expected created punishment id")
	}
	if created.TriggeringRuleID != nil {
		t.Fatalf("expected no triggering rule for manual creation, got %+v", created.TriggeringRuleID)
	}
	if created.TriggeringRuleName != nil {
		t.Fatalf("expected no triggering rule name for manual creation, got %+v", created.TriggeringRuleName)
	}
	if created.Automated {
		t.Fatalf("expected automated=false for manual creation, got %+v", created)
	}
	if created.ResolvedAt != nil {
		t.Fatalf("expected unresolved punishment, got %+v", created.ResolvedAt)
	}
	if created.StudentFirstName != "Jean" || created.StudentLastName != "Dupont" || created.PunishmentTypeName != "Heure de colle" {
		t.Fatalf("expected enriched create payload, got %+v", created)
	}

	listReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/punishments/?state=pending", userID, cfg)
	listRR := httptest.NewRecorder()
	router.ServeHTTP(listRR, listReq)

	if listRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listRR.Code)
	}

	listResp := httpx.DecodeJSONResponse[paginatedPunishmentResponse](t, listRR)
	if listResp.TotalCount != 1 || len(listResp.Data) != 1 {
		t.Fatalf("unexpected list response: %+v", listResp)
	}
	if listResp.Data[0].Automated {
		t.Fatalf("expected automated=false for manual punishment in list, got %+v", listResp.Data[0])
	}
	if listResp.Data[0].StudentFirstName != "Jean" || listResp.Data[0].StudentLastName != "Dupont" || listResp.Data[0].PunishmentTypeName != "Heure de colle" {
		t.Fatalf("expected enriched list payload, got %+v", listResp.Data[0])
	}

	getReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/punishments/"+created.ID.String(), userID, cfg)
	getRR := httptest.NewRecorder()
	router.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getRR.Code)
	}

	getResp := httpx.DecodeJSONResponse[punishmentResponse](t, getRR)
	if getResp.ID != created.ID {
		t.Fatalf("expected punishment id %s, got %s", created.ID, getResp.ID)
	}
	if getResp.Automated {
		t.Fatalf("expected automated=false for manual punishment on get, got %+v", getResp)
	}
	if getResp.StudentFirstName != "Jean" || getResp.StudentLastName != "Dupont" || getResp.PunishmentTypeName != "Heure de colle" {
		t.Fatalf("expected enriched get payload, got %+v", getResp)
	}

	resolveReq := handlertest.NewAuthorizedRequest(t, http.MethodPost, "/v1/punishments/"+created.ID.String()+"/resolve", userID, cfg)
	resolveRR := httptest.NewRecorder()
	router.ServeHTTP(resolveRR, resolveReq)

	if resolveRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resolveRR.Code)
	}

	resolved := httpx.DecodeJSONResponse[punishmentResponse](t, resolveRR)
	if resolved.ResolvedAt == nil {
		t.Fatal("expected resolved_at to be set")
	}
	if resolved.Automated {
		t.Fatalf("expected automated=false after resolving manual punishment, got %+v", resolved)
	}
	if resolved.StudentFirstName != "Jean" || resolved.StudentLastName != "Dupont" || resolved.PunishmentTypeName != "Heure de colle" {
		t.Fatalf("expected enriched resolve payload, got %+v", resolved)
	}

	listResolvedReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/punishments/?state=resolved", userID, cfg)
	listResolvedRR := httptest.NewRecorder()
	router.ServeHTTP(listResolvedRR, listResolvedReq)

	if listResolvedRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listResolvedRR.Code)
	}

	listResolvedResp := httpx.DecodeJSONResponse[paginatedPunishmentResponse](t, listResolvedRR)
	if listResolvedResp.TotalCount != 1 || len(listResolvedResp.Data) != 1 {
		t.Fatalf("unexpected resolved list response: %+v", listResolvedResp)
	}

	listByStudentReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+studentID.String()+"/punishments?state=resolved", userID, cfg)
	listByStudentRR := httptest.NewRecorder()
	router.ServeHTTP(listByStudentRR, listByStudentReq)

	if listByStudentRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listByStudentRR.Code)
	}

	listByStudentResp := httpx.DecodeJSONResponse[paginatedPunishmentResponse](t, listByStudentRR)
	if listByStudentResp.TotalCount != 1 || len(listByStudentResp.Data) != 1 {
		t.Fatalf("unexpected by-student list response: %+v", listByStudentResp)
	}
	if listByStudentResp.Data[0].StudentFirstName != "Jean" || listByStudentResp.Data[0].StudentLastName != "Dupont" || listByStudentResp.Data[0].PunishmentTypeName != "Heure de colle" {
		t.Fatalf("expected enriched by-student payload, got %+v", listByStudentResp.Data[0])
	}

	deleteReq := handlertest.NewAuthorizedRequest(t, http.MethodDelete, "/v1/punishments/"+created.ID.String(), userID, cfg)
	deleteRR := httptest.NewRecorder()
	router.ServeHTTP(deleteRR, deleteReq)

	if deleteRR.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, deleteRR.Code)
	}

	getDeletedReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/punishments/"+created.ID.String(), userID, cfg)
	getDeletedRR := httptest.NewRecorder()
	router.ServeHTTP(getDeletedRR, getDeletedReq)

	if getDeletedRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, getDeletedRR.Code)
	}

	getDeletedResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, getDeletedRR)
	if getDeletedResp.Error != api.ErrPunishmentNotFound.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrPunishmentNotFound.Error(), getDeletedResp.Error)
	}
}

func TestPunishmentHandlerValidationsAndDecodeErrors(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newPunishmentRouter(repo, cfg)
	userID := uuid.New()

	t.Run("create_validations", func(t *testing.T) {
		req := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/punishments/", map[string]any{}, userID, cfg)
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
		shared.AssertHasErrorDetail(t, resp.ErrorDetails, "punishment_type_id", api.KeyValidationFieldRequired)
		shared.AssertHasErrorDetail(t, resp.ErrorDetails, "due_at", api.KeyValidationFieldRequired)
	})

	t.Run("decode_unknown_field", func(t *testing.T) {
		req := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/punishments/", map[string]any{
			"student_id":         uuid.New().String(),
			"punishment_type_id": uuid.New().String(),
			"due_at":             time.Now().UTC().Format(time.RFC3339),
			"unknown":            "x",
		}, userID, cfg)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrInvalidRequestBody.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrInvalidRequestBody.Error(), resp.Error)
		}

		shared.AssertHasErrorDetail(t, resp.ErrorDetails, "unknown", api.KeyValidationUnknownField)
	})

	t.Run("decode_malformed_parameter", func(t *testing.T) {
		req := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/punishments/", map[string]any{
			"student_id":         uuid.New().String(),
			"punishment_type_id": uuid.New().String(),
			"due_at":             123,
		}, userID, cfg)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrMalformedParameter.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrMalformedParameter.Error(), resp.Error)
		}

		shared.AssertHasErrorDetail(t, resp.ErrorDetails, "due_at", "validation_malformed_parameter:expected_string")
	})

	t.Run("invalid_due_at_format", func(t *testing.T) {
		req := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/punishments/", map[string]any{
			"student_id":         uuid.New().String(),
			"punishment_type_id": uuid.New().String(),
			"due_at":             "not-a-date",
		}, userID, cfg)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrInvalidRequestBody.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrInvalidRequestBody.Error(), resp.Error)
		}
		shared.AssertHasErrorDetail(t, resp.ErrorDetails, "due_at", "validation_malformed_parameter:expected_rfc3339_datetime")
	})

	t.Run("malformed_ids_and_query_param", func(t *testing.T) {
		getReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/punishments/not-a-uuid", userID, cfg)
		getRR := httptest.NewRecorder()
		router.ServeHTTP(getRR, getReq)
		if getRR.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, getRR.Code)
		}

		resolveReq := handlertest.NewAuthorizedRequest(t, http.MethodPost, "/v1/punishments/not-a-uuid/resolve", userID, cfg)
		resolveRR := httptest.NewRecorder()
		router.ServeHTTP(resolveRR, resolveReq)
		if resolveRR.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resolveRR.Code)
		}

		listBadStateReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/punishments/?state=bad", userID, cfg)
		listBadStateRR := httptest.NewRecorder()
		router.ServeHTTP(listBadStateRR, listBadStateReq)
		if listBadStateRR.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, listBadStateRR.Code)
		}
		listBadStateResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, listBadStateRR)
		shared.AssertHasErrorDetail(t, listBadStateResp.ErrorDetails, "state", "validation_malformed_parameter:expected_resolved_or_pending")

		listByStudentBadIDReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/not-a-uuid/punishments", userID, cfg)
		listByStudentBadIDRR := httptest.NewRecorder()
		router.ServeHTTP(listByStudentBadIDRR, listByStudentBadIDReq)
		if listByStudentBadIDRR.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, listByStudentBadIDRR.Code)
		}
	})
}

func TestPunishmentHandlerListSearch(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newPunishmentRouter(repo, cfg)
	userID := uuid.New()
	studentMatchID := uuid.New()
	studentOtherID := uuid.New()
	punishmentTypeID := uuid.New()
	now := time.Now()

	repo.SeedStudent(repository.Student{ID: studentMatchID, UserID: userID, FirstName: "Jean", LastName: "Dupont"})
	repo.SeedStudent(repository.Student{ID: studentOtherID, UserID: userID, FirstName: "Lucas", LastName: "Martin"})
	repo.SeedPunishmentType(repository.PunishmentType{ID: punishmentTypeID, UserID: userID, Name: "Heure de colle"})

	usedAt := now.Add(3 * time.Minute)
	repo.SeedPunishment(repository.Punishment{ID: uuid.New(), UserID: userID, StudentID: studentMatchID, PunishmentTypeID: punishmentTypeID, CreatedAt: now.Add(1 * time.Minute), DueAt: now.Add(2 * time.Hour)})
	repo.SeedPunishment(repository.Punishment{ID: uuid.New(), UserID: userID, StudentID: studentMatchID, PunishmentTypeID: punishmentTypeID, CreatedAt: now.Add(2 * time.Minute), DueAt: now.Add(2 * time.Hour), ResolvedAt: &usedAt})
	repo.SeedPunishment(repository.Punishment{ID: uuid.New(), UserID: userID, StudentID: studentOtherID, PunishmentTypeID: punishmentTypeID, CreatedAt: now.Add(4 * time.Minute), DueAt: now.Add(2 * time.Hour)})

	searchReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/punishments/?search=%20%20jean%20%20dupont%20%20", userID, cfg)
	searchRR := httptest.NewRecorder()
	router.ServeHTTP(searchRR, searchReq)

	if searchRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, searchRR.Code)
	}

	searchResp := httpx.DecodeJSONResponse[paginatedPunishmentResponse](t, searchRR)
	if searchResp.TotalCount != 2 || len(searchResp.Data) != 2 {
		t.Fatalf("unexpected search response: %+v", searchResp)
	}

	stateSearchReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/punishments/?state=pending&search=jean%20dupont", userID, cfg)
	stateSearchRR := httptest.NewRecorder()
	router.ServeHTTP(stateSearchRR, stateSearchReq)

	if stateSearchRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, stateSearchRR.Code)
	}

	stateSearchResp := httpx.DecodeJSONResponse[paginatedPunishmentResponse](t, stateSearchRR)
	if stateSearchResp.TotalCount != 1 || len(stateSearchResp.Data) != 1 {
		t.Fatalf("unexpected state+search response: %+v", stateSearchResp)
	}
	if stateSearchResp.Data[0].StudentFirstName != "Jean" || stateSearchResp.Data[0].StudentLastName != "Dupont" {
		t.Fatalf("expected Jean Dupont result, got %+v", stateSearchResp.Data[0])
	}
}

func TestPunishmentHandlerBusinessAndInternalErrors(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newPunishmentRouter(repo, cfg)
	userID := uuid.New()
	studentID := uuid.New()
	punishmentTypeID := uuid.New()

	repo.SeedStudent(repository.Student{
		ID:        studentID,
		UserID:    userID,
		FirstName: "Jean",
		LastName:  "Dupont",
	})
	repo.SeedPunishmentType(repository.PunishmentType{
		ID:     punishmentTypeID,
		UserID: userID,
		Name:   "Heure de colle",
	})

	dueAt := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	createMissingStudentReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/punishments/", map[string]any{
		"student_id":         uuid.New().String(),
		"punishment_type_id": punishmentTypeID.String(),
		"due_at":             dueAt,
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

	createMissingTypeReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/punishments/", map[string]any{
		"student_id":         studentID.String(),
		"punishment_type_id": uuid.New().String(),
		"due_at":             dueAt,
	}, userID, cfg)
	createMissingTypeRR := httptest.NewRecorder()
	router.ServeHTTP(createMissingTypeRR, createMissingTypeReq)

	if createMissingTypeRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, createMissingTypeRR.Code)
	}

	createMissingTypeResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, createMissingTypeRR)
	if createMissingTypeResp.Error != api.ErrPunishmentTypeNotFound.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrPunishmentTypeNotFound.Error(), createMissingTypeResp.Error)
	}

	createReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/punishments/", map[string]any{
		"student_id":         studentID.String(),
		"punishment_type_id": punishmentTypeID.String(),
		"due_at":             dueAt,
	}, userID, cfg)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createRR.Code)
	}
	created := httpx.DecodeJSONResponse[punishmentResponse](t, createRR)

	resolveReq := handlertest.NewAuthorizedRequest(t, http.MethodPost, "/v1/punishments/"+created.ID.String()+"/resolve", userID, cfg)
	resolveRR := httptest.NewRecorder()
	router.ServeHTTP(resolveRR, resolveReq)
	if resolveRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resolveRR.Code)
	}

	resolveTwiceReq := handlertest.NewAuthorizedRequest(t, http.MethodPost, "/v1/punishments/"+created.ID.String()+"/resolve", userID, cfg)
	resolveTwiceRR := httptest.NewRecorder()
	router.ServeHTTP(resolveTwiceRR, resolveTwiceReq)

	if resolveTwiceRR.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, resolveTwiceRR.Code)
	}

	resolveTwiceResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, resolveTwiceRR)
	if resolveTwiceResp.Error != api.ErrPunishmentAlreadyResolved.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrPunishmentAlreadyResolved.Error(), resolveTwiceResp.Error)
	}

	listByMissingStudentReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+uuid.New().String()+"/punishments", userID, cfg)
	listByMissingStudentRR := httptest.NewRecorder()
	router.ServeHTTP(listByMissingStudentRR, listByMissingStudentReq)

	if listByMissingStudentRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, listByMissingStudentRR.Code)
	}

	listByMissingStudentResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, listByMissingStudentRR)
	if listByMissingStudentResp.Error != api.ErrStudentNotFound.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrStudentNotFound.Error(), listByMissingStudentResp.Error)
	}

	repo.SetError(inmemory.OpCreatePunishment, errors.New("database unavailable"))
	createInternalReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/punishments/", map[string]any{
		"student_id":         studentID.String(),
		"punishment_type_id": punishmentTypeID.String(),
		"due_at":             dueAt,
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

	repo.ClearError(inmemory.OpCreatePunishment)
	repo.SetError(inmemory.OpListPunishmentsByUser, errors.New("database unavailable"))

	listReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/punishments/", userID, cfg)
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

func newPunishmentRouter(repo *inmemory.Repository, cfg config.JWTConfig) http.Handler {
	svc := service.NewPunishmentService(repo)
	h := handler.NewPunishmentHandler(svc)

	r := chi.NewRouter()
	r.Use(platformauth.AuthMiddleware(cfg.AccessSecret, cfg.Issuer, cfg.Audience))

	r.Route("/v1/punishments", func(r chi.Router) {
		r.Post("/", h.CreatePunishment)
		r.Get("/", h.ListPunishments)
		r.Get("/{id}", h.GetPunishment)
		r.Post("/{id}/resolve", h.ResolvePunishment)
		r.Delete("/{id}", h.DeletePunishment)
	})

	r.Route("/v1/students", func(r chi.Router) {
		r.Get("/{id}/punishments", h.ListPunishmentsByStudent)
	})

	return r
}
