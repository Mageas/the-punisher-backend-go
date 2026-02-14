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

func (r *Repository) CreatePunishment(_ context.Context, arg repository.CreatePunishmentParams) (repository.Punishment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpCreatePunishment); err != nil {
		return repository.Punishment{}, err
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

	return punishment, nil
}

func (r *Repository) CreatePunishmentFromRule(_ context.Context, arg repository.CreatePunishmentFromRuleParams) (repository.Punishment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpCreatePunishmentFromRule); err != nil {
		return repository.Punishment{}, err
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

	return punishment, nil
}

func (r *Repository) GetPunishmentByUser(_ context.Context, arg repository.GetPunishmentByUserParams) (repository.Punishment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpGetPunishmentByUser); err != nil {
		return repository.Punishment{}, err
	}

	punishment, ok := r.punishments[arg.ID]
	if !ok || punishment.UserID != arg.UserID {
		return repository.Punishment{}, pgx.ErrNoRows
	}

	return punishment, nil
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

func (r *Repository) ListPunishmentsByUser(_ context.Context, arg repository.ListPunishmentsByUserParams) ([]repository.Punishment, error) {
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
	return paginate(items, arg.QueryOffset, arg.QueryLimit), nil
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

func (r *Repository) ListPunishmentsByStudent(_ context.Context, arg repository.ListPunishmentsByStudentParams) ([]repository.Punishment, error) {
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
	return paginate(items, arg.QueryOffset, arg.QueryLimit), nil
}

func (r *Repository) ResolvePunishment(_ context.Context, arg repository.ResolvePunishmentParams) (repository.Punishment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpResolvePunishment); err != nil {
		return repository.Punishment{}, err
	}

	punishment, ok := r.punishments[arg.ID]
	if !ok || punishment.UserID != arg.UserID || punishment.ResolvedAt.Valid {
		return repository.Punishment{}, pgx.ErrNoRows
	}

	punishment.ResolvedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	r.punishments[arg.ID] = punishment

	return punishment, nil
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
