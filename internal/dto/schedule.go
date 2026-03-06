package dto

import (
	"time"

	"github.com/google/uuid"
)

type RequestScheduleSlotDto struct {
	Weekday      string   `json:"weekday" validate:"required,oneof=monday tuesday wednesday thursday friday saturday sunday"`
	StartTime    string   `json:"start_time" validate:"required"`
	EndTime      string   `json:"end_time" validate:"required"`
	WeekPattern  string   `json:"week_pattern" validate:"required,oneof=every_week even_weeks odd_weeks"`
	ClassroomIDs []string `json:"classroom_ids" validate:"required,min=1,dive,uuid"`
}

type UpdateScheduleSlotDto struct {
	Weekday      *string   `json:"weekday" validate:"omitempty,oneof=monday tuesday wednesday thursday friday saturday sunday"`
	StartTime    *string   `json:"start_time" validate:"omitempty"`
	EndTime      *string   `json:"end_time" validate:"omitempty"`
	WeekPattern  *string   `json:"week_pattern" validate:"omitempty,oneof=every_week even_weeks odd_weeks"`
	ClassroomIDs *[]string `json:"classroom_ids" validate:"omitempty,min=1,dive,uuid"`
}

type ScheduleSlotClassroomDto struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type ReturnScheduleSlotDto struct {
	ID          uuid.UUID                  `json:"id"`
	Weekday     string                     `json:"weekday"`
	StartTime   string                     `json:"start_time"`
	EndTime     string                     `json:"end_time"`
	WeekPattern string                     `json:"week_pattern"`
	Classrooms  []ScheduleSlotClassroomDto `json:"classrooms"`
	CreatedAt   time.Time                  `json:"created_at"`
	UpdatedAt   time.Time                  `json:"updated_at"`
}

type RequestScheduleExceptionDto struct {
	Type      string `json:"type" validate:"required,oneof=vacation public_holiday"`
	StartDate string `json:"start_date" validate:"required"`
	EndDate   string `json:"end_date" validate:"required"`
}

type UpdateScheduleExceptionDto struct {
	Type      *string `json:"type" validate:"omitempty,oneof=vacation public_holiday"`
	StartDate *string `json:"start_date" validate:"omitempty"`
	EndDate   *string `json:"end_date" validate:"omitempty"`
}

type ReturnScheduleExceptionDto struct {
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"type"`
	StartDate string    `json:"start_date"`
	EndDate   string    `json:"end_date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NextLessonDto struct {
	Date      string `json:"date"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}
