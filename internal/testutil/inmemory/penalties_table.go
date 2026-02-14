package inmemory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

const (
	OpCreatePenalty                  = "CreatePenalty"
	OpGetPenaltyByUser               = "GetPenaltyByUser"
	OpCountPenaltiesByUser           = "CountPenaltiesByUser"
	OpListPenaltiesByUser            = "ListPenaltiesByUser"
	OpCountPenaltiesByStudent        = "CountPenaltiesByStudent"
	OpListPenaltiesByStudent         = "ListPenaltiesByStudent"
	OpCountPenaltiesByStudentAndType = "CountPenaltiesByStudentAndType"
	OpDeletePenaltyByUser            = "DeletePenaltyByUser"
)

func (r *Repository) SeedPenalty(penalty repository.Penalty) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if penalty.ID == uuid.Nil {
		penalty.ID = uuid.New()
	}
	if penalty.CreatedAt.IsZero() {
		penalty.CreatedAt = time.Now()
	}

	r.penalties[penalty.ID] = penalty
}

func (r *Repository) CreatePenalty(_ context.Context, arg repository.CreatePenaltyParams) (repository.Penalty, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpCreatePenalty); err != nil {
		return repository.Penalty{}, err
	}

	penalty := repository.Penalty{
		ID:            uuid.New(),
		UserID:        arg.UserID,
		StudentID:     arg.StudentID,
		PenaltyTypeID: arg.PenaltyTypeID,
		CreatedAt:     time.Now(),
	}
	r.penalties[penalty.ID] = penalty

	return penalty, nil
}

func (r *Repository) GetPenaltyByUser(_ context.Context, arg repository.GetPenaltyByUserParams) (repository.Penalty, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpGetPenaltyByUser); err != nil {
		return repository.Penalty{}, err
	}

	penalty, ok := r.penalties[arg.ID]
	if !ok || penalty.UserID != arg.UserID {
		return repository.Penalty{}, pgx.ErrNoRows
	}

	return penalty, nil
}

func (r *Repository) CountPenaltiesByUser(_ context.Context, userID uuid.UUID) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpCountPenaltiesByUser); err != nil {
		return 0, err
	}

	var count int64
	for _, penalty := range r.penalties {
		if penalty.UserID == userID {
			count++
		}
	}

	return count, nil
}

func (r *Repository) ListPenaltiesByUser(_ context.Context, arg repository.ListPenaltiesByUserParams) ([]repository.Penalty, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListPenaltiesByUser); err != nil {
		return nil, err
	}

	items := make([]repository.Penalty, 0)
	for _, penalty := range r.penalties {
		if penalty.UserID == arg.UserID {
			items = append(items, penalty)
		}
	}

	sortPenaltiesByCreatedAtDesc(items)
	return paginate(items, arg.QueryOffset, arg.QueryLimit), nil
}

func (r *Repository) CountPenaltiesByStudent(_ context.Context, arg repository.CountPenaltiesByStudentParams) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpCountPenaltiesByStudent); err != nil {
		return 0, err
	}

	var count int64
	for _, penalty := range r.penalties {
		if penalty.StudentID == arg.StudentID && penalty.UserID == arg.UserID {
			count++
		}
	}

	return count, nil
}

func (r *Repository) ListPenaltiesByStudent(_ context.Context, arg repository.ListPenaltiesByStudentParams) ([]repository.Penalty, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListPenaltiesByStudent); err != nil {
		return nil, err
	}

	items := make([]repository.Penalty, 0)
	for _, penalty := range r.penalties {
		if penalty.StudentID == arg.StudentID && penalty.UserID == arg.UserID {
			items = append(items, penalty)
		}
	}

	sortPenaltiesByCreatedAtDesc(items)
	return paginate(items, arg.QueryOffset, arg.QueryLimit), nil
}

func (r *Repository) CountPenaltiesByStudentAndType(_ context.Context, arg repository.CountPenaltiesByStudentAndTypeParams) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpCountPenaltiesByStudentAndType); err != nil {
		return 0, err
	}

	var count int64
	for _, penalty := range r.penalties {
		if penalty.StudentID == arg.StudentID && penalty.UserID == arg.UserID && penalty.PenaltyTypeID == arg.PenaltyTypeID {
			count++
		}
	}

	return count, nil
}

func (r *Repository) DeletePenaltyByUser(_ context.Context, arg repository.DeletePenaltyByUserParams) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpDeletePenaltyByUser); err != nil {
		return 0, err
	}

	penalty, ok := r.penalties[arg.ID]
	if !ok || penalty.UserID != arg.UserID {
		return 0, nil
	}

	delete(r.penalties, arg.ID)
	return 1, nil
}
