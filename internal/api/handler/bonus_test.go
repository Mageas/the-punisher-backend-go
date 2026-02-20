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

type bonusResponse struct {
	ID               uuid.UUID  `json:"id"`
	StudentID        uuid.UUID  `json:"student_id"`
	StudentFirstName string     `json:"student_first_name"`
	StudentLastName  string     `json:"student_last_name"`
	BonusTypeID      uuid.UUID  `json:"bonus_type_id"`
	BonusTypeName    string     `json:"bonus_type_name"`
	Points           float64    `json:"points"`
	UsedAt           *time.Time `json:"used_at"`
}

type paginatedBonusResponse struct {
	Page       int             `json:"page"`
	TotalCount int64           `json:"total_count"`
	Data       []bonusResponse `json:"data"`
}

func TestBonusHandlerCRUDSuccess(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newBonusRouter(repo, cfg)
	userID := uuid.New()
	studentID := uuid.New()
	bonusTypeID := uuid.New()

	repo.SeedStudent(repository.Student{
		ID:        studentID,
		UserID:    userID,
		FirstName: "Jean",
		LastName:  "Dupont",
	})
	repo.SeedBonusType(repository.BonusType{
		ID:     bonusTypeID,
		UserID: userID,
		Name:   "Participation",
	})

	createReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/bonuses/", map[string]any{
		"student_id":    studentID.String(),
		"bonus_type_id": bonusTypeID.String(),
		"points":        1.5,
	}, userID, cfg)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)

	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createRR.Code)
	}

	created := httpx.DecodeJSONResponse[bonusResponse](t, createRR)
	if created.ID == uuid.Nil {
		t.Fatal("expected created bonus id")
	}
	if created.Points != 1.5 {
		t.Fatalf("expected points %f, got %f", 1.5, created.Points)
	}
	if created.UsedAt != nil {
		t.Fatalf("expected unused bonus, got used_at=%v", created.UsedAt)
	}
	if created.StudentFirstName != "Jean" || created.StudentLastName != "Dupont" || created.BonusTypeName != "Participation" {
		t.Fatalf("expected enriched create payload, got %+v", created)
	}

	listReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/bonuses/", userID, cfg)
	listRR := httptest.NewRecorder()
	router.ServeHTTP(listRR, listReq)

	if listRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listRR.Code)
	}

	listResp := httpx.DecodeJSONResponse[paginatedBonusResponse](t, listRR)
	if listResp.TotalCount != 1 || len(listResp.Data) != 1 {
		t.Fatalf("unexpected list response: %+v", listResp)
	}
	if listResp.Data[0].StudentFirstName != "Jean" || listResp.Data[0].StudentLastName != "Dupont" || listResp.Data[0].BonusTypeName != "Participation" {
		t.Fatalf("expected enriched list payload, got %+v", listResp.Data[0])
	}

	listUnusedReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/bonuses/?state=unused", userID, cfg)
	listUnusedRR := httptest.NewRecorder()
	router.ServeHTTP(listUnusedRR, listUnusedReq)

	if listUnusedRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listUnusedRR.Code)
	}

	listUnusedResp := httpx.DecodeJSONResponse[paginatedBonusResponse](t, listUnusedRR)
	if listUnusedResp.TotalCount != 1 || len(listUnusedResp.Data) != 1 {
		t.Fatalf("unexpected unused list response: %+v", listUnusedResp)
	}

	useReq := handlertest.NewAuthorizedRequest(t, http.MethodPost, "/v1/bonuses/"+created.ID.String()+"/use", userID, cfg)
	useRR := httptest.NewRecorder()
	router.ServeHTTP(useRR, useReq)

	if useRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, useRR.Code)
	}

	used := httpx.DecodeJSONResponse[bonusResponse](t, useRR)
	if used.UsedAt == nil {
		t.Fatal("expected used_at to be set")
	}
	if used.StudentFirstName != "Jean" || used.StudentLastName != "Dupont" || used.BonusTypeName != "Participation" {
		t.Fatalf("expected enriched use payload, got %+v", used)
	}

	listUsedReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/bonuses/?state=used", userID, cfg)
	listUsedRR := httptest.NewRecorder()
	router.ServeHTTP(listUsedRR, listUsedReq)

	if listUsedRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listUsedRR.Code)
	}

	listUsedResp := httpx.DecodeJSONResponse[paginatedBonusResponse](t, listUsedRR)
	if listUsedResp.TotalCount != 1 || len(listUsedResp.Data) != 1 {
		t.Fatalf("unexpected used list response: %+v", listUsedResp)
	}

	getReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/bonuses/"+created.ID.String(), userID, cfg)
	getRR := httptest.NewRecorder()
	router.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getRR.Code)
	}

	getResp := httpx.DecodeJSONResponse[bonusResponse](t, getRR)
	if getResp.ID != created.ID {
		t.Fatalf("expected bonus id %s, got %s", created.ID, getResp.ID)
	}
	if getResp.StudentFirstName != "Jean" || getResp.StudentLastName != "Dupont" || getResp.BonusTypeName != "Participation" {
		t.Fatalf("expected enriched get payload, got %+v", getResp)
	}

	listByStudentReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+studentID.String()+"/bonuses?state=used", userID, cfg)
	listByStudentRR := httptest.NewRecorder()
	router.ServeHTTP(listByStudentRR, listByStudentReq)

	if listByStudentRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listByStudentRR.Code)
	}

	listByStudentResp := httpx.DecodeJSONResponse[paginatedBonusResponse](t, listByStudentRR)
	if listByStudentResp.TotalCount != 1 || len(listByStudentResp.Data) != 1 {
		t.Fatalf("unexpected list by student response: %+v", listByStudentResp)
	}
	if listByStudentResp.Data[0].StudentFirstName != "Jean" || listByStudentResp.Data[0].StudentLastName != "Dupont" || listByStudentResp.Data[0].BonusTypeName != "Participation" {
		t.Fatalf("expected enriched list-by-student payload, got %+v", listByStudentResp.Data[0])
	}

	deleteReq := handlertest.NewAuthorizedRequest(t, http.MethodDelete, "/v1/bonuses/"+created.ID.String(), userID, cfg)
	deleteRR := httptest.NewRecorder()
	router.ServeHTTP(deleteRR, deleteReq)

	if deleteRR.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, deleteRR.Code)
	}

	getDeletedReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/bonuses/"+created.ID.String(), userID, cfg)
	getDeletedRR := httptest.NewRecorder()
	router.ServeHTTP(getDeletedRR, getDeletedReq)

	if getDeletedRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, getDeletedRR.Code)
	}

	getDeletedResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, getDeletedRR)
	if getDeletedResp.Error != api.ErrBonusNotFound.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrBonusNotFound.Error(), getDeletedResp.Error)
	}
}

func TestBonusHandlerValidationsAndDecodeErrors(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newBonusRouter(repo, cfg)
	userID := uuid.New()

	t.Run("create_validations", func(t *testing.T) {
		req := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/bonuses/", map[string]any{
			"points": 0,
		}, userID, cfg)
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
		shared.AssertHasErrorDetail(t, resp.ErrorDetails, "bonus_type_id", api.KeyValidationFieldRequired)
		shared.AssertHasErrorDetail(t, resp.ErrorDetails, "points", api.KeyValidationFieldRequired)
	})

	t.Run("decode_unknown_field", func(t *testing.T) {
		req := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/bonuses/", map[string]any{
			"student_id":    uuid.New().String(),
			"bonus_type_id": uuid.New().String(),
			"points":        1.0,
			"unknown":       "x",
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
		req := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/bonuses/", map[string]any{
			"student_id":    uuid.New().String(),
			"bonus_type_id": uuid.New().String(),
			"points":        "invalid",
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

		shared.AssertHasErrorDetail(t, resp.ErrorDetails, "points", "validation_malformed_parameter:expected_float64")
	})

	t.Run("malformed_ids_and_query_param", func(t *testing.T) {
		getReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/bonuses/not-a-uuid", userID, cfg)
		getRR := httptest.NewRecorder()
		router.ServeHTTP(getRR, getReq)
		if getRR.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, getRR.Code)
		}

		useReq := handlertest.NewAuthorizedRequest(t, http.MethodPost, "/v1/bonuses/not-a-uuid/use", userID, cfg)
		useRR := httptest.NewRecorder()
		router.ServeHTTP(useRR, useReq)
		if useRR.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, useRR.Code)
		}

		listBadStateReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/bonuses/?state=bad", userID, cfg)
		listBadStateRR := httptest.NewRecorder()
		router.ServeHTTP(listBadStateRR, listBadStateReq)
		if listBadStateRR.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, listBadStateRR.Code)
		}
		listBadStateResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, listBadStateRR)
		shared.AssertHasErrorDetail(t, listBadStateResp.ErrorDetails, "state", "validation_malformed_parameter:expected_used_or_unused")

		listByStudentBadIDReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/not-a-uuid/bonuses", userID, cfg)
		listByStudentBadIDRR := httptest.NewRecorder()
		router.ServeHTTP(listByStudentBadIDRR, listByStudentBadIDReq)
		if listByStudentBadIDRR.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, listByStudentBadIDRR.Code)
		}
	})
}

func TestBonusHandlerListSearch(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newBonusRouter(repo, cfg)
	userID := uuid.New()
	studentMatchID := uuid.New()
	studentOtherID := uuid.New()
	bonusTypeID := uuid.New()
	now := time.Now()

	repo.SeedStudent(repository.Student{ID: studentMatchID, UserID: userID, FirstName: "Jean", LastName: "Dupont"})
	repo.SeedStudent(repository.Student{ID: studentOtherID, UserID: userID, FirstName: "Lucas", LastName: "Martin"})
	repo.SeedBonusType(repository.BonusType{ID: bonusTypeID, UserID: userID, Name: "Participation"})

	usedAt := now.Add(3 * time.Minute)
	repo.SeedBonus(repository.Bonus{ID: uuid.New(), UserID: userID, StudentID: studentMatchID, BonusTypeID: bonusTypeID, Points: 1.5, CreatedAt: now.Add(1 * time.Minute)})
	repo.SeedBonus(repository.Bonus{ID: uuid.New(), UserID: userID, StudentID: studentMatchID, BonusTypeID: bonusTypeID, Points: 2.5, CreatedAt: now.Add(2 * time.Minute), UsedAt: &usedAt})
	repo.SeedBonus(repository.Bonus{ID: uuid.New(), UserID: userID, StudentID: studentOtherID, BonusTypeID: bonusTypeID, Points: 3.5, CreatedAt: now.Add(4 * time.Minute)})

	searchReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/bonuses/?search=%20%20jean%20%20%20dupont%20%20", userID, cfg)
	searchRR := httptest.NewRecorder()
	router.ServeHTTP(searchRR, searchReq)

	if searchRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, searchRR.Code)
	}

	searchResp := httpx.DecodeJSONResponse[paginatedBonusResponse](t, searchRR)
	if searchResp.TotalCount != 2 || len(searchResp.Data) != 2 {
		t.Fatalf("unexpected search response: %+v", searchResp)
	}

	stateSearchReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/bonuses/?state=unused&search=jean%20dupont", userID, cfg)
	stateSearchRR := httptest.NewRecorder()
	router.ServeHTTP(stateSearchRR, stateSearchReq)

	if stateSearchRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, stateSearchRR.Code)
	}

	stateSearchResp := httpx.DecodeJSONResponse[paginatedBonusResponse](t, stateSearchRR)
	if stateSearchResp.TotalCount != 1 || len(stateSearchResp.Data) != 1 {
		t.Fatalf("unexpected state+search response: %+v", stateSearchResp)
	}
	if stateSearchResp.Data[0].StudentFirstName != "Jean" || stateSearchResp.Data[0].StudentLastName != "Dupont" {
		t.Fatalf("expected Jean Dupont result, got %+v", stateSearchResp.Data[0])
	}
}

func TestBonusHandlerBusinessAndInternalErrors(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newBonusRouter(repo, cfg)
	userID := uuid.New()
	studentID := uuid.New()
	bonusTypeID := uuid.New()

	repo.SeedStudent(repository.Student{
		ID:        studentID,
		UserID:    userID,
		FirstName: "Jean",
		LastName:  "Dupont",
	})
	repo.SeedBonusType(repository.BonusType{
		ID:     bonusTypeID,
		UserID: userID,
		Name:   "Participation",
	})

	createMissingStudentReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/bonuses/", map[string]any{
		"student_id":    uuid.New().String(),
		"bonus_type_id": bonusTypeID.String(),
		"points":        1.0,
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

	createMissingTypeReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/bonuses/", map[string]any{
		"student_id":    studentID.String(),
		"bonus_type_id": uuid.New().String(),
		"points":        1.0,
	}, userID, cfg)
	createMissingTypeRR := httptest.NewRecorder()
	router.ServeHTTP(createMissingTypeRR, createMissingTypeReq)

	if createMissingTypeRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, createMissingTypeRR.Code)
	}

	createMissingTypeResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, createMissingTypeRR)
	if createMissingTypeResp.Error != api.ErrBonusTypeNotFound.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrBonusTypeNotFound.Error(), createMissingTypeResp.Error)
	}

	createReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/bonuses/", map[string]any{
		"student_id":    studentID.String(),
		"bonus_type_id": bonusTypeID.String(),
		"points":        1.0,
	}, userID, cfg)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)

	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createRR.Code)
	}

	created := httpx.DecodeJSONResponse[bonusResponse](t, createRR)

	useReq := handlertest.NewAuthorizedRequest(t, http.MethodPost, "/v1/bonuses/"+created.ID.String()+"/use", userID, cfg)
	useRR := httptest.NewRecorder()
	router.ServeHTTP(useRR, useReq)
	if useRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, useRR.Code)
	}

	useTwiceReq := handlertest.NewAuthorizedRequest(t, http.MethodPost, "/v1/bonuses/"+created.ID.String()+"/use", userID, cfg)
	useTwiceRR := httptest.NewRecorder()
	router.ServeHTTP(useTwiceRR, useTwiceReq)

	if useTwiceRR.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, useTwiceRR.Code)
	}

	useTwiceResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, useTwiceRR)
	if useTwiceResp.Error != api.ErrBonusAlreadyUsed.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrBonusAlreadyUsed.Error(), useTwiceResp.Error)
	}

	listByMissingStudentReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+uuid.New().String()+"/bonuses", userID, cfg)
	listByMissingStudentRR := httptest.NewRecorder()
	router.ServeHTTP(listByMissingStudentRR, listByMissingStudentReq)

	if listByMissingStudentRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, listByMissingStudentRR.Code)
	}

	listByMissingStudentResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, listByMissingStudentRR)
	if listByMissingStudentResp.Error != api.ErrStudentNotFound.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrStudentNotFound.Error(), listByMissingStudentResp.Error)
	}

	repo.SetError(inmemory.OpCreateBonus, errors.New("database unavailable"))
	createInternalReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/bonuses/", map[string]any{
		"student_id":    studentID.String(),
		"bonus_type_id": bonusTypeID.String(),
		"points":        2.0,
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

	repo.ClearError(inmemory.OpCreateBonus)
	repo.SetError(inmemory.OpListBonusesByUser, errors.New("database unavailable"))

	listReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/bonuses/", userID, cfg)
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

func newBonusRouter(repo *inmemory.Repository, cfg config.JWTConfig) http.Handler {
	svc := service.NewBonusService(repo)
	h := handler.NewBonusHandler(svc)

	r := chi.NewRouter()
	r.Use(platformauth.AuthMiddleware(cfg.AccessSecret, cfg.Issuer, cfg.Audience))

	r.Route("/v1/bonuses", func(r chi.Router) {
		r.Post("/", h.CreateBonus)
		r.Get("/", h.ListBonuses)
		r.Get("/{bonus_id}", h.GetBonus)
		r.Post("/{bonus_id}/use", h.UseBonus)
		r.Delete("/{bonus_id}", h.DeleteBonus)
	})

	r.Route("/v1/students", func(r chi.Router) {
		r.Get("/{student_id}/bonuses", h.ListBonusesByStudent)
	})

	return r
}
