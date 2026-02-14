package shared

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/platform/config"
	"github.com/mageas/the-punisher-backend/internal/testutil/handlertest"
	"github.com/mageas/the-punisher-backend/internal/testutil/httpx"
	"github.com/mageas/the-punisher-backend/internal/testutil/inmemory"
)

type typeResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type paginatedTypeResponse struct {
	Page       int            `json:"page"`
	TotalCount int64          `json:"total_count"`
	Data       []typeResponse `json:"data"`
}

type TypeSeed struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Name   string
}

type ManagedTypeSuite struct {
	BasePath      string
	NotFoundError string
	OpCreate      string
	OpList        string
	Seed          func(repo *inmemory.Repository, seed TypeSeed)
	NewRouter     func(repo *inmemory.Repository, cfg config.JWTConfig) http.Handler
}

func RunTypeHandlerCRUDSuccess(t *testing.T, suite ManagedTypeSuite) {
	t.Helper()

	repo := inmemory.NewRepository()
	cfg := TestJWTConfig()
	router := suite.NewRouter(repo, cfg)
	userID := uuid.New()

	createReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, suite.BasePath+"/", map[string]any{
		"name": "Type Alpha",
	}, userID, cfg)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)

	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createRR.Code)
	}

	created := httpx.DecodeJSONResponse[typeResponse](t, createRR)
	if created.ID == uuid.Nil {
		t.Fatal("expected created resource id")
	}
	if created.Name != "Type Alpha" {
		t.Fatalf("expected created name %q, got %q", "Type Alpha", created.Name)
	}

	listReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, suite.BasePath+"/", userID, cfg)
	listRR := httptest.NewRecorder()
	router.ServeHTTP(listRR, listReq)

	if listRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listRR.Code)
	}

	listResp := httpx.DecodeJSONResponse[paginatedTypeResponse](t, listRR)
	if listResp.TotalCount != 1 {
		t.Fatalf("expected total_count=1, got %d", listResp.TotalCount)
	}
	if len(listResp.Data) != 1 {
		t.Fatalf("expected one item in list, got %d", len(listResp.Data))
	}
	if listResp.Data[0].ID != created.ID {
		t.Fatalf("expected listed id %s, got %s", created.ID, listResp.Data[0].ID)
	}

	getReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, suite.BasePath+"/"+created.ID.String(), userID, cfg)
	getRR := httptest.NewRecorder()
	router.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getRR.Code)
	}

	getResp := httpx.DecodeJSONResponse[typeResponse](t, getRR)
	if getResp.ID != created.ID {
		t.Fatalf("expected retrieved id %s, got %s", created.ID, getResp.ID)
	}

	updateReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPut, suite.BasePath+"/"+created.ID.String(), map[string]any{
		"name": "Type Beta",
	}, userID, cfg)
	updateRR := httptest.NewRecorder()
	router.ServeHTTP(updateRR, updateReq)

	if updateRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, updateRR.Code)
	}

	updated := httpx.DecodeJSONResponse[typeResponse](t, updateRR)
	if updated.Name != "Type Beta" {
		t.Fatalf("expected updated name %q, got %q", "Type Beta", updated.Name)
	}

	updateEmptyReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPut, suite.BasePath+"/"+created.ID.String(), map[string]any{}, userID, cfg)
	updateEmptyRR := httptest.NewRecorder()
	router.ServeHTTP(updateEmptyRR, updateEmptyReq)

	if updateEmptyRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, updateEmptyRR.Code)
	}

	updateEmptyResp := httpx.DecodeJSONResponse[typeResponse](t, updateEmptyRR)
	if updateEmptyResp.Name != "Type Beta" {
		t.Fatalf("expected name to remain %q, got %q", "Type Beta", updateEmptyResp.Name)
	}

	deleteReq := handlertest.NewAuthorizedRequest(t, http.MethodDelete, suite.BasePath+"/"+created.ID.String(), userID, cfg)
	deleteRR := httptest.NewRecorder()
	router.ServeHTTP(deleteRR, deleteReq)

	if deleteRR.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, deleteRR.Code)
	}

	getDeletedReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, suite.BasePath+"/"+created.ID.String(), userID, cfg)
	getDeletedRR := httptest.NewRecorder()
	router.ServeHTTP(getDeletedRR, getDeletedReq)

	if getDeletedRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, getDeletedRR.Code)
	}

	errResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, getDeletedRR)
	if errResp.Error != suite.NotFoundError {
		t.Fatalf("expected error %q, got %q", suite.NotFoundError, errResp.Error)
	}
}

func RunTypeHandlerDTOValidations(t *testing.T, suite ManagedTypeSuite) {
	t.Helper()

	t.Run("create_validations", func(t *testing.T) {
		repo := inmemory.NewRepository()
		cfg := TestJWTConfig()
		router := suite.NewRouter(repo, cfg)
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
				name: "name_min_length",
				payload: map[string]any{
					"name": "a",
				},
				expectedField: "name",
				expectedError: "validation_min_length:2",
			},
			{
				name: "name_max_length",
				payload: map[string]any{
					"name": strings.Repeat("a", 101),
				},
				expectedField: "name",
				expectedError: "validation_max_length:100",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				req := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, suite.BasePath+"/", tc.payload, userID, cfg)
				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				if rr.Code != http.StatusBadRequest {
					t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
				}

				resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
				if resp.Error != api.ErrValidationFailed.Error() {
					t.Fatalf("expected error %q, got %q", api.ErrValidationFailed.Error(), resp.Error)
				}

				AssertHasErrorDetail(t, resp.ErrorDetails, tc.expectedField, tc.expectedError)
			})
		}
	})

	t.Run("update_validations", func(t *testing.T) {
		repo := inmemory.NewRepository()
		cfg := TestJWTConfig()
		router := suite.NewRouter(repo, cfg)
		userID := uuid.New()
		resourceID := uuid.New()

		suite.Seed(repo, TypeSeed{ID: resourceID, UserID: userID, Name: "Existing"})

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
					"name": "a",
				},
				expectedCode:  http.StatusBadRequest,
				expectedField: "name",
				expectedError: "validation_min_length:2",
			},
			{
				name: "name_max_length",
				payload: map[string]any{
					"name": strings.Repeat("b", 101),
				},
				expectedCode:  http.StatusBadRequest,
				expectedField: "name",
				expectedError: "validation_max_length:100",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				req := handlertest.NewAuthorizedJSONRequest(t, http.MethodPut, suite.BasePath+"/"+resourceID.String(), tc.payload, userID, cfg)
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

					AssertHasErrorDetail(t, resp.ErrorDetails, tc.expectedField, tc.expectedError)
				}
			})
		}
	})
}

func RunTypeHandlerDecodeAndIDErrors(t *testing.T, suite ManagedTypeSuite) {
	t.Helper()

	repo := inmemory.NewRepository()
	cfg := TestJWTConfig()
	router := suite.NewRouter(repo, cfg)
	userID := uuid.New()
	resourceID := uuid.New()

	suite.Seed(repo, TypeSeed{ID: resourceID, UserID: userID, Name: "Existing"})

	createUnknownReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, suite.BasePath+"/", map[string]any{
		"name":    "Valid name",
		"unknown": "x",
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
	AssertHasErrorDetail(t, createUnknownResp.ErrorDetails, "unknown", api.KeyValidationUnknownField)

	createMalformedReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, suite.BasePath+"/", map[string]any{
		"name": 123,
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
	AssertHasErrorDetail(t, createMalformedResp.ErrorDetails, "name", "validation_malformed_parameter:expected_string")

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete}
	for _, method := range methods {
		var req *http.Request
		if method == http.MethodPut {
			req = handlertest.NewAuthorizedJSONRequest(t, method, suite.BasePath+"/not-a-uuid", map[string]any{"name": "Valid"}, userID, cfg)
		} else {
			req = handlertest.NewAuthorizedRequest(t, method, suite.BasePath+"/not-a-uuid", userID, cfg)
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

func RunTypeHandlerNotFoundAndInternalErrors(t *testing.T, suite ManagedTypeSuite) {
	t.Helper()

	repo := inmemory.NewRepository()
	cfg := TestJWTConfig()
	router := suite.NewRouter(repo, cfg)
	userID := uuid.New()
	missingID := uuid.New()

	getReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, suite.BasePath+"/"+missingID.String(), userID, cfg)
	getRR := httptest.NewRecorder()
	router.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, getRR.Code)
	}

	getResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, getRR)
	if getResp.Error != suite.NotFoundError {
		t.Fatalf("expected error %q, got %q", suite.NotFoundError, getResp.Error)
	}

	updateReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPut, suite.BasePath+"/"+missingID.String(), map[string]any{
		"name": "Updated",
	}, userID, cfg)
	updateRR := httptest.NewRecorder()
	router.ServeHTTP(updateRR, updateReq)

	if updateRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, updateRR.Code)
	}

	updateResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, updateRR)
	if updateResp.Error != suite.NotFoundError {
		t.Fatalf("expected error %q, got %q", suite.NotFoundError, updateResp.Error)
	}

	deleteReq := handlertest.NewAuthorizedRequest(t, http.MethodDelete, suite.BasePath+"/"+missingID.String(), userID, cfg)
	deleteRR := httptest.NewRecorder()
	router.ServeHTTP(deleteRR, deleteReq)

	if deleteRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, deleteRR.Code)
	}

	deleteResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, deleteRR)
	if deleteResp.Error != suite.NotFoundError {
		t.Fatalf("expected error %q, got %q", suite.NotFoundError, deleteResp.Error)
	}

	repo.SetError(suite.OpCreate, errors.New("database unavailable"))

	createReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, suite.BasePath+"/", map[string]any{
		"name": "Type Alpha",
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

	repo.ClearError(suite.OpCreate)
	repo.SetError(suite.OpList, errors.New("database unavailable"))

	listReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, suite.BasePath+"/", userID, cfg)
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

func TestJWTConfig() config.JWTConfig {
	return config.JWTConfig{
		AccessSecret:      "test-access-secret",
		AccessExpiration:  15 * time.Minute,
		RefreshSecret:     "test-refresh-secret",
		RefreshExpiration: 7 * 24 * time.Hour,
		Issuer:            "the-punisher-tests",
		Audience:          "the-punisher-tests",
	}
}

func AssertHasErrorDetail(t *testing.T, details []api.ErrorDetail, field, errKey string) {
	t.Helper()

	for _, detail := range details {
		if detail.Field == field && detail.Error == errKey {
			return
		}
	}

	t.Fatalf("expected error detail field=%q error=%q, got %+v", field, errKey, details)
}
