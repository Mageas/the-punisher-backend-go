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
	OpCreateBonus           = "CreateBonus"
	OpGetBonusByUser        = "GetBonusByUser"
	OpCountBonusesByUser    = "CountBonusesByUser"
	OpListBonusesByUser     = "ListBonusesByUser"
	OpCountBonusesByStudent = "CountBonusesByStudent"
	OpListBonusesByStudent  = "ListBonusesByStudent"
	OpUseBonus              = "UseBonus"
	OpDeleteBonusByUser     = "DeleteBonusByUser"
)

func (r *Repository) SeedBonus(bonus repository.Bonus) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if bonus.ID == uuid.Nil {
		bonus.ID = uuid.New()
	}
	if bonus.CreatedAt.IsZero() {
		bonus.CreatedAt = now
	}

	r.bonuses[bonus.ID] = bonus
}

func (r *Repository) CreateBonus(_ context.Context, arg repository.CreateBonusParams) (repository.Bonus, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpCreateBonus); err != nil {
		return repository.Bonus{}, err
	}

	bonus := repository.Bonus{
		ID:          uuid.New(),
		UserID:      arg.UserID,
		StudentID:   arg.StudentID,
		BonusTypeID: arg.BonusTypeID,
		Points:      arg.Points,
		CreatedAt:   time.Now(),
	}
	r.bonuses[bonus.ID] = bonus

	return bonus, nil
}

func (r *Repository) GetBonusByUser(_ context.Context, arg repository.GetBonusByUserParams) (repository.Bonus, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpGetBonusByUser); err != nil {
		return repository.Bonus{}, err
	}

	bonus, ok := r.bonuses[arg.ID]
	if !ok || bonus.UserID != arg.UserID {
		return repository.Bonus{}, pgx.ErrNoRows
	}

	return bonus, nil
}

func (r *Repository) CountBonusesByUser(_ context.Context, arg repository.CountBonusesByUserParams) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpCountBonusesByUser); err != nil {
		return 0, err
	}

	var count int64
	for _, bonus := range r.bonuses {
		if bonus.UserID != arg.UserID {
			continue
		}

		isUsed := bonus.UsedAt.Valid
		if matchesOptionalBool(arg.Used, isUsed) {
			count++
		}
	}

	return count, nil
}

func (r *Repository) ListBonusesByUser(_ context.Context, arg repository.ListBonusesByUserParams) ([]repository.Bonus, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListBonusesByUser); err != nil {
		return nil, err
	}

	items := make([]repository.Bonus, 0)
	for _, bonus := range r.bonuses {
		if bonus.UserID != arg.UserID {
			continue
		}

		isUsed := bonus.UsedAt.Valid
		if matchesOptionalBool(arg.Used, isUsed) {
			items = append(items, bonus)
		}
	}

	sortBonusesByCreatedAtDesc(items)
	return paginate(items, arg.QueryOffset, arg.QueryLimit), nil
}

func (r *Repository) CountBonusesByStudent(_ context.Context, arg repository.CountBonusesByStudentParams) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpCountBonusesByStudent); err != nil {
		return 0, err
	}

	var count int64
	for _, bonus := range r.bonuses {
		if bonus.StudentID != arg.StudentID || bonus.UserID != arg.UserID {
			continue
		}

		isUsed := bonus.UsedAt.Valid
		if matchesOptionalBool(arg.Used, isUsed) {
			count++
		}
	}

	return count, nil
}

func (r *Repository) ListBonusesByStudent(_ context.Context, arg repository.ListBonusesByStudentParams) ([]repository.Bonus, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListBonusesByStudent); err != nil {
		return nil, err
	}

	items := make([]repository.Bonus, 0)
	for _, bonus := range r.bonuses {
		if bonus.StudentID != arg.StudentID || bonus.UserID != arg.UserID {
			continue
		}

		isUsed := bonus.UsedAt.Valid
		if matchesOptionalBool(arg.Used, isUsed) {
			items = append(items, bonus)
		}
	}

	sortBonusesByCreatedAtDesc(items)
	return paginate(items, arg.QueryOffset, arg.QueryLimit), nil
}

func (r *Repository) UseBonus(_ context.Context, arg repository.UseBonusParams) (repository.Bonus, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpUseBonus); err != nil {
		return repository.Bonus{}, err
	}

	bonus, ok := r.bonuses[arg.ID]
	if !ok || bonus.UserID != arg.UserID || bonus.UsedAt.Valid {
		return repository.Bonus{}, pgx.ErrNoRows
	}

	bonus.UsedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	r.bonuses[arg.ID] = bonus

	return bonus, nil
}

func (r *Repository) DeleteBonusByUser(_ context.Context, arg repository.DeleteBonusByUserParams) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpDeleteBonusByUser); err != nil {
		return 0, err
	}

	bonus, ok := r.bonuses[arg.ID]
	if !ok || bonus.UserID != arg.UserID {
		return 0, nil
	}

	delete(r.bonuses, arg.ID)
	return 1, nil
}
