package dto

import (
	"time"

	"github.com/google/uuid"
)

type RequestPenaltyDto struct {
	StudentID     string `json:"student_id" validate:"required,uuid"`
	PenaltyTypeID string `json:"penalty_type_id" validate:"required,uuid"`
}

type ReturnPenaltyDto struct {
	ID               uuid.UUID `json:"id"`
	StudentID        uuid.UUID `json:"student_id"`
	StudentFirstName string    `json:"student_first_name"`
	StudentLastName  string    `json:"student_last_name"`
	PenaltyTypeID    uuid.UUID `json:"penalty_type_id"`
	PenaltyTypeName  string    `json:"penalty_type_name"`
	CreatedAt        time.Time `json:"created_at"`
}
