package inmemory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

const (
	OpCreateBonusType       = "CreateBonusType"
	OpCountBonusTypesByUser = "CountBonusTypesByUser"
	OpListBonusTypesByUser  = "ListBonusTypesByUser"
	OpGetBonusTypeByUser    = "GetBonusTypeByUser"
	OpUpdateBonusTypeByUser = "UpdateBonusTypeByUser"
	OpDeleteBonusTypeByUser = "DeleteBonusTypeByUser"
)

func (r *Repository) SeedBonusType(bt repository.BonusType) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if bt.ID == uuid.Nil {
		bt.ID = uuid.New()
	}
	if bt.CreatedAt.IsZero() {
		bt.CreatedAt = now
	}
	if bt.UpdatedAt.IsZero() {
		bt.UpdatedAt = bt.CreatedAt
	}

	r.bonusTypes[bt.ID] = bt
}

func (r *Repository) CreateBonusType(_ context.Context, arg repository.CreateBonusTypeParams) (repository.BonusType, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpCreateBonusType); err != nil {
		return repository.BonusType{}, err
	}

	now := time.Now()
	bt := repository.BonusType{
		ID:        uuid.New(),
		UserID:    arg.UserID,
		Name:      arg.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.bonusTypes[bt.ID] = bt

	return bt, nil
}

func (r *Repository) CountBonusTypesByUser(_ context.Context, userID uuid.UUID) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpCountBonusTypesByUser); err != nil {
		return 0, err
	}

	var count int64
	for _, bt := range r.bonusTypes {
		if bt.UserID == userID {
			count++
		}
	}

	return count, nil
}

func (r *Repository) ListBonusTypesByUser(_ context.Context, arg repository.ListBonusTypesByUserParams) ([]repository.BonusType, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListBonusTypesByUser); err != nil {
		return nil, err
	}

	items := make([]repository.BonusType, 0)
	for _, bt := range r.bonusTypes {
		if bt.UserID == arg.UserID {
			items = append(items, bt)
		}
	}

	sortBonusTypesByCreatedAtDesc(items)
	return paginate(items, arg.QueryOffset, arg.QueryLimit), nil
}

func (r *Repository) GetBonusTypeByUser(_ context.Context, arg repository.GetBonusTypeByUserParams) (repository.BonusType, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpGetBonusTypeByUser); err != nil {
		return repository.BonusType{}, err
	}

	bt, ok := r.bonusTypes[arg.ID]
	if !ok || bt.UserID != arg.UserID {
		return repository.BonusType{}, pgx.ErrNoRows
	}

	return bt, nil
}

func (r *Repository) UpdateBonusTypeByUser(_ context.Context, arg repository.UpdateBonusTypeByUserParams) (repository.BonusType, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpUpdateBonusTypeByUser); err != nil {
		return repository.BonusType{}, err
	}

	bt, ok := r.bonusTypes[arg.ID]
	if !ok || bt.UserID != arg.UserID {
		return repository.BonusType{}, pgx.ErrNoRows
	}

	if arg.Name != nil {
		bt.Name = *arg.Name
	}
	bt.UpdatedAt = time.Now()
	r.bonusTypes[arg.ID] = bt

	return bt, nil
}

func (r *Repository) DeleteBonusTypeByUser(_ context.Context, arg repository.DeleteBonusTypeByUserParams) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpDeleteBonusTypeByUser); err != nil {
		return 0, err
	}

	bt, ok := r.bonusTypes[arg.ID]
	if !ok || bt.UserID != arg.UserID {
		return 0, nil
	}

	delete(r.bonusTypes, arg.ID)
	return 1, nil
}
