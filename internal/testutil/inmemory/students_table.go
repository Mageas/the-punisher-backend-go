package inmemory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

const (
	OpCreateStudent                 = "CreateStudent"
	OpGetStudentByUser              = "GetStudentByUser"
	OpCountStudentsByUser           = "CountStudentsByUser"
	OpListStudentsByUser            = "ListStudentsByUser"
	OpUpdateStudentByUser           = "UpdateStudentByUser"
	OpDeleteStudentByUser           = "DeleteStudentByUser"
	OpListClassroomRefsByStudentIDs = "ListClassroomRefsByStudentIDs"
)

func (r *Repository) studentAggregateFieldsLocked(student repository.Student) (float64, int64) {
	var availableBonusPoints float64
	for _, bonus := range r.bonuses {
		if bonus.StudentID == student.ID && bonus.UserID == student.UserID && !hasTime(bonus.UsedAt) {
			availableBonusPoints += bonus.Points
		}
	}

	var penaltyCount int64
	for _, penalty := range r.penalties {
		if penalty.StudentID == student.ID && penalty.UserID == student.UserID {
			penaltyCount++
		}
	}

	return availableBonusPoints, penaltyCount
}

func (r *Repository) buildCreateStudentRowLocked(student repository.Student) repository.CreateStudentRow {
	availableBonusPoints, penaltyCount := r.studentAggregateFieldsLocked(student)

	return repository.CreateStudentRow{
		ID:                   student.ID,
		UserID:               student.UserID,
		FirstName:            student.FirstName,
		LastName:             student.LastName,
		CreatedAt:            student.CreatedAt,
		UpdatedAt:            student.UpdatedAt,
		AvailableBonusPoints: availableBonusPoints,
		PenaltyCount:         penaltyCount,
	}
}

func (r *Repository) buildGetStudentRowLocked(student repository.Student) repository.GetStudentByUserRow {
	availableBonusPoints, penaltyCount := r.studentAggregateFieldsLocked(student)

	return repository.GetStudentByUserRow{
		ID:                   student.ID,
		UserID:               student.UserID,
		FirstName:            student.FirstName,
		LastName:             student.LastName,
		CreatedAt:            student.CreatedAt,
		UpdatedAt:            student.UpdatedAt,
		AvailableBonusPoints: availableBonusPoints,
		PenaltyCount:         penaltyCount,
	}
}

func (r *Repository) buildListStudentByUserRowLocked(student repository.Student) repository.ListStudentsByUserRow {
	availableBonusPoints, penaltyCount := r.studentAggregateFieldsLocked(student)

	return repository.ListStudentsByUserRow{
		ID:                   student.ID,
		UserID:               student.UserID,
		FirstName:            student.FirstName,
		LastName:             student.LastName,
		CreatedAt:            student.CreatedAt,
		UpdatedAt:            student.UpdatedAt,
		AvailableBonusPoints: availableBonusPoints,
		PenaltyCount:         penaltyCount,
	}
}

func (r *Repository) buildListStudentByClassroomRowLocked(student repository.Student) repository.ListStudentsByClassroomRow {
	availableBonusPoints, penaltyCount := r.studentAggregateFieldsLocked(student)

	return repository.ListStudentsByClassroomRow{
		ID:                   student.ID,
		UserID:               student.UserID,
		FirstName:            student.FirstName,
		LastName:             student.LastName,
		CreatedAt:            student.CreatedAt,
		UpdatedAt:            student.UpdatedAt,
		AvailableBonusPoints: availableBonusPoints,
		PenaltyCount:         penaltyCount,
	}
}

func (r *Repository) buildUpdateStudentRowLocked(student repository.Student) repository.UpdateStudentByUserRow {
	availableBonusPoints, penaltyCount := r.studentAggregateFieldsLocked(student)

	return repository.UpdateStudentByUserRow{
		ID:                   student.ID,
		UserID:               student.UserID,
		FirstName:            student.FirstName,
		LastName:             student.LastName,
		CreatedAt:            student.CreatedAt,
		UpdatedAt:            student.UpdatedAt,
		AvailableBonusPoints: availableBonusPoints,
		PenaltyCount:         penaltyCount,
	}
}

func (r *Repository) SeedStudent(student repository.Student) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if student.ID == uuid.Nil {
		student.ID = uuid.New()
	}
	if student.CreatedAt.IsZero() {
		student.CreatedAt = now
	}
	if student.UpdatedAt.IsZero() {
		student.UpdatedAt = student.CreatedAt
	}

	r.students[student.ID] = student
}

func (r *Repository) CreateStudent(_ context.Context, arg repository.CreateStudentParams) (repository.CreateStudentRow, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpCreateStudent); err != nil {
		return repository.CreateStudentRow{}, err
	}

	now := time.Now()
	student := repository.Student{
		ID:        uuid.New(),
		UserID:    arg.UserID,
		FirstName: arg.FirstName,
		LastName:  arg.LastName,
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.students[student.ID] = student

	return r.buildCreateStudentRowLocked(student), nil
}

func (r *Repository) GetStudentByUser(_ context.Context, arg repository.GetStudentByUserParams) (repository.GetStudentByUserRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpGetStudentByUser); err != nil {
		return repository.GetStudentByUserRow{}, err
	}

	student, ok := r.students[arg.ID]
	if !ok || student.UserID != arg.UserID {
		return repository.GetStudentByUserRow{}, pgx.ErrNoRows
	}

	return r.buildGetStudentRowLocked(student), nil
}

func (r *Repository) CountStudentsByUser(_ context.Context, arg repository.CountStudentsByUserParams) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpCountStudentsByUser); err != nil {
		return 0, err
	}

	var count int64
	for _, student := range r.students {
		if student.UserID != arg.UserID {
			continue
		}

		if matchesOptionalStudentSearch(arg.Search, student.FirstName, student.LastName) {
			count++
		}
	}

	return count, nil
}

func (r *Repository) ListStudentsByUser(_ context.Context, arg repository.ListStudentsByUserParams) ([]repository.ListStudentsByUserRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListStudentsByUser); err != nil {
		return nil, err
	}

	items := make([]repository.Student, 0)
	for _, student := range r.students {
		if student.UserID != arg.UserID {
			continue
		}

		if matchesOptionalStudentSearch(arg.Search, student.FirstName, student.LastName) {
			items = append(items, student)
		}
	}

	sortStudentsByCreatedAtDesc(items)
	paginated := paginate(items, arg.QueryOffset, arg.QueryLimit)

	rows := make([]repository.ListStudentsByUserRow, 0, len(paginated))
	for _, student := range paginated {
		rows = append(rows, r.buildListStudentByUserRowLocked(student))
	}

	return rows, nil
}

func (r *Repository) UpdateStudentByUser(_ context.Context, arg repository.UpdateStudentByUserParams) (repository.UpdateStudentByUserRow, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpUpdateStudentByUser); err != nil {
		return repository.UpdateStudentByUserRow{}, err
	}

	student, ok := r.students[arg.ID]
	if !ok || student.UserID != arg.UserID {
		return repository.UpdateStudentByUserRow{}, pgx.ErrNoRows
	}

	if arg.FirstName != nil {
		student.FirstName = *arg.FirstName
	}
	if arg.LastName != nil {
		student.LastName = *arg.LastName
	}

	student.UpdatedAt = time.Now()
	r.students[arg.ID] = student

	return r.buildUpdateStudentRowLocked(student), nil
}

func (r *Repository) DeleteStudentByUser(_ context.Context, arg repository.DeleteStudentByUserParams) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpDeleteStudentByUser); err != nil {
		return 0, err
	}

	student, ok := r.students[arg.ID]
	if !ok || student.UserID != arg.UserID {
		return 0, nil
	}

	delete(r.students, arg.ID)

	// Simulate ON DELETE CASCADE for join table.
	for key, relation := range r.studentClassrooms {
		if relation.StudentID == arg.ID {
			delete(r.studentClassrooms, key)
		}
	}

	return 1, nil
}

func (r *Repository) ListClassroomRefsByStudentIDs(_ context.Context, arg repository.ListClassroomRefsByStudentIDsParams) ([]repository.ListClassroomRefsByStudentIDsRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListClassroomRefsByStudentIDs); err != nil {
		return nil, err
	}

	studentIDs := make(map[uuid.UUID]struct{}, len(arg.StudentIds))
	for _, studentID := range arg.StudentIds {
		studentIDs[studentID] = struct{}{}
	}

	items := make([]repository.ListClassroomRefsByStudentIDsRow, 0)
	for _, relation := range r.studentClassrooms {
		if _, ok := studentIDs[relation.StudentID]; !ok {
			continue
		}

		student, studentExists := r.students[relation.StudentID]
		if !studentExists || student.UserID != arg.UserID {
			continue
		}

		classroom, classroomExists := r.classrooms[relation.ClassroomID]
		if !classroomExists {
			continue
		}

		items = append(items, repository.ListClassroomRefsByStudentIDsRow{
			StudentID:     relation.StudentID,
			ClassroomID:   classroom.ID,
			ClassroomName: classroom.Name,
		})
	}

	return items, nil
}
