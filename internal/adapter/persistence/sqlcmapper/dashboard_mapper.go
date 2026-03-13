package sqlcmapper

import (
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func DashboardFromRow(
	kpis *repository.GetDashboardKpisRow,
) *dto.ReturnDashboardDto {
	response := &dto.ReturnDashboardDto{}

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
