package inmemory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

const (
	OpCreateStudent       = "CreateStudent"
	OpGetStudentByUser    = "GetStudentByUser"
	OpCountStudentsByUser = "CountStudentsByUser"
	OpListStudentsByUser  = "ListStudentsByUser"
	OpUpdateStudentByUser = "UpdateStudentByUser"
	OpDeleteStudentByUser = "DeleteStudentByUser"
)

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

func (r *Repository) CreateStudent(_ context.Context, arg repository.CreateStudentParams) (repository.Student, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpCreateStudent); err != nil {
		return repository.Student{}, err
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

	return student, nil
}

func (r *Repository) GetStudentByUser(_ context.Context, arg repository.GetStudentByUserParams) (repository.Student, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpGetStudentByUser); err != nil {
		return repository.Student{}, err
	}

	student, ok := r.students[arg.ID]
	if !ok || student.UserID != arg.UserID {
		return repository.Student{}, pgx.ErrNoRows
	}

	return student, nil
}

func (r *Repository) CountStudentsByUser(_ context.Context, userID uuid.UUID) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpCountStudentsByUser); err != nil {
		return 0, err
	}

	var count int64
	for _, student := range r.students {
		if student.UserID == userID {
			count++
		}
	}

	return count, nil
}

func (r *Repository) ListStudentsByUser(_ context.Context, arg repository.ListStudentsByUserParams) ([]repository.Student, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListStudentsByUser); err != nil {
		return nil, err
	}

	items := make([]repository.Student, 0)
	for _, student := range r.students {
		if student.UserID == arg.UserID {
			items = append(items, student)
		}
	}

	sortStudentsByCreatedAtDesc(items)
	return paginate(items, arg.QueryOffset, arg.QueryLimit), nil
}

func (r *Repository) UpdateStudentByUser(_ context.Context, arg repository.UpdateStudentByUserParams) (repository.Student, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpUpdateStudentByUser); err != nil {
		return repository.Student{}, err
	}

	student, ok := r.students[arg.ID]
	if !ok || student.UserID != arg.UserID {
		return repository.Student{}, pgx.ErrNoRows
	}

	if arg.FirstName.Valid {
		student.FirstName = arg.FirstName.String
	}
	if arg.LastName.Valid {
		student.LastName = arg.LastName.String
	}

	student.UpdatedAt = time.Now()
	r.students[arg.ID] = student

	return student, nil
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
