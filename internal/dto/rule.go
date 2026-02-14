package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type RequestRuleDto struct {
	Name                      string `json:"name" validate:"required,min=2,max=120"`
	ResultingPunishmentTypeID string `json:"resulting_punishment_type_id" validate:"required,uuid"`
	PenaltyTypeID             string `json:"penalty_type_id" validate:"required,uuid"`
	Threshold                 int32  `json:"threshold" validate:"required,min=1"`
	DueAtAfterDays            int32  `json:"due_at_after_days" validate:"min=0"`
	Mode                      string `json:"mode" validate:"required,oneof=after at every"`
	IsActive                  *bool  `json:"is_active" validate:"omitempty"`
}

type UpdateRuleDto struct {
	Name                      *string `json:"name" validate:"omitempty,min=2,max=120"`
	ResultingPunishmentTypeID *string `json:"resulting_punishment_type_id" validate:"omitempty,uuid"`
	PenaltyTypeID             *string `json:"penalty_type_id" validate:"omitempty,uuid"`
	Threshold                 *int32  `json:"threshold" validate:"omitempty,min=1"`
	DueAtAfterDays            *int32  `json:"due_at_after_days" validate:"omitempty,min=0"`
	Mode                      *string `json:"mode" validate:"omitempty,oneof=after at every"`
	IsActive                  *bool   `json:"is_active" validate:"omitempty"`
}

type ReturnRuleDto struct {
	ID                        uuid.UUID `json:"id"`
	Name                      string    `json:"name"`
	ResultingPunishmentTypeID uuid.UUID `json:"resulting_punishment_type_id"`
	PenaltyTypeID             uuid.UUID `json:"penalty_type_id"`
	Threshold                 int32     `json:"threshold"`
	DueAtAfterDays            int32     `json:"due_at_after_days"`
	Mode                      string    `json:"mode"`
	IsActive                  bool      `json:"is_active"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

func RuleFromRepository(rule *repository.Rule) *ReturnRuleDto {
	if rule == nil {
		return nil
	}

	return &ReturnRuleDto{
		ID:                        rule.ID,
		Name:                      rule.Name,
		ResultingPunishmentTypeID: rule.ResultingPunishmentTypeID,
		PenaltyTypeID:             rule.PenaltyTypeID,
		Threshold:                 rule.Threshold,
		DueAtAfterDays:            rule.DueAtAfterDays,
		Mode:                      rule.Mode,
		IsActive:                  rule.IsActive,
		CreatedAt:                 rule.CreatedAt,
		UpdatedAt:                 rule.UpdatedAt,
	}
}

func RuleListFromRepository(rules []repository.Rule) []*ReturnRuleDto {
	dtos := make([]*ReturnRuleDto, 0, len(rules))

	for _, rule := range rules {
		if dto := RuleFromRepository(&rule); dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
}
