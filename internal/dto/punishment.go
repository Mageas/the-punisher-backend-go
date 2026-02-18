package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type RequestPunishmentDto struct {
	StudentID        string `json:"student_id" validate:"required,uuid"`
	PunishmentTypeID string `json:"punishment_type_id" validate:"required,uuid"`
	DueAt            string `json:"due_at" validate:"required"`
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
	CreatedAt          time.Time  `json:"created_at"`
	DueAt              time.Time  `json:"due_at"`
	ResolvedAt         *time.Time `json:"resolved_at"`
}

func buildReturnPunishmentDto(
	id uuid.UUID,
	studentID uuid.UUID,
	studentFirstName string,
	studentLastName string,
	punishmentTypeID uuid.UUID,
	punishmentTypeName string,
	triggeringRuleID pgtype.UUID,
	triggeringRuleName *string,
	createdAt time.Time,
	dueAt time.Time,
	resolvedAt pgtype.Timestamptz,
) *ReturnPunishmentDto {
	dto := &ReturnPunishmentDto{
		ID:                 id,
		StudentID:          studentID,
		StudentFirstName:   studentFirstName,
		StudentLastName:    studentLastName,
		PunishmentTypeID:   punishmentTypeID,
		PunishmentTypeName: punishmentTypeName,
		TriggeringRuleID:   punishmentTriggeringRuleID(triggeringRuleID),
		TriggeringRuleName: triggeringRuleName,
		CreatedAt:          createdAt,
		DueAt:              dueAt,
	}

	if resolvedAtValue := punishmentResolvedAt(resolvedAt); resolvedAtValue != nil {
		dto.ResolvedAt = resolvedAtValue
	}

	return dto
}

func PunishmentFromCreateRow(p *repository.CreatePunishmentRow) *ReturnPunishmentDto {
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
		p.CreatedAt,
		p.DueAt,
		p.ResolvedAt,
	)
}

func PunishmentFromGetRow(p *repository.GetPunishmentByUserRow) *ReturnPunishmentDto {
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
		p.CreatedAt,
		p.DueAt,
		p.ResolvedAt,
	)
}

func PunishmentListFromListByUserRows(punishments []repository.ListPunishmentsByUserRow) []*ReturnPunishmentDto {
	dtos := make([]*ReturnPunishmentDto, 0, len(punishments))

	for _, punishment := range punishments {
		dto := buildReturnPunishmentDto(
			punishment.ID,
			punishment.StudentID,
			punishment.StudentFirstName,
			punishment.StudentLastName,
			punishment.PunishmentTypeID,
			punishment.PunishmentTypeName,
			punishment.TriggeringRuleID,
			punishmentTriggeringRuleNameFromText(punishment.TriggeringRuleName),
			punishment.CreatedAt,
			punishment.DueAt,
			punishment.ResolvedAt,
		)
		if dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
}

func PunishmentListFromListByStudentRows(punishments []repository.ListPunishmentsByStudentRow) []*ReturnPunishmentDto {
	dtos := make([]*ReturnPunishmentDto, 0, len(punishments))

	for _, punishment := range punishments {
		dto := buildReturnPunishmentDto(
			punishment.ID,
			punishment.StudentID,
			punishment.StudentFirstName,
			punishment.StudentLastName,
			punishment.PunishmentTypeID,
			punishment.PunishmentTypeName,
			punishment.TriggeringRuleID,
			punishmentTriggeringRuleNameFromText(punishment.TriggeringRuleName),
			punishment.CreatedAt,
			punishment.DueAt,
			punishment.ResolvedAt,
		)
		if dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
}

func PunishmentFromResolveRow(p *repository.ResolvePunishmentRow) *ReturnPunishmentDto {
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
		p.CreatedAt,
		p.DueAt,
		p.ResolvedAt,
	)
}

func punishmentResolvedAt(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	return &value.Time
}

func punishmentTriggeringRuleID(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}

	id := uuid.UUID(value.Bytes)
	return &id
}

func punishmentTriggeringRuleNameFromText(value pgtype.Text) *string {
	if !value.Valid || value.String == "" {
		return nil
	}

	name := value.String
	return &name
}
