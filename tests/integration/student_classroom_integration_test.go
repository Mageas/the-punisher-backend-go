//go:build integration

package integration

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
	. "github.com/mageas/the-punisher-backend/internal/service"
)

func TestStudentService_CRUD_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	svc := NewStudentService(repo)

	created, err := svc.CreateStudent(ctx, user.ID, dto.RequestStudentDto{FirstName: "Alice", LastName: "DUPONT"})
	if err != nil {
		t.Fatalf("CreateStudent returned error: %v", err)
	}
	if created.ID == uuid.Nil || created.FirstName != "Alice" {
		t.Fatalf("unexpected created student: %+v", created)
	}

	got, err := svc.GetStudent(ctx, user.ID, created.ID)
	if err != nil {
		t.Fatalf("GetStudent returned error: %v", err)
	}
	if got.ID != created.ID {
		t.Fatalf("unexpected student id: %s", got.ID)
	}

	search := "Alice"
	list, total, err := svc.ListStudents(ctx, user.ID, &search, 20, 0)
	if err != nil {
		t.Fatalf("ListStudents returned error: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("expected one student, got total=%d len=%d", total, len(list))
	}

	first := "Alicia"
	last := "MARTIN"
	updated, err := svc.UpdateStudent(ctx, user.ID, created.ID, dto.UpdateStudentDto{FirstName: &first, LastName: &last})
	if err != nil {
		t.Fatalf("UpdateStudent returned error: %v", err)
	}
	if updated.FirstName != "Alicia" || updated.LastName != "MARTIN" {
		t.Fatalf("unexpected updated student: %+v", updated)
	}

	if err := svc.DeleteStudent(ctx, user.ID, created.ID); err != nil {
		t.Fatalf("DeleteStudent returned error: %v", err)
	}

	_, err = svc.GetStudent(ctx, user.ID, created.ID)
	if !errors.Is(err, api.ErrStudentNotFound) {
		t.Fatalf("expected ErrStudentNotFound, got %v", err)
	}
}

func TestStudentService_DeleteAllStudents_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	otherUser := mustCreateUserRecord(t, repo, ctx)

	student1, err := repo.CreateStudent(ctx, repository.CreateStudentParams{
		UserID:    user.ID,
		FirstName: "Alice",
		LastName:  "DUPONT",
	})
	if err != nil {
		t.Fatalf("failed to create first student fixture: %v", err)
	}
	if _, err := repo.CreateStudent(ctx, repository.CreateStudentParams{
		UserID:    user.ID,
		FirstName: "Bob",
		LastName:  "MARTIN",
	}); err != nil {
		t.Fatalf("failed to create second student fixture: %v", err)
	}
	if _, err := repo.CreateStudent(ctx, repository.CreateStudentParams{
		UserID:    otherUser.ID,
		FirstName: "Charlie",
		LastName:  "DURAND",
	}); err != nil {
		t.Fatalf("failed to create other-user student fixture: %v", err)
	}

	svc := NewStudentService(repo)
	if err := svc.DeleteAllStudents(ctx, user.ID); err != nil {
		t.Fatalf("DeleteAllStudents returned error: %v", err)
	}

	students, total, err := svc.ListStudents(ctx, user.ID, nil, 20, 0)
	if err != nil {
		t.Fatalf("ListStudents returned error after bulk delete: %v", err)
	}
	if total != 0 || len(students) != 0 {
		t.Fatalf("expected no student for deleted user, got total=%d len=%d", total, len(students))
	}

	_, err = svc.GetStudent(ctx, user.ID, student1.ID)
	if !errors.Is(err, api.ErrStudentNotFound) {
		t.Fatalf("expected ErrStudentNotFound for deleted student, got %v", err)
	}

	otherUserStudents, otherUserTotal, err := svc.ListStudents(ctx, otherUser.ID, nil, 20, 0)
	if err != nil {
		t.Fatalf("ListStudents returned error for other user: %v", err)
	}
	if otherUserTotal != 1 || len(otherUserStudents) != 1 {
		t.Fatalf("expected other user students to remain, got total=%d len=%d", otherUserTotal, len(otherUserStudents))
	}
}

func TestStudentService_ImportStudents_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	existingClassroom, err := repo.CreateClassroom(ctx, repository.CreateClassroomParams{
		UserID: user.ID,
		Name:   "6A",
	})
	if err != nil {
		t.Fatalf("failed to create existing classroom fixture: %v", err)
	}
	existingStudent, err := repo.CreateStudent(ctx, repository.CreateStudentParams{
		UserID:    user.ID,
		FirstName: "Jean",
		LastName:  "DUPONT",
	})
	if err != nil {
		t.Fatalf("failed to create existing student fixture: %v", err)
	}
	if _, err := repo.AddStudentToClassroom(ctx, repository.AddStudentToClassroomParams{
		StudentID:   existingStudent.ID,
		ClassroomID: existingClassroom.ID,
		UserID:      user.ID,
	}); err != nil {
		t.Fatalf("failed to create existing student-classroom link fixture: %v", err)
	}

	csvContent := strings.Join([]string{
		"eleves,classes",
		"DUPONT Jean,6A;6B",
		"MARTIN Alice,6B",
	}, "\n")

	svc := NewStudentService(repo)
	result, err := svc.ImportStudents(ctx, user.ID, strings.NewReader(csvContent), "students.csv")
	if err != nil {
		t.Fatalf("ImportStudents returned error: %v", err)
	}

	if result.Summary.RowsTotal != 2 || result.Summary.RowsProcessed != 2 || result.Summary.RowsFailed != 0 {
		t.Fatalf("unexpected rows summary: %+v", result.Summary)
	}
	if result.Summary.ClassroomsCreated != 1 || result.Summary.ClassroomsExisting != 1 {
		t.Fatalf("unexpected classroom summary: %+v", result.Summary)
	}
	if result.Summary.StudentsCreated != 1 || result.Summary.StudentsExisting != 1 {
		t.Fatalf("unexpected student summary: %+v", result.Summary)
	}
	if result.Summary.LinksCreated != 2 || result.Summary.LinksExisting != 1 {
		t.Fatalf("unexpected links summary: %+v", result.Summary)
	}
	if len(result.Errors) != 0 {
		t.Fatalf("expected no row errors, got %+v", result.Errors)
	}

	importedStudents, totalStudents, err := svc.ListStudents(ctx, user.ID, nil, 20, 0)
	if err != nil {
		t.Fatalf("ListStudents returned error after import: %v", err)
	}
	if totalStudents != 2 || len(importedStudents) != 2 {
		t.Fatalf("expected 2 students after import, got total=%d len=%d", totalStudents, len(importedStudents))
	}

	studentClassCount := map[string]int{}
	for _, student := range importedStudents {
		key := student.LastName + " " + student.FirstName
		studentClassCount[key] = len(student.Classrooms)
	}
	if studentClassCount["DUPONT Jean"] != 2 {
		t.Fatalf("expected DUPONT Jean to be linked to 2 classrooms, got %d", studentClassCount["DUPONT Jean"])
	}
	if studentClassCount["MARTIN Alice"] != 1 {
		t.Fatalf("expected MARTIN Alice to be linked to 1 classroom, got %d", studentClassCount["MARTIN Alice"])
	}
}

func TestStudentService_KpisAndHistory_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	student := mustCreateStudentRecord(t, repo, ctx, user.ID)
	bonusType := mustCreateBonusTypeRecord(t, repo, ctx, user.ID)
	penaltyType := mustCreatePenaltyTypeRecord(t, repo, ctx, user.ID)
	punishmentType := mustCreatePunishmentTypeRecord(t, repo, ctx, user.ID)

	unusedBonus := mustCreateBonusRecord(t, repo, ctx, user.ID, student.ID, bonusType.ID, 5)
	usedBonus := mustCreateBonusRecord(t, repo, ctx, user.ID, student.ID, bonusType.ID, 2)
	if _, err := repo.UseBonus(ctx, repository.UseBonusParams{ID: usedBonus.ID, UserID: user.ID}); err != nil {
		t.Fatalf("failed to use bonus fixture: %v", err)
	}

	_ = mustCreatePenaltyRecord(t, repo, ctx, user.ID, student.ID, penaltyType.ID)
	_ = mustCreatePunishmentRecord(t, repo, ctx, user.ID, student.ID, punishmentType.ID, time.Now().UTC().Add(-2*time.Hour))

	svc := NewStudentService(repo)
	kpis, err := svc.GetStudentKpis(ctx, user.ID, student.ID)
	if err != nil {
		t.Fatalf("GetStudentKpis returned error: %v", err)
	}
	if kpis.AvailableBonusPoints != 5 || kpis.TotalBonusPoints != 7 {
		t.Fatalf("unexpected bonus kpis: %+v", kpis)
	}
	if kpis.PenaltyCount != 1 || kpis.TotalPunishmentCount != 1 {
		t.Fatalf("unexpected penalties/punishments kpis: %+v", kpis)
	}
	if kpis.OverduePunishmentCount != 1 || kpis.PendingPunishmentCount != 1 {
		t.Fatalf("unexpected pending/overdue kpis: %+v", kpis)
	}

	history, total, err := svc.ListStudentHistory(ctx, user.ID, student.ID, 20, 0)
	if err != nil {
		t.Fatalf("ListStudentHistory returned error: %v", err)
	}
	if total != 4 || len(history) != 4 {
		t.Fatalf("expected 4 history entries, got total=%d len=%d", total, len(history))
	}

	seenBonus := false
	for _, item := range history {
		if item.Type == "bonus" && item.ID == unusedBonus.ID {
			seenBonus = true
		}
	}
	if !seenBonus {
		t.Fatalf("expected to find created bonus in history")
	}
}

func TestStudentService_History_UsesOccurredAtOrder_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	student := mustCreateStudentRecord(t, repo, ctx, user.ID)
	bonusType := mustCreateBonusTypeRecord(t, repo, ctx, user.ID)
	penaltyType := mustCreatePenaltyTypeRecord(t, repo, ctx, user.ID)

	recentOccurred := time.Now().UTC()
	backdatedOccurred := recentOccurred.AddDate(0, 0, -5)

	recentBonus, err := repo.CreateBonus(ctx, repository.CreateBonusParams{
		UserID:      user.ID,
		StudentID:   student.ID,
		BonusTypeID: bonusType.ID,
		Points:      3,
		OccurredAt:  &recentOccurred,
	})
	if err != nil {
		t.Fatalf("failed to create recent bonus fixture: %v", err)
	}

	backdatedPenalty, err := repo.CreatePenalty(ctx, repository.CreatePenaltyParams{
		UserID:        user.ID,
		StudentID:     student.ID,
		PenaltyTypeID: penaltyType.ID,
		OccurredAt:    &backdatedOccurred,
	})
	if err != nil {
		t.Fatalf("failed to create backdated penalty fixture: %v", err)
	}

	svc := NewStudentService(repo)
	history, total, err := svc.ListStudentHistory(ctx, user.ID, student.ID, 20, 0)
	if err != nil {
		t.Fatalf("ListStudentHistory returned error: %v", err)
	}
	if total != 2 || len(history) != 2 {
		t.Fatalf("expected 2 history entries, got total=%d len=%d", total, len(history))
	}
	if history[0].ID != recentBonus.ID || history[0].Type != "bonus" {
		t.Fatalf("expected bonus with recent occurred_at first, got %+v", history[0])
	}
	assertTimeEqualToPostgresPrecision(t, "history[0].occurred_at", history[0].OccurredAt, recentOccurred)
	if history[1].ID != backdatedPenalty.ID || history[1].Type != "penalty" {
		t.Fatalf("expected backdated penalty second, got %+v", history[1])
	}
}

func TestStudentService_NotFoundBranches_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	svc := NewStudentService(repo)
	userID := uuid.New()
	studentID := uuid.New()

	if _, err := svc.GetStudent(ctx, userID, studentID); !errors.Is(err, api.ErrStudentNotFound) {
		t.Fatalf("expected ErrStudentNotFound from GetStudent, got %v", err)
	}
	if _, err := svc.GetStudentKpis(ctx, userID, studentID); !errors.Is(err, api.ErrStudentNotFound) {
		t.Fatalf("expected ErrStudentNotFound from GetStudentKpis, got %v", err)
	}
	if _, _, err := svc.ListStudentHistory(ctx, userID, studentID, 20, 0); !errors.Is(err, api.ErrStudentNotFound) {
		t.Fatalf("expected ErrStudentNotFound from ListStudentHistory, got %v", err)
	}
	if err := svc.DeleteStudent(ctx, userID, studentID); !errors.Is(err, api.ErrStudentNotFound) {
		t.Fatalf("expected ErrStudentNotFound from DeleteStudent, got %v", err)
	}
}

func TestClassroomService_CRUDMembershipAndList_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	student := mustCreateStudentRecord(t, repo, ctx, user.ID)

	classroomSvc := NewClassroomService(repo)
	studentSvc := NewStudentService(repo)

	year := "2025"
	teacher := "Mme Dupont"
	created, err := classroomSvc.CreateClassroom(ctx, user.ID, dto.RequestClassroomDto{Name: "6A", Year: &year, MainTeacher: &teacher})
	if err != nil {
		t.Fatalf("CreateClassroom returned error: %v", err)
	}

	got, err := classroomSvc.GetClassroom(ctx, user.ID, created.ID)
	if err != nil {
		t.Fatalf("GetClassroom returned error: %v", err)
	}
	if got.Name != "6A" || got.Year == nil || *got.Year != year {
		t.Fatalf("unexpected classroom: %+v", got)
	}

	if err := classroomSvc.AddStudentToClassroom(ctx, user.ID, created.ID, student.ID); err != nil {
		t.Fatalf("AddStudentToClassroom returned error: %v", err)
	}

	students, totalStudents, err := classroomSvc.ListStudentsByClassroom(ctx, user.ID, created.ID, nil, 20, 0)
	if err != nil {
		t.Fatalf("ListStudentsByClassroom returned error: %v", err)
	}
	if totalStudents != 1 || len(students) != 1 {
		t.Fatalf("expected one student in classroom, got total=%d len=%d", totalStudents, len(students))
	}

	classroomsByStudent, totalClassrooms, err := classroomSvc.ListClassroomsByStudent(ctx, user.ID, student.ID, 20, 0)
	if err != nil {
		t.Fatalf("ListClassroomsByStudent returned error: %v", err)
	}
	if totalClassrooms != 1 || len(classroomsByStudent) != 1 {
		t.Fatalf("expected one classroom for student, got total=%d len=%d", totalClassrooms, len(classroomsByStudent))
	}

	studentFromSvc, err := studentSvc.GetStudent(ctx, user.ID, student.ID)
	if err != nil {
		t.Fatalf("GetStudent returned error: %v", err)
	}
	if len(studentFromSvc.Classrooms) != 1 {
		t.Fatalf("expected one student-classroom relation, got %d", len(studentFromSvc.Classrooms))
	}

	if err := classroomSvc.RemoveStudentFromClassroom(ctx, user.ID, created.ID, student.ID); err != nil {
		t.Fatalf("RemoveStudentFromClassroom returned error: %v", err)
	}
	if err := classroomSvc.RemoveStudentFromClassroom(ctx, user.ID, created.ID, student.ID); !errors.Is(err, api.ErrStudentOrClassroomNotFound) {
		t.Fatalf("expected ErrStudentOrClassroomNotFound, got %v", err)
	}

	newName := "6A - updated"
	updated, err := classroomSvc.UpdateClassroom(ctx, user.ID, created.ID, dto.UpdateClassroomDto{Name: &newName})
	if err != nil {
		t.Fatalf("UpdateClassroom returned error: %v", err)
	}
	if updated.Name != newName {
		t.Fatalf("expected updated name %q, got %q", newName, updated.Name)
	}

	if err := classroomSvc.DeleteClassroom(ctx, user.ID, created.ID); err != nil {
		t.Fatalf("DeleteClassroom returned error: %v", err)
	}
	if err := classroomSvc.DeleteClassroom(ctx, user.ID, created.ID); !errors.Is(err, api.ErrClassroomNotFound) {
		t.Fatalf("expected ErrClassroomNotFound, got %v", err)
	}
}

func TestClassroomService_ListStudentsByClassroom_Search_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	classroom := mustCreateClassroomRecord(t, repo, ctx, user.ID)

	alice, err := repo.CreateStudent(ctx, repository.CreateStudentParams{
		UserID:    user.ID,
		FirstName: "Alice",
		LastName:  "DUPONT",
	})
	if err != nil {
		t.Fatalf("failed to create Alice fixture: %v", err)
	}
	bob, err := repo.CreateStudent(ctx, repository.CreateStudentParams{
		UserID:    user.ID,
		FirstName: "Bob",
		LastName:  "MARTIN",
	})
	if err != nil {
		t.Fatalf("failed to create Bob fixture: %v", err)
	}

	if _, err := repo.AddStudentToClassroom(ctx, repository.AddStudentToClassroomParams{
		StudentID:   alice.ID,
		ClassroomID: classroom.ID,
		UserID:      user.ID,
	}); err != nil {
		t.Fatalf("failed to link Alice to classroom: %v", err)
	}
	if _, err := repo.AddStudentToClassroom(ctx, repository.AddStudentToClassroomParams{
		StudentID:   bob.ID,
		ClassroomID: classroom.ID,
		UserID:      user.ID,
	}); err != nil {
		t.Fatalf("failed to link Bob to classroom: %v", err)
	}

	svc := NewClassroomService(repo)
	search := "alice dup"
	students, total, err := svc.ListStudentsByClassroom(ctx, user.ID, classroom.ID, &search, 20, 0)
	if err != nil {
		t.Fatalf("ListStudentsByClassroom (with search) returned error: %v", err)
	}
	if total != 1 || len(students) != 1 {
		t.Fatalf("expected one student with search, got total=%d len=%d", total, len(students))
	}
	if students[0].FirstName != "Alice" || students[0].LastName != "DUPONT" {
		t.Fatalf("unexpected searched student: %+v", students[0])
	}
}

func TestClassroomService_ListClassrooms_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	student := mustCreateStudentRecord(t, repo, ctx, user.ID)

	svc := NewClassroomService(repo)
	firstClassroom, err := repo.CreateClassroom(ctx, repository.CreateClassroomParams{
		UserID: user.ID,
		Name:   "6A",
	})
	if err != nil {
		t.Fatalf("failed to create first classroom fixture: %v", err)
	}
	secondClassroom, err := repo.CreateClassroom(ctx, repository.CreateClassroomParams{
		UserID: user.ID,
		Name:   "6B",
	})
	if err != nil {
		t.Fatalf("failed to create second classroom fixture: %v", err)
	}
	if _, err := repo.AddStudentToClassroom(ctx, repository.AddStudentToClassroomParams{
		StudentID:   student.ID,
		ClassroomID: firstClassroom.ID,
		UserID:      user.ID,
	}); err != nil {
		t.Fatalf("failed to create student-classroom relation fixture: %v", err)
	}

	classrooms, total, err := svc.ListClassrooms(ctx, user.ID, nil, 20, 0)
	if err != nil {
		t.Fatalf("ListClassrooms returned error: %v", err)
	}
	if total != 2 || len(classrooms) != 2 {
		t.Fatalf("expected 2 classrooms, got total=%d len=%d", total, len(classrooms))
	}

	classroomsByID := make(map[uuid.UUID]*dto.ReturnClassroomDto, len(classrooms))
	for _, classroom := range classrooms {
		classroomsByID[classroom.ID] = classroom
	}

	firstListed, found := classroomsByID[firstClassroom.ID]
	if !found {
		t.Fatalf("first classroom missing from list")
	}
	if firstListed.StudentCount != 1 || len(firstListed.StudentsPreview) != 1 {
		t.Fatalf("unexpected first classroom counters/preview: %+v", firstListed)
	}

	secondListed, found := classroomsByID[secondClassroom.ID]
	if !found {
		t.Fatalf("second classroom missing from list")
	}
	if secondListed.StudentCount != 0 || len(secondListed.StudentsPreview) != 0 {
		t.Fatalf("unexpected second classroom counters/preview: %+v", secondListed)
	}

	search := "6A"
	filtered, filteredTotal, err := svc.ListClassrooms(ctx, user.ID, &search, 20, 0)
	if err != nil {
		t.Fatalf("ListClassrooms (with search) returned error: %v", err)
	}
	if filteredTotal != 1 || len(filtered) != 1 {
		t.Fatalf("expected one classroom with search, got total=%d len=%d", filteredTotal, len(filtered))
	}
	if filtered[0].ID != firstClassroom.ID {
		t.Fatalf("unexpected searched classroom: %+v", filtered[0])
	}
}

func TestClassroomService_AddStudentDuplicateRelation_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	student := mustCreateStudentRecord(t, repo, ctx, user.ID)
	classroom := mustCreateClassroomRecord(t, repo, ctx, user.ID)

	svc := NewClassroomService(repo)
	if err := svc.AddStudentToClassroom(ctx, user.ID, classroom.ID, student.ID); err != nil {
		t.Fatalf("first AddStudentToClassroom returned error: %v", err)
	}

	if err := svc.AddStudentToClassroom(ctx, user.ID, classroom.ID, student.ID); !errors.Is(err, api.ErrStudentClassroomRelationExists) {
		t.Fatalf("expected ErrStudentClassroomRelationExists, got %v", err)
	}
}

func TestClassroomService_KpisAndBulkDelete_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	student := mustCreateStudentRecord(t, repo, ctx, user.ID)
	classroom := mustCreateClassroomRecord(t, repo, ctx, user.ID)
	if _, err := repo.AddStudentToClassroom(ctx, repository.AddStudentToClassroomParams{StudentID: student.ID, ClassroomID: classroom.ID, UserID: user.ID}); err != nil {
		t.Fatalf("failed to add relation fixture: %v", err)
	}

	bonusType := mustCreateBonusTypeRecord(t, repo, ctx, user.ID)
	penaltyType := mustCreatePenaltyTypeRecord(t, repo, ctx, user.ID)
	_ = mustCreateBonusRecord(t, repo, ctx, user.ID, student.ID, bonusType.ID, 3)
	_ = mustCreatePenaltyRecord(t, repo, ctx, user.ID, student.ID, penaltyType.ID)

	svc := NewClassroomService(repo)
	kpis, err := svc.GetClassroomKpis(ctx, user.ID, classroom.ID)
	if err != nil {
		t.Fatalf("GetClassroomKpis returned error: %v", err)
	}
	if kpis.StudentCount != 1 || kpis.PenaltyCount != 1 {
		t.Fatalf("unexpected classroom kpis: %+v", kpis)
	}

	classroom2 := mustCreateClassroomRecord(t, repo, ctx, user.ID)
	if err := svc.DeleteAllClassrooms(ctx, user.ID); err != nil {
		t.Fatalf("DeleteAllClassrooms returned error: %v", err)
	}

	_, err = svc.GetClassroom(ctx, user.ID, classroom2.ID)
	if !errors.Is(err, api.ErrClassroomNotFound) {
		t.Fatalf("expected ErrClassroomNotFound after bulk delete, got %v", err)
	}
}

func TestClassroomService_NotFoundBranches_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	svc := NewClassroomService(repo)
	userID := uuid.New()
	classroomID := uuid.New()
	studentID := uuid.New()

	if err := svc.AddStudentToClassroom(ctx, userID, classroomID, studentID); !errors.Is(err, api.ErrStudentOrClassroomNotFound) {
		t.Fatalf("expected ErrStudentOrClassroomNotFound from AddStudentToClassroom, got %v", err)
	}
	if _, _, err := svc.ListStudentsByClassroom(ctx, userID, classroomID, nil, 20, 0); !errors.Is(err, api.ErrClassroomNotFound) {
		t.Fatalf("expected ErrClassroomNotFound from ListStudentsByClassroom, got %v", err)
	}
	if _, _, err := svc.ListClassroomsByStudent(ctx, userID, studentID, 20, 0); !errors.Is(err, api.ErrStudentNotFound) {
		t.Fatalf("expected ErrStudentNotFound from ListClassroomsByStudent, got %v", err)
	}
}
