package inmemory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

const (
	OpCreatePunishmentType       = "CreatePunishmentType"
	OpCountPunishmentTypesByUser = "CountPunishmentTypesByUser"
	OpListPunishmentTypesByUser  = "ListPunishmentTypesByUser"
	OpGetPunishmentTypeByUser    = "GetPunishmentTypeByUser"
	OpUpdatePunishmentTypeByUser = "UpdatePunishmentTypeByUser"
	OpDeletePunishmentTypeByUser = "DeletePunishmentTypeByUser"
)

func (r *Repository) SeedPunishmentType(pt repository.PunishmentType) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if pt.ID == uuid.Nil {
		pt.ID = uuid.New()
	}
	if pt.CreatedAt.IsZero() {
		pt.CreatedAt = now
	}
	if pt.UpdatedAt.IsZero() {
		pt.UpdatedAt = pt.CreatedAt
	}

	r.punishmentTypes[pt.ID] = pt
}

func (r *Repository) CreatePunishmentType(_ context.Context, arg repository.CreatePunishmentTypeParams) (repository.PunishmentType, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpCreatePunishmentType); err != nil {
		return repository.PunishmentType{}, err
	}

	now := time.Now()
	pt := repository.PunishmentType{
		ID:        uuid.New(),
		UserID:    arg.UserID,
		Name:      arg.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.punishmentTypes[pt.ID] = pt

	return pt, nil
}

func (r *Repository) CountPunishmentTypesByUser(_ context.Context, userID uuid.UUID) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpCountPunishmentTypesByUser); err != nil {
		return 0, err
	}

	var count int64
	for _, pt := range r.punishmentTypes {
		if pt.UserID == userID {
			count++
		}
	}

	return count, nil
}

func (r *Repository) ListPunishmentTypesByUser(_ context.Context, arg repository.ListPunishmentTypesByUserParams) ([]repository.PunishmentType, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListPunishmentTypesByUser); err != nil {
		return nil, err
	}

	items := make([]repository.PunishmentType, 0)
	for _, pt := range r.punishmentTypes {
		if pt.UserID == arg.UserID {
			items = append(items, pt)
		}
	}

	sortPunishmentTypesByCreatedAtDesc(items)
	return paginate(items, arg.QueryOffset, arg.QueryLimit), nil
}

func (r *Repository) GetPunishmentTypeByUser(_ context.Context, arg repository.GetPunishmentTypeByUserParams) (repository.PunishmentType, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpGetPunishmentTypeByUser); err != nil {
		return repository.PunishmentType{}, err
	}

	pt, ok := r.punishmentTypes[arg.ID]
	if !ok || pt.UserID != arg.UserID {
		return repository.PunishmentType{}, pgx.ErrNoRows
	}

	return pt, nil
}

func (r *Repository) UpdatePunishmentTypeByUser(_ context.Context, arg repository.UpdatePunishmentTypeByUserParams) (repository.PunishmentType, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpUpdatePunishmentTypeByUser); err != nil {
		return repository.PunishmentType{}, err
	}

	pt, ok := r.punishmentTypes[arg.ID]
	if !ok || pt.UserID != arg.UserID {
		return repository.PunishmentType{}, pgx.ErrNoRows
	}

	if arg.Name != nil {
		pt.Name = *arg.Name
	}
	pt.UpdatedAt = time.Now()
	r.punishmentTypes[arg.ID] = pt

	return pt, nil
}

func (r *Repository) DeletePunishmentTypeByUser(_ context.Context, arg repository.DeletePunishmentTypeByUserParams) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpDeletePunishmentTypeByUser); err != nil {
		return 0, err
	}

	pt, ok := r.punishmentTypes[arg.ID]
	if !ok || pt.UserID != arg.UserID {
		return 0, nil
	}

	delete(r.punishmentTypes, arg.ID)
	return 1, nil
}
