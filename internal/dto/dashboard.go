package dto

type DashboardKpisDto struct {
	StudentCount           int64   `json:"student_count"`
	AvailableBonusPoints   float64 `json:"available_bonus_points"`
	TotalBonusPoints       float64 `json:"total_bonus_points"`
	UnusedBonusCount       int64   `json:"unused_bonus_count"`
	PenaltyCount           int64   `json:"penalty_count"`
	TotalPunishmentCount   int64   `json:"total_punishment_count"`
	OverduePunishmentCount int64   `json:"overdue_punishment_count"`
	PendingPunishmentCount int64   `json:"pending_punishment_count"`
}

type ReturnDashboardDto struct {
	Kpis               DashboardKpisDto       `json:"kpis"`
	RecentPenalties    []*ReturnPenaltyDto    `json:"recent_penalties"`
	RecentBonuses      []*ReturnBonusDto      `json:"recent_bonuses"`
	PendingPunishments []*ReturnPunishmentDto `json:"pending_punishments"`
}
