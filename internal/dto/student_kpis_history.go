package dto

import (
	"time"

	"github.com/google/uuid"
)

type StudentKpisDto struct {
	AvailableBonusPoints   float64 `json:"available_bonus_points"`
	TotalBonusPoints       float64 `json:"total_bonus_points"`
	ActiveBonusCount       int64   `json:"active_bonus_count"`
	PenaltyCount           int64   `json:"penalty_count"`
	TotalPenaltyCount      int64   `json:"total_penalty_count"`
	TotalPunishmentCount   int64   `json:"total_punishment_count"`
	OverduePunishmentCount int64   `json:"overdue_punishment_count"`
	PendingPunishmentCount int64   `json:"pending_punishment_count"`
}

type StudentHistoryItemDto struct {
	Type               string     `json:"type"`
	ID                 uuid.UUID  `json:"id"`
	CreatedAt          time.Time  `json:"created_at"`
	OccurredAt         time.Time  `json:"occurred_at"`
	EvaluationLabel    *string    `json:"evaluation_label,omitempty"`
	PenaltyTypeID      *uuid.UUID `json:"penalty_type_id,omitempty"`
	PenaltyTypeName    *string    `json:"penalty_type_name,omitempty"`
	BonusTypeID        *uuid.UUID `json:"bonus_type_id,omitempty"`
	BonusTypeName      *string    `json:"bonus_type_name,omitempty"`
	Points             *float64   `json:"points,omitempty"`
	UsedAt             *time.Time `json:"used_at,omitempty"`
	PunishmentTypeID   *uuid.UUID `json:"punishment_type_id,omitempty"`
	PunishmentTypeName *string    `json:"punishment_type_name,omitempty"`
	TriggeringRuleID   *uuid.UUID `json:"triggering_rule_id,omitempty"`
	TriggeringRuleName *string    `json:"triggering_rule_name,omitempty"`
	Automated          *bool      `json:"automated,omitempty"`
	DueAt              *time.Time `json:"due_at,omitempty"`
	ResolvedAt         *time.Time `json:"resolved_at,omitempty"`
}
