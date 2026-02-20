package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/repository"
	"github.com/mageas/the-punisher-backend/internal/testutil/handlertest"
	"github.com/mageas/the-punisher-backend/internal/testutil/httpx"
	"github.com/mageas/the-punisher-backend/internal/testutil/inmemory"
	shared "github.com/mageas/the-punisher-backend/internal/testutil/shared"
)

func TestClassroomHandlerDTOValidations(t *testing.T) {
	t.Run("create_validations", func(t *testing.T) {
		repo := inmemory.NewRepository()
		cfg := shared.TestJWTConfig()
		router := newClassroomRouter(repo, cfg)
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
					"name": "A",
				},
				expectedField: "name",
				expectedError: "validation_min_length:2",
			},
			{
				name: "name_max_length",
				payload: map[string]any{
					"name": strings.Repeat("A", 101),
				},
				expectedField: "name",
				expectedError: "validation_max_length:100",
			},
			{
				name: "year_max_length",
				payload: map[string]any{
					"name": "CM1 A",
					"year": strings.Repeat("2", 21),
				},
				expectedField: "year",
				expectedError: "validation_max_length:20",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				req := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/classrooms/", tc.payload, userID, cfg)
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
		router := newClassroomRouter(repo, cfg)
		userID := uuid.New()
		classroomID := uuid.New()
		repo.SeedClassroom(repository.Classroom{
			ID:     classroomID,
			UserID: userID,
			Name:   "CM1 A",
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
				req := handlertest.NewAuthorizedJSONRequest(t, http.MethodPut, "/v1/classrooms/"+classroomID.String(), tc.payload, userID, cfg)
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

	t.Run("add_student_validation", func(t *testing.T) {
		repo := inmemory.NewRepository()
		cfg := shared.TestJWTConfig()
		router := newClassroomRouter(repo, cfg)
		userID := uuid.New()
		classroomID := uuid.New()
		repo.SeedClassroom(repository.Classroom{
			ID:     classroomID,
			UserID: userID,
			Name:   "CM1 A",
		})

		req := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/classrooms/"+classroomID.String()+"/students", map[string]any{}, userID, cfg)
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
	})
}

func TestClassroomHandlerDecodeAndIDErrors(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newClassroomRouter(repo, cfg)
	userID := uuid.New()

	createUnknownReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/classrooms/", map[string]any{
		"name":    "CM1 A",
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
	shared.AssertHasErrorDetail(t, createUnknownResp.ErrorDetails, "unknown", api.KeyValidationUnknownField)

	createMalformedReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/classrooms/", map[string]any{
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
	shared.AssertHasErrorDetail(t, createMalformedResp.ErrorDetails, "name", "validation_malformed_parameter:expected_string")

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete}
	for _, method := range methods {
		var req *http.Request
		if method == http.MethodPut {
			req = handlertest.NewAuthorizedJSONRequest(t, method, "/v1/classrooms/not-a-uuid", map[string]any{"name": "CM1 A"}, userID, cfg)
		} else {
			req = handlertest.NewAuthorizedRequest(t, method, "/v1/classrooms/not-a-uuid", userID, cfg)
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
		shared.AssertHasErrorDetail(t, resp.ErrorDetails, "classroom_id", "validation_malformed_parameter:expected_uuid")
	}

	addBadClassroomReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/classrooms/not-a-uuid/students", map[string]any{
		"student_id": uuid.New().String(),
	}, userID, cfg)
	addBadClassroomRR := httptest.NewRecorder()
	router.ServeHTTP(addBadClassroomRR, addBadClassroomReq)

	if addBadClassroomRR.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, addBadClassroomRR.Code)
	}
	addBadClassroomResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, addBadClassroomRR)
	shared.AssertHasErrorDetail(t, addBadClassroomResp.ErrorDetails, "classroom_id", "validation_malformed_parameter:expected_uuid")

	removeBadStudentReq := handlertest.NewAuthorizedRequest(t, http.MethodDelete, "/v1/classrooms/"+uuid.New().String()+"/students/not-a-uuid", userID, cfg)
	removeBadStudentRR := httptest.NewRecorder()
	router.ServeHTTP(removeBadStudentRR, removeBadStudentReq)

	if removeBadStudentRR.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, removeBadStudentRR.Code)
	}
	removeBadStudentResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, removeBadStudentRR)
	shared.AssertHasErrorDetail(t, removeBadStudentResp.ErrorDetails, "student_id", "validation_malformed_parameter:expected_uuid")

	listByBadStudentReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/not-a-uuid/classrooms", userID, cfg)
	listByBadStudentRR := httptest.NewRecorder()
	router.ServeHTTP(listByBadStudentRR, listByBadStudentReq)

	if listByBadStudentRR.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, listByBadStudentRR.Code)
	}
	listByBadStudentResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, listByBadStudentRR)
	shared.AssertHasErrorDetail(t, listByBadStudentResp.ErrorDetails, "student_id", "validation_malformed_parameter:expected_uuid")
}
