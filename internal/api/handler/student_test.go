package handler_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/api/handler"
	"github.com/mageas/the-punisher-backend/internal/dto"
	platformauth "github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/config"
	"github.com/mageas/the-punisher-backend/internal/platform/web"
	"github.com/mageas/the-punisher-backend/internal/repository"
	"github.com/mageas/the-punisher-backend/internal/service"
	"github.com/mageas/the-punisher-backend/internal/testutil/handlertest"
	"github.com/mageas/the-punisher-backend/internal/testutil/httpx"
	"github.com/mageas/the-punisher-backend/internal/testutil/inmemory"
	shared "github.com/mageas/the-punisher-backend/internal/testutil/shared"
)

func newStudentRouter(repo *inmemory.Repository, cfg config.JWTConfig) http.Handler {
	svc := service.NewStudentService(repo)
	h := handler.NewStudentHandler(svc)

	r := chi.NewRouter()
	r.Use(platformauth.AuthMiddleware(cfg.AccessSecret, cfg.Issuer, cfg.Audience))
	r.Route("/v1/students", func(r chi.Router) {
		r.Post("/", h.CreateStudent)
		r.Get("/", h.ListStudents)
		r.Delete("/", h.DeleteAllStudents)
		r.Get("/{student_id}", h.GetStudent)
		r.Get("/{student_id}/kpis", h.GetStudentKpis)
		r.Get("/{student_id}/history", h.GetStudentHistory)
		r.Put("/{student_id}", h.UpdateStudent)
		r.Delete("/{student_id}", h.DeleteStudent)
	})

	return r
}

func inmemoryStudent(id, userID uuid.UUID, firstName, lastName string) repository.Student {
	return repository.Student{
		ID:        id,
		UserID:    userID,
		FirstName: firstName,
		LastName:  lastName,
	}
}

// --- CRUD Tests ---

func TestStudentHandlerUnauthorized(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newStudentRouter(repo, cfg)

	req := httptest.NewRequest(http.MethodGet, "/v1/students/", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestStudentHandlerCRUDSuccess(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newStudentRouter(repo, cfg)
	userID := uuid.New()

	createReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/students/", map[string]any{
		"first_name": "Jean",
		"last_name":  "Dupont",
	}, userID, cfg)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)

	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createRR.Code)
	}

	created := httpx.DecodeJSONResponse[dto.ReturnStudentDto](t, createRR)
	if created.ID == uuid.Nil {
		t.Fatal("expected created student id")
	}
	if created.FirstName != "Jean" || created.LastName != "Dupont" {
		t.Fatalf("unexpected student payload: %+v", created)
	}
	if len(created.Classrooms) != 0 {
		t.Fatalf("expected no classrooms on create, got %+v", created.Classrooms)
	}
	if created.AvailableBonusPoints != 0 || created.PenaltyCount != 0 {
		t.Fatalf("expected zero aggregates on create, got bonus=%v penalties=%d", created.AvailableBonusPoints, created.PenaltyCount)
	}

	listReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/", userID, cfg)
	listRR := httptest.NewRecorder()
	router.ServeHTTP(listRR, listReq)

	if listRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listRR.Code)
	}

	listResp := httpx.DecodeJSONResponse[web.PaginatedResponse[*dto.ReturnStudentDto]](t, listRR)
	if listResp.TotalCount != 1 {
		t.Fatalf("expected total_count=1, got %d", listResp.TotalCount)
	}
	if len(listResp.Data) != 1 {
		t.Fatalf("expected one item in list, got %d", len(listResp.Data))
	}
	if listResp.Data[0].ID != created.ID {
		t.Fatalf("expected listed id %s, got %s", created.ID, listResp.Data[0].ID)
	}
	if listResp.Data[0].AvailableBonusPoints != 0 || listResp.Data[0].PenaltyCount != 0 {
		t.Fatalf("expected zero aggregates in list, got %+v", listResp.Data[0])
	}

	getReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+created.ID.String(), userID, cfg)
	getRR := httptest.NewRecorder()
	router.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getRR.Code)
	}

	getResp := httpx.DecodeJSONResponse[dto.ReturnStudentDto](t, getRR)
	if getResp.ID != created.ID {
		t.Fatalf("expected student id %s, got %s", created.ID, getResp.ID)
	}
	if getResp.AvailableBonusPoints != 0 || getResp.PenaltyCount != 0 {
		t.Fatalf("expected zero aggregates in get, got %+v", getResp)
	}

	updateReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPut, "/v1/students/"+created.ID.String(), map[string]any{
		"first_name": "Jeanne",
	}, userID, cfg)
	updateRR := httptest.NewRecorder()
	router.ServeHTTP(updateRR, updateReq)

	if updateRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, updateRR.Code)
	}

	updated := httpx.DecodeJSONResponse[dto.ReturnStudentDto](t, updateRR)
	if updated.FirstName != "Jeanne" || updated.LastName != "Dupont" {
		t.Fatalf("expected updated student, got %+v", updated)
	}
	if updated.AvailableBonusPoints != 0 || updated.PenaltyCount != 0 {
		t.Fatalf("expected zero aggregates in update, got %+v", updated)
	}

	updateEmptyReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPut, "/v1/students/"+created.ID.String(), map[string]any{}, userID, cfg)
	updateEmptyRR := httptest.NewRecorder()
	router.ServeHTTP(updateEmptyRR, updateEmptyReq)

	if updateEmptyRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, updateEmptyRR.Code)
	}

	updateEmptyResp := httpx.DecodeJSONResponse[dto.ReturnStudentDto](t, updateEmptyRR)
	if updateEmptyResp.FirstName != "Jeanne" || updateEmptyResp.LastName != "Dupont" {
		t.Fatalf("expected student unchanged, got %+v", updateEmptyResp)
	}

	deleteReq := handlertest.NewAuthorizedRequest(t, http.MethodDelete, "/v1/students/"+created.ID.String(), userID, cfg)
	deleteRR := httptest.NewRecorder()
	router.ServeHTTP(deleteRR, deleteReq)

	if deleteRR.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, deleteRR.Code)
	}

	getDeletedReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+created.ID.String(), userID, cfg)
	getDeletedRR := httptest.NewRecorder()
	router.ServeHTTP(getDeletedRR, getDeletedReq)

	if getDeletedRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, getDeletedRR.Code)
	}

	errResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, getDeletedRR)
	if errResp.Error != api.ErrStudentNotFound.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrStudentNotFound.Error(), errResp.Error)
	}
}

func TestStudentHandlerDeleteAllSuccessAndTenantIsolation(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newStudentRouter(repo, cfg)

	userID := uuid.New()
	otherUserID := uuid.New()
	otherStudentID := uuid.New()

	repo.SeedStudent(inmemoryStudent(uuid.New(), userID, "Jean", "Dupont"))
	repo.SeedStudent(inmemoryStudent(uuid.New(), userID, "Emma", "Martin"))
	repo.SeedStudent(inmemoryStudent(otherStudentID, otherUserID, "Lucas", "Outside"))

	deleteAllReq := handlertest.NewAuthorizedRequest(t, http.MethodDelete, "/v1/students/", userID, cfg)
	deleteAllRR := httptest.NewRecorder()
	router.ServeHTTP(deleteAllRR, deleteAllReq)

	if deleteAllRR.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, deleteAllRR.Code)
	}

	userListReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/", userID, cfg)
	userListRR := httptest.NewRecorder()
	router.ServeHTTP(userListRR, userListReq)

	if userListRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, userListRR.Code)
	}

	userListResp := httpx.DecodeJSONResponse[web.PaginatedResponse[*dto.ReturnStudentDto]](t, userListRR)
	if userListResp.TotalCount != 0 || len(userListResp.Data) != 0 {
		t.Fatalf("expected empty list after bulk delete, got %+v", userListResp)
	}

	otherListReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/", otherUserID, cfg)
	otherListRR := httptest.NewRecorder()
	router.ServeHTTP(otherListRR, otherListReq)

	if otherListRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, otherListRR.Code)
	}

	otherListResp := httpx.DecodeJSONResponse[web.PaginatedResponse[*dto.ReturnStudentDto]](t, otherListRR)
	if otherListResp.TotalCount != 1 || len(otherListResp.Data) != 1 {
		t.Fatalf("expected tenant isolation with one remaining external student, got %+v", otherListResp)
	}
	if otherListResp.Data[0].ID != otherStudentID {
		t.Fatalf("expected external student id %s, got %s", otherStudentID, otherListResp.Data[0].ID)
	}
}

func TestStudentHandlerDTOValidations(t *testing.T) {
	t.Run("create_validations", func(t *testing.T) {
		repo := inmemory.NewRepository()
		cfg := shared.TestJWTConfig()
		router := newStudentRouter(repo, cfg)
		userID := uuid.New()

		tests := []struct {
			name          string
			payload       map[string]any
			expectedField string
			expectedError string
		}{
			{
				name: "first_name_required",
				payload: map[string]any{
					"last_name": "Dupont",
				},
				expectedField: "first_name",
				expectedError: api.KeyValidationFieldRequired,
			},
			{
				name: "last_name_required",
				payload: map[string]any{
					"first_name": "Jean",
				},
				expectedField: "last_name",
				expectedError: api.KeyValidationFieldRequired,
			},
			{
				name: "first_name_min_length",
				payload: map[string]any{
					"first_name": "J",
					"last_name":  "Dupont",
				},
				expectedField: "first_name",
				expectedError: "validation_min_length:2",
			},
			{
				name: "last_name_max_length",
				payload: map[string]any{
					"first_name": "Jean",
					"last_name":  strings.Repeat("D", 71),
				},
				expectedField: "last_name",
				expectedError: "validation_max_length:70",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				req := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/students/", tc.payload, userID, cfg)
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
		router := newStudentRouter(repo, cfg)
		userID := uuid.New()
		studentID := uuid.New()
		repo.SeedStudent(inmemoryStudent(studentID, userID, "Existing", "Student"))

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
				name: "first_name_min_length",
				payload: map[string]any{
					"first_name": "A",
				},
				expectedCode:  http.StatusBadRequest,
				expectedField: "first_name",
				expectedError: "validation_min_length:2",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				req := handlertest.NewAuthorizedJSONRequest(t, http.MethodPut, "/v1/students/"+studentID.String(), tc.payload, userID, cfg)
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

func TestStudentHandlerDecodeAndIDErrors(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newStudentRouter(repo, cfg)
	userID := uuid.New()

	createUnknownReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/students/", map[string]any{
		"first_name": "Jean",
		"last_name":  "Dupont",
		"unknown":    "x",
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

	createMalformedReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/students/", map[string]any{
		"first_name": 123,
		"last_name":  "Dupont",
	}, userID, cfg)
	createMalformedRR := httptest.NewRecorder()
	router.ServeHTTP(createMalformedRR, createMalformedReq)

	if createMalformedRR.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, createMalformedRR.Code)
	}

	createMalformedResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, createMalformedRR)
	if createMalformedResp.Error != api.ErrInvalidRequestBody.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrInvalidRequestBody.Error(), createMalformedResp.Error)
	}
	shared.AssertHasErrorDetail(t, createMalformedResp.ErrorDetails, "first_name", "validation_malformed_parameter:expected_string")

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete}
	for _, method := range methods {
		var req *http.Request
		if method == http.MethodPut {
			req = handlertest.NewAuthorizedJSONRequest(t, method, "/v1/students/not-a-uuid", map[string]any{"first_name": "Jean"}, userID, cfg)
		} else {
			req = handlertest.NewAuthorizedRequest(t, method, "/v1/students/not-a-uuid", userID, cfg)
		}

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrNotFound.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrNotFound.Error(), resp.Error)
		}
		shared.AssertHasErrorDetail(t, resp.ErrorDetails, "student_id", "validation_malformed_parameter:expected_uuid")
	}
}

func TestStudentHandlerListSearch(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newStudentRouter(repo, cfg)
	userID := uuid.New()
	otherUserID := uuid.New()

	for i := 0; i < 21; i++ {
		repo.SeedStudent(inmemoryStudent(uuid.New(), userID, "Lucas", fmt.Sprintf("Dupont %02d", i)))
	}
	repo.SeedStudent(inmemoryStudent(uuid.New(), userID, "Jean", "Martin"))
	repo.SeedStudent(inmemoryStudent(uuid.New(), otherUserID, "Lucas", "Dupont Externe"))

	pageOneReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/?search=%20%20lucas%20%20%20dupont%20%20", userID, cfg)
	pageOneRR := httptest.NewRecorder()
	router.ServeHTTP(pageOneRR, pageOneReq)

	if pageOneRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, pageOneRR.Code)
	}

	pageOneResp := httpx.DecodeJSONResponse[web.PaginatedResponse[*dto.ReturnStudentDto]](t, pageOneRR)
	if pageOneResp.Page != 1 || pageOneResp.TotalCount != 21 || len(pageOneResp.Data) != 20 {
		t.Fatalf("unexpected page 1 search response: %+v", pageOneResp)
	}

	pageTwoReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/?page=2&search=lucas%20dupont", userID, cfg)
	pageTwoRR := httptest.NewRecorder()
	router.ServeHTTP(pageTwoRR, pageTwoReq)

	if pageTwoRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, pageTwoRR.Code)
	}

	pageTwoResp := httpx.DecodeJSONResponse[web.PaginatedResponse[*dto.ReturnStudentDto]](t, pageTwoRR)
	if pageTwoResp.Page != 2 || pageTwoResp.TotalCount != 21 || len(pageTwoResp.Data) != 1 {
		t.Fatalf("unexpected page 2 search response: %+v", pageTwoResp)
	}
}

func TestStudentHandlerNotFoundAndInternalErrors(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newStudentRouter(repo, cfg)
	userID := uuid.New()
	missingID := uuid.New()

	getReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+missingID.String(), userID, cfg)
	getRR := httptest.NewRecorder()
	router.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, getRR.Code)
	}

	getResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, getRR)
	if getResp.Error != api.ErrStudentNotFound.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrStudentNotFound.Error(), getResp.Error)
	}

	updateReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPut, "/v1/students/"+missingID.String(), map[string]any{
		"first_name": "Updated",
	}, userID, cfg)
	updateRR := httptest.NewRecorder()
	router.ServeHTTP(updateRR, updateReq)

	if updateRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, updateRR.Code)
	}

	updateResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, updateRR)
	if updateResp.Error != api.ErrStudentNotFound.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrStudentNotFound.Error(), updateResp.Error)
	}

	deleteReq := handlertest.NewAuthorizedRequest(t, http.MethodDelete, "/v1/students/"+missingID.String(), userID, cfg)
	deleteRR := httptest.NewRecorder()
	router.ServeHTTP(deleteRR, deleteReq)

	if deleteRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, deleteRR.Code)
	}

	deleteResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, deleteRR)
	if deleteResp.Error != api.ErrStudentNotFound.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrStudentNotFound.Error(), deleteResp.Error)
	}

	repo.SetError(inmemory.OpCreateStudent, errors.New("database unavailable"))
	createReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/students/", map[string]any{
		"first_name": "Jean",
		"last_name":  "Dupont",
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

	repo.ClearError(inmemory.OpCreateStudent)
	repo.SetError(inmemory.OpListStudentsByUser, errors.New("database unavailable"))

	listReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/", userID, cfg)
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

func TestStudentHandlerDeleteAllInternalError(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newStudentRouter(repo, cfg)
	userID := uuid.New()

	repo.SetError(inmemory.OpDeleteAllStudentsByUser, errors.New("database unavailable"))

	req := handlertest.NewAuthorizedRequest(t, http.MethodDelete, "/v1/students/", userID, cfg)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
	if resp.Error != api.ErrInternalError.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrInternalError.Error(), resp.Error)
	}
}

// --- KPIs & History Tests ---

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
	now := time.Now().UTC()
	overdueDueAt := now.Add(-24 * time.Hour)
	futureDueAt := now.Add(24 * time.Hour)

	repo.SeedStudent(inmemoryStudent(studentID, userID, "Lucas", "Dubois"))
	repo.SeedBonusType(repository.BonusType{ID: bonusTypeID, UserID: userID, Name: "Participation"})
	repo.SeedPenaltyType(repository.PenaltyType{ID: penaltyTypeID, UserID: userID, Name: "Bavardage"})
	repo.SeedPunishmentType(repository.PunishmentType{ID: punishmentTypeID, UserID: userID, Name: "Retenue"})

	usedAt := base.Add(3 * time.Hour)
	repo.SeedBonus(repository.Bonus{ID: uuid.New(), UserID: userID, StudentID: studentID, BonusTypeID: bonusTypeID, Points: 2, CreatedAt: base.Add(1 * time.Hour)})
	repo.SeedBonus(repository.Bonus{ID: uuid.New(), UserID: userID, StudentID: studentID, BonusTypeID: bonusTypeID, Points: 1, CreatedAt: base.Add(2 * time.Hour), UsedAt: &usedAt})
	repo.SeedPenalty(repository.Penalty{ID: uuid.New(), UserID: userID, StudentID: studentID, PenaltyTypeID: penaltyTypeID, CreatedAt: base.Add(4 * time.Hour)})
	repo.SeedPunishment(repository.Punishment{ID: uuid.New(), UserID: userID, StudentID: studentID, PunishmentTypeID: punishmentTypeID, CreatedAt: base.Add(5 * time.Hour), DueAt: overdueDueAt})
	resolvedAt := base.Add(8 * time.Hour)
	repo.SeedPunishment(repository.Punishment{ID: uuid.New(), UserID: userID, StudentID: studentID, PunishmentTypeID: punishmentTypeID, CreatedAt: base.Add(6 * time.Hour), DueAt: futureDueAt, ResolvedAt: &resolvedAt})

	req := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+studentID.String()+"/kpis", userID, cfg)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	resp := httpx.DecodeJSONResponse[dto.StudentKpisDto](t, rr)
	if resp.AvailableBonusPoints != 2 ||
		resp.TotalBonusPoints != 3 ||
		resp.ActiveBonusCount != 1 ||
		resp.PenaltyCount != 1 ||
		resp.TotalPenaltyCount != 1 ||
		resp.TotalPunishmentCount != 2 ||
		resp.OverduePunishmentCount != 1 ||
		resp.PendingPunishmentCount != 1 {
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

	usedAt := base.Add(3 * time.Hour)
	repo.SeedBonus(repository.Bonus{ID: uuid.New(), UserID: userID, StudentID: studentID, BonusTypeID: bonusTypeID, Points: 2, CreatedAt: base.Add(1 * time.Hour)})
	repo.SeedBonus(repository.Bonus{ID: uuid.New(), UserID: userID, StudentID: studentID, BonusTypeID: bonusTypeID, Points: 1, CreatedAt: base.Add(2 * time.Hour), UsedAt: &usedAt})
	repo.SeedPenalty(repository.Penalty{ID: uuid.New(), UserID: userID, StudentID: studentID, PenaltyTypeID: penaltyTypeID, CreatedAt: base.Add(4 * time.Hour)})
	repo.SeedPunishment(repository.Punishment{ID: uuid.New(), UserID: userID, StudentID: studentID, PunishmentTypeID: punishmentTypeID, TriggeringRuleID: &ruleID, Automated: true, CreatedAt: base.Add(5 * time.Hour), DueAt: base.Add(24 * time.Hour)})
	resolvedAt := base.Add(8 * time.Hour)
	repo.SeedPunishment(repository.Punishment{ID: uuid.New(), UserID: userID, StudentID: studentID, PunishmentTypeID: punishmentTypeID, CreatedAt: base.Add(6 * time.Hour), DueAt: base.Add(24 * time.Hour), ResolvedAt: &resolvedAt})

	req := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+studentID.String()+"/history", userID, cfg)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	resp := httpx.DecodeJSONResponse[web.PaginatedResponse[dto.StudentHistoryItemDto]](t, rr)
	if resp.TotalCount != 5 {
		t.Fatalf("expected total_count=5, got %d", resp.TotalCount)
	}
	if len(resp.Data) != 5 {
		t.Fatalf("expected 5 history items, got %d (%+v)", len(resp.Data), resp.Data)
	}
	if resp.Data[0].Type != "punishment" {
		t.Fatalf("expected first history item to be latest punishment, got %+v", resp.Data[0])
	}
	if resp.Data[0].PunishmentTypeID == nil || resp.Data[0].PunishmentTypeName == nil || resp.Data[0].DueAt == nil {
		t.Fatalf("expected punishment fields on first item, got %+v", resp.Data[0])
	}

	typesCount := map[string]int{}
	automatedTrueCount := 0
	automatedFalseCount := 0
	for _, item := range resp.Data {
		typesCount[item.Type]++
		if item.Type == "punishment" {
			if item.Automated == nil {
				t.Fatalf("expected automated field on punishment history item, got %+v", item)
			}
			if *item.Automated {
				automatedTrueCount++
			} else {
				automatedFalseCount++
			}
			continue
		}
		if item.Automated != nil {
			t.Fatalf("expected automated to be omitted for non-punishment history item, got %+v", item)
		}
	}
	if typesCount["bonus"] != 2 || typesCount["penalty"] != 1 || typesCount["punishment"] != 2 {
		t.Fatalf("unexpected history type distribution: %+v", typesCount)
	}
	if automatedTrueCount != 1 || automatedFalseCount != 1 {
		t.Fatalf("expected one automated and one manual punishment in history, got true=%d false=%d", automatedTrueCount, automatedFalseCount)
	}
}

func TestStudentHistoryHandlerPagination(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newStudentRouter(repo, cfg)
	userID := uuid.New()
	studentID := uuid.New()
	bonusTypeID := uuid.New()
	base := time.Date(2026, 2, 2, 10, 0, 0, 0, time.UTC)

	repo.SeedStudent(inmemoryStudent(studentID, userID, "Jean", "Dupont"))
	repo.SeedBonusType(repository.BonusType{ID: bonusTypeID, UserID: userID, Name: "Participation"})
	for i := 0; i < 21; i++ {
		repo.SeedBonus(repository.Bonus{
			ID:          uuid.New(),
			UserID:      userID,
			StudentID:   studentID,
			BonusTypeID: bonusTypeID,
			Points:      1,
			CreatedAt:   base.Add(time.Duration(i) * time.Minute),
		})
	}

	req := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+studentID.String()+"/history?page=2", userID, cfg)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	resp := httpx.DecodeJSONResponse[web.PaginatedResponse[dto.StudentHistoryItemDto]](t, rr)
	if resp.Page != 2 {
		t.Fatalf("expected page=2, got %d", resp.Page)
	}
	if resp.TotalCount != 21 {
		t.Fatalf("expected total_count=21, got %d", resp.TotalCount)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected exactly one history item on page 2, got %d", len(resp.Data))
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

		if rrKpis.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, rrKpis.Code)
		}
		respKpis := httpx.DecodeJSONResponse[api.ErrorResponse](t, rrKpis)
		shared.AssertHasErrorDetail(t, respKpis.ErrorDetails, "student_id", "validation_malformed_parameter:expected_uuid")

		reqHistory := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/not-a-uuid/history", userID, cfg)
		rrHistory := httptest.NewRecorder()
		router.ServeHTTP(rrHistory, reqHistory)

		if rrHistory.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, rrHistory.Code)
		}
		respHistory := httpx.DecodeJSONResponse[api.ErrorResponse](t, rrHistory)
		shared.AssertHasErrorDetail(t, respHistory.ErrorDetails, "student_id", "validation_malformed_parameter:expected_uuid")
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
