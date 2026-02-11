package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type RequestStudentDto struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=70"`
	LastName  string `json:"last_name" validate:"required,min=2,max=70"`
}

type UpdateStudentDto struct {
	FirstName *string `json:"first_name" validate:"omitempty,min=2,max=70"`
	LastName  *string `json:"last_name" validate:"omitempty,min=2,max=70"`
}

type ReturnStudentDto struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func StudentFromRepository(s *repository.Student) *ReturnStudentDto {
	if s == nil {
		return nil
	}

	return &ReturnStudentDto{
		ID:        s.ID,
		FirstName: s.FirstName,
		LastName:  s.LastName,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func StudentListFromRepository(students []repository.Student) []*ReturnStudentDto {
	dtos := make([]*ReturnStudentDto, 0, len(students))

	for _, s := range students {
		if dto := StudentFromRepository(&s); dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
}
