package service

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/repository"
	"github.com/mageas/the-punisher-backend/internal/testutil/inmemory"
	"github.com/xuri/excelize/v2"
)

func TestStudentServiceImportStudentsCSVSuccess(t *testing.T) {
	t.Parallel()

	repo := inmemory.NewRepository()
	svc := NewStudentService(repo)

	userID := uuid.New()
	existingStudentID := uuid.New()
	existingClassroomID := uuid.New()

	repo.SeedStudent(repository.Student{
		ID:        existingStudentID,
		UserID:    userID,
		FirstName: "Anna",
		LastName:  "BRUN",
	})
	repo.SeedClassroom(repository.Classroom{
		ID:     existingClassroomID,
		UserID: userID,
		Name:   "6eme3",
	})
	rowsAffected, err := repo.AddStudentToClassroom(context.Background(), repository.AddStudentToClassroomParams{
		StudentID:   existingStudentID,
		ClassroomID: existingClassroomID,
		UserID:      userID,
	})
	if err != nil || rowsAffected != 1 {
		t.Fatalf("failed to seed existing student-classroom relation: rows=%d err=%v", rowsAffected, err)
	}

	csvPayload := strings.Join([]string{
		"Eleves,Classes",
		"\"BRUN Anna\",\"6eme3;6eme3 latin\"",
		"\"BRUN Anna\",\"6eme3\"",
		"\"DUPONT Jean\",\"6eme3,5eme1\"",
	}, "\n")

	result, err := svc.ImportStudents(context.Background(), userID, strings.NewReader(csvPayload), "students.csv")
	if err != nil {
		t.Fatalf("expected import success, got err=%v", err)
	}

	if result.Summary.RowsTotal != 3 || result.Summary.RowsProcessed != 3 {
		t.Fatalf("unexpected rows summary: %+v", result.Summary)
	}
	if result.Summary.ClassroomsExisting != 1 || result.Summary.ClassroomsCreated != 2 {
		t.Fatalf("unexpected classroom summary: %+v", result.Summary)
	}
	if result.Summary.StudentsExisting != 2 || result.Summary.StudentsCreated != 1 {
		t.Fatalf("unexpected student summary: %+v", result.Summary)
	}
	if result.Summary.LinksCreated != 3 || result.Summary.LinksExisting != 2 {
		t.Fatalf("unexpected links summary: %+v", result.Summary)
	}
	if len(result.Errors) != 0 {
		t.Fatalf("expected no row errors, got %+v", result.Errors)
	}
}

func TestStudentServiceImportStudentsIsIdempotentOnSecondRun(t *testing.T) {
	t.Parallel()

	repo := inmemory.NewRepository()
	svc := NewStudentService(repo)
	userID := uuid.New()

	csvPayload := strings.Join([]string{
		"Eleves,Classes",
		"\"BRUN Anna\",\"6eme3;6eme3 latin\"",
		"\"DUPONT Jean\",\"6eme3,5eme1\"",
	}, "\n")

	firstResult, err := svc.ImportStudents(context.Background(), userID, strings.NewReader(csvPayload), "students.csv")
	if err != nil {
		t.Fatalf("first import failed: %v", err)
	}
	if firstResult.Summary.LinksCreated == 0 {
		t.Fatalf("expected links created on first import, got %+v", firstResult.Summary)
	}

	secondResult, err := svc.ImportStudents(context.Background(), userID, strings.NewReader(csvPayload), "students.csv")
	if err != nil {
		t.Fatalf("second import failed: %v", err)
	}

	if secondResult.Summary.ClassroomsCreated != 0 || secondResult.Summary.StudentsCreated != 0 || secondResult.Summary.LinksCreated != 0 {
		t.Fatalf("expected no creations on second import, got %+v", secondResult.Summary)
	}
	if secondResult.Summary.ClassroomsExisting != 3 {
		t.Fatalf("expected 3 existing classrooms on second import, got %+v", secondResult.Summary)
	}
	if secondResult.Summary.StudentsExisting != 2 {
		t.Fatalf("expected 2 existing students on second import, got %+v", secondResult.Summary)
	}
	if secondResult.Summary.LinksExisting != 4 {
		t.Fatalf("expected 4 existing links on second import, got %+v", secondResult.Summary)
	}
}

func TestStudentServiceImportStudentsValidationFailsBeforeWrite(t *testing.T) {
	t.Parallel()

	repo := inmemory.NewRepository()
	svc := NewStudentService(repo)
	userID := uuid.New()

	csvPayload := strings.Join([]string{
		"Eleves,Classes",
		"\"Anna BRUN\",\"6eme3\"",
	}, "\n")

	_, err := svc.ImportStudents(context.Background(), userID, strings.NewReader(csvPayload), "students.csv")
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	apiErr, ok := err.(*api.APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Message != api.ErrImportValidationFailed.Message {
		t.Fatalf("expected error=%q, got=%q", api.ErrImportValidationFailed.Message, apiErr.Message)
	}
	if len(apiErr.Details) == 0 {
		t.Fatal("expected validation details, got none")
	}
	if apiErr.Details[0].Row == nil || *apiErr.Details[0].Row != 2 {
		t.Fatalf("expected row=2 detail, got %+v", apiErr.Details[0])
	}

	studentCount, err := repo.CountStudentsByUser(context.Background(), repository.CountStudentsByUserParams{
		UserID: userID,
	})
	if err != nil {
		t.Fatalf("failed to count students: %v", err)
	}
	if studentCount != 0 {
		t.Fatalf("expected no student write on validation error, got %d", studentCount)
	}

	classroomCount, err := repo.CountClassroomsByUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("failed to count classrooms: %v", err)
	}
	if classroomCount != 0 {
		t.Fatalf("expected no classroom write on validation error, got %d", classroomCount)
	}
}

func TestStudentServiceImportStudentsXLSXSuccess(t *testing.T) {
	t.Parallel()

	repo := inmemory.NewRepository()
	svc := NewStudentService(repo)
	userID := uuid.New()

	workbook := excelize.NewFile()
	sheetName := workbook.GetSheetName(0)

	if err := workbook.SetCellValue(sheetName, "A1", "Eleves"); err != nil {
		t.Fatalf("failed to set header A1: %v", err)
	}
	if err := workbook.SetCellValue(sheetName, "B1", "Classes"); err != nil {
		t.Fatalf("failed to set header B1: %v", err)
	}
	if err := workbook.SetCellValue(sheetName, "A2", "DURAND Lea"); err != nil {
		t.Fatalf("failed to set value A2: %v", err)
	}
	if err := workbook.SetCellValue(sheetName, "B2", "6eme2;6eme2 latin"); err != nil {
		t.Fatalf("failed to set value B2: %v", err)
	}

	var payload bytes.Buffer
	if err := workbook.Write(&payload); err != nil {
		t.Fatalf("failed to serialize workbook: %v", err)
	}
	if err := workbook.Close(); err != nil {
		t.Fatalf("failed to close workbook: %v", err)
	}

	result, err := svc.ImportStudents(context.Background(), userID, bytes.NewReader(payload.Bytes()), "students.xlsx")
	if err != nil {
		t.Fatalf("expected import success, got err=%v", err)
	}

	if result.Summary.RowsTotal != 1 || result.Summary.RowsProcessed != 1 {
		t.Fatalf("unexpected rows summary: %+v", result.Summary)
	}
	if result.Summary.ClassroomsCreated != 2 || result.Summary.StudentsCreated != 1 {
		t.Fatalf("unexpected creation summary: %+v", result.Summary)
	}
	if result.Summary.LinksCreated != 2 {
		t.Fatalf("unexpected links summary: %+v", result.Summary)
	}
}
