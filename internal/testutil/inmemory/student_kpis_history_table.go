package inmemory

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

const (
	OpGetStudentKpis     = "GetStudentKpis"
	OpListStudentHistory = "ListStudentHistory"
)

var studentHistoryFallbackDueAt = time.Unix(0, 0).UTC()

func (r *Repository) GetStudentKpis(_ context.Context, arg repository.GetStudentKpisParams) (repository.GetStudentKpisRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpGetStudentKpis); err != nil {
		return repository.GetStudentKpisRow{}, err
	}

	row := repository.GetStudentKpisRow{}

	for _, bonus := range r.bonuses {
		if bonus.StudentID != arg.StudentID || bonus.UserID != arg.UserID || bonus.UsedAt.Valid {
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
		if punishment.StudentID != arg.StudentID || punishment.UserID != arg.UserID || punishment.ResolvedAt.Valid {
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
		items = append(items, repository.ListStudentHistoryRow{
			Type:               "punishment",
			ID:                 punishment.ID,
			CreatedAt:          punishment.CreatedAt,
			PenaltyTypeID:      uuid.Nil,
			PenaltyTypeName:    "",
			BonusTypeID:        uuid.Nil,
			BonusTypeName:      "",
			Points:             0,
			UsedAt:             studentHistoryFallbackDueAt,
			PunishmentTypeID:   punishment.PunishmentTypeID,
			PunishmentTypeName: r.punishmentTypeNameLocked(punishment.PunishmentTypeID),
			TriggeringRuleID:   pgtype.UUID{Bytes: uuid.Nil, Valid: true},
			TriggeringRuleName: "",
			Automated:          punishment.Automated,
			DueAt:              punishment.DueAt,
			ResolvedAt:         pgtype.Timestamptz{Time: studentHistoryFallbackDueAt, Valid: true},
		})
		if punishment.TriggeringRuleID.Valid {
			items[len(items)-1].TriggeringRuleID = punishment.TriggeringRuleID
			items[len(items)-1].TriggeringRuleName = triggeringRuleName
		}
		if punishment.ResolvedAt.Valid {
			items[len(items)-1].ResolvedAt = punishment.ResolvedAt
		}
	}

	for _, penalty := range r.penalties {
		if penalty.StudentID != arg.StudentID || penalty.UserID != arg.UserID {
			continue
		}

		items = append(items, repository.ListStudentHistoryRow{
			Type:               "penalty",
			ID:                 penalty.ID,
			CreatedAt:          penalty.CreatedAt,
			PenaltyTypeID:      penalty.PenaltyTypeID,
			PenaltyTypeName:    r.penaltyTypeNameForPenaltyLocked(penalty.PenaltyTypeID),
			BonusTypeID:        uuid.Nil,
			BonusTypeName:      "",
			Points:             0,
			UsedAt:             studentHistoryFallbackDueAt,
			PunishmentTypeID:   uuid.Nil,
			PunishmentTypeName: "",
			TriggeringRuleID:   pgtype.UUID{Bytes: uuid.Nil, Valid: true},
			TriggeringRuleName: "",
			Automated:          false,
			DueAt:              studentHistoryFallbackDueAt,
			ResolvedAt:         pgtype.Timestamptz{Time: studentHistoryFallbackDueAt, Valid: true},
		})
	}

	for _, bonus := range r.bonuses {
		if bonus.StudentID != arg.StudentID || bonus.UserID != arg.UserID {
			continue
		}

		items = append(items, repository.ListStudentHistoryRow{
			Type:               "bonus",
			ID:                 bonus.ID,
			CreatedAt:          bonus.CreatedAt,
			PenaltyTypeID:      uuid.Nil,
			PenaltyTypeName:    "",
			BonusTypeID:        bonus.BonusTypeID,
			BonusTypeName:      r.bonusTypeNameForBonusLocked(bonus.BonusTypeID),
			Points:             bonus.Points,
			UsedAt:             studentHistoryFallbackDueAt,
			PunishmentTypeID:   uuid.Nil,
			PunishmentTypeName: "",
			TriggeringRuleID:   pgtype.UUID{Bytes: uuid.Nil, Valid: true},
			TriggeringRuleName: "",
			Automated:          false,
			DueAt:              studentHistoryFallbackDueAt,
			ResolvedAt:         pgtype.Timestamptz{Time: studentHistoryFallbackDueAt, Valid: true},
		})
		if bonus.UsedAt.Valid {
			items[len(items)-1].UsedAt = bonus.UsedAt.Time
		}
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
