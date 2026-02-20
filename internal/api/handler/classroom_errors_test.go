package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/repository"
	"github.com/mageas/the-punisher-backend/internal/testutil/handlertest"
	"github.com/mageas/the-punisher-backend/internal/testutil/httpx"
	"github.com/mageas/the-punisher-backend/internal/testutil/inmemory"
	shared "github.com/mageas/the-punisher-backend/internal/testutil/shared"
)

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
