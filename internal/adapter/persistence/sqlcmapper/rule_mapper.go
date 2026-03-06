package sqlcmapper

import (
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func buildReturnRuleDto(
	id uuid.UUID,
	name string,
	resultingPunishmentTypeID uuid.UUID,
	resultingPunishmentTypeName string,
	penaltyTypeID uuid.UUID,
	penaltyTypeName string,
	threshold int32,
	dueAtAfterDays int32,
	mode string,
	isActive bool,
	createdAt time.Time,
	updatedAt time.Time,
) *dto.ReturnRuleDto {
	return &dto.ReturnRuleDto{
		ID:                          id,
		Name:                        name,
		ResultingPunishmentTypeID:   resultingPunishmentTypeID,
		ResultingPunishmentTypeName: resultingPunishmentTypeName,
		PenaltyTypeID:               penaltyTypeID,
		PenaltyTypeName:             penaltyTypeName,
		Threshold:                   threshold,
		DueAtAfterDays:              dueAtAfterDays,
		Mode:                        mode,
		IsActive:                    isActive,
		CreatedAt:                   normalizeAPITime(createdAt),
		UpdatedAt:                   normalizeAPITime(updatedAt),
	}
}

func RuleFromCreateRow(rule *repository.CreateRuleRow) *dto.ReturnRuleDto {
	if rule == nil {
		return nil
	}

	return buildReturnRuleDto(
		rule.ID,
		rule.Name,
		rule.ResultingPunishmentTypeID,
		rule.ResultingPunishmentTypeName,
		rule.PenaltyTypeID,
		rule.PenaltyTypeName,
		rule.Threshold,
		rule.DueAtAfterDays,
		rule.Mode,
		rule.IsActive,
		rule.CreatedAt,
		rule.UpdatedAt,
	)
}

func RuleFromGetRow(rule *repository.GetRuleByUserRow) *dto.ReturnRuleDto {
	if rule == nil {
		return nil
	}

	return buildReturnRuleDto(
		rule.ID,
		rule.Name,
		rule.ResultingPunishmentTypeID,
		rule.ResultingPunishmentTypeName,
		rule.PenaltyTypeID,
		rule.PenaltyTypeName,
		rule.Threshold,
		rule.DueAtAfterDays,
		rule.Mode,
		rule.IsActive,
		rule.CreatedAt,
		rule.UpdatedAt,
	)
}

func RuleListFromListByUserRows(rules []repository.ListRulesByUserRow) []*dto.ReturnRuleDto {
	responses := make([]*dto.ReturnRuleDto, 0, len(rules))

	for _, rule := range rules {
		response := buildReturnRuleDto(
			rule.ID,
			rule.Name,
			rule.ResultingPunishmentTypeID,
			rule.ResultingPunishmentTypeName,
			rule.PenaltyTypeID,
			rule.PenaltyTypeName,
			rule.Threshold,
			rule.DueAtAfterDays,
			rule.Mode,
			rule.IsActive,
			rule.CreatedAt,
			rule.UpdatedAt,
		)
		if response != nil {
			responses = append(responses, response)
		}
	}

	return responses
}

func RuleFromUpdateRow(rule *repository.UpdateRuleByUserRow) *dto.ReturnRuleDto {
	if rule == nil {
		return nil
	}

	return buildReturnRuleDto(
		rule.ID,
		rule.Name,
		rule.ResultingPunishmentTypeID,
		rule.ResultingPunishmentTypeName,
		rule.PenaltyTypeID,
		rule.PenaltyTypeName,
		rule.Threshold,
		rule.DueAtAfterDays,
		rule.Mode,
		rule.IsActive,
		rule.CreatedAt,
		rule.UpdatedAt,
	)
}
