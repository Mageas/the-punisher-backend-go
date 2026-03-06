package sqlcmapper

import (
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func buildReturnPunishmentDto(
	id uuid.UUID,
	studentID uuid.UUID,
	studentFirstName string,
	studentLastName string,
	punishmentTypeID uuid.UUID,
	punishmentTypeName string,
	triggeringRuleID *uuid.UUID,
	triggeringRuleName *string,
	automated bool,
	createdAt time.Time,
	occurredAt time.Time,
	evaluationLabel *string,
	dueAt time.Time,
	resolvedAt *time.Time,
) *dto.ReturnPunishmentDto {
	response := &dto.ReturnPunishmentDto{
		ID:                 id,
		StudentID:          studentID,
		StudentFirstName:   studentFirstName,
		StudentLastName:    studentLastName,
		PunishmentTypeID:   punishmentTypeID,
		PunishmentTypeName: punishmentTypeName,
		TriggeringRuleID:   punishmentTriggeringRuleID(triggeringRuleID),
		TriggeringRuleName: triggeringRuleName,
		Automated:          automated,
		CreatedAt:          normalizeAPITime(createdAt),
		OccurredAt:         normalizeAPITime(occurredAt),
		EvaluationLabel:    bonusEvaluationLabel(evaluationLabel),
		DueAt:              normalizeAPITime(dueAt),
	}

	if resolvedAtValue := punishmentResolvedAt(resolvedAt); resolvedAtValue != nil {
		response.ResolvedAt = resolvedAtValue
	}

	return response
}

func PunishmentFromCreateRow(p *repository.CreatePunishmentRow) *dto.ReturnPunishmentDto {
	if p == nil {
		return nil
	}

	return buildReturnPunishmentDto(
		p.ID,
		p.StudentID,
		p.StudentFirstName,
		p.StudentLastName,
		p.PunishmentTypeID,
		p.PunishmentTypeName,
		p.TriggeringRuleID,
		punishmentTriggeringRuleNameFromText(p.TriggeringRuleName),
		p.Automated,
		p.CreatedAt,
		p.OccurredAt,
		p.EvaluationLabel,
		p.DueAt,
		p.ResolvedAt,
	)
}

func PunishmentFromGetRow(p *repository.GetPunishmentByUserRow) *dto.ReturnPunishmentDto {
	if p == nil {
		return nil
	}

	return buildReturnPunishmentDto(
		p.ID,
		p.StudentID,
		p.StudentFirstName,
		p.StudentLastName,
		p.PunishmentTypeID,
		p.PunishmentTypeName,
		p.TriggeringRuleID,
		punishmentTriggeringRuleNameFromText(p.TriggeringRuleName),
		p.Automated,
		p.CreatedAt,
		p.OccurredAt,
		p.EvaluationLabel,
		p.DueAt,
		p.ResolvedAt,
	)
}

func PunishmentListFromListByUserRows(punishments []repository.ListPunishmentsByUserRow) []*dto.ReturnPunishmentDto {
	responses := make([]*dto.ReturnPunishmentDto, 0, len(punishments))

	for _, punishment := range punishments {
		response := buildReturnPunishmentDto(
			punishment.ID,
			punishment.StudentID,
			punishment.StudentFirstName,
			punishment.StudentLastName,
			punishment.PunishmentTypeID,
			punishment.PunishmentTypeName,
			punishment.TriggeringRuleID,
			punishmentTriggeringRuleNameFromText(punishment.TriggeringRuleName),
			punishment.Automated,
			punishment.CreatedAt,
			punishment.OccurredAt,
			punishment.EvaluationLabel,
			punishment.DueAt,
			punishment.ResolvedAt,
		)
		if response != nil {
			responses = append(responses, response)
		}
	}

	return responses
}

func PunishmentListFromListByStudentRows(punishments []repository.ListPunishmentsByStudentRow) []*dto.ReturnPunishmentDto {
	responses := make([]*dto.ReturnPunishmentDto, 0, len(punishments))

	for _, punishment := range punishments {
		response := buildReturnPunishmentDto(
			punishment.ID,
			punishment.StudentID,
			punishment.StudentFirstName,
			punishment.StudentLastName,
			punishment.PunishmentTypeID,
			punishment.PunishmentTypeName,
			punishment.TriggeringRuleID,
			punishmentTriggeringRuleNameFromText(punishment.TriggeringRuleName),
			punishment.Automated,
			punishment.CreatedAt,
			punishment.OccurredAt,
			punishment.EvaluationLabel,
			punishment.DueAt,
			punishment.ResolvedAt,
		)
		if response != nil {
			responses = append(responses, response)
		}
	}

	return responses
}

func PunishmentFromResolveRow(p *repository.ResolvePunishmentRow) *dto.ReturnPunishmentDto {
	if p == nil {
		return nil
	}

	return buildReturnPunishmentDto(
		p.ID,
		p.StudentID,
		p.StudentFirstName,
		p.StudentLastName,
		p.PunishmentTypeID,
		p.PunishmentTypeName,
		p.TriggeringRuleID,
		punishmentTriggeringRuleNameFromText(p.TriggeringRuleName),
		p.Automated,
		p.CreatedAt,
		p.OccurredAt,
		p.EvaluationLabel,
		p.DueAt,
		p.ResolvedAt,
	)
}

func PunishmentFromUpdateRow(p *repository.UpdatePunishmentByUserRow) *dto.ReturnPunishmentDto {
	if p == nil {
		return nil
	}

	return buildReturnPunishmentDto(
		p.ID,
		p.StudentID,
		p.StudentFirstName,
		p.StudentLastName,
		p.PunishmentTypeID,
		p.PunishmentTypeName,
		p.TriggeringRuleID,
		punishmentTriggeringRuleNameFromText(p.TriggeringRuleName),
		p.Automated,
		p.CreatedAt,
		p.OccurredAt,
		p.EvaluationLabel,
		p.DueAt,
		p.ResolvedAt,
	)
}

func punishmentResolvedAt(value *time.Time) *time.Time {
	return normalizeOptionalAPITime(value)
}

func punishmentTriggeringRuleID(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}

	id := *value
	return &id
}

func punishmentTriggeringRuleNameFromText(value *string) *string {
	if value == nil || *value == "" {
		return nil
	}

	name := *value
	return &name
}
