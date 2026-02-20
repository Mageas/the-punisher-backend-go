package inmemory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

const (
	OpCreatePenaltyType       = "CreatePenaltyType"
	OpCountPenaltyTypesByUser = "CountPenaltyTypesByUser"
	OpListPenaltyTypesByUser  = "ListPenaltyTypesByUser"
	OpGetPenaltyTypeByUser    = "GetPenaltyTypeByUser"
	OpUpdatePenaltyTypeByUser = "UpdatePenaltyTypeByUser"
	OpDeletePenaltyTypeByUser = "DeletePenaltyTypeByUser"
)

func (r *Repository) SeedPenaltyType(pt repository.PenaltyType) {
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

	r.penaltyTypes[pt.ID] = pt
}

func (r *Repository) CreatePenaltyType(_ context.Context, arg repository.CreatePenaltyTypeParams) (repository.PenaltyType, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpCreatePenaltyType); err != nil {
		return repository.PenaltyType{}, err
	}

	now := time.Now()
	pt := repository.PenaltyType{
		ID:        uuid.New(),
		UserID:    arg.UserID,
		Name:      arg.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.penaltyTypes[pt.ID] = pt

	return pt, nil
}

func (r *Repository) CountPenaltyTypesByUser(_ context.Context, userID uuid.UUID) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpCountPenaltyTypesByUser); err != nil {
		return 0, err
	}

	var count int64
	for _, pt := range r.penaltyTypes {
		if pt.UserID == userID {
			count++
		}
	}

	return count, nil
}

func (r *Repository) ListPenaltyTypesByUser(_ context.Context, arg repository.ListPenaltyTypesByUserParams) ([]repository.PenaltyType, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListPenaltyTypesByUser); err != nil {
		return nil, err
	}

	items := make([]repository.PenaltyType, 0)
	for _, pt := range r.penaltyTypes {
		if pt.UserID == arg.UserID {
			items = append(items, pt)
		}
	}

	sortPenaltyTypesByCreatedAtDesc(items)
	return paginate(items, arg.QueryOffset, arg.QueryLimit), nil
}

func (r *Repository) GetPenaltyTypeByUser(_ context.Context, arg repository.GetPenaltyTypeByUserParams) (repository.PenaltyType, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpGetPenaltyTypeByUser); err != nil {
		return repository.PenaltyType{}, err
	}

	pt, ok := r.penaltyTypes[arg.ID]
	if !ok || pt.UserID != arg.UserID {
		return repository.PenaltyType{}, pgx.ErrNoRows
	}

	return pt, nil
}

func (r *Repository) UpdatePenaltyTypeByUser(_ context.Context, arg repository.UpdatePenaltyTypeByUserParams) (repository.PenaltyType, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpUpdatePenaltyTypeByUser); err != nil {
		return repository.PenaltyType{}, err
	}

	pt, ok := r.penaltyTypes[arg.ID]
	if !ok || pt.UserID != arg.UserID {
		return repository.PenaltyType{}, pgx.ErrNoRows
	}

	if arg.Name != nil {
		pt.Name = *arg.Name
	}
	pt.UpdatedAt = time.Now()
	r.penaltyTypes[arg.ID] = pt

	return pt, nil
}

func (r *Repository) DeletePenaltyTypeByUser(_ context.Context, arg repository.DeletePenaltyTypeByUserParams) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpDeletePenaltyTypeByUser); err != nil {
		return 0, err
	}

	pt, ok := r.penaltyTypes[arg.ID]
	if !ok || pt.UserID != arg.UserID {
		return 0, nil
	}

	delete(r.penaltyTypes, arg.ID)
	return 1, nil
}
