package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mageas/the-punisher-backend/internal/adapter/persistence/sqlcmapper"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

const (
	scheduleTimeLayout     = "15:04"
	scheduleDateLayout     = "2006-01-02"
	nextLessonsLimit       = 5
	nextLessonsMaxDaysScan = 3660
)

type ScheduleService interface {
	CreateScheduleSlot(ctx context.Context, userID uuid.UUID, req dto.RequestScheduleSlotDto) (*dto.ReturnScheduleSlotDto, error)
	GetScheduleSlot(ctx context.Context, userID, scheduleSlotID uuid.UUID) (*dto.ReturnScheduleSlotDto, error)
	ListScheduleSlots(ctx context.Context, userID uuid.UUID) ([]*dto.ReturnScheduleSlotDto, error)
	UpdateScheduleSlot(ctx context.Context, userID, scheduleSlotID uuid.UUID, req dto.UpdateScheduleSlotDto) (*dto.ReturnScheduleSlotDto, error)
	DeleteScheduleSlot(ctx context.Context, userID, scheduleSlotID uuid.UUID) error

	CreateScheduleException(ctx context.Context, userID uuid.UUID, req dto.RequestScheduleExceptionDto) (*dto.ReturnScheduleExceptionDto, error)
	GetScheduleException(ctx context.Context, userID, scheduleExceptionID uuid.UUID) (*dto.ReturnScheduleExceptionDto, error)
	ListScheduleExceptions(ctx context.Context, userID uuid.UUID) ([]*dto.ReturnScheduleExceptionDto, error)
	UpdateScheduleException(ctx context.Context, userID, scheduleExceptionID uuid.UUID, req dto.UpdateScheduleExceptionDto) (*dto.ReturnScheduleExceptionDto, error)
	DeleteScheduleException(ctx context.Context, userID, scheduleExceptionID uuid.UUID) error

	ListNextLessons(ctx context.Context, userID, classroomID uuid.UUID) ([]dto.NextLessonDto, error)
}

type scheduleService struct {
	repo repository.Querier
	now  func() time.Time
}

type transactionalScheduleRepo interface {
	repository.Querier
	WithinTransaction(ctx context.Context, fn func(repository.Querier) error) error
}

type parsedScheduleTime struct {
	Text    string
	Minutes int
	DBValue pgtype.Time
}

type scheduleDateRange struct {
	Start time.Time
	End   time.Time
}

func NewScheduleService(repo repository.Querier) ScheduleService {
	return &scheduleService{
		repo: repo,
		now:  time.Now,
	}
}

func (s *scheduleService) CreateScheduleSlot(ctx context.Context, userID uuid.UUID, req dto.RequestScheduleSlotDto) (*dto.ReturnScheduleSlotDto, error) {
	txRepo, ok := s.repo.(transactionalScheduleRepo)
	if !ok {
		return nil, fmt.Errorf("schedule repository does not support transactions")
	}

	weekday, err := scheduleWeekdayISOFromText(req.Weekday)
	if err != nil {
		return nil, err
	}

	startTime, err := parseScheduleTime(req.StartTime, "start_time")
	if err != nil {
		return nil, err
	}

	endTime, err := parseScheduleTime(req.EndTime, "end_time")
	if err != nil {
		return nil, err
	}

	if err := validateScheduleTimeRange(startTime, endTime); err != nil {
		return nil, err
	}

	classroomIDs, err := parseUniqueUUIDs(req.ClassroomIDs, "classroom_ids")
	if err != nil {
		return nil, err
	}

	var slot repository.ScheduleSlot
	err = txRepo.WithinTransaction(ctx, func(txQuerier repository.Querier) error {
		if err := ensureClassroomsExist(ctx, txQuerier, userID, classroomIDs); err != nil {
			return err
		}

		if err := ensureNoScheduleSlotConflict(ctx, txQuerier, userID, nil, weekday, startTime, endTime, req.WeekPattern); err != nil {
			return err
		}

		createdSlot, createErr := txQuerier.CreateScheduleSlot(ctx, repository.CreateScheduleSlotParams{
			UserID:      userID,
			Weekday:     weekday,
			StartTime:   startTime.Text,
			EndTime:     endTime.Text,
			WeekPattern: req.WeekPattern,
		})
		if createErr != nil {
			return fmt.Errorf("failed to create schedule slot: %w", createErr)
		}

		if err := setScheduleSlotClassrooms(ctx, txQuerier, userID, createdSlot.ID, classroomIDs); err != nil {
			return err
		}

		slot = createdSlot
		return nil
	})
	if err != nil {
		return nil, err
	}

	slog.Info("schedule slot created", "schedule_slot_id", slot.ID, "user_id", userID)

	response := sqlcmapper.ScheduleSlotFromModel(&slot)
	if err := attachClassroomsToScheduleSlots(ctx, s.repo, userID, []*dto.ReturnScheduleSlotDto{response}); err != nil {
		return nil, fmt.Errorf("failed to list schedule slot classrooms: %w", err)
	}

	return response, nil
}

func (s *scheduleService) GetScheduleSlot(ctx context.Context, userID, scheduleSlotID uuid.UUID) (*dto.ReturnScheduleSlotDto, error) {
	slot, err := s.repo.GetScheduleSlotByUser(ctx, repository.GetScheduleSlotByUserParams{
		ID:     scheduleSlotID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrScheduleSlotNotFound
		}
		return nil, fmt.Errorf("failed to get schedule slot: %w", err)
	}

	response := sqlcmapper.ScheduleSlotFromModel(&slot)
	if err := attachClassroomsToScheduleSlots(ctx, s.repo, userID, []*dto.ReturnScheduleSlotDto{response}); err != nil {
		return nil, fmt.Errorf("failed to list schedule slot classrooms: %w", err)
	}

	return response, nil
}

func (s *scheduleService) ListScheduleSlots(ctx context.Context, userID uuid.UUID) ([]*dto.ReturnScheduleSlotDto, error) {
	slots, err := s.repo.ListScheduleSlotsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list schedule slots: %w", err)
	}

	response := sqlcmapper.ScheduleSlotListFromModels(slots)
	if err := attachClassroomsToScheduleSlots(ctx, s.repo, userID, response); err != nil {
		return nil, fmt.Errorf("failed to list schedule slot classrooms: %w", err)
	}

	return response, nil
}

func (s *scheduleService) UpdateScheduleSlot(ctx context.Context, userID, scheduleSlotID uuid.UUID, req dto.UpdateScheduleSlotDto) (*dto.ReturnScheduleSlotDto, error) {
	txRepo, ok := s.repo.(transactionalScheduleRepo)
	if !ok {
		return nil, fmt.Errorf("schedule repository does not support transactions")
	}

	existingSlot, err := s.repo.GetScheduleSlotByUser(ctx, repository.GetScheduleSlotByUserParams{
		ID:     scheduleSlotID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrScheduleSlotNotFound
		}
		return nil, fmt.Errorf("failed to get schedule slot: %w", err)
	}

	weekday := existingSlot.Weekday
	if req.Weekday != nil {
		parsedWeekday, err := scheduleWeekdayISOFromText(*req.Weekday)
		if err != nil {
			return nil, err
		}
		weekday = parsedWeekday
	}

	startTime, err := parseScheduleTime(existingSlot.StartTime, "start_time")
	if err != nil {
		return nil, err
	}
	if req.StartTime != nil {
		startTime, err = parseScheduleTime(*req.StartTime, "start_time")
		if err != nil {
			return nil, err
		}
	}

	endTime, err := parseScheduleTime(existingSlot.EndTime, "end_time")
	if err != nil {
		return nil, err
	}
	if req.EndTime != nil {
		endTime, err = parseScheduleTime(*req.EndTime, "end_time")
		if err != nil {
			return nil, err
		}
	}

	if err := validateScheduleTimeRange(startTime, endTime); err != nil {
		return nil, err
	}

	weekPattern := existingSlot.WeekPattern
	if req.WeekPattern != nil {
		weekPattern = *req.WeekPattern
	}

	var classroomIDs []uuid.UUID
	if req.ClassroomIDs != nil {
		classroomIDs, err = parseUniqueUUIDs(*req.ClassroomIDs, "classroom_ids")
		if err != nil {
			return nil, err
		}
	}

	if err := ensureNoScheduleSlotConflict(ctx, s.repo, userID, &scheduleSlotID, weekday, startTime, endTime, weekPattern); err != nil {
		return nil, err
	}

	var updatedSlot repository.ScheduleSlot
	err = txRepo.WithinTransaction(ctx, func(txQuerier repository.Querier) error {
		if req.ClassroomIDs != nil {
			if err := ensureClassroomsExist(ctx, txQuerier, userID, classroomIDs); err != nil {
				return err
			}
		}

		slot, updateErr := txQuerier.UpdateScheduleSlotByUser(ctx, repository.UpdateScheduleSlotByUserParams{
			Weekday:     &weekday,
			StartTime:   &startTime.Text,
			EndTime:     &endTime.Text,
			WeekPattern: &weekPattern,
			ID:          scheduleSlotID,
			UserID:      userID,
		})
		if updateErr != nil {
			if errors.Is(updateErr, repository.ErrNoRows) {
				return api.ErrScheduleSlotNotFound
			}
			return fmt.Errorf("failed to update schedule slot: %w", updateErr)
		}

		if req.ClassroomIDs != nil {
			if _, err := txQuerier.DeleteScheduleSlotClassroomRelationsBySlot(ctx, repository.DeleteScheduleSlotClassroomRelationsBySlotParams{
				UserID:         userID,
				ScheduleSlotID: scheduleSlotID,
			}); err != nil {
				return fmt.Errorf("failed to clear schedule slot classrooms: %w", err)
			}

			if err := setScheduleSlotClassrooms(ctx, txQuerier, userID, scheduleSlotID, classroomIDs); err != nil {
				return err
			}
		}

		updatedSlot = slot
		return nil
	})
	if err != nil {
		return nil, err
	}

	response := sqlcmapper.ScheduleSlotFromModel(&updatedSlot)
	if err := attachClassroomsToScheduleSlots(ctx, s.repo, userID, []*dto.ReturnScheduleSlotDto{response}); err != nil {
		return nil, fmt.Errorf("failed to list schedule slot classrooms: %w", err)
	}

	return response, nil
}

func (s *scheduleService) DeleteScheduleSlot(ctx context.Context, userID, scheduleSlotID uuid.UUID) error {
	rowsAffected, err := s.repo.DeleteScheduleSlotByUser(ctx, repository.DeleteScheduleSlotByUserParams{
		ID:     scheduleSlotID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete schedule slot: %w", err)
	}

	if rowsAffected == 0 {
		return api.ErrScheduleSlotNotFound
	}

	slog.Info("schedule slot deleted", "schedule_slot_id", scheduleSlotID, "user_id", userID)

	return nil
}

func (s *scheduleService) CreateScheduleException(ctx context.Context, userID uuid.UUID, req dto.RequestScheduleExceptionDto) (*dto.ReturnScheduleExceptionDto, error) {
	location, err := resolveUserLocation(ctx, s.repo, userID)
	if err != nil {
		return nil, err
	}

	startDate, err := parseScheduleDate(req.StartDate, "start_date", location)
	if err != nil {
		return nil, err
	}

	endDate, err := parseScheduleDate(req.EndDate, "end_date", location)
	if err != nil {
		return nil, err
	}

	if err := validateScheduleDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	if err := ensureNoScheduleExceptionOverlap(ctx, s.repo, userID, nil, startDate, endDate); err != nil {
		return nil, err
	}

	exception, err := s.repo.CreateScheduleException(ctx, repository.CreateScheduleExceptionParams{
		UserID:        userID,
		ExceptionType: req.Type,
		StartDate:     startDate,
		EndDate:       endDate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create schedule exception: %w", err)
	}

	slog.Info("schedule exception created", "schedule_exception_id", exception.ID, "user_id", userID)

	return sqlcmapper.ScheduleExceptionFromCreateRow(&exception), nil
}

func (s *scheduleService) GetScheduleException(ctx context.Context, userID, scheduleExceptionID uuid.UUID) (*dto.ReturnScheduleExceptionDto, error) {
	exception, err := s.repo.GetScheduleExceptionByUser(ctx, repository.GetScheduleExceptionByUserParams{
		ID:     scheduleExceptionID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrScheduleExceptionNotFound
		}
		return nil, fmt.Errorf("failed to get schedule exception: %w", err)
	}

	return sqlcmapper.ScheduleExceptionFromGetRow(&exception), nil
}

func (s *scheduleService) ListScheduleExceptions(ctx context.Context, userID uuid.UUID) ([]*dto.ReturnScheduleExceptionDto, error) {
	exceptions, err := s.repo.ListScheduleExceptionsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list schedule exceptions: %w", err)
	}

	return sqlcmapper.ScheduleExceptionListFromRows(exceptions), nil
}

func (s *scheduleService) UpdateScheduleException(ctx context.Context, userID, scheduleExceptionID uuid.UUID, req dto.UpdateScheduleExceptionDto) (*dto.ReturnScheduleExceptionDto, error) {
	existingException, err := s.repo.GetScheduleExceptionByUser(ctx, repository.GetScheduleExceptionByUserParams{
		ID:     scheduleExceptionID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrScheduleExceptionNotFound
		}
		return nil, fmt.Errorf("failed to get schedule exception: %w", err)
	}

	location, err := resolveUserLocation(ctx, s.repo, userID)
	if err != nil {
		return nil, err
	}

	exceptionType := existingException.Type
	if req.Type != nil {
		exceptionType = *req.Type
	}

	startDate := normalizeScheduleDate(existingException.StartDate, location)
	if req.StartDate != nil {
		startDate, err = parseScheduleDate(*req.StartDate, "start_date", location)
		if err != nil {
			return nil, err
		}
	}

	endDate := normalizeScheduleDate(existingException.EndDate, location)
	if req.EndDate != nil {
		endDate, err = parseScheduleDate(*req.EndDate, "end_date", location)
		if err != nil {
			return nil, err
		}
	}

	if err := validateScheduleDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	if err := ensureNoScheduleExceptionOverlap(ctx, s.repo, userID, &scheduleExceptionID, startDate, endDate); err != nil {
		return nil, err
	}

	exception, err := s.repo.UpdateScheduleExceptionByUser(ctx, repository.UpdateScheduleExceptionByUserParams{
		ExceptionType: &exceptionType,
		StartDate:     &startDate,
		EndDate:       &endDate,
		ID:            scheduleExceptionID,
		UserID:        userID,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrScheduleExceptionNotFound
		}
		return nil, fmt.Errorf("failed to update schedule exception: %w", err)
	}

	return sqlcmapper.ScheduleExceptionFromUpdateRow(&exception), nil
}

func (s *scheduleService) DeleteScheduleException(ctx context.Context, userID, scheduleExceptionID uuid.UUID) error {
	rowsAffected, err := s.repo.DeleteScheduleExceptionByUser(ctx, repository.DeleteScheduleExceptionByUserParams{
		ID:     scheduleExceptionID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete schedule exception: %w", err)
	}

	if rowsAffected == 0 {
		return api.ErrScheduleExceptionNotFound
	}

	slog.Info("schedule exception deleted", "schedule_exception_id", scheduleExceptionID, "user_id", userID)

	return nil
}

func (s *scheduleService) ListNextLessons(ctx context.Context, userID, classroomID uuid.UUID) ([]dto.NextLessonDto, error) {
	if _, err := s.repo.GetClassroomByUser(ctx, repository.GetClassroomByUserParams{
		ID:     classroomID,
		UserID: userID,
	}); err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrClassroomNotFound
		}
		return nil, fmt.Errorf("failed to get classroom: %w", err)
	}

	location, err := resolveUserLocation(ctx, s.repo, userID)
	if err != nil {
		return nil, err
	}

	lessons, err := listNextLessonOccurrences(ctx, s.repo, userID, classroomID, s.now(), nextLessonsLimit, location)
	if err != nil {
		return nil, err
	}

	return nextLessonOccurrencesToDTO(lessons), nil
}

func attachClassroomsToScheduleSlots(ctx context.Context, repo repository.Querier, userID uuid.UUID, slots []*dto.ReturnScheduleSlotDto) error {
	if len(slots) == 0 {
		return nil
	}

	slotIDs := make([]uuid.UUID, 0, len(slots))
	for _, slot := range slots {
		if slot == nil {
			continue
		}
		slotIDs = append(slotIDs, slot.ID)
	}

	if len(slotIDs) == 0 {
		return nil
	}

	rows, err := repo.ListScheduleSlotClassroomRefsBySlotIDs(ctx, repository.ListScheduleSlotClassroomRefsBySlotIDsParams{
		UserID:          userID,
		ScheduleSlotIds: slotIDs,
	})
	if err != nil {
		return err
	}

	classroomsBySlot := sqlcmapper.ScheduleSlotClassroomsBySlotFromRows(rows)
	for _, slot := range slots {
		if slot == nil {
			continue
		}

		classrooms := classroomsBySlot[slot.ID]
		if classrooms == nil {
			classrooms = []dto.ScheduleSlotClassroomDto{}
		}
		slot.Classrooms = classrooms
	}

	return nil
}

func ensureClassroomsExist(ctx context.Context, repo repository.Querier, userID uuid.UUID, classroomIDs []uuid.UUID) error {
	if len(classroomIDs) == 0 {
		return newScheduleMinItemsError("classroom_ids", 1)
	}

	count, err := repo.CountClassroomsByIDsAndUser(ctx, repository.CountClassroomsByIDsAndUserParams{
		UserID:       userID,
		ClassroomIds: classroomIDs,
	})
	if err != nil {
		return fmt.Errorf("failed to count classrooms: %w", err)
	}

	if count != int64(len(classroomIDs)) {
		return api.ErrClassroomNotFound
	}

	return nil
}

func ensureNoScheduleSlotConflict(
	ctx context.Context,
	repo repository.Querier,
	userID uuid.UUID,
	excludedID *uuid.UUID,
	weekday int32,
	startTime parsedScheduleTime,
	endTime parsedScheduleTime,
	weekPattern string,
) error {
	conflictCount, err := repo.CountScheduleSlotConflicts(ctx, repository.CountScheduleSlotConflictsParams{
		UserID:      userID,
		Weekday:     weekday,
		ExcludedID:  excludedID,
		StartTime:   startTime.DBValue,
		EndTime:     endTime.DBValue,
		WeekPattern: weekPattern,
	})
	if err != nil {
		return fmt.Errorf("failed to check schedule slot conflicts: %w", err)
	}

	if conflictCount > 0 {
		return api.ErrScheduleSlotConflict
	}

	return nil
}

func ensureNoScheduleExceptionOverlap(
	ctx context.Context,
	repo repository.Querier,
	userID uuid.UUID,
	excludedID *uuid.UUID,
	startDate time.Time,
	endDate time.Time,
) error {
	overlapCount, err := repo.CountScheduleExceptionOverlaps(ctx, repository.CountScheduleExceptionOverlapsParams{
		UserID:     userID,
		ExcludedID: excludedID,
		StartDate:  startDate,
		EndDate:    endDate,
	})
	if err != nil {
		return fmt.Errorf("failed to check schedule exception overlaps: %w", err)
	}

	if overlapCount > 0 {
		return api.ErrScheduleExceptionOverlap
	}

	return nil
}

func setScheduleSlotClassrooms(ctx context.Context, repo repository.Querier, userID, scheduleSlotID uuid.UUID, classroomIDs []uuid.UUID) error {
	for _, classroomID := range classroomIDs {
		if _, err := repo.CreateScheduleSlotClassroomRelation(ctx, repository.CreateScheduleSlotClassroomRelationParams{
			UserID:         userID,
			ScheduleSlotID: scheduleSlotID,
			ClassroomID:    classroomID,
		}); err != nil {
			return fmt.Errorf("failed to attach classroom to schedule slot: %w", err)
		}
	}

	return nil
}

func parseUniqueUUIDs(rawValues []string, field string) ([]uuid.UUID, error) {
	trimmedValues := make([]string, 0, len(rawValues))
	for _, rawValue := range rawValues {
		trimmedValue := strings.TrimSpace(rawValue)
		if trimmedValue != "" {
			trimmedValues = append(trimmedValues, trimmedValue)
		}
	}

	if len(trimmedValues) == 0 {
		return nil, newScheduleMinItemsError(field, 1)
	}

	seen := make(map[uuid.UUID]struct{}, len(trimmedValues))
	values := make([]uuid.UUID, 0, len(trimmedValues))
	for _, rawValue := range trimmedValues {
		parsedValue, err := uuid.Parse(rawValue)
		if err != nil {
			return nil, api.NewAPIError(http.StatusBadRequest, "invalid_request_body", api.ErrorDetail{
				Field: field,
				Error: fmt.Sprintf(api.KeyValidationMalformedParameter, "uuid"),
			})
		}

		if _, exists := seen[parsedValue]; exists {
			continue
		}

		seen[parsedValue] = struct{}{}
		values = append(values, parsedValue)
	}

	return values, nil
}

func parseScheduleTime(rawValue string, field string) (parsedScheduleTime, error) {
	trimmedValue := strings.TrimSpace(rawValue)
	parsedValue, err := time.Parse(scheduleTimeLayout, trimmedValue)
	if err != nil || parsedValue.Format(scheduleTimeLayout) != trimmedValue {
		return parsedScheduleTime{}, api.NewAPIError(http.StatusBadRequest, "invalid_request_body", api.ErrorDetail{
			Field: field,
			Error: fmt.Sprintf(api.KeyValidationMalformedParameter, "hh:mm"),
		})
	}

	totalMinutes := parsedValue.Hour()*60 + parsedValue.Minute()

	return parsedScheduleTime{
		Text:    parsedValue.Format(scheduleTimeLayout),
		Minutes: totalMinutes,
		DBValue: pgtype.Time{
			Microseconds: int64(totalMinutes) * int64(time.Minute/time.Microsecond),
			Valid:        true,
		},
	}, nil
}

func parseScheduleDate(rawValue string, field string, location *time.Location) (time.Time, error) {
	trimmedValue := strings.TrimSpace(rawValue)
	parsedValue, err := time.ParseInLocation(scheduleDateLayout, trimmedValue, location)
	if err != nil || parsedValue.Format(scheduleDateLayout) != trimmedValue {
		return time.Time{}, api.NewAPIError(http.StatusBadRequest, "invalid_request_body", api.ErrorDetail{
			Field: field,
			Error: fmt.Sprintf(api.KeyValidationMalformedParameter, "yyyy-mm-dd"),
		})
	}

	return normalizeScheduleDate(parsedValue, location), nil
}

func validateScheduleTimeRange(startTime, endTime parsedScheduleTime) error {
	if endTime.Minutes <= startTime.Minutes {
		return api.NewAPIError(http.StatusBadRequest, "invalid_request_body", api.ErrorDetail{
			Field: "end_time",
			Error: "schedule_end_time_must_be_after_start_time",
		})
	}

	return nil
}

func validateScheduleDateRange(startDate, endDate time.Time) error {
	if endDate.Before(startDate) {
		return api.NewAPIError(http.StatusBadRequest, "invalid_request_body", api.ErrorDetail{
			Field: "end_date",
			Error: "schedule_end_date_must_be_on_or_after_start_date",
		})
	}

	return nil
}

func newScheduleMinItemsError(field string, min int) error {
	return api.NewAPIError(http.StatusBadRequest, "validation_failed", api.ErrorDetail{
		Field: field,
		Error: fmt.Sprintf(api.KeyValidationMinLength, fmt.Sprintf("%d", min)),
	})
}

func scheduleWeekdayISOFromText(value string) (int32, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "monday":
		return 1, nil
	case "tuesday":
		return 2, nil
	case "wednesday":
		return 3, nil
	case "thursday":
		return 4, nil
	case "friday":
		return 5, nil
	case "saturday":
		return 6, nil
	case "sunday":
		return 7, nil
	default:
		return 0, api.NewAPIError(http.StatusBadRequest, "invalid_request_body", api.ErrorDetail{
			Field: "weekday",
			Error: "schedule_weekday_invalid",
		})
	}
}

func normalizeScheduleDate(value time.Time, location *time.Location) time.Time {
	return calendarDateInLocation(value, location)
}

func startOfScheduleDay(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}

func isoWeekdayFromTime(value time.Weekday) int32 {
	switch value {
	case time.Monday:
		return 1
	case time.Tuesday:
		return 2
	case time.Wednesday:
		return 3
	case time.Thursday:
		return 4
	case time.Friday:
		return 5
	case time.Saturday:
		return 6
	default:
		return 7
	}
}

func scheduleSlotAppliesToISOWeek(weekPattern string, isoWeek int) bool {
	switch weekPattern {
	case "even_weeks":
		return isoWeek%2 == 0
	case "odd_weeks":
		return isoWeek%2 != 0
	default:
		return true
	}
}

func isDateBlockedByScheduleException(day time.Time, exceptionRanges []scheduleDateRange) bool {
	for _, exceptionRange := range exceptionRanges {
		if (day.Equal(exceptionRange.Start) || day.After(exceptionRange.Start)) &&
			(day.Equal(exceptionRange.End) || day.Before(exceptionRange.End)) {
			return true
		}
	}

	return false
}
