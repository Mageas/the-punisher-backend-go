package dto

import (
	"time"

	"github.com/google/uuid"
)

type RequestStudentDto struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=70"`
	LastName  string `json:"last_name" validate:"required,min=2,max=70"`
}

type UpdateStudentDto struct {
	FirstName *string `json:"first_name" validate:"omitempty,min=2,max=70"`
	LastName  *string `json:"last_name" validate:"omitempty,min=2,max=70"`
}

type StudentClassroomDto struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type ReturnStudentDto struct {
	ID                   uuid.UUID             `json:"id"`
	FirstName            string                `json:"first_name"`
	LastName             string                `json:"last_name"`
	Classrooms           []StudentClassroomDto `json:"classrooms"`
	AvailableBonusPoints float64               `json:"available_bonus_points"`
	PenaltyCount         int64                 `json:"penalty_count"`
	CreatedAt            time.Time             `json:"created_at"`
	UpdatedAt            time.Time             `json:"updated_at"`
}
