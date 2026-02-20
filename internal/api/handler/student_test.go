package handler_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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

type studentResponse struct {
	ID                   uuid.UUID                  `json:"id"`
	FirstName            string                     `json:"first_name"`
	LastName             string                     `json:"last_name"`
	Classrooms           []studentClassroomResponse `json:"classrooms"`
	AvailableBonusPoints float64                    `json:"available_bonus_points"`
	PenaltyCount         int64                      `json:"penalty_count"`
}

type studentClassroomResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type paginatedStudentResponse struct {
	Page       int               `json:"page"`
	TotalCount int64             `json:"total_count"`
	Data       []studentResponse `json:"data"`
}

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

	created := httpx.DecodeJSONResponse[studentResponse](t, createRR)
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

	listResp := httpx.DecodeJSONResponse[paginatedStudentResponse](t, listRR)
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

	getResp := httpx.DecodeJSONResponse[studentResponse](t, getRR)
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

	updated := httpx.DecodeJSONResponse[studentResponse](t, updateRR)
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

	updateEmptyResp := httpx.DecodeJSONResponse[studentResponse](t, updateEmptyRR)
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
	if createMalformedResp.Error != api.ErrMalformedParameter.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrMalformedParameter.Error(), createMalformedResp.Error)
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

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}

		resp := httpx.DecodeJSONResponse[api.ErrorResponse](t, rr)
		if resp.Error != api.ErrMalformedParameter.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrMalformedParameter.Error(), resp.Error)
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

	pageOneResp := httpx.DecodeJSONResponse[paginatedStudentResponse](t, pageOneRR)
	if pageOneResp.Page != 1 || pageOneResp.TotalCount != 21 || len(pageOneResp.Data) != 20 {
		t.Fatalf("unexpected page 1 search response: %+v", pageOneResp)
	}

	pageTwoReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/?page=2&search=lucas%20dupont", userID, cfg)
	pageTwoRR := httptest.NewRecorder()
	router.ServeHTTP(pageTwoRR, pageTwoReq)

	if pageTwoRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, pageTwoRR.Code)
	}

	pageTwoResp := httpx.DecodeJSONResponse[paginatedStudentResponse](t, pageTwoRR)
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

func newStudentRouter(repo *inmemory.Repository, cfg config.JWTConfig) http.Handler {
	svc := service.NewStudentService(repo)
	h := handler.NewStudentHandler(svc)

	r := chi.NewRouter()
	r.Use(platformauth.AuthMiddleware(cfg.AccessSecret, cfg.Issuer, cfg.Audience))
	r.Route("/v1/students", func(r chi.Router) {
		r.Post("/", h.CreateStudent)
		r.Get("/", h.ListStudents)
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
