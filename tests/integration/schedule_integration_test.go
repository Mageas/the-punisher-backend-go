//go:build integration

package integration

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	. "github.com/mageas/the-punisher-backend/internal/service"
)

func TestScheduleService_ScheduleSlotCRUD_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	classroomA := mustCreateClassroomRecord(t, repo, ctx, user.ID)
	classroomB := mustCreateClassroomRecord(t, repo, ctx, user.ID)

	svc := NewScheduleService(repo)

	_, err := svc.CreateScheduleSlot(ctx, user.ID, dto.RequestScheduleSlotDto{
		Weekday:      "monday",
		StartTime:    "08:00",
		EndTime:      "09:00",
		WeekPattern:  "every_week",
		ClassroomIDs: []string{},
	})
	assertAPIError(t, err, http.StatusBadRequest, "validation_failed")

	created, err := svc.CreateScheduleSlot(ctx, user.ID, dto.RequestScheduleSlotDto{
		Weekday:      "monday",
		StartTime:    "08:15",
		EndTime:      "09:00",
		WeekPattern:  "every_week",
		ClassroomIDs: []string{classroomA.ID.String(), classroomA.ID.String(), classroomB.ID.String()},
	})
	if err != nil {
		t.Fatalf("CreateScheduleSlot returned error: %v", err)
	}
	if created.Weekday != "monday" || created.StartTime != "08:15" || created.EndTime != "09:00" {
		t.Fatalf("unexpected created schedule slot: %+v", created)
	}
	if len(created.Classrooms) != 2 {
		t.Fatalf("expected 2 unique classrooms on created slot, got %d", len(created.Classrooms))
	}

	listed, err := svc.ListScheduleSlots(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListScheduleSlots returned error: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 schedule slot, got %d", len(listed))
	}

	updatedClassroomIDs := []string{classroomB.ID.String()}
	updated, err := svc.UpdateScheduleSlot(ctx, user.ID, created.ID, dto.UpdateScheduleSlotDto{
		Weekday:      ptr("tuesday"),
		StartTime:    ptr("14:15"),
		EndTime:      ptr("15:00"),
		WeekPattern:  ptr("odd_weeks"),
		ClassroomIDs: &updatedClassroomIDs,
	})
	if err != nil {
		t.Fatalf("UpdateScheduleSlot returned error: %v", err)
	}
	if updated.Weekday != "tuesday" || updated.StartTime != "14:15" || updated.EndTime != "15:00" || updated.WeekPattern != "odd_weeks" {
		t.Fatalf("unexpected updated schedule slot: %+v", updated)
	}
	if len(updated.Classrooms) != 1 || updated.Classrooms[0].ID != classroomB.ID {
		t.Fatalf("expected updated slot to keep only classroom B, got %+v", updated.Classrooms)
	}

	got, err := svc.GetScheduleSlot(ctx, user.ID, created.ID)
	if err != nil {
		t.Fatalf("GetScheduleSlot returned error: %v", err)
	}
	if got.ID != created.ID || got.Weekday != "tuesday" {
		t.Fatalf("unexpected retrieved schedule slot: %+v", got)
	}

	if err := svc.DeleteScheduleSlot(ctx, user.ID, created.ID); err != nil {
		t.Fatalf("DeleteScheduleSlot returned error: %v", err)
	}

	_, err = svc.GetScheduleSlot(ctx, user.ID, created.ID)
	if !errors.Is(err, api.ErrScheduleSlotNotFound) {
		t.Fatalf("expected ErrScheduleSlotNotFound after delete, got %v", err)
	}
}

func TestScheduleService_ScheduleSlotConflicts_WeekPatternsAndAdjacency_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	classroomA := mustCreateClassroomRecord(t, repo, ctx, user.ID)
	classroomB := mustCreateClassroomRecord(t, repo, ctx, user.ID)

	svc := NewScheduleService(repo)

	if _, err := svc.CreateScheduleSlot(ctx, user.ID, dto.RequestScheduleSlotDto{
		Weekday:      "monday",
		StartTime:    "10:00",
		EndTime:      "11:00",
		WeekPattern:  "even_weeks",
		ClassroomIDs: []string{classroomA.ID.String()},
	}); err != nil {
		t.Fatalf("failed to create even-weeks slot: %v", err)
	}

	if _, err := svc.CreateScheduleSlot(ctx, user.ID, dto.RequestScheduleSlotDto{
		Weekday:      "monday",
		StartTime:    "10:00",
		EndTime:      "11:00",
		WeekPattern:  "odd_weeks",
		ClassroomIDs: []string{classroomB.ID.String()},
	}); err != nil {
		t.Fatalf("expected even/odd slots to coexist, got %v", err)
	}

	_, err := svc.CreateScheduleSlot(ctx, user.ID, dto.RequestScheduleSlotDto{
		Weekday:      "monday",
		StartTime:    "10:00",
		EndTime:      "11:00",
		WeekPattern:  "every_week",
		ClassroomIDs: []string{classroomA.ID.String()},
	})
	if !errors.Is(err, api.ErrScheduleSlotConflict) {
		t.Fatalf("expected every-week slot to conflict with even/odd slots, got %v", err)
	}

	_, err = svc.CreateScheduleSlot(ctx, user.ID, dto.RequestScheduleSlotDto{
		Weekday:      "monday",
		StartTime:    "10:30",
		EndTime:      "11:30",
		WeekPattern:  "even_weeks",
		ClassroomIDs: []string{classroomB.ID.String()},
	})
	if !errors.Is(err, api.ErrScheduleSlotConflict) {
		t.Fatalf("expected overlapping even-weeks slot to conflict, got %v", err)
	}

	adjacent, err := svc.CreateScheduleSlot(ctx, user.ID, dto.RequestScheduleSlotDto{
		Weekday:      "monday",
		StartTime:    "11:00",
		EndTime:      "12:00",
		WeekPattern:  "every_week",
		ClassroomIDs: []string{classroomA.ID.String()},
	})
	if err != nil {
		t.Fatalf("expected adjacent schedule slot to be allowed, got %v", err)
	}
	if adjacent.StartTime != "11:00" || adjacent.EndTime != "12:00" {
		t.Fatalf("unexpected adjacent slot payload: %+v", adjacent)
	}
}

func TestScheduleService_ScheduleExceptionsCRUDAndOverlap_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	svc := NewScheduleService(repo)

	first, err := svc.CreateScheduleException(ctx, user.ID, dto.RequestScheduleExceptionDto{
		Type:      "vacation",
		StartDate: "2026-04-01",
		EndDate:   "2026-04-05",
	})
	if err != nil {
		t.Fatalf("CreateScheduleException returned error: %v", err)
	}
	if first.Type != "vacation" || first.StartDate != "2026-04-01" || first.EndDate != "2026-04-05" {
		t.Fatalf("unexpected created schedule exception: %+v", first)
	}

	second, err := svc.CreateScheduleException(ctx, user.ID, dto.RequestScheduleExceptionDto{
		Type:      "public_holiday",
		StartDate: "2026-04-06",
		EndDate:   "2026-04-06",
	})
	if err != nil {
		t.Fatalf("expected non-overlapping exception to be created, got %v", err)
	}

	_, err = svc.CreateScheduleException(ctx, user.ID, dto.RequestScheduleExceptionDto{
		Type:      "public_holiday",
		StartDate: "2026-04-05",
		EndDate:   "2026-04-07",
	})
	if !errors.Is(err, api.ErrScheduleExceptionOverlap) {
		t.Fatalf("expected overlapping exception creation to fail, got %v", err)
	}

	listed, err := svc.ListScheduleExceptions(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListScheduleExceptions returned error: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("expected 2 schedule exceptions, got %d", len(listed))
	}

	_, err = svc.UpdateScheduleException(ctx, user.ID, second.ID, dto.UpdateScheduleExceptionDto{
		StartDate: ptr("2026-04-04"),
		EndDate:   ptr("2026-04-06"),
	})
	if !errors.Is(err, api.ErrScheduleExceptionOverlap) {
		t.Fatalf("expected overlapping exception update to fail, got %v", err)
	}

	if err := svc.DeleteScheduleException(ctx, user.ID, first.ID); err != nil {
		t.Fatalf("DeleteScheduleException returned error: %v", err)
	}

	_, err = svc.GetScheduleException(ctx, user.ID, first.ID)
	if !errors.Is(err, api.ErrScheduleExceptionNotFound) {
		t.Fatalf("expected ErrScheduleExceptionNotFound after delete, got %v", err)
	}
}

func TestClassroomService_DeleteClassroom_CleansScheduleRelationsAndOrphans_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	classroomA := mustCreateClassroomRecord(t, repo, ctx, user.ID)
	classroomB := mustCreateClassroomRecord(t, repo, ctx, user.ID)

	scheduleSvc := NewScheduleService(repo)
	classroomSvc := NewClassroomService(repo)

	sharedSlot := mustCreateScheduleSlot(t, scheduleSvc, ctx, user.ID, dto.RequestScheduleSlotDto{
		Weekday:      "monday",
		StartTime:    "08:00",
		EndTime:      "09:00",
		WeekPattern:  "every_week",
		ClassroomIDs: []string{classroomA.ID.String(), classroomB.ID.String()},
	})
	orphanSlot := mustCreateScheduleSlot(t, scheduleSvc, ctx, user.ID, dto.RequestScheduleSlotDto{
		Weekday:      "tuesday",
		StartTime:    "09:30",
		EndTime:      "10:30",
		WeekPattern:  "every_week",
		ClassroomIDs: []string{classroomA.ID.String()},
	})

	if err := classroomSvc.DeleteClassroom(ctx, user.ID, classroomA.ID); err != nil {
		t.Fatalf("DeleteClassroom returned error: %v", err)
	}

	sharedAfterDelete, err := scheduleSvc.GetScheduleSlot(ctx, user.ID, sharedSlot.ID)
	if err != nil {
		t.Fatalf("expected shared slot to remain after classroom deletion, got %v", err)
	}
	if len(sharedAfterDelete.Classrooms) != 1 || sharedAfterDelete.Classrooms[0].ID != classroomB.ID {
		t.Fatalf("expected shared slot to keep only classroom B, got %+v", sharedAfterDelete.Classrooms)
	}

	_, err = scheduleSvc.GetScheduleSlot(ctx, user.ID, orphanSlot.ID)
	if !errors.Is(err, api.ErrScheduleSlotNotFound) {
		t.Fatalf("expected orphan slot to be deleted after classroom deletion, got %v", err)
	}

	remainingSlots, err := scheduleSvc.ListScheduleSlots(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListScheduleSlots returned error after classroom deletion: %v", err)
	}
	if len(remainingSlots) != 1 || remainingSlots[0].ID != sharedSlot.ID {
		t.Fatalf("unexpected remaining schedule slots: %+v", remainingSlots)
	}
}

func TestScheduleService_ListNextLessons_RespectsExceptionsAndParity_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := mustCreateUserRecord(t, repo, ctx)
	targetClassroom := mustCreateClassroomRecord(t, repo, ctx, user.ID)
	otherClassroom := mustCreateClassroomRecord(t, repo, ctx, user.ID)

	svc := NewScheduleService(repo)

	now := time.Now().In(time.Local)
	today := startOfLocalScheduleDay(now)
	tomorrow := today.AddDate(0, 0, 1)
	tomorrowWeekday := weekdayTextFromTime(tomorrow.Weekday())
	mustCreateScheduleSlot(t, svc, ctx, user.ID, dto.RequestScheduleSlotDto{
		Weekday:      tomorrowWeekday,
		StartTime:    "09:00",
		EndTime:      "10:00",
		WeekPattern:  "every_week",
		ClassroomIDs: []string{targetClassroom.ID.String()},
	})
	mustCreateScheduleSlot(t, svc, ctx, user.ID, dto.RequestScheduleSlotDto{
		Weekday:      tomorrowWeekday,
		StartTime:    "11:00",
		EndTime:      "12:00",
		WeekPattern:  "even_weeks",
		ClassroomIDs: []string{targetClassroom.ID.String()},
	})
	mustCreateScheduleSlot(t, svc, ctx, user.ID, dto.RequestScheduleSlotDto{
		Weekday:      tomorrowWeekday,
		StartTime:    "13:00",
		EndTime:      "14:00",
		WeekPattern:  "odd_weeks",
		ClassroomIDs: []string{targetClassroom.ID.String()},
	})
	mustCreateScheduleSlot(t, svc, ctx, user.ID, dto.RequestScheduleSlotDto{
		Weekday:      tomorrowWeekday,
		StartTime:    "16:00",
		EndTime:      "17:00",
		WeekPattern:  "every_week",
		ClassroomIDs: []string{otherClassroom.ID.String()},
	})

	_, err := svc.CreateScheduleException(ctx, user.ID, dto.RequestScheduleExceptionDto{
		Type:      "vacation",
		StartDate: tomorrow.Format("2006-01-02"),
		EndDate:   tomorrow.Format("2006-01-02"),
	})
	if err != nil {
		t.Fatalf("failed to create blocking exception: %v", err)
	}

	nextLessons, err := svc.ListNextLessons(ctx, user.ID, targetClassroom.ID)
	if err != nil {
		t.Fatalf("ListNextLessons returned error: %v", err)
	}
	if len(nextLessons) != 5 {
		t.Fatalf("expected 5 next lessons, got %d", len(nextLessons))
	}

	firstWeeklyDate := tomorrow.AddDate(0, 0, 7)
	secondWeeklyDate := firstWeeklyDate.AddDate(0, 0, 7)
	thirdWeeklyDate := secondWeeklyDate.AddDate(0, 0, 7)
	expected := []dto.NextLessonDto{
		{Date: firstWeeklyDate.Format("2006-01-02"), StartTime: "09:00", EndTime: "10:00"},
		{Date: firstWeeklyDate.Format("2006-01-02"), StartTime: paritySlotStartTime(firstWeeklyDate), EndTime: paritySlotEndTime(firstWeeklyDate)},
		{Date: secondWeeklyDate.Format("2006-01-02"), StartTime: "09:00", EndTime: "10:00"},
		{Date: secondWeeklyDate.Format("2006-01-02"), StartTime: paritySlotStartTime(secondWeeklyDate), EndTime: paritySlotEndTime(secondWeeklyDate)},
		{Date: thirdWeeklyDate.Format("2006-01-02"), StartTime: "09:00", EndTime: "10:00"},
	}

	for i, lesson := range nextLessons {
		if lesson != expected[i] {
			t.Fatalf("unexpected next lesson at index %d: expected %+v, got %+v", i, expected[i], lesson)
		}
		if lesson.Date == today.Format("2006-01-02") {
			t.Fatalf("expected current day to be excluded from next lessons, got %+v", lesson)
		}
		if lesson.Date == tomorrow.Format("2006-01-02") {
			t.Fatalf("expected blocked day to be excluded from next lessons, got %+v", lesson)
		}
		if lesson.StartTime == "16:00" {
			t.Fatalf("expected next lessons to exclude slots from other classrooms, got %+v", lesson)
		}
	}
}

func mustCreateScheduleSlot(
	t *testing.T,
	svc ScheduleService,
	ctx context.Context,
	userID uuid.UUID,
	req dto.RequestScheduleSlotDto,
) *dto.ReturnScheduleSlotDto {
	t.Helper()

	slot, err := svc.CreateScheduleSlot(ctx, userID, req)
	if err != nil {
		t.Fatalf("failed to create schedule slot fixture: %v", err)
	}

	return slot
}

func assertAPIError(t *testing.T, err error, expectedStatus int, expectedMessage string) {
	t.Helper()

	var apiErr *api.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %v", err)
	}
	if apiErr.StatusCode != expectedStatus || apiErr.Message != expectedMessage {
		t.Fatalf("expected APIError status=%d message=%s, got status=%d message=%s", expectedStatus, expectedMessage, apiErr.StatusCode, apiErr.Message)
	}
}

func startOfLocalScheduleDay(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}

func weekdayTextFromTime(value time.Weekday) string {
	switch value {
	case time.Monday:
		return "monday"
	case time.Tuesday:
		return "tuesday"
	case time.Wednesday:
		return "wednesday"
	case time.Thursday:
		return "thursday"
	case time.Friday:
		return "friday"
	case time.Saturday:
		return "saturday"
	default:
		return "sunday"
	}
}

func paritySlotStartTime(value time.Time) string {
	_, isoWeek := value.ISOWeek()
	if isoWeek%2 == 0 {
		return "11:00"
	}

	return "13:00"
}

func paritySlotEndTime(value time.Time) string {
	_, isoWeek := value.ISOWeek()
	if isoWeek%2 == 0 {
		return "12:00"
	}

	return "14:00"
}
