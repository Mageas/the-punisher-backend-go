package dto

import (
	"time"

	"github.com/google/uuid"
)

type RequestBonusTypeDto struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

type UpdateBonusTypeDto struct {
	Name *string `json:"name" validate:"omitempty,min=2,max=100"`
}

type ReturnBonusTypeDto struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (d ReturnBonusTypeDto) GetID() uuid.UUID {
	return d.ID
}

func (d ReturnBonusTypeDto) GetName() string {
	return d.Name
}
