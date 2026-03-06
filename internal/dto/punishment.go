package dto

import (
	"time"

	"github.com/google/uuid"
)

type RequestPunishmentDto struct {
	StudentID        string  `json:"student_id" validate:"required,uuid"`
	PunishmentTypeID string  `json:"punishment_type_id" validate:"required,uuid"`
	DueAt            string  `json:"due_at" validate:"required"`
	OccurredAt       *string `json:"occurred_at" validate:"omitempty"`
	EvaluationLabel  *string `json:"evaluation_label" validate:"omitempty"`
}

type UpdatePunishmentDto struct {
	OccurredAt      *string             `json:"occurred_at" validate:"omitempty"`
	EvaluationLabel NullableStringField `json:"evaluation_label"`
}

type ReturnPunishmentDto struct {
	ID                 uuid.UUID  `json:"id"`
	StudentID          uuid.UUID  `json:"student_id"`
	StudentFirstName   string     `json:"student_first_name"`
	StudentLastName    string     `json:"student_last_name"`
	PunishmentTypeID   uuid.UUID  `json:"punishment_type_id"`
	PunishmentTypeName string     `json:"punishment_type_name"`
	TriggeringRuleID   *uuid.UUID `json:"triggering_rule_id"`
	TriggeringRuleName *string    `json:"triggering_rule_name"`
	Automated          bool       `json:"automated"`
	CreatedAt          time.Time  `json:"created_at"`
	OccurredAt         time.Time  `json:"occurred_at"`
	EvaluationLabel    *string    `json:"evaluation_label,omitempty"`
	DueAt              time.Time  `json:"due_at"`
	ResolvedAt         *time.Time `json:"resolved_at"`
}
