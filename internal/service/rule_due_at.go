package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

const (
	ruleDueAtModeDays        = "days"
	ruleDueAtModeNextLessons = "next_lessons"

	ruleDueAtAfterLessonsMin = 1
	ruleDueAtAfterLessonsMax = nextLessonsLimit
)

type nextLessonOccurrence struct {
	Date      time.Time
	StartTime string
	EndTime   string
}

func listNextLessonOccurrences(
	ctx context.Context,
	repo repository.Querier,
	userID, classroomID uuid.UUID,
	now time.Time,
	limit int,
	location *time.Location,
) ([]nextLessonOccurrence, error) {
	if limit <= 0 {
		return []nextLessonOccurrence{}, nil
	}

	slots, err := repo.ListScheduleSlotsByClassroom(ctx, repository.ListScheduleSlotsByClassroomParams{
		UserID:      userID,
		ClassroomID: classroomID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list schedule slots by classroom: %w", err)
	}

	exceptions, err := repo.ListScheduleExceptionsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list schedule exceptions: %w", err)
	}

	slotsByWeekday := make(map[int32][]repository.ScheduleSlot)
	for _, slot := range slots {
		slotsByWeekday[slot.Weekday] = append(slotsByWeekday[slot.Weekday], slot)
	}

	exceptionRanges := make([]scheduleDateRange, 0, len(exceptions))
	for _, exception := range exceptions {
		exceptionRanges = append(exceptionRanges, scheduleDateRange{
			Start: normalizeScheduleDate(exception.StartDate, location),
			End:   normalizeScheduleDate(exception.EndDate, location),
		})
	}

	currentDay := startOfDayForInstant(now, location)
	results := make([]nextLessonOccurrence, 0, limit)
	for day, scanned := currentDay.AddDate(0, 0, 1), 0; len(results) < limit && scanned < nextLessonsMaxDaysScan; day, scanned = day.AddDate(0, 0, 1), scanned+1 {
		if isDateBlockedByScheduleException(day, exceptionRanges) {
			continue
		}

		weekday := isoWeekdayFromTime(day.Weekday())
		daySlots := slotsByWeekday[weekday]
		if len(daySlots) == 0 {
			continue
		}

		_, isoWeek := day.ISOWeek()
		for _, slot := range daySlots {
			if !scheduleSlotAppliesToISOWeek(slot.WeekPattern, isoWeek) {
				continue
			}

			results = append(results, nextLessonOccurrence{
				Date:      day,
				StartTime: slot.StartTime,
				EndTime:   slot.EndTime,
			})
			if len(results) == limit {
				break
			}
		}
	}

	return results, nil
}

func nextLessonOccurrencesToDTO(lessons []nextLessonOccurrence) []dto.NextLessonDto {
	results := make([]dto.NextLessonDto, 0, len(lessons))
	for _, lesson := range lessons {
		results = append(results, dto.NextLessonDto{
			Date:      lesson.Date.Format(scheduleDateLayout),
			StartTime: lesson.StartTime,
			EndTime:   lesson.EndTime,
		})
	}

	return results
}

func computeRuleDueAt(
	ctx context.Context,
	repo repository.Querier,
	userID uuid.UUID,
	rule repository.Rule,
	classroomID *uuid.UUID,
	now time.Time,
	location *time.Location,
) (time.Time, error) {
	switch rule.DueAtMode {
	case "", ruleDueAtModeDays:
		if rule.DueAtAfterDays == nil {
			return time.Time{}, api.ErrRuleDueAtNotComputable
		}

		return now.In(location).AddDate(0, 0, int(*rule.DueAtAfterDays)).UTC(), nil
	case ruleDueAtModeNextLessons:
		if classroomID == nil || rule.DueAtAfterLessons == nil {
			return time.Time{}, api.ErrRuleDueAtNotComputable
		}

		lessonCount := int(*rule.DueAtAfterLessons)
		if lessonCount < ruleDueAtAfterLessonsMin || lessonCount > ruleDueAtAfterLessonsMax {
			return time.Time{}, api.ErrRuleDueAtNotComputable
		}

		lessons, err := listNextLessonOccurrences(ctx, repo, userID, *classroomID, now, lessonCount, location)
		if err != nil {
			return time.Time{}, err
		}
		if len(lessons) < lessonCount {
			return time.Time{}, api.ErrRuleDueAtNotComputable
		}

		return lessonStartsAt(lessons[lessonCount-1], location)
	default:
		return time.Time{}, api.ErrRuleDueAtNotComputable
	}
}

func resolvePunishmentClassroomID(
	ctx context.Context,
	repo repository.Querier,
	userID, studentID uuid.UUID,
	requestedClassroomID *uuid.UUID,
) (*uuid.UUID, error) {
	if requestedClassroomID != nil {
		if _, err := repo.GetClassroomByUser(ctx, repository.GetClassroomByUserParams{
			ID:     *requestedClassroomID,
			UserID: userID,
		}); err != nil {
			if errors.Is(err, repository.ErrNoRows) {
				return nil, api.ErrClassroomNotFound
			}
			return nil, fmt.Errorf("failed to get classroom: %w", err)
		}
	}

	rows, err := repo.ListClassroomRefsByStudentIDs(ctx, repository.ListClassroomRefsByStudentIDsParams{
		UserID:     userID,
		StudentIds: []uuid.UUID{studentID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list classrooms by student: %w", err)
	}

	if requestedClassroomID != nil {
		for _, row := range rows {
			if row.ClassroomID == *requestedClassroomID {
				classroomID := row.ClassroomID
				return &classroomID, nil
			}
		}

		return nil, api.ErrPunishmentStudentNotInClassroom
	}

	if len(rows) != 1 {
		return nil, api.ErrPunishmentClassroomNotResolved
	}

	classroomID := rows[0].ClassroomID
	return &classroomID, nil
}

func lessonStartsAt(lesson nextLessonOccurrence, location *time.Location) (time.Time, error) {
	startTime, err := time.Parse(scheduleTimeLayout, lesson.StartTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse schedule start time: %w", err)
	}

	return time.Date(
		lesson.Date.Year(),
		lesson.Date.Month(),
		lesson.Date.Day(),
		startTime.Hour(),
		startTime.Minute(),
		0,
		0,
		location,
	), nil
}
