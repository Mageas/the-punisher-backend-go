package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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

type ruleResponse struct {
	ID                          uuid.UUID `json:"id"`
	Name                        string    `json:"name"`
	ResultingPunishmentTypeID   uuid.UUID `json:"resulting_punishment_type_id"`
	ResultingPunishmentTypeName string    `json:"resulting_punishment_type_name"`
	PenaltyTypeID               uuid.UUID `json:"penalty_type_id"`
	PenaltyTypeName             string    `json:"penalty_type_name"`
	Threshold                   int32     `json:"threshold"`
	DueAtAfterDays              int32     `json:"due_at_after_days"`
	Mode                        string    `json:"mode"`
	IsActive                    bool      `json:"is_active"`
}

type paginatedRuleResponse struct {
	Page       int            `json:"page"`
	TotalCount int64          `json:"total_count"`
	Data       []ruleResponse `json:"data"`
}

func TestRuleHandlerCRUDSuccess(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newRuleRouter(repo, cfg)
	userID := uuid.New()

	penaltyTypeID := uuid.New()
	punishmentTypeID := uuid.New()
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

	createReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/rules/", map[string]any{
		"name":                         "3 retards => colle",
		"resulting_punishment_type_id": punishmentTypeID.String(),
		"penalty_type_id":              penaltyTypeID.String(),
		"threshold":                    3,
		"due_at_after_days":            2,
		"mode":                         "at",
		"is_active":                    true,
	}, userID, cfg)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)

	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createRR.Code)
	}

	created := httpx.DecodeJSONResponse[ruleResponse](t, createRR)
	if created.ID == uuid.Nil {
		t.Fatal("expected created rule id")
	}
	if created.Mode != "at" || !created.IsActive {
		t.Fatalf("unexpected rule payload: %+v", created)
	}
	if created.PenaltyTypeName != "Retard" {
		t.Fatalf("expected penalty_type_name %q, got %q", "Retard", created.PenaltyTypeName)
	}
	if created.ResultingPunishmentTypeName != "Heure de colle" {
		t.Fatalf("expected resulting_punishment_type_name %q, got %q", "Heure de colle", created.ResultingPunishmentTypeName)
	}

	listReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/rules/", userID, cfg)
	listRR := httptest.NewRecorder()
	router.ServeHTTP(listRR, listReq)

	if listRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listRR.Code)
	}

	listResp := httpx.DecodeJSONResponse[paginatedRuleResponse](t, listRR)
	if listResp.TotalCount != 1 || len(listResp.Data) != 1 {
		t.Fatalf("unexpected list response: %+v", listResp)
	}
	if listResp.Data[0].ID != created.ID {
		t.Fatalf("expected listed id %s, got %s", created.ID, listResp.Data[0].ID)
	}
	if listResp.Data[0].PenaltyTypeName != "Retard" || listResp.Data[0].ResultingPunishmentTypeName != "Heure de colle" {
		t.Fatalf("expected enriched list payload, got %+v", listResp.Data[0])
	}

	getReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/rules/"+created.ID.String(), userID, cfg)
	getRR := httptest.NewRecorder()
	router.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getRR.Code)
	}

	getResp := httpx.DecodeJSONResponse[ruleResponse](t, getRR)
	if getResp.ID != created.ID {
		t.Fatalf("expected rule id %s, got %s", created.ID, getResp.ID)
	}
	if getResp.PenaltyTypeName != "Retard" || getResp.ResultingPunishmentTypeName != "Heure de colle" {
		t.Fatalf("expected enriched get payload, got %+v", getResp)
	}

	updateReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPut, "/v1/rules/"+created.ID.String(), map[string]any{
		"threshold": 4,
		"mode":      "every",
		"is_active": false,
	}, userID, cfg)
	updateRR := httptest.NewRecorder()
	router.ServeHTTP(updateRR, updateReq)

	if updateRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, updateRR.Code)
	}

	updated := httpx.DecodeJSONResponse[ruleResponse](t, updateRR)
	if updated.Threshold != 4 || updated.Mode != "every" || updated.IsActive {
		t.Fatalf("unexpected updated payload: %+v", updated)
	}
	if updated.PenaltyTypeName != "Retard" || updated.ResultingPunishmentTypeName != "Heure de colle" {
		t.Fatalf("expected enriched update payload, got %+v", updated)
	}

	deleteReq := handlertest.NewAuthorizedRequest(t, http.MethodDelete, "/v1/rules/"+created.ID.String(), userID, cfg)
	deleteRR := httptest.NewRecorder()
	router.ServeHTTP(deleteRR, deleteReq)

	if deleteRR.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, deleteRR.Code)
	}

	getDeletedReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/rules/"+created.ID.String(), userID, cfg)
	getDeletedRR := httptest.NewRecorder()
	router.ServeHTTP(getDeletedRR, getDeletedReq)

	if getDeletedRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, getDeletedRR.Code)
	}

	getDeletedResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, getDeletedRR)
	if getDeletedResp.Error != api.ErrRuleNotFound.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrRuleNotFound.Error(), getDeletedResp.Error)
	}
}

func TestRuleHandlerDTOValidations(t *testing.T) {
	t.Run("create_validations", func(t *testing.T) {
		repo := inmemory.NewRepository()
		cfg := shared.TestJWTConfig()
		router := newRuleRouter(repo, cfg)
		userID := uuid.New()

		tests := []struct {
			name          string
			payload       map[string]any
			expectedField string
			expectedError string
		}{
			{
				name:          "name_required",
				payload:       map[string]any{},
				expectedField: "name",
				expectedError: api.KeyValidationFieldRequired,
			},
			{
				name: "threshold_min",
				payload: map[string]any{
					"name":                         "Rule",
					"resulting_punishment_type_id": uuid.New().String(),
					"penalty_type_id":              uuid.New().String(),
					"threshold":                    0,
					"mode":                         "at",
				},
				expectedField: "threshold",
				expectedError: api.KeyValidationFieldRequired,
			},
			{
				name: "mode_invalid",
				payload: map[string]any{
					"name":                         "Rule",
					"resulting_punishment_type_id": uuid.New().String(),
					"penalty_type_id":              uuid.New().String(),
					"threshold":                    1,
					"mode":                         "invalid",
				},
				expectedField: "mode",
				expectedError: "validation_error:oneof",
			},
			{
				name: "due_at_after_days_min",
				payload: map[string]any{
					"name":                         "Rule",
					"resulting_punishment_type_id": uuid.New().String(),
					"penalty_type_id":              uuid.New().String(),
					"threshold":                    1,
					"due_at_after_days":            -1,
					"mode":                         "at",
				},
				expectedField: "due_at_after_days",
				expectedError: "validation_min_length:0",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				req := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/rules/", tc.payload, userID, cfg)
				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				if rr.Code != http.StatusBadRequest {
					t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
				}

				resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
				if resp.Error != api.ErrValidationFailed.Error() {
					t.Fatalf("expected error %q, got %q", api.ErrValidationFailed.Error(), resp.Error)
				}

				shared.AssertHasErrorDetail(t, resp.ErrorDetails, tc.expectedField, tc.expectedError)
			})
		}
	})

	t.Run("update_validations", func(t *testing.T) {
		repo := inmemory.NewRepository()
		cfg := shared.TestJWTConfig()
		router := newRuleRouter(repo, cfg)
		userID := uuid.New()
		ruleID := uuid.New()

		repo.SeedRule(repository.Rule{
			ID:                        ruleID,
			UserID:                    userID,
			Name:                      "Rule",
			ResultingPunishmentTypeID: uuid.New(),
			PenaltyTypeID:             uuid.New(),
			Threshold:                 2,
			DueAtAfterDays:            1,
			Mode:                      "at",
			IsActive:                  true,
		})

		tests := []struct {
			name          string
			payload       map[string]any
			expectedCode  int
			expectedField string
			expectedError string
		}{
			{
				name:         "empty_object_is_valid_with_omitempty",
				payload:      map[string]any{},
				expectedCode: http.StatusOK,
			},
			{
				name: "name_min_length",
				payload: map[string]any{
					"name": "A",
				},
				expectedCode:  http.StatusBadRequest,
				expectedField: "name",
				expectedError: "validation_min_length:2",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				req := handlertest.NewAuthorizedJSONRequest(t, http.MethodPut, "/v1/rules/"+ruleID.String(), tc.payload, userID, cfg)
				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				if rr.Code != tc.expectedCode {
					t.Fatalf("expected status %d, got %d", tc.expectedCode, rr.Code)
				}

				if tc.expectedCode == http.StatusBadRequest {
					resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
					if resp.Error != api.ErrValidationFailed.Error() {
						t.Fatalf("expected error %q, got %q", api.ErrValidationFailed.Error(), resp.Error)
					}

					shared.AssertHasErrorDetail(t, resp.ErrorDetails, tc.expectedField, tc.expectedError)
				}
			})
		}
	})
}

func TestRuleHandlerDecodeAndIDErrors(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newRuleRouter(repo, cfg)
	userID := uuid.New()

	createUnknownReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/rules/", map[string]any{
		"name":                         "Rule",
		"resulting_punishment_type_id": uuid.New().String(),
		"penalty_type_id":              uuid.New().String(),
		"threshold":                    1,
		"mode":                         "at",
		"unknown":                      "x",
	}, userID, cfg)
	createUnknownRR := httptest.NewRecorder()
	router.ServeHTTP(createUnknownRR, createUnknownReq)

	if createUnknownRR.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, createUnknownRR.Code)
	}

	createUnknownResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, createUnknownRR)
	if createUnknownResp.Error != api.ErrInvalidRequestBody.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrInvalidRequestBody.Error(), createUnknownResp.Error)
	}
	shared.AssertHasErrorDetail(t, createUnknownResp.ErrorDetails, "unknown", api.KeyValidationUnknownField)

	createMalformedReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/rules/", map[string]any{
		"name":                         "Rule",
		"resulting_punishment_type_id": uuid.New().String(),
		"penalty_type_id":              uuid.New().String(),
		"threshold":                    "invalid",
		"mode":                         "at",
	}, userID, cfg)
	createMalformedRR := httptest.NewRecorder()
	router.ServeHTTP(createMalformedRR, createMalformedReq)

	if createMalformedRR.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, createMalformedRR.Code)
	}

	createMalformedResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, createMalformedRR)
	if createMalformedResp.Error != api.ErrMalformedParameter.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrMalformedParameter.Error(), createMalformedResp.Error)
	}
	shared.AssertHasErrorDetail(t, createMalformedResp.ErrorDetails, "threshold", "validation_malformed_parameter:expected_int32")

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete}
	for _, method := range methods {
		var req *http.Request
		if method == http.MethodPut {
			req = handlertest.NewAuthorizedJSONRequest(t, method, "/v1/rules/not-a-uuid", map[string]any{}, userID, cfg)
		} else {
			req = handlertest.NewAuthorizedRequest(t, method, "/v1/rules/not-a-uuid", userID, cfg)
		}

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrMalformedParameter.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrMalformedParameter.Error(), resp.Error)
		}
	}
}

func TestRuleHandlerBusinessAndInternalErrors(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newRuleRouter(repo, cfg)
	userID := uuid.New()
	missingRuleID := uuid.New()

	getReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/rules/"+missingRuleID.String(), userID, cfg)
	getRR := httptest.NewRecorder()
	router.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, getRR.Code)
	}

	getResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, getRR)
	if getResp.Error != api.ErrRuleNotFound.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrRuleNotFound.Error(), getResp.Error)
	}

	createMissingPunishmentTypeReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/rules/", map[string]any{
		"name":                         "Rule",
		"resulting_punishment_type_id": uuid.New().String(),
		"penalty_type_id":              uuid.New().String(),
		"threshold":                    1,
		"mode":                         "at",
	}, userID, cfg)
	createMissingPunishmentTypeRR := httptest.NewRecorder()
	router.ServeHTTP(createMissingPunishmentTypeRR, createMissingPunishmentTypeReq)

	if createMissingPunishmentTypeRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, createMissingPunishmentTypeRR.Code)
	}

	createMissingPunishmentTypeResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, createMissingPunishmentTypeRR)
	if createMissingPunishmentTypeResp.Error != api.ErrPunishmentTypeNotFound.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrPunishmentTypeNotFound.Error(), createMissingPunishmentTypeResp.Error)
	}

	penaltyTypeID := uuid.New()
	punishmentTypeID := uuid.New()
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

	repo.SetError(inmemory.OpCreateRule, errors.New("database unavailable"))
	createReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/rules/", map[string]any{
		"name":                         "Rule",
		"resulting_punishment_type_id": punishmentTypeID.String(),
		"penalty_type_id":              penaltyTypeID.String(),
		"threshold":                    1,
		"mode":                         "at",
	}, userID, cfg)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)

	if createRR.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, createRR.Code)
	}

	createResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, createRR)
	if createResp.Error != api.ErrInternalError.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrInternalError.Error(), createResp.Error)
	}

	repo.ClearError(inmemory.OpCreateRule)
	repo.SetError(inmemory.OpListRulesByUser, errors.New("database unavailable"))

	listReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/rules/", userID, cfg)
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

func newRuleRouter(repo *inmemory.Repository, cfg config.JWTConfig) http.Handler {
	svc := service.NewRuleService(repo)
	h := handler.NewRuleHandler(svc)

	r := chi.NewRouter()
	r.Use(platformauth.AuthMiddleware(cfg.AccessSecret, cfg.Issuer, cfg.Audience))
	r.Route("/v1/rules", func(r chi.Router) {
		r.Post("/", h.CreateRule)
		r.Get("/", h.ListRules)
		r.Get("/{rule_id}", h.GetRule)
		r.Put("/{rule_id}", h.UpdateRule)
		r.Delete("/{rule_id}", h.DeleteRule)
	})

	return r
}
