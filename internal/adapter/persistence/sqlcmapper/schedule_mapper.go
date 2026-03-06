package sqlcmapper

import (
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

const scheduleDateLayout = "2006-01-02"

func buildReturnScheduleSlotDto(
	id uuid.UUID,
	weekday int32,
	startTime string,
	endTime string,
	weekPattern string,
	createdAt time.Time,
	updatedAt time.Time,
) *dto.ReturnScheduleSlotDto {
	return &dto.ReturnScheduleSlotDto{
		ID:          id,
		Weekday:     scheduleWeekdayTextFromISO(weekday),
		StartTime:   startTime,
		EndTime:     endTime,
		WeekPattern: weekPattern,
		Classrooms:  []dto.ScheduleSlotClassroomDto{},
		CreatedAt:   normalizeAPITime(createdAt),
		UpdatedAt:   normalizeAPITime(updatedAt),
	}
}

func ScheduleSlotFromModel(slot *repository.ScheduleSlot) *dto.ReturnScheduleSlotDto {
	if slot == nil {
		return nil
	}

	return buildReturnScheduleSlotDto(
		slot.ID,
		slot.Weekday,
		slot.StartTime,
		slot.EndTime,
		slot.WeekPattern,
		slot.CreatedAt,
		slot.UpdatedAt,
	)
}

func ScheduleSlotListFromModels(slots []repository.ScheduleSlot) []*dto.ReturnScheduleSlotDto {
	responses := make([]*dto.ReturnScheduleSlotDto, 0, len(slots))

	for _, slot := range slots {
		response := ScheduleSlotFromModel(&slot)
		if response != nil {
			responses = append(responses, response)
		}
	}

	return responses
}

func ScheduleSlotClassroomsBySlotFromRows(rows []repository.ListScheduleSlotClassroomRefsBySlotIDsRow) map[uuid.UUID][]dto.ScheduleSlotClassroomDto {
	classroomsBySlot := make(map[uuid.UUID][]dto.ScheduleSlotClassroomDto)

	for _, row := range rows {
		classroomsBySlot[row.ScheduleSlotID] = append(classroomsBySlot[row.ScheduleSlotID], dto.ScheduleSlotClassroomDto{
			ID:   row.ClassroomID,
			Name: row.ClassroomName,
		})
	}

	return classroomsBySlot
}

func buildReturnScheduleExceptionDto(
	id uuid.UUID,
	exceptionType string,
	startDate time.Time,
	endDate time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) *dto.ReturnScheduleExceptionDto {
	return &dto.ReturnScheduleExceptionDto{
		ID:        id,
		Type:      exceptionType,
		StartDate: startDate.Format(scheduleDateLayout),
		EndDate:   endDate.Format(scheduleDateLayout),
		CreatedAt: normalizeAPITime(createdAt),
		UpdatedAt: normalizeAPITime(updatedAt),
	}
}

func ScheduleExceptionFromCreateRow(row *repository.CreateScheduleExceptionRow) *dto.ReturnScheduleExceptionDto {
	if row == nil {
		return nil
	}

	return buildReturnScheduleExceptionDto(
		row.ID,
		row.Type,
		row.StartDate,
		row.EndDate,
		row.CreatedAt,
		row.UpdatedAt,
	)
}

func ScheduleExceptionFromGetRow(row *repository.GetScheduleExceptionByUserRow) *dto.ReturnScheduleExceptionDto {
	if row == nil {
		return nil
	}

	return buildReturnScheduleExceptionDto(
		row.ID,
		row.Type,
		row.StartDate,
		row.EndDate,
		row.CreatedAt,
		row.UpdatedAt,
	)
}

func ScheduleExceptionFromUpdateRow(row *repository.UpdateScheduleExceptionByUserRow) *dto.ReturnScheduleExceptionDto {
	if row == nil {
		return nil
	}

	return buildReturnScheduleExceptionDto(
		row.ID,
		row.Type,
		row.StartDate,
		row.EndDate,
		row.CreatedAt,
		row.UpdatedAt,
	)
}

func ScheduleExceptionListFromRows(rows []repository.ListScheduleExceptionsByUserRow) []*dto.ReturnScheduleExceptionDto {
	responses := make([]*dto.ReturnScheduleExceptionDto, 0, len(rows))

	for _, row := range rows {
		response := buildReturnScheduleExceptionDto(
			row.ID,
			row.Type,
			row.StartDate,
			row.EndDate,
			row.CreatedAt,
			row.UpdatedAt,
		)
		if response != nil {
			responses = append(responses, response)
		}
	}

	return responses
}

func scheduleWeekdayTextFromISO(value int32) string {
	switch value {
	case 1:
		return "monday"
	case 2:
		return "tuesday"
	case 3:
		return "wednesday"
	case 4:
		return "thursday"
	case 5:
		return "friday"
	case 6:
		return "saturday"
	case 7:
		return "sunday"
	default:
		return ""
	}
}
