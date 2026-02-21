package dto

import (
	"time"

	"github.com/google/uuid"
)

type RequestPunishmentTypeDto struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

type UpdatePunishmentTypeDto struct {
	Name *string `json:"name" validate:"omitempty,min=2,max=100"`
}

type ReturnPunishmentTypeDto struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (d ReturnPunishmentTypeDto) GetID() uuid.UUID {
	return d.ID
}

func (d ReturnPunishmentTypeDto) GetName() string {
	return d.Name
}
