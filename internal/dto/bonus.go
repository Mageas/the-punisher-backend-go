package dto

import (
	"time"

	"github.com/google/uuid"
)

type RequestBonusDto struct {
	StudentID   string  `json:"student_id" validate:"required,uuid"`
	BonusTypeID string  `json:"bonus_type_id" validate:"required,uuid"`
	Points      float64 `json:"points" validate:"required,gt=0"`
}

type ReturnBonusDto struct {
	ID               uuid.UUID  `json:"id"`
	StudentID        uuid.UUID  `json:"student_id"`
	StudentFirstName string     `json:"student_first_name"`
	StudentLastName  string     `json:"student_last_name"`
	BonusTypeID      uuid.UUID  `json:"bonus_type_id"`
	BonusTypeName    string     `json:"bonus_type_name"`
	Points           float64    `json:"points"`
	CreatedAt        time.Time  `json:"created_at"`
	UsedAt           *time.Time `json:"used_at"`
}
