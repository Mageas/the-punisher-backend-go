package sqlcmapper

import (
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
			item.PenaltyTypeID = row.PenaltyTypeID
			item.PenaltyTypeName = row.PenaltyTypeName
		case "bonus":
			item.BonusTypeID = row.BonusTypeID
			item.BonusTypeName = row.BonusTypeName
			item.Points = row.Points
			item.UsedAt = row.UsedAt
		case "punishment":
			item.PunishmentTypeID = &row.PunishmentTypeID
			item.PunishmentTypeName = &row.PunishmentTypeName
			item.TriggeringRuleID = row.TriggeringRuleID
			item.TriggeringRuleName = row.TriggeringRuleName
			item.Automated = &row.Automated
			item.DueAt = &row.DueAt
			item.ResolvedAt = row.ResolvedAt
		}

		history = append(history, item)
	}

	return history
}
