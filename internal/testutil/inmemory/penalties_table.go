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
	OpDeletePenaltiesByTypeByUser    = "DeletePenaltiesByTypeByUser"
)

func (r *Repository) studentNamesLocked(studentID uuid.UUID) (string, string) {
	if student, ok := r.students[studentID]; ok {
		return student.FirstName, student.LastName
	}

	return "", ""
}

func (r *Repository) penaltyTypeNameForPenaltyLocked(penaltyTypeID uuid.UUID) string {
	if penaltyType, ok := r.penaltyTypes[penaltyTypeID]; ok {
		return penaltyType.Name
	}

	return ""
}

func (r *Repository) buildCreatePenaltyRowLocked(penalty repository.Penalty) repository.CreatePenaltyRow {
	studentFirstName, studentLastName := r.studentNamesLocked(penalty.StudentID)

	return repository.CreatePenaltyRow{
		ID:               penalty.ID,
		UserID:           penalty.UserID,
		StudentID:        penalty.StudentID,
		PenaltyTypeID:    penalty.PenaltyTypeID,
		CreatedAt:        penalty.CreatedAt,
		StudentFirstName: studentFirstName,
		StudentLastName:  studentLastName,
		PenaltyTypeName:  r.penaltyTypeNameForPenaltyLocked(penalty.PenaltyTypeID),
	}
}

func (r *Repository) buildGetPenaltyRowLocked(penalty repository.Penalty) repository.GetPenaltyByUserRow {
	studentFirstName, studentLastName := r.studentNamesLocked(penalty.StudentID)

	return repository.GetPenaltyByUserRow{
		ID:               penalty.ID,
		UserID:           penalty.UserID,
		StudentID:        penalty.StudentID,
		PenaltyTypeID:    penalty.PenaltyTypeID,
		CreatedAt:        penalty.CreatedAt,
		StudentFirstName: studentFirstName,
		StudentLastName:  studentLastName,
		PenaltyTypeName:  r.penaltyTypeNameForPenaltyLocked(penalty.PenaltyTypeID),
	}
}

func (r *Repository) buildListPenaltyByUserRowLocked(penalty repository.Penalty) repository.ListPenaltiesByUserRow {
	studentFirstName, studentLastName := r.studentNamesLocked(penalty.StudentID)

	return repository.ListPenaltiesByUserRow{
		ID:               penalty.ID,
		UserID:           penalty.UserID,
		StudentID:        penalty.StudentID,
		PenaltyTypeID:    penalty.PenaltyTypeID,
		CreatedAt:        penalty.CreatedAt,
		StudentFirstName: studentFirstName,
		StudentLastName:  studentLastName,
		PenaltyTypeName:  r.penaltyTypeNameForPenaltyLocked(penalty.PenaltyTypeID),
	}
}

func (r *Repository) buildListPenaltyByStudentRowLocked(penalty repository.Penalty) repository.ListPenaltiesByStudentRow {
	studentFirstName, studentLastName := r.studentNamesLocked(penalty.StudentID)

	return repository.ListPenaltiesByStudentRow{
		ID:               penalty.ID,
		UserID:           penalty.UserID,
		StudentID:        penalty.StudentID,
		PenaltyTypeID:    penalty.PenaltyTypeID,
		CreatedAt:        penalty.CreatedAt,
		StudentFirstName: studentFirstName,
		StudentLastName:  studentLastName,
		PenaltyTypeName:  r.penaltyTypeNameForPenaltyLocked(penalty.PenaltyTypeID),
	}
}

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

func (r *Repository) CreatePenalty(_ context.Context, arg repository.CreatePenaltyParams) (repository.CreatePenaltyRow, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpCreatePenalty); err != nil {
		return repository.CreatePenaltyRow{}, err
	}

	penalty := repository.Penalty{
		ID:            uuid.New(),
		UserID:        arg.UserID,
		StudentID:     arg.StudentID,
		PenaltyTypeID: arg.PenaltyTypeID,
		CreatedAt:     time.Now(),
	}
	r.penalties[penalty.ID] = penalty

	return r.buildCreatePenaltyRowLocked(penalty), nil
}

func (r *Repository) GetPenaltyByUser(_ context.Context, arg repository.GetPenaltyByUserParams) (repository.GetPenaltyByUserRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpGetPenaltyByUser); err != nil {
		return repository.GetPenaltyByUserRow{}, err
	}

	penalty, ok := r.penalties[arg.ID]
	if !ok || penalty.UserID != arg.UserID {
		return repository.GetPenaltyByUserRow{}, pgx.ErrNoRows
	}

	return r.buildGetPenaltyRowLocked(penalty), nil
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

func (r *Repository) ListPenaltiesByUser(_ context.Context, arg repository.ListPenaltiesByUserParams) ([]repository.ListPenaltiesByUserRow, error) {
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
	paginated := paginate(items, arg.QueryOffset, arg.QueryLimit)

	rows := make([]repository.ListPenaltiesByUserRow, 0, len(paginated))
	for _, penalty := range paginated {
		rows = append(rows, r.buildListPenaltyByUserRowLocked(penalty))
	}

	return rows, nil
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

func (r *Repository) ListPenaltiesByStudent(_ context.Context, arg repository.ListPenaltiesByStudentParams) ([]repository.ListPenaltiesByStudentRow, error) {
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
	paginated := paginate(items, arg.QueryOffset, arg.QueryLimit)

	rows := make([]repository.ListPenaltiesByStudentRow, 0, len(paginated))
	for _, penalty := range paginated {
		rows = append(rows, r.buildListPenaltyByStudentRowLocked(penalty))
	}

	return rows, nil
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

func (r *Repository) DeletePenaltiesByTypeByUser(_ context.Context, arg repository.DeletePenaltiesByTypeByUserParams) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpDeletePenaltiesByTypeByUser); err != nil {
		return 0, err
	}

	var deleted int64
	for id, penalty := range r.penalties {
		if penalty.UserID != arg.UserID || penalty.PenaltyTypeID != arg.PenaltyTypeID {
			continue
		}
		delete(r.penalties, id)
		deleted++
	}

	return deleted, nil
}
