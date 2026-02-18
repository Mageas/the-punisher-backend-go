package inmemory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

const (
	OpCreatePunishment          = "CreatePunishment"
	OpCreatePunishmentFromRule  = "CreatePunishmentFromRule"
	OpGetPunishmentByUser       = "GetPunishmentByUser"
	OpCountPunishmentsByUser    = "CountPunishmentsByUser"
	OpListPunishmentsByUser     = "ListPunishmentsByUser"
	OpCountPunishmentsByStudent = "CountPunishmentsByStudent"
	OpListPunishmentsByStudent  = "ListPunishmentsByStudent"
	OpResolvePunishment         = "ResolvePunishment"
	OpDeletePunishmentByUser    = "DeletePunishmentByUser"
)

func (r *Repository) triggeringRuleNameForPunishmentLocked(triggeringRuleID pgtype.UUID) string {
	if !triggeringRuleID.Valid {
		return ""
	}

	ruleID := uuid.UUID(triggeringRuleID.Bytes)
	if rule, ok := r.rules[ruleID]; ok {
		return rule.Name
	}

	return ""
}

func triggeringRuleNameAsText(name string) pgtype.Text {
	if name == "" {
		return pgtype.Text{}
	}

	return pgtype.Text{String: name, Valid: true}
}

func (r *Repository) buildCreatePunishmentRowLocked(punishment repository.Punishment) repository.CreatePunishmentRow {
	studentFirstName, studentLastName := r.studentNamesLocked(punishment.StudentID)
	triggeringRuleName := r.triggeringRuleNameForPunishmentLocked(punishment.TriggeringRuleID)

	return repository.CreatePunishmentRow{
		ID:                 punishment.ID,
		UserID:             punishment.UserID,
		StudentID:          punishment.StudentID,
		PunishmentTypeID:   punishment.PunishmentTypeID,
		TriggeringRuleID:   punishment.TriggeringRuleID,
		CreatedAt:          punishment.CreatedAt,
		DueAt:              punishment.DueAt,
		ResolvedAt:         punishment.ResolvedAt,
		StudentFirstName:   studentFirstName,
		StudentLastName:    studentLastName,
		PunishmentTypeName: r.punishmentTypeNameLocked(punishment.PunishmentTypeID),
		TriggeringRuleName: triggeringRuleNameAsText(triggeringRuleName),
	}
}

func (r *Repository) buildCreatePunishmentFromRuleRowLocked(punishment repository.Punishment) repository.CreatePunishmentFromRuleRow {
	studentFirstName, studentLastName := r.studentNamesLocked(punishment.StudentID)
	triggeringRuleName := r.triggeringRuleNameForPunishmentLocked(punishment.TriggeringRuleID)

	return repository.CreatePunishmentFromRuleRow{
		ID:                 punishment.ID,
		UserID:             punishment.UserID,
		StudentID:          punishment.StudentID,
		PunishmentTypeID:   punishment.PunishmentTypeID,
		TriggeringRuleID:   punishment.TriggeringRuleID,
		CreatedAt:          punishment.CreatedAt,
		DueAt:              punishment.DueAt,
		ResolvedAt:         punishment.ResolvedAt,
		StudentFirstName:   studentFirstName,
		StudentLastName:    studentLastName,
		PunishmentTypeName: r.punishmentTypeNameLocked(punishment.PunishmentTypeID),
		TriggeringRuleName: triggeringRuleNameAsText(triggeringRuleName),
	}
}

func (r *Repository) buildGetPunishmentRowLocked(punishment repository.Punishment) repository.GetPunishmentByUserRow {
	studentFirstName, studentLastName := r.studentNamesLocked(punishment.StudentID)
	triggeringRuleName := r.triggeringRuleNameForPunishmentLocked(punishment.TriggeringRuleID)

	return repository.GetPunishmentByUserRow{
		ID:                 punishment.ID,
		UserID:             punishment.UserID,
		StudentID:          punishment.StudentID,
		PunishmentTypeID:   punishment.PunishmentTypeID,
		TriggeringRuleID:   punishment.TriggeringRuleID,
		CreatedAt:          punishment.CreatedAt,
		DueAt:              punishment.DueAt,
		ResolvedAt:         punishment.ResolvedAt,
		StudentFirstName:   studentFirstName,
		StudentLastName:    studentLastName,
		PunishmentTypeName: r.punishmentTypeNameLocked(punishment.PunishmentTypeID),
		TriggeringRuleName: triggeringRuleNameAsText(triggeringRuleName),
	}
}

func (r *Repository) buildListPunishmentByUserRowLocked(punishment repository.Punishment) repository.ListPunishmentsByUserRow {
	studentFirstName, studentLastName := r.studentNamesLocked(punishment.StudentID)
	triggeringRuleName := r.triggeringRuleNameForPunishmentLocked(punishment.TriggeringRuleID)

	return repository.ListPunishmentsByUserRow{
		ID:                 punishment.ID,
		UserID:             punishment.UserID,
		StudentID:          punishment.StudentID,
		PunishmentTypeID:   punishment.PunishmentTypeID,
		TriggeringRuleID:   punishment.TriggeringRuleID,
		CreatedAt:          punishment.CreatedAt,
		DueAt:              punishment.DueAt,
		ResolvedAt:         punishment.ResolvedAt,
		StudentFirstName:   studentFirstName,
		StudentLastName:    studentLastName,
		PunishmentTypeName: r.punishmentTypeNameLocked(punishment.PunishmentTypeID),
		TriggeringRuleName: triggeringRuleNameAsText(triggeringRuleName),
	}
}

func (r *Repository) buildListPunishmentByStudentRowLocked(punishment repository.Punishment) repository.ListPunishmentsByStudentRow {
	studentFirstName, studentLastName := r.studentNamesLocked(punishment.StudentID)
	triggeringRuleName := r.triggeringRuleNameForPunishmentLocked(punishment.TriggeringRuleID)

	return repository.ListPunishmentsByStudentRow{
		ID:                 punishment.ID,
		UserID:             punishment.UserID,
		StudentID:          punishment.StudentID,
		PunishmentTypeID:   punishment.PunishmentTypeID,
		TriggeringRuleID:   punishment.TriggeringRuleID,
		CreatedAt:          punishment.CreatedAt,
		DueAt:              punishment.DueAt,
		ResolvedAt:         punishment.ResolvedAt,
		StudentFirstName:   studentFirstName,
		StudentLastName:    studentLastName,
		PunishmentTypeName: r.punishmentTypeNameLocked(punishment.PunishmentTypeID),
		TriggeringRuleName: triggeringRuleNameAsText(triggeringRuleName),
	}
}

func (r *Repository) buildResolvePunishmentRowLocked(punishment repository.Punishment) repository.ResolvePunishmentRow {
	studentFirstName, studentLastName := r.studentNamesLocked(punishment.StudentID)
	triggeringRuleName := r.triggeringRuleNameForPunishmentLocked(punishment.TriggeringRuleID)

	return repository.ResolvePunishmentRow{
		ID:                 punishment.ID,
		UserID:             punishment.UserID,
		StudentID:          punishment.StudentID,
		PunishmentTypeID:   punishment.PunishmentTypeID,
		TriggeringRuleID:   punishment.TriggeringRuleID,
		CreatedAt:          punishment.CreatedAt,
		DueAt:              punishment.DueAt,
		ResolvedAt:         punishment.ResolvedAt,
		StudentFirstName:   studentFirstName,
		StudentLastName:    studentLastName,
		PunishmentTypeName: r.punishmentTypeNameLocked(punishment.PunishmentTypeID),
		TriggeringRuleName: triggeringRuleNameAsText(triggeringRuleName),
	}
}

func (r *Repository) SeedPunishment(punishment repository.Punishment) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if punishment.ID == uuid.Nil {
		punishment.ID = uuid.New()
	}
	if punishment.CreatedAt.IsZero() {
		punishment.CreatedAt = now
	}
	if punishment.DueAt.IsZero() {
		punishment.DueAt = now.Add(24 * time.Hour)
	}

	r.punishments[punishment.ID] = punishment
}

func (r *Repository) CreatePunishment(_ context.Context, arg repository.CreatePunishmentParams) (repository.CreatePunishmentRow, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpCreatePunishment); err != nil {
		return repository.CreatePunishmentRow{}, err
	}

	punishment := repository.Punishment{
		ID:               uuid.New(),
		UserID:           arg.UserID,
		StudentID:        arg.StudentID,
		PunishmentTypeID: arg.PunishmentTypeID,
		CreatedAt:        time.Now(),
		DueAt:            arg.DueAt,
	}
	r.punishments[punishment.ID] = punishment

	return r.buildCreatePunishmentRowLocked(punishment), nil
}

func (r *Repository) CreatePunishmentFromRule(_ context.Context, arg repository.CreatePunishmentFromRuleParams) (repository.CreatePunishmentFromRuleRow, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpCreatePunishmentFromRule); err != nil {
		return repository.CreatePunishmentFromRuleRow{}, err
	}

	punishment := repository.Punishment{
		ID:               uuid.New(),
		UserID:           arg.UserID,
		StudentID:        arg.StudentID,
		PunishmentTypeID: arg.PunishmentTypeID,
		TriggeringRuleID: arg.TriggeringRuleID,
		CreatedAt:        time.Now(),
		DueAt:            arg.DueAt,
	}
	r.punishments[punishment.ID] = punishment

	return r.buildCreatePunishmentFromRuleRowLocked(punishment), nil
}

func (r *Repository) GetPunishmentByUser(_ context.Context, arg repository.GetPunishmentByUserParams) (repository.GetPunishmentByUserRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpGetPunishmentByUser); err != nil {
		return repository.GetPunishmentByUserRow{}, err
	}

	punishment, ok := r.punishments[arg.ID]
	if !ok || punishment.UserID != arg.UserID {
		return repository.GetPunishmentByUserRow{}, pgx.ErrNoRows
	}

	return r.buildGetPunishmentRowLocked(punishment), nil
}

func (r *Repository) CountPunishmentsByUser(_ context.Context, arg repository.CountPunishmentsByUserParams) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpCountPunishmentsByUser); err != nil {
		return 0, err
	}

	var count int64
	for _, punishment := range r.punishments {
		if punishment.UserID != arg.UserID {
			continue
		}

		isResolved := punishment.ResolvedAt.Valid
		if matchesOptionalBool(arg.Resolved, isResolved) {
			count++
		}
	}

	return count, nil
}

func (r *Repository) ListPunishmentsByUser(_ context.Context, arg repository.ListPunishmentsByUserParams) ([]repository.ListPunishmentsByUserRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListPunishmentsByUser); err != nil {
		return nil, err
	}

	items := make([]repository.Punishment, 0)
	for _, punishment := range r.punishments {
		if punishment.UserID != arg.UserID {
			continue
		}

		isResolved := punishment.ResolvedAt.Valid
		if matchesOptionalBool(arg.Resolved, isResolved) {
			items = append(items, punishment)
		}
	}

	sortPunishmentsByCreatedAtDesc(items)
	paginated := paginate(items, arg.QueryOffset, arg.QueryLimit)

	rows := make([]repository.ListPunishmentsByUserRow, 0, len(paginated))
	for _, punishment := range paginated {
		rows = append(rows, r.buildListPunishmentByUserRowLocked(punishment))
	}

	return rows, nil
}

func (r *Repository) CountPunishmentsByStudent(_ context.Context, arg repository.CountPunishmentsByStudentParams) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpCountPunishmentsByStudent); err != nil {
		return 0, err
	}

	var count int64
	for _, punishment := range r.punishments {
		if punishment.StudentID != arg.StudentID || punishment.UserID != arg.UserID {
			continue
		}

		isResolved := punishment.ResolvedAt.Valid
		if matchesOptionalBool(arg.Resolved, isResolved) {
			count++
		}
	}

	return count, nil
}

func (r *Repository) ListPunishmentsByStudent(_ context.Context, arg repository.ListPunishmentsByStudentParams) ([]repository.ListPunishmentsByStudentRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListPunishmentsByStudent); err != nil {
		return nil, err
	}

	items := make([]repository.Punishment, 0)
	for _, punishment := range r.punishments {
		if punishment.StudentID != arg.StudentID || punishment.UserID != arg.UserID {
			continue
		}

		isResolved := punishment.ResolvedAt.Valid
		if matchesOptionalBool(arg.Resolved, isResolved) {
			items = append(items, punishment)
		}
	}

	sortPunishmentsByCreatedAtDesc(items)
	paginated := paginate(items, arg.QueryOffset, arg.QueryLimit)

	rows := make([]repository.ListPunishmentsByStudentRow, 0, len(paginated))
	for _, punishment := range paginated {
		rows = append(rows, r.buildListPunishmentByStudentRowLocked(punishment))
	}

	return rows, nil
}

func (r *Repository) ResolvePunishment(_ context.Context, arg repository.ResolvePunishmentParams) (repository.ResolvePunishmentRow, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpResolvePunishment); err != nil {
		return repository.ResolvePunishmentRow{}, err
	}

	punishment, ok := r.punishments[arg.ID]
	if !ok || punishment.UserID != arg.UserID || punishment.ResolvedAt.Valid {
		return repository.ResolvePunishmentRow{}, pgx.ErrNoRows
	}

	punishment.ResolvedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	r.punishments[arg.ID] = punishment

	return r.buildResolvePunishmentRowLocked(punishment), nil
}

func (r *Repository) DeletePunishmentByUser(_ context.Context, arg repository.DeletePunishmentByUserParams) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpDeletePunishmentByUser); err != nil {
		return 0, err
	}

	punishment, ok := r.punishments[arg.ID]
	if !ok || punishment.UserID != arg.UserID {
		return 0, nil
	}

	delete(r.punishments, arg.ID)
	return 1, nil
}
