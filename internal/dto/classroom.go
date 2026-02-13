package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
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
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Year        *string   `json:"year"`
	MainTeacher *string   `json:"main_teacher"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type StudentClassroomRequestDto struct {
	StudentID string `json:"student_id" validate:"required,uuid"`
}

func ClassroomFromRepository(c *repository.Classroom) *ReturnClassroomDto {
	if c == nil {
		return nil
	}

	dto := &ReturnClassroomDto{
		ID:        c.ID,
		Name:      c.Name,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}

	if c.Year.Valid {
		dto.Year = &c.Year.String
	}
	if c.MainTeacher.Valid {
		dto.MainTeacher = &c.MainTeacher.String
	}

	return dto
}

func ClassroomListFromRepository(classrooms []repository.Classroom) []*ReturnClassroomDto {
	dtos := make([]*ReturnClassroomDto, 0, len(classrooms))

	for _, c := range classrooms {
		if dto := ClassroomFromRepository(&c); dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
}
