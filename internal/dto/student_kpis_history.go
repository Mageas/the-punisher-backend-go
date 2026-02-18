package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

var studentHistoryTimeSentinel = time.Unix(0, 0).UTC()

type StudentKpisDto struct {
	AvailableBonusPoints   float64 `json:"available_bonus_points"`
	ActiveBonusCount       int64   `json:"active_bonus_count"`
	TotalPenaltyCount      int64   `json:"total_penalty_count"`
	PendingPunishmentCount int64   `json:"pending_punishment_count"`
}

type StudentHistoryItemDto struct {
	Type               string     `json:"type"`
	ID                 uuid.UUID  `json:"id"`
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
	DueAt              *time.Time `json:"due_at,omitempty"`
	ResolvedAt         *time.Time `json:"resolved_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

func StudentKpisFromRow(kpis *repository.GetStudentKpisRow) *StudentKpisDto {
	if kpis == nil {
		return nil
	}

	return &StudentKpisDto{
		AvailableBonusPoints:   kpis.AvailableBonusPoints,
		ActiveBonusCount:       kpis.ActiveBonusCount,
		TotalPenaltyCount:      kpis.TotalPenaltyCount,
		PendingPunishmentCount: kpis.PendingPunishmentCount,
	}
}

func StudentHistoryFromRows(rows []repository.ListStudentHistoryRow) []StudentHistoryItemDto {
	return studentHistoryFromRows(rows)
}

func studentHistoryFromRows(rows []repository.ListStudentHistoryRow) []StudentHistoryItemDto {
	history := make([]StudentHistoryItemDto, 0, len(rows))

	for _, row := range rows {
		item := StudentHistoryItemDto{
			Type:      row.Type,
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
		}

		switch row.Type {
		case "penalty":
			item.PenaltyTypeID = studentHistoryUUIDPtrFromSentinel(row.PenaltyTypeID)
			item.PenaltyTypeName = studentHistoryTextPtrFromString(row.PenaltyTypeName)
		case "bonus":
			item.BonusTypeID = studentHistoryUUIDPtrFromSentinel(row.BonusTypeID)
			item.BonusTypeName = studentHistoryTextPtrFromString(row.BonusTypeName)
			item.Points = studentHistoryFloatPtrFromFloat(row.Points)
			item.UsedAt = studentHistoryTimePtrFromSentinel(row.UsedAt)
		case "punishment":
			punishmentTypeID := row.PunishmentTypeID
			punishmentTypeName := row.PunishmentTypeName
			dueAt := row.DueAt
			item.PunishmentTypeID = &punishmentTypeID
			item.PunishmentTypeName = &punishmentTypeName
			item.TriggeringRuleID = studentHistoryUUIDPtr(row.TriggeringRuleID)
			item.TriggeringRuleName = studentHistoryTextPtrFromString(row.TriggeringRuleName)
			item.DueAt = &dueAt
			item.ResolvedAt = studentHistoryTimePtrFromSentinelPg(row.ResolvedAt)
		}

		history = append(history, item)
	}

	return history
}

func studentHistoryUUIDPtr(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}

	id := uuid.UUID(value.Bytes)
	if id == uuid.Nil {
		return nil
	}

	return &id
}

func studentHistoryUUIDPtrFromSentinel(value uuid.UUID) *uuid.UUID {
	if value == uuid.Nil {
		return nil
	}

	id := value
	return &id
}

func studentHistoryTextPtrFromString(value string) *string {
	if value == "" {
		return nil
	}

	text := value
	return &text
}

func studentHistoryTimePtrFromSentinel(value time.Time) *time.Time {
	if value.Equal(studentHistoryTimeSentinel) {
		return nil
	}

	timeValue := value
	return &timeValue
}

func studentHistoryTimePtrFromSentinelPg(value pgtype.Timestamptz) *time.Time {
	if !value.Valid || value.Time.Equal(studentHistoryTimeSentinel) {
		return nil
	}

	timeValue := value.Time
	return &timeValue
}

func studentHistoryFloatPtrFromFloat(value float64) *float64 {
	floatValue := value
	return &floatValue
}
