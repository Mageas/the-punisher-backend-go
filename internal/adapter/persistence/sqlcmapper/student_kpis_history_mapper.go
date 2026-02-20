package sqlcmapper

import (
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func StudentKpisFromRow(kpis *repository.GetStudentKpisRow) *dto.StudentKpisDto {
	if kpis == nil {
		return nil
	}

	return &dto.StudentKpisDto{
		AvailableBonusPoints:   kpis.AvailableBonusPoints,
		ActiveBonusCount:       kpis.ActiveBonusCount,
		TotalPenaltyCount:      kpis.TotalPenaltyCount,
		PendingPunishmentCount: kpis.PendingPunishmentCount,
	}
}

func StudentHistoryFromRows(rows []repository.ListStudentHistoryRow) []dto.StudentHistoryItemDto {
	history := make([]dto.StudentHistoryItemDto, 0, len(rows))

	for _, row := range rows {
		item := dto.StudentHistoryItemDto{
			Type:      row.Type,
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
		}

		switch row.Type {
		case "penalty":
			item.PenaltyTypeID = cloneUUIDPtr(row.PenaltyTypeID)
			item.PenaltyTypeName = cloneStringPtr(row.PenaltyTypeName)
		case "bonus":
			item.BonusTypeID = cloneUUIDPtr(row.BonusTypeID)
			item.BonusTypeName = cloneStringPtr(row.BonusTypeName)
			item.Points = cloneFloat64Ptr(row.Points)
			item.UsedAt = cloneTimePtrFromDouble(row.UsedAt)
		case "punishment":
			item.PunishmentTypeID = cloneUUIDValue(row.PunishmentTypeID)
			item.PunishmentTypeName = cloneStringValue(row.PunishmentTypeName)
			item.TriggeringRuleID = cloneUUIDPtr(row.TriggeringRuleID)
			item.TriggeringRuleName = cloneStringPtr(row.TriggeringRuleName)
			item.Automated = cloneBoolValue(row.Automated)
			item.DueAt = cloneTimeValue(row.DueAt)
			item.ResolvedAt = cloneTimePtrFromDouble(row.ResolvedAt)
		}

		history = append(history, item)
	}

	return history
}

func cloneUUIDPtr(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}

	cloned := *value
	return &cloned
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}

	cloned := *value
	if cloned == "" {
		return nil
	}
	return &cloned
}

func cloneStringValue(value string) *string {
	if value == "" {
		return nil
	}

	cloned := value
	return &cloned
}

func cloneFloat64Ptr(value *float64) *float64 {
	if value == nil {
		return nil
	}

	cloned := *value
	return &cloned
}

func cloneBoolValue(value bool) *bool {
	cloned := value
	return &cloned
}

func cloneTimeValue(value time.Time) *time.Time {
	cloned := value
	return &cloned
}

func cloneUUIDValue(value uuid.UUID) *uuid.UUID {
	cloned := value
	return &cloned
}

func cloneTimePtrFromDouble(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}

	cloned := *value
	return &cloned
}
