package inmemory

import (
	"context"
	"sort"

	"github.com/mageas/the-punisher-backend/internal/repository"
)

const (
	OpGetStudentKpis     = "GetStudentKpis"
	OpListStudentHistory = "ListStudentHistory"
)

func (r *Repository) GetStudentKpis(_ context.Context, arg repository.GetStudentKpisParams) (repository.GetStudentKpisRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpGetStudentKpis); err != nil {
		return repository.GetStudentKpisRow{}, err
	}

	row := repository.GetStudentKpisRow{}

	for _, bonus := range r.bonuses {
		if bonus.StudentID != arg.StudentID || bonus.UserID != arg.UserID || hasTime(bonus.UsedAt) {
			continue
		}
		row.AvailableBonusPoints += bonus.Points
		row.ActiveBonusCount++
	}

	for _, penalty := range r.penalties {
		if penalty.StudentID == arg.StudentID && penalty.UserID == arg.UserID {
			row.TotalPenaltyCount++
		}
	}

	for _, punishment := range r.punishments {
		if punishment.StudentID != arg.StudentID || punishment.UserID != arg.UserID || hasTime(punishment.ResolvedAt) {
			continue
		}
		row.PendingPunishmentCount++
	}

	return row, nil
}

func (r *Repository) ListStudentHistory(_ context.Context, arg repository.ListStudentHistoryParams) ([]repository.ListStudentHistoryRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListStudentHistory); err != nil {
		return nil, err
	}

	items := make([]repository.ListStudentHistoryRow, 0)

	for _, punishment := range r.punishments {
		if punishment.StudentID != arg.StudentID || punishment.UserID != arg.UserID {
			continue
		}

		triggeringRuleName := r.triggeringRuleNameForPunishmentLocked(punishment.TriggeringRuleID)
		item := repository.ListStudentHistoryRow{
			Type:               "punishment",
			ID:                 punishment.ID,
			CreatedAt:          punishment.CreatedAt,
			PenaltyTypeID:      nil,
			PenaltyTypeName:    nil,
			BonusTypeID:        nil,
			BonusTypeName:      nil,
			Points:             nil,
			UsedAt:             nil,
			PunishmentTypeID:   punishment.PunishmentTypeID,
			PunishmentTypeName: r.punishmentTypeNameLocked(punishment.PunishmentTypeID),
			TriggeringRuleID:   punishment.TriggeringRuleID,
			TriggeringRuleName: nil,
			Automated:          punishment.Automated,
			DueAt:              punishment.DueAt,
			ResolvedAt:         punishment.ResolvedAt,
		}
		if item.TriggeringRuleID != nil && triggeringRuleName != "" {
			item.TriggeringRuleName = &triggeringRuleName
		}

		items = append(items, item)
	}

	for _, penalty := range r.penalties {
		if penalty.StudentID != arg.StudentID || penalty.UserID != arg.UserID {
			continue
		}

		penaltyTypeName := r.penaltyTypeNameForPenaltyLocked(penalty.PenaltyTypeID)
		items = append(items, repository.ListStudentHistoryRow{
			Type:               "penalty",
			ID:                 penalty.ID,
			CreatedAt:          penalty.CreatedAt,
			PenaltyTypeID:      &penalty.PenaltyTypeID,
			PenaltyTypeName:    &penaltyTypeName,
			BonusTypeID:        nil,
			BonusTypeName:      nil,
			Points:             nil,
			UsedAt:             nil,
			PunishmentTypeID:   penalty.PenaltyTypeID,
			PunishmentTypeName: penaltyTypeName,
			TriggeringRuleID:   nil,
			TriggeringRuleName: nil,
			Automated:          false,
			DueAt:              penalty.CreatedAt,
			ResolvedAt:         nil,
		})
	}

	for _, bonus := range r.bonuses {
		if bonus.StudentID != arg.StudentID || bonus.UserID != arg.UserID {
			continue
		}

		bonusTypeName := r.bonusTypeNameForBonusLocked(bonus.BonusTypeID)
		items = append(items, repository.ListStudentHistoryRow{
			Type:               "bonus",
			ID:                 bonus.ID,
			CreatedAt:          bonus.CreatedAt,
			PenaltyTypeID:      nil,
			PenaltyTypeName:    nil,
			BonusTypeID:        &bonus.BonusTypeID,
			BonusTypeName:      &bonusTypeName,
			Points:             &bonus.Points,
			UsedAt:             bonus.UsedAt,
			PunishmentTypeID:   bonus.BonusTypeID,
			PunishmentTypeName: bonusTypeName,
			TriggeringRuleID:   nil,
			TriggeringRuleName: nil,
			Automated:          false,
			DueAt:              bonus.CreatedAt,
			ResolvedAt:         nil,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].ID.String() > items[j].ID.String()
		}
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	paginated := paginate(items, arg.QueryOffset, arg.QueryLimit)
	return paginated, nil
}
