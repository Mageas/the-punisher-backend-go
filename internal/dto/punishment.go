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
	ID               uuid.UUID  `json:"id"`
	StudentID        uuid.UUID  `json:"student_id"`
	PunishmentTypeID uuid.UUID  `json:"punishment_type_id"`
	TriggeringRuleID *uuid.UUID `json:"triggering_rule_id"`
	CreatedAt        time.Time  `json:"created_at"`
	DueAt            time.Time  `json:"due_at"`
	ResolvedAt       *time.Time `json:"resolved_at"`
}

func PunishmentFromRepository(p *repository.Punishment) *ReturnPunishmentDto {
	if p == nil {
		return nil
	}

	dto := &ReturnPunishmentDto{
		ID:               p.ID,
		StudentID:        p.StudentID,
		PunishmentTypeID: p.PunishmentTypeID,
		TriggeringRuleID: punishmentTriggeringRuleID(p.TriggeringRuleID),
		CreatedAt:        p.CreatedAt,
		DueAt:            p.DueAt,
	}

	if resolvedAt := punishmentResolvedAt(p.ResolvedAt); resolvedAt != nil {
		dto.ResolvedAt = resolvedAt
	}

	return dto
}

func PunishmentListFromRepository(punishments []repository.Punishment) []*ReturnPunishmentDto {
	dtos := make([]*ReturnPunishmentDto, 0, len(punishments))

	for _, punishment := range punishments {
		if dto := PunishmentFromRepository(&punishment); dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
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
