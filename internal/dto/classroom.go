package dto

import (
	"time"

	"github.com/google/uuid"
)

type RequestClassroomDto struct {
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Year        *string `json:"year" validate:"omitempty,max=20"`
	MainTeacher *string `json:"main_teacher" validate:"omitempty,max=100"`
}

type UpdateClassroomDto struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=100"`
	Year        *string `json:"year" validate:"omitempty,max=20"`
	MainTeacher *string `json:"main_teacher" validate:"omitempty,max=100"`
}

type ReturnClassroomDto struct {
	ID              uuid.UUID                    `json:"id"`
	Name            string                       `json:"name"`
	Year            *string                      `json:"year"`
	MainTeacher     *string                      `json:"main_teacher"`
	StudentCount    int64                        `json:"student_count"`
	StudentsPreview []ClassroomStudentPreviewDto `json:"students_preview"`
	CreatedAt       time.Time                    `json:"created_at"`
	UpdatedAt       time.Time                    `json:"updated_at"`
}

type StudentClassroomRequestDto struct {
	StudentID string `json:"student_id" validate:"required,uuid"`
}

type ClassroomStudentPreviewDto struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
}
