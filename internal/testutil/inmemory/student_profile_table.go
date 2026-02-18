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
	OpGetStudentProfileKpis                = "GetStudentProfileKpis"
	OpListStudentProfileClassrooms         = "ListStudentProfileClassrooms"
	OpListStudentProfilePendingPunishments = "ListStudentProfilePendingPunishments"
	OpListStudentProfileAvailableBonuses   = "ListStudentProfileAvailableBonuses"
	OpListStudentProfileHistory            = "ListStudentProfileHistory"
)

var studentProfileHistoryFallbackDueAt = time.Unix(0, 0).UTC()

func (r *Repository) GetStudentProfileKpis(_ context.Context, arg repository.GetStudentProfileKpisParams) (repository.GetStudentProfileKpisRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpGetStudentProfileKpis); err != nil {
		return repository.GetStudentProfileKpisRow{}, err
	}

	row := repository.GetStudentProfileKpisRow{}

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

func (r *Repository) ListStudentProfileClassrooms(_ context.Context, arg repository.ListStudentProfileClassroomsParams) ([]repository.ListStudentProfileClassroomsRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListStudentProfileClassrooms); err != nil {
		return nil, err
	}

	items := make([]repository.Classroom, 0)
	for _, relation := range r.studentClassrooms {
		if relation.StudentID != arg.StudentID {
			continue
		}

		classroom, classroomExists := r.classrooms[relation.ClassroomID]
		if !classroomExists || classroom.UserID != arg.UserID {
			continue
		}

		items = append(items, classroom)
	}

	sortClassroomsByCreatedAtDesc(items)

	rows := make([]repository.ListStudentProfileClassroomsRow, 0, len(items))
	for _, classroom := range items {
		rows = append(rows, repository.ListStudentProfileClassroomsRow{
			ID:   classroom.ID,
			Name: classroom.Name,
		})
	}

	return rows, nil
}

func (r *Repository) ListStudentProfilePendingPunishments(_ context.Context, arg repository.ListStudentProfilePendingPunishmentsParams) ([]repository.ListStudentProfilePendingPunishmentsRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListStudentProfilePendingPunishments); err != nil {
		return nil, err
	}

	items := make([]repository.Punishment, 0)
	for _, punishment := range r.punishments {
		if punishment.StudentID != arg.StudentID || punishment.UserID != arg.UserID || punishment.ResolvedAt.Valid {
			continue
		}
		items = append(items, punishment)
	}

	sortPunishmentsByCreatedAtDesc(items)

	rows := make([]repository.ListStudentProfilePendingPunishmentsRow, 0, len(items))
	for _, punishment := range items {
		triggeringRuleName := r.triggeringRuleNameForPunishmentLocked(punishment.TriggeringRuleID)
		rows = append(rows, repository.ListStudentProfilePendingPunishmentsRow{
			ID:                 punishment.ID,
			PunishmentTypeID:   punishment.PunishmentTypeID,
			TriggeringRuleID:   punishment.TriggeringRuleID,
			CreatedAt:          punishment.CreatedAt,
			DueAt:              punishment.DueAt,
			PunishmentTypeName: r.punishmentTypeNameLocked(punishment.PunishmentTypeID),
			TriggeringRuleName: triggeringRuleNameAsText(triggeringRuleName),
		})
	}

	return rows, nil
}

func (r *Repository) ListStudentProfileAvailableBonuses(_ context.Context, arg repository.ListStudentProfileAvailableBonusesParams) ([]repository.ListStudentProfileAvailableBonusesRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListStudentProfileAvailableBonuses); err != nil {
		return nil, err
	}

	items := make([]repository.Bonus, 0)
	for _, bonus := range r.bonuses {
		if bonus.StudentID != arg.StudentID || bonus.UserID != arg.UserID || bonus.UsedAt.Valid {
			continue
		}
		items = append(items, bonus)
	}

	sortBonusesByCreatedAtDesc(items)

	rows := make([]repository.ListStudentProfileAvailableBonusesRow, 0, len(items))
	for _, bonus := range items {
		rows = append(rows, repository.ListStudentProfileAvailableBonusesRow{
			ID:            bonus.ID,
			BonusTypeID:   bonus.BonusTypeID,
			Points:        bonus.Points,
			CreatedAt:     bonus.CreatedAt,
			BonusTypeName: r.bonusTypeNameForBonusLocked(bonus.BonusTypeID),
		})
	}

	return rows, nil
}

func (r *Repository) ListStudentProfileHistory(_ context.Context, arg repository.ListStudentProfileHistoryParams) ([]repository.ListStudentProfileHistoryRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListStudentProfileHistory); err != nil {
		return nil, err
	}

	items := make([]repository.ListStudentProfileHistoryRow, 0)

	for _, punishment := range r.punishments {
		if punishment.StudentID != arg.StudentID || punishment.UserID != arg.UserID {
			continue
		}

		triggeringRuleName := r.triggeringRuleNameForPunishmentLocked(punishment.TriggeringRuleID)
		items = append(items, repository.ListStudentProfileHistoryRow{
			Type:               "punishment",
			ID:                 punishment.ID,
			CreatedAt:          punishment.CreatedAt,
			PenaltyTypeID:      uuid.Nil,
			PenaltyTypeName:    "",
			BonusTypeID:        uuid.Nil,
			BonusTypeName:      "",
			Points:             0,
			UsedAt:             studentProfileHistoryFallbackDueAt,
			PunishmentTypeID:   punishment.PunishmentTypeID,
			PunishmentTypeName: r.punishmentTypeNameLocked(punishment.PunishmentTypeID),
			TriggeringRuleID:   pgtype.UUID{Bytes: uuid.Nil, Valid: true},
			TriggeringRuleName: "",
			DueAt:              punishment.DueAt,
			ResolvedAt:         pgtype.Timestamptz{Time: studentProfileHistoryFallbackDueAt, Valid: true},
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

		items = append(items, repository.ListStudentProfileHistoryRow{
			Type:               "penalty",
			ID:                 penalty.ID,
			CreatedAt:          penalty.CreatedAt,
			PenaltyTypeID:      penalty.PenaltyTypeID,
			PenaltyTypeName:    r.penaltyTypeNameForPenaltyLocked(penalty.PenaltyTypeID),
			BonusTypeID:        uuid.Nil,
			BonusTypeName:      "",
			Points:             0,
			UsedAt:             studentProfileHistoryFallbackDueAt,
			PunishmentTypeID:   uuid.Nil,
			PunishmentTypeName: "",
			TriggeringRuleID:   pgtype.UUID{Bytes: uuid.Nil, Valid: true},
			TriggeringRuleName: "",
			DueAt:              studentProfileHistoryFallbackDueAt,
			ResolvedAt:         pgtype.Timestamptz{Time: studentProfileHistoryFallbackDueAt, Valid: true},
		})
	}

	for _, bonus := range r.bonuses {
		if bonus.StudentID != arg.StudentID || bonus.UserID != arg.UserID {
			continue
		}

		items = append(items, repository.ListStudentProfileHistoryRow{
			Type:               "bonus",
			ID:                 bonus.ID,
			CreatedAt:          bonus.CreatedAt,
			PenaltyTypeID:      uuid.Nil,
			PenaltyTypeName:    "",
			BonusTypeID:        bonus.BonusTypeID,
			BonusTypeName:      r.bonusTypeNameForBonusLocked(bonus.BonusTypeID),
			Points:             bonus.Points,
			UsedAt:             studentProfileHistoryFallbackDueAt,
			PunishmentTypeID:   uuid.Nil,
			PunishmentTypeName: "",
			TriggeringRuleID:   pgtype.UUID{Bytes: uuid.Nil, Valid: true},
			TriggeringRuleName: "",
			DueAt:              studentProfileHistoryFallbackDueAt,
			ResolvedAt:         pgtype.Timestamptz{Time: studentProfileHistoryFallbackDueAt, Valid: true},
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
