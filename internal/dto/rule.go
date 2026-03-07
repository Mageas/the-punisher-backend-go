package dto

import (
	"time"

	"github.com/google/uuid"
)

type RequestRuleDto struct {
	Name                      string `json:"name" validate:"required,min=2,max=120"`
	ResultingPunishmentTypeID string `json:"resulting_punishment_type_id" validate:"required,uuid"`
	PenaltyTypeID             string `json:"penalty_type_id" validate:"required,uuid"`
	Threshold                 int32  `json:"threshold" validate:"required,min=1"`
	DueAtAfterDays            *int32 `json:"due_at_after_days" validate:"omitempty,min=0"`
	DueAtMode                 string `json:"due_at_mode" validate:"required,oneof=days next_lessons"`
	DueAtAfterLessons         *int32 `json:"due_at_after_lessons" validate:"omitempty,min=1,max=5"`
	Mode                      string `json:"mode" validate:"required,oneof=after at every"`
	IsActive                  *bool  `json:"is_active" validate:"omitempty"`
}

type UpdateRuleDto struct {
	Name                      *string `json:"name" validate:"omitempty,min=2,max=120"`
	ResultingPunishmentTypeID *string `json:"resulting_punishment_type_id" validate:"omitempty,uuid"`
	PenaltyTypeID             *string `json:"penalty_type_id" validate:"omitempty,uuid"`
	Threshold                 *int32  `json:"threshold" validate:"omitempty,min=1"`
	DueAtAfterDays            *int32  `json:"due_at_after_days" validate:"omitempty,min=0"`
	DueAtMode                 *string `json:"due_at_mode" validate:"omitempty,oneof=days next_lessons"`
	DueAtAfterLessons         *int32  `json:"due_at_after_lessons" validate:"omitempty,min=1,max=5"`
	Mode                      *string `json:"mode" validate:"omitempty,oneof=after at every"`
	IsActive                  *bool   `json:"is_active" validate:"omitempty"`
}

type ReturnRuleDto struct {
	ID                          uuid.UUID `json:"id"`
	Name                        string    `json:"name"`
	ResultingPunishmentTypeID   uuid.UUID `json:"resulting_punishment_type_id"`
	ResultingPunishmentTypeName string    `json:"resulting_punishment_type_name"`
	PenaltyTypeID               uuid.UUID `json:"penalty_type_id"`
	PenaltyTypeName             string    `json:"penalty_type_name"`
	Threshold                   int32     `json:"threshold"`
	DueAtAfterDays              *int32    `json:"due_at_after_days"`
	DueAtMode                   string    `json:"due_at_mode"`
	DueAtAfterLessons           *int32    `json:"due_at_after_lessons"`
	Mode                        string    `json:"mode"`
	IsActive                    bool      `json:"is_active"`
	CreatedAt                   time.Time `json:"created_at"`
	UpdatedAt                   time.Time `json:"updated_at"`
}
