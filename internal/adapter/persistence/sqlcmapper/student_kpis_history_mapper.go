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
		TotalBonusPoints:       kpis.TotalBonusPoints,
		ActiveBonusCount:       kpis.ActiveBonusCount,
		PenaltyCount:           kpis.PenaltyCount,
		TotalPenaltyCount:      kpis.TotalPenaltyCount,
		TotalPunishmentCount:   kpis.TotalPunishmentCount,
		OverduePunishmentCount: kpis.OverduePunishmentCount,
		PendingPunishmentCount: kpis.PendingPunishmentCount,
	}
}

func StudentHistoryFromRows(rows []repository.ListStudentHistoryRow) []dto.StudentHistoryItemDto {
	history := make([]dto.StudentHistoryItemDto, 0, len(rows))

	for _, row := range rows {
		item := dto.StudentHistoryItemDto{
			Type:            row.Type,
			ID:              row.ID,
			CreatedAt:       normalizeAPITime(row.CreatedAt),
			OccurredAt:      normalizeAPITime(row.OccurredAt),
			EvaluationLabel: row.EvaluationLabel,
		}

		switch row.Type {
		case "penalty":
			item.PenaltyTypeID = row.PenaltyTypeID
			item.PenaltyTypeName = row.PenaltyTypeName
		case "bonus":
			item.BonusTypeID = row.BonusTypeID
			item.BonusTypeName = row.BonusTypeName
			item.Points = row.Points
			item.UsedAt = normalizeOptionalAPITime(row.UsedAt)
		case "punishment":
			punishmentTypeID := row.PunishmentTypeID
			punishmentTypeName := row.PunishmentTypeName
			automated := row.Automated
			dueAt := normalizeAPITime(row.DueAt)

			item.PunishmentTypeID = &punishmentTypeID
			item.PunishmentTypeName = &punishmentTypeName
			item.TriggeringRuleID = punishmentTriggeringRuleID(row.TriggeringRuleID)
			item.TriggeringRuleName = punishmentTriggeringRuleNameFromText(row.TriggeringRuleName)
			item.Automated = &automated
			item.DueAt = &dueAt
			item.ResolvedAt = normalizeOptionalAPITime(row.ResolvedAt)
		}

		history = append(history, item)
	}

	return history
}
