package dto

import "github.com/mageas/the-punisher-backend/internal/repository"

type DashboardKpisDto struct {
	StudentCount           int64   `json:"student_count"`
	AvailableBonusPoints   float64 `json:"available_bonus_points"`
	UnusedBonusCount       int64   `json:"unused_bonus_count"`
	PenaltyCount           int64   `json:"penalty_count"`
	PendingPunishmentCount int64   `json:"pending_punishment_count"`
}

type ReturnDashboardDto struct {
	Kpis               DashboardKpisDto       `json:"kpis"`
	RecentPenalties    []*ReturnPenaltyDto    `json:"recent_penalties"`
	RecentBonuses      []*ReturnBonusDto      `json:"recent_bonuses"`
	PendingPunishments []*ReturnPunishmentDto `json:"pending_punishments"`
}

func DashboardFromRows(
	kpis *repository.GetDashboardKpisRow,
	penalties []repository.ListDashboardRecentPenaltiesRow,
	bonuses []repository.ListDashboardRecentBonusesRow,
	punishments []repository.ListDashboardPendingPunishmentsRow,
) *ReturnDashboardDto {
	response := &ReturnDashboardDto{
		RecentPenalties:    dashboardPenaltiesFromRows(penalties),
		RecentBonuses:      dashboardBonusesFromRows(bonuses),
		PendingPunishments: dashboardPendingPunishmentsFromRows(punishments),
	}

	if kpis != nil {
		response.Kpis = DashboardKpisDto{
			StudentCount:           kpis.StudentCount,
			AvailableBonusPoints:   kpis.AvailableBonusPoints,
			UnusedBonusCount:       kpis.UnusedBonusCount,
			PenaltyCount:           kpis.PenaltyCount,
			PendingPunishmentCount: kpis.PendingPunishmentCount,
		}
	}

	return response
}

func dashboardPenaltiesFromRows(rows []repository.ListDashboardRecentPenaltiesRow) []*ReturnPenaltyDto {
	dtos := make([]*ReturnPenaltyDto, 0, len(rows))

	for _, row := range rows {
		dto := buildReturnPenaltyDto(
			row.ID,
			row.StudentID,
			row.StudentFirstName,
			row.StudentLastName,
			row.PenaltyTypeID,
			row.PenaltyTypeName,
			row.CreatedAt,
		)
		if dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
}

func dashboardBonusesFromRows(rows []repository.ListDashboardRecentBonusesRow) []*ReturnBonusDto {
	dtos := make([]*ReturnBonusDto, 0, len(rows))

	for _, row := range rows {
		dto := buildReturnBonusDto(
			row.ID,
			row.StudentID,
			row.StudentFirstName,
			row.StudentLastName,
			row.BonusTypeID,
			row.BonusTypeName,
			row.Points,
			row.CreatedAt,
			row.UsedAt,
		)
		if dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
}

func dashboardPendingPunishmentsFromRows(rows []repository.ListDashboardPendingPunishmentsRow) []*ReturnPunishmentDto {
	dtos := make([]*ReturnPunishmentDto, 0, len(rows))

	for _, row := range rows {
		dto := buildReturnPunishmentDto(
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
			row.DueAt,
			row.ResolvedAt,
		)
		if dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
}
