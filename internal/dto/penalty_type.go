package dto

import (
	"time"

	"github.com/google/uuid"
)

type RequestPenaltyTypeDto struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

type UpdatePenaltyTypeDto struct {
	Name *string `json:"name" validate:"omitempty,min=2,max=100"`
}

type ReturnPenaltyTypeDto struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (d ReturnPenaltyTypeDto) GetID() uuid.UUID {
	return d.ID
}

func (d ReturnPenaltyTypeDto) GetName() string {
	return d.Name
}
