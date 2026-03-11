package sqlcmapper

import (
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func DashboardFromRows(
	kpis *repository.GetDashboardKpisRow,
	penalties []repository.ListDashboardRecentPenaltiesRow,
	bonuses []repository.ListDashboardRecentBonusesRow,
	punishments []repository.ListDashboardPendingPunishmentsRow,
) *dto.ReturnDashboardDto {
	response := &dto.ReturnDashboardDto{
		RecentPenalties:    dashboardPenaltiesFromRows(penalties),
		RecentBonuses:      dashboardBonusesFromRows(bonuses),
		PendingPunishments: dashboardPendingPunishmentsFromRows(punishments),
	}

	if mappedKpis := DashboardKpisFromRow(kpis); mappedKpis != nil {
		response.Kpis = *mappedKpis
	}

	return response
}

func DashboardKpisFromRow(kpis *repository.GetDashboardKpisRow) *dto.DashboardKpisDto {
	if kpis == nil {
		return nil
	}

	return &dto.DashboardKpisDto{
		StudentCount:           kpis.StudentCount,
		AvailableBonusPoints:   roundResponseFloat(kpis.AvailableBonusPoints),
		TotalBonusPoints:       roundResponseFloat(kpis.TotalBonusPoints),
		UnusedBonusCount:       kpis.UnusedBonusCount,
		PenaltyCount:           kpis.PenaltyCount,
		TotalPunishmentCount:   kpis.TotalPunishmentCount,
		OverduePunishmentCount: kpis.OverduePunishmentCount,
		PendingPunishmentCount: kpis.PendingPunishmentCount,
	}
}

func dashboardPenaltiesFromRows(rows []repository.ListDashboardRecentPenaltiesRow) []*dto.ReturnPenaltyDto {
	responses := make([]*dto.ReturnPenaltyDto, 0, len(rows))

	for _, row := range rows {
		response := buildReturnPenaltyDto(
			row.ID,
			row.StudentID,
			row.StudentFirstName,
			row.StudentLastName,
			row.PenaltyTypeID,
			row.PenaltyTypeName,
			row.CreatedAt,
			row.OccurredAt,
			row.EvaluationLabel,
		)
		if response != nil {
			responses = append(responses, response)
		}
	}

	return responses
}

func dashboardBonusesFromRows(rows []repository.ListDashboardRecentBonusesRow) []*dto.ReturnBonusDto {
	responses := make([]*dto.ReturnBonusDto, 0, len(rows))

	for _, row := range rows {
		response := buildReturnBonusDto(
			row.ID,
			row.StudentID,
			row.StudentFirstName,
			row.StudentLastName,
			row.BonusTypeID,
			row.BonusTypeName,
			row.Points,
			row.CreatedAt,
			row.OccurredAt,
			row.EvaluationLabel,
			row.UsedAt,
		)
		if response != nil {
			responses = append(responses, response)
		}
	}

	return responses
}

func dashboardPendingPunishmentsFromRows(rows []repository.ListDashboardPendingPunishmentsRow) []*dto.ReturnPunishmentDto {
	responses := make([]*dto.ReturnPunishmentDto, 0, len(rows))

	for _, row := range rows {
		response := buildReturnPunishmentDto(
			row.ID,
			row.StudentID,
			row.StudentFirstName,
			row.StudentLastName,
			row.PunishmentTypeID,
			row.PunishmentTypeName,
			row.TriggeringRuleID,
			punishmentTriggeringRuleNameFromText(row.TriggeringRuleName),
			row.Automated,
			row.CreatedAt,
			row.OccurredAt,
			row.EvaluationLabel,
			row.DueAt,
			row.ResolvedAt,
		)
		if response != nil {
			responses = append(responses, response)
		}
	}

	return responses
}
