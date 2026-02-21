package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

// --- Shared Helper Functions ---

func newClassroomRouter(repo *inmemory.Repository, cfg config.JWTConfig) http.Handler {
	classroomSvc := service.NewClassroomService(repo)
	classroomHandler := handler.NewClassroomHandler(classroomSvc)

	r := chi.NewRouter()
	r.Use(platformauth.AuthMiddleware(cfg.AccessSecret, cfg.Issuer, cfg.Audience))

	r.Route("/v1/classrooms", func(r chi.Router) {
		r.Post("/", classroomHandler.CreateClassroom)
		r.Get("/", classroomHandler.ListClassrooms)
		r.Get("/{classroom_id}", classroomHandler.GetClassroom)
		r.Put("/{classroom_id}", classroomHandler.UpdateClassroom)
		r.Delete("/{classroom_id}", classroomHandler.DeleteClassroom)
		r.Post("/{classroom_id}/students", classroomHandler.AddStudentToClassroom)
		r.Delete("/{classroom_id}/students/{student_id}", classroomHandler.RemoveStudentFromClassroom)
		r.Get("/{classroom_id}/students", classroomHandler.ListStudentsByClassroom)
	})

	r.Route("/v1/students", func(r chi.Router) {
		r.Get("/{student_id}/classrooms", classroomHandler.ListClassroomsByStudent)
	})

	return r
}

func TestClassroomHandlerCRUDAndRelationsSuccess(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newClassroomRouter(repo, cfg)
	userID := uuid.New()
	studentID := uuid.New()

	repo.SeedStudent(repository.Student{
		ID:        studentID,
		UserID:    userID,
		FirstName: "Jean",
		LastName:  "Dupont",
	})

	createReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/classrooms/", map[string]any{
		"name":         "CM1 A",
		"year":         "2025",
		"main_teacher": "Mme Martin",
	}, userID, cfg)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)

	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createRR.Code)
	}

	created := httpx.DecodeJSONResponse[dto.ReturnClassroomDto](t, createRR)
	if created.ID == uuid.Nil {
		t.Fatal("expected created classroom id")
	}
	if created.Name != "CM1 A" {
		t.Fatalf("expected classroom name %q, got %q", "CM1 A", created.Name)
	}
	if created.Year == nil || *created.Year != "2025" {
		t.Fatalf("expected year %q, got %+v", "2025", created.Year)
	}
	if created.MainTeacher == nil || *created.MainTeacher != "Mme Martin" {
		t.Fatalf("expected teacher %q, got %+v", "Mme Martin", created.MainTeacher)
	}
	if created.StudentCount != 0 {
		t.Fatalf("expected student_count %d, got %d", 0, created.StudentCount)
	}
	if len(created.StudentsPreview) != 0 {
		t.Fatalf("expected empty students_preview, got %+v", created.StudentsPreview)
	}
	if created.TotalBonusPoints != 0 || created.TotalPenaltyCount != 0 {
		t.Fatalf("expected zero classroom aggregates, got %+v", created)
	}

	listReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/classrooms/", userID, cfg)
	listRR := httptest.NewRecorder()
	router.ServeHTTP(listRR, listReq)

	if listRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listRR.Code)
	}

	listResp := httpx.DecodeJSONResponse[web.PaginatedResponse[*dto.ReturnClassroomDto]](t, listRR)
	if listResp.TotalCount != 1 || len(listResp.Data) != 1 {
		t.Fatalf("unexpected list response: %+v", listResp)
	}
	if listResp.Data[0].ID != created.ID {
		t.Fatalf("expected listed id %s, got %s", created.ID, listResp.Data[0].ID)
	}
	if listResp.Data[0].StudentCount != 0 || len(listResp.Data[0].StudentsPreview) != 0 {
		t.Fatalf("expected empty classroom preview before link, got %+v", listResp.Data[0])
	}

	getReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/classrooms/"+created.ID.String(), userID, cfg)
	getRR := httptest.NewRecorder()
	router.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getRR.Code)
	}

	getResp := httpx.DecodeJSONResponse[dto.ReturnClassroomDto](t, getRR)
	if getResp.ID != created.ID {
		t.Fatalf("expected classroom id %s, got %s", created.ID, getResp.ID)
	}
	if getResp.StudentCount != 0 || len(getResp.StudentsPreview) != 0 {
		t.Fatalf("expected empty classroom preview before link, got %+v", getResp)
	}

	updateReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPut, "/v1/classrooms/"+created.ID.String(), map[string]any{
		"name": "CM1 B",
	}, userID, cfg)
	updateRR := httptest.NewRecorder()
	router.ServeHTTP(updateRR, updateReq)

	if updateRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, updateRR.Code)
	}

	updated := httpx.DecodeJSONResponse[dto.ReturnClassroomDto](t, updateRR)
	if updated.Name != "CM1 B" {
		t.Fatalf("expected updated name %q, got %q", "CM1 B", updated.Name)
	}
	if updated.StudentCount != 0 || len(updated.StudentsPreview) != 0 {
		t.Fatalf("expected empty classroom preview on update before link, got %+v", updated)
	}

	addStudentReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/classrooms/"+created.ID.String()+"/students", map[string]any{
		"student_id": studentID.String(),
	}, userID, cfg)
	addStudentRR := httptest.NewRecorder()
	router.ServeHTTP(addStudentRR, addStudentReq)

	if addStudentRR.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, addStudentRR.Code)
	}

	listStudentsReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/classrooms/"+created.ID.String()+"/students", userID, cfg)
	listStudentsRR := httptest.NewRecorder()
	router.ServeHTTP(listStudentsRR, listStudentsReq)

	if listStudentsRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listStudentsRR.Code)
	}

	listStudentsResp := httpx.DecodeJSONResponse[web.PaginatedResponse[*dto.ReturnStudentDto]](t, listStudentsRR)
	if listStudentsResp.TotalCount != 1 || len(listStudentsResp.Data) != 1 {
		t.Fatalf("unexpected students by classroom response: %+v", listStudentsResp)
	}
	if listStudentsResp.Data[0].ID != studentID {
		t.Fatalf("expected student id %s, got %s", studentID, listStudentsResp.Data[0].ID)
	}
	if len(listStudentsResp.Data[0].Classrooms) != 1 || listStudentsResp.Data[0].Classrooms[0].ID != created.ID {
		t.Fatalf("expected student classroom badges to include current classroom, got %+v", listStudentsResp.Data[0].Classrooms)
	}
	if listStudentsResp.Data[0].AvailableBonusPoints != 0 || listStudentsResp.Data[0].PenaltyCount != 0 {
		t.Fatalf("expected zero student aggregates, got %+v", listStudentsResp.Data[0])
	}

	listClassroomsByStudentReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+studentID.String()+"/classrooms", userID, cfg)
	listClassroomsByStudentRR := httptest.NewRecorder()
	router.ServeHTTP(listClassroomsByStudentRR, listClassroomsByStudentReq)

	if listClassroomsByStudentRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listClassroomsByStudentRR.Code)
	}

	listClassroomsByStudentResp := httpx.DecodeJSONResponse[web.PaginatedResponse[*dto.ReturnClassroomDto]](t, listClassroomsByStudentRR)
	if listClassroomsByStudentResp.TotalCount != 1 || len(listClassroomsByStudentResp.Data) != 1 {
		t.Fatalf("unexpected classrooms by student response: %+v", listClassroomsByStudentResp)
	}
	if listClassroomsByStudentResp.Data[0].ID != created.ID {
		t.Fatalf("expected classroom id %s, got %s", created.ID, listClassroomsByStudentResp.Data[0].ID)
	}
	if listClassroomsByStudentResp.Data[0].StudentCount != 1 {
		t.Fatalf("expected student_count %d, got %d", 1, listClassroomsByStudentResp.Data[0].StudentCount)
	}
	if len(listClassroomsByStudentResp.Data[0].StudentsPreview) != 1 {
		t.Fatalf("expected one students_preview entry, got %+v", listClassroomsByStudentResp.Data[0].StudentsPreview)
	}
	if listClassroomsByStudentResp.Data[0].StudentsPreview[0].ID != studentID {
		t.Fatalf("expected preview student id %s, got %s", studentID, listClassroomsByStudentResp.Data[0].StudentsPreview[0].ID)
	}
	if listClassroomsByStudentResp.Data[0].StudentsPreview[0].FirstName != "Jean" || listClassroomsByStudentResp.Data[0].StudentsPreview[0].LastName != "Dupont" {
		t.Fatalf("unexpected student preview names: %+v", listClassroomsByStudentResp.Data[0].StudentsPreview[0])
	}
	if listClassroomsByStudentResp.Data[0].TotalBonusPoints != 0 || listClassroomsByStudentResp.Data[0].TotalPenaltyCount != 0 {
		t.Fatalf("expected zero classroom aggregates, got %+v", listClassroomsByStudentResp.Data[0])
	}

	removeReq := handlertest.NewAuthorizedRequest(t, http.MethodDelete, "/v1/classrooms/"+created.ID.String()+"/students/"+studentID.String(), userID, cfg)
	removeRR := httptest.NewRecorder()
	router.ServeHTTP(removeRR, removeReq)

	if removeRR.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, removeRR.Code)
	}

	listStudentsAfterRemovalReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/classrooms/"+created.ID.String()+"/students", userID, cfg)
	listStudentsAfterRemovalRR := httptest.NewRecorder()
	router.ServeHTTP(listStudentsAfterRemovalRR, listStudentsAfterRemovalReq)

	if listStudentsAfterRemovalRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listStudentsAfterRemovalRR.Code)
	}

	listStudentsAfterRemovalResp := httpx.DecodeJSONResponse[web.PaginatedResponse[*dto.ReturnStudentDto]](t, listStudentsAfterRemovalRR)
	if listStudentsAfterRemovalResp.TotalCount != 0 || len(listStudentsAfterRemovalResp.Data) != 0 {
		t.Fatalf("expected no student relation, got %+v", listStudentsAfterRemovalResp)
	}

	deleteReq := handlertest.NewAuthorizedRequest(t, http.MethodDelete, "/v1/classrooms/"+created.ID.String(), userID, cfg)
	deleteRR := httptest.NewRecorder()
	router.ServeHTTP(deleteRR, deleteReq)

	if deleteRR.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, deleteRR.Code)
	}

	getDeletedReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/classrooms/"+created.ID.String(), userID, cfg)
	getDeletedRR := httptest.NewRecorder()
	router.ServeHTTP(getDeletedRR, getDeletedReq)

	if getDeletedRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, getDeletedRR.Code)
	}

	errResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, getDeletedRR)
	if errResp.Error != api.ErrClassroomNotFound.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrClassroomNotFound.Error(), errResp.Error)
	}
}

// --- Business & Internal Errors Tests ---

func TestClassroomHandlerBusinessAndInternalErrors(t *testing.T) {
	repo := inmemory.NewRepository()
	cfg := shared.TestJWTConfig()
	router := newClassroomRouter(repo, cfg)
	userID := uuid.New()
	missingClassroomID := uuid.New()
	studentID := uuid.New()

	getReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/classrooms/"+missingClassroomID.String(), userID, cfg)
	getRR := httptest.NewRecorder()
	router.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, getRR.Code)
	}

	getResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, getRR)
	if getResp.Error != api.ErrClassroomNotFound.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrClassroomNotFound.Error(), getResp.Error)
	}

	repo.SeedStudent(repository.Student{
		ID:        studentID,
		UserID:    userID,
		FirstName: "Jean",
		LastName:  "Dupont",
	})

	addMissingReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/classrooms/"+missingClassroomID.String()+"/students", map[string]any{
		"student_id": studentID.String(),
	}, userID, cfg)
	addMissingRR := httptest.NewRecorder()
	router.ServeHTTP(addMissingRR, addMissingReq)

	if addMissingRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, addMissingRR.Code)
	}

	addMissingResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, addMissingRR)
	if addMissingResp.Error != api.ErrStudentOrClassroomNotFound.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrStudentOrClassroomNotFound.Error(), addMissingResp.Error)
	}

	classroomID := uuid.New()
	repo.SeedClassroom(repository.Classroom{
		ID:     classroomID,
		UserID: userID,
		Name:   "CM1 A",
	})

	addReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/classrooms/"+classroomID.String()+"/students", map[string]any{
		"student_id": studentID.String(),
	}, userID, cfg)
	addRR := httptest.NewRecorder()
	router.ServeHTTP(addRR, addReq)

	if addRR.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, addRR.Code)
	}

	addDuplicateReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/classrooms/"+classroomID.String()+"/students", map[string]any{
		"student_id": studentID.String(),
	}, userID, cfg)
	addDuplicateRR := httptest.NewRecorder()
	router.ServeHTTP(addDuplicateRR, addDuplicateReq)

	if addDuplicateRR.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, addDuplicateRR.Code)
	}

	addDuplicateResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, addDuplicateRR)
	if addDuplicateResp.Error != api.ErrStudentClassroomRelationExists.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrStudentClassroomRelationExists.Error(), addDuplicateResp.Error)
	}

	removeMissingRelationReq := handlertest.NewAuthorizedRequest(t, http.MethodDelete, "/v1/classrooms/"+classroomID.String()+"/students/"+uuid.New().String(), userID, cfg)
	removeMissingRelationRR := httptest.NewRecorder()
	router.ServeHTTP(removeMissingRelationRR, removeMissingRelationReq)

	if removeMissingRelationRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, removeMissingRelationRR.Code)
	}

	removeMissingRelationResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, removeMissingRelationRR)
	if removeMissingRelationResp.Error != api.ErrStudentOrClassroomNotFound.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrStudentOrClassroomNotFound.Error(), removeMissingRelationResp.Error)
	}

	listByMissingStudentReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/students/"+uuid.New().String()+"/classrooms", userID, cfg)
	listByMissingStudentRR := httptest.NewRecorder()
	router.ServeHTTP(listByMissingStudentRR, listByMissingStudentReq)

	if listByMissingStudentRR.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, listByMissingStudentRR.Code)
	}

	listByMissingStudentResp := httpx.DecodeJSONResponse[api.ErrorResponse](t, listByMissingStudentRR)
	if listByMissingStudentResp.Error != api.ErrStudentNotFound.Error() {
		t.Fatalf("expected error %q, got %q", api.ErrStudentNotFound.Error(), listByMissingStudentResp.Error)
	}

	repo.SetError(inmemory.OpCreateClassroom, errors.New("database unavailable"))
	createReq := handlertest.NewAuthorizedJSONRequest(t, http.MethodPost, "/v1/classrooms/", map[string]any{
		"name": "CM1 A",
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

	repo.ClearError(inmemory.OpCreateClassroom)
	repo.SetError(inmemory.OpListClassroomsByUser, errors.New("database unavailable"))

	listReq := handlertest.NewAuthorizedRequest(t, http.MethodGet, "/v1/classrooms/", userID, cfg)
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

// --- Validation Tests ---

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
