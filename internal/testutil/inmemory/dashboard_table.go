package inmemory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

const (
	OpGetDashboardKpis                = "GetDashboardKpis"
	OpListDashboardRecentPenalties    = "ListDashboardRecentPenalties"
	OpListDashboardRecentBonuses      = "ListDashboardRecentBonuses"
	OpListDashboardPendingPunishments = "ListDashboardPendingPunishments"
)

func (r *Repository) dashboardStudentIDsLocked(userID uuid.UUID, classroomID *uuid.UUID) map[uuid.UUID]struct{} {
	studentIDs := make(map[uuid.UUID]struct{})

	if classroomID == nil {
		for _, student := range r.students {
			if student.UserID == userID {
				studentIDs[student.ID] = struct{}{}
			}
		}

		return studentIDs
	}

	filteredClassroomID := *classroomID
	classroom, classroomExists := r.classrooms[filteredClassroomID]
	if !classroomExists || classroom.UserID != userID {
		return studentIDs
	}

	for _, relation := range r.studentClassrooms {
		if relation.ClassroomID != filteredClassroomID {
			continue
		}

		student, studentExists := r.students[relation.StudentID]
		if !studentExists || student.UserID != userID {
			continue
		}

		studentIDs[student.ID] = struct{}{}
	}

	return studentIDs
}

func (r *Repository) GetDashboardKpis(_ context.Context, arg repository.GetDashboardKpisParams) (repository.GetDashboardKpisRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpGetDashboardKpis); err != nil {
		return repository.GetDashboardKpisRow{}, err
	}

	studentIDs := r.dashboardStudentIDsLocked(arg.UserID, arg.ClassroomID)
	row := repository.GetDashboardKpisRow{
		StudentCount: int64(len(studentIDs)),
	}

	for _, bonus := range r.bonuses {
		if bonus.UserID != arg.UserID {
			continue
		}
		if _, ok := studentIDs[bonus.StudentID]; !ok {
			continue
		}
		row.TotalBonusPoints += bonus.Points
		if !hasTime(bonus.UsedAt) {
			row.AvailableBonusPoints += bonus.Points
			row.UnusedBonusCount++
		}
	}

	for _, penalty := range r.penalties {
		if penalty.UserID != arg.UserID {
			continue
		}
		if _, ok := studentIDs[penalty.StudentID]; ok {
			row.PenaltyCount++
		}
	}

	for _, punishment := range r.punishments {
		if punishment.UserID != arg.UserID {
			continue
		}
		if _, ok := studentIDs[punishment.StudentID]; !ok {
			continue
		}
		row.TotalPunishmentCount++
		if !hasTime(punishment.ResolvedAt) {
			if punishment.DueAt.Before(time.Now().UTC()) {
				row.OverduePunishmentCount++
			}
			row.PendingPunishmentCount++
		}
	}

	return row, nil
}

func (r *Repository) ListDashboardRecentPenalties(_ context.Context, arg repository.ListDashboardRecentPenaltiesParams) ([]repository.ListDashboardRecentPenaltiesRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListDashboardRecentPenalties); err != nil {
		return nil, err
	}

	studentIDs := r.dashboardStudentIDsLocked(arg.UserID, arg.ClassroomID)
	items := make([]repository.Penalty, 0)
	for _, penalty := range r.penalties {
		if penalty.UserID != arg.UserID {
			continue
		}
		if _, ok := studentIDs[penalty.StudentID]; ok {
			items = append(items, penalty)
		}
	}

	sortPenaltiesByCreatedAtDesc(items)
	paginated := paginate(items, 0, arg.QueryLimit)

	rows := make([]repository.ListDashboardRecentPenaltiesRow, 0, len(paginated))
	for _, penalty := range paginated {
		studentFirstName, studentLastName := r.studentNamesLocked(penalty.StudentID)
		rows = append(rows, repository.ListDashboardRecentPenaltiesRow{
			ID:               penalty.ID,
			UserID:           penalty.UserID,
			StudentID:        penalty.StudentID,
			PenaltyTypeID:    penalty.PenaltyTypeID,
			CreatedAt:        penalty.CreatedAt,
			StudentFirstName: studentFirstName,
			StudentLastName:  studentLastName,
			PenaltyTypeName:  r.penaltyTypeNameForPenaltyLocked(penalty.PenaltyTypeID),
		})
	}

	return rows, nil
}

func (r *Repository) ListDashboardRecentBonuses(_ context.Context, arg repository.ListDashboardRecentBonusesParams) ([]repository.ListDashboardRecentBonusesRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListDashboardRecentBonuses); err != nil {
		return nil, err
	}

	studentIDs := r.dashboardStudentIDsLocked(arg.UserID, arg.ClassroomID)
	items := make([]repository.Bonus, 0)
	for _, bonus := range r.bonuses {
		if bonus.UserID != arg.UserID {
			continue
		}
		if _, ok := studentIDs[bonus.StudentID]; ok {
			items = append(items, bonus)
		}
	}

	sortBonusesByCreatedAtDesc(items)
	paginated := paginate(items, 0, arg.QueryLimit)

	rows := make([]repository.ListDashboardRecentBonusesRow, 0, len(paginated))
	for _, bonus := range paginated {
		studentFirstName, studentLastName := r.studentNamesLocked(bonus.StudentID)
		rows = append(rows, repository.ListDashboardRecentBonusesRow{
			ID:               bonus.ID,
			UserID:           bonus.UserID,
			StudentID:        bonus.StudentID,
			BonusTypeID:      bonus.BonusTypeID,
			Points:           bonus.Points,
			CreatedAt:        bonus.CreatedAt,
			UsedAt:           bonus.UsedAt,
			StudentFirstName: studentFirstName,
			StudentLastName:  studentLastName,
			BonusTypeName:    r.bonusTypeNameForBonusLocked(bonus.BonusTypeID),
		})
	}

	return rows, nil
}

func (r *Repository) ListDashboardPendingPunishments(_ context.Context, arg repository.ListDashboardPendingPunishmentsParams) ([]repository.ListDashboardPendingPunishmentsRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListDashboardPendingPunishments); err != nil {
		return nil, err
	}

	studentIDs := r.dashboardStudentIDsLocked(arg.UserID, arg.ClassroomID)
	items := make([]repository.Punishment, 0)
	for _, punishment := range r.punishments {
		if punishment.UserID != arg.UserID || hasTime(punishment.ResolvedAt) {
			continue
		}
		if _, ok := studentIDs[punishment.StudentID]; ok {
			items = append(items, punishment)
		}
	}

	sortPunishmentsByCreatedAtDesc(items)
	paginated := paginate(items, 0, arg.QueryLimit)

	rows := make([]repository.ListDashboardPendingPunishmentsRow, 0, len(paginated))
	for _, punishment := range paginated {
		studentFirstName, studentLastName := r.studentNamesLocked(punishment.StudentID)
		triggeringRuleName := r.triggeringRuleNameForPunishmentLocked(punishment.TriggeringRuleID)
		rows = append(rows, repository.ListDashboardPendingPunishmentsRow{
			ID:                 punishment.ID,
			UserID:             punishment.UserID,
			StudentID:          punishment.StudentID,
			PunishmentTypeID:   punishment.PunishmentTypeID,
			TriggeringRuleID:   punishment.TriggeringRuleID,
			Automated:          punishment.Automated,
			CreatedAt:          punishment.CreatedAt,
			DueAt:              punishment.DueAt,
			ResolvedAt:         punishment.ResolvedAt,
			StudentFirstName:   studentFirstName,
			StudentLastName:    studentLastName,
			PunishmentTypeName: r.punishmentTypeNameLocked(punishment.PunishmentTypeID),
			TriggeringRuleName: triggeringRuleNameAsText(triggeringRuleName),
		})
	}

	return rows, nil
}
