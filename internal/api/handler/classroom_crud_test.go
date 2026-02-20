package handler_test

import (
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

	created := httpx.DecodeJSONResponse[classroomResponse](t, createRR)
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

	listResp := httpx.DecodeJSONResponse[paginatedClassroomResponse](t, listRR)
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

	getResp := httpx.DecodeJSONResponse[classroomResponse](t, getRR)
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

	updated := httpx.DecodeJSONResponse[classroomResponse](t, updateRR)
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

	listStudentsResp := httpx.DecodeJSONResponse[paginatedStudentResponse](t, listStudentsRR)
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

	listClassroomsByStudentResp := httpx.DecodeJSONResponse[paginatedClassroomResponse](t, listClassroomsByStudentRR)
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

	listStudentsAfterRemovalResp := httpx.DecodeJSONResponse[paginatedStudentResponse](t, listStudentsAfterRemovalRR)
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
