package inmemory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

const (
	OpCreateClassroom            = "CreateClassroom"
	OpGetClassroomByUser         = "GetClassroomByUser"
	OpCountClassroomsByUser      = "CountClassroomsByUser"
	OpListClassroomsByUser       = "ListClassroomsByUser"
	OpUpdateClassroomByUser      = "UpdateClassroomByUser"
	OpDeleteClassroomByUser      = "DeleteClassroomByUser"
	OpAddStudentToClassroom      = "AddStudentToClassroom"
	OpRemoveStudentFromClassroom = "RemoveStudentFromClassroom"
	OpCountStudentsByClassroom   = "CountStudentsByClassroom"
	OpListStudentsByClassroom    = "ListStudentsByClassroom"
	OpCountClassroomsByStudent   = "CountClassroomsByStudent"
	OpListClassroomsByStudent    = "ListClassroomsByStudent"
)

func (r *Repository) SeedClassroom(classroom repository.Classroom) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if classroom.ID == uuid.Nil {
		classroom.ID = uuid.New()
	}
	if classroom.CreatedAt.IsZero() {
		classroom.CreatedAt = now
	}
	if classroom.UpdatedAt.IsZero() {
		classroom.UpdatedAt = classroom.CreatedAt
	}

	r.classrooms[classroom.ID] = classroom
}

func (r *Repository) CreateClassroom(_ context.Context, arg repository.CreateClassroomParams) (repository.Classroom, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpCreateClassroom); err != nil {
		return repository.Classroom{}, err
	}

	now := time.Now()
	classroom := repository.Classroom{
		ID:          uuid.New(),
		UserID:      arg.UserID,
		Name:        arg.Name,
		Year:        arg.Year,
		MainTeacher: arg.MainTeacher,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	r.classrooms[classroom.ID] = classroom

	return classroom, nil
}

func (r *Repository) GetClassroomByUser(_ context.Context, arg repository.GetClassroomByUserParams) (repository.Classroom, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpGetClassroomByUser); err != nil {
		return repository.Classroom{}, err
	}

	classroom, ok := r.classrooms[arg.ID]
	if !ok || classroom.UserID != arg.UserID {
		return repository.Classroom{}, pgx.ErrNoRows
	}

	return classroom, nil
}

func (r *Repository) CountClassroomsByUser(_ context.Context, userID uuid.UUID) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpCountClassroomsByUser); err != nil {
		return 0, err
	}

	var count int64
	for _, classroom := range r.classrooms {
		if classroom.UserID == userID {
			count++
		}
	}

	return count, nil
}

func (r *Repository) ListClassroomsByUser(_ context.Context, arg repository.ListClassroomsByUserParams) ([]repository.Classroom, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListClassroomsByUser); err != nil {
		return nil, err
	}

	items := make([]repository.Classroom, 0)
	for _, classroom := range r.classrooms {
		if classroom.UserID == arg.UserID {
			items = append(items, classroom)
		}
	}

	sortClassroomsByCreatedAtDesc(items)
	return paginate(items, arg.QueryOffset, arg.QueryLimit), nil
}

func (r *Repository) UpdateClassroomByUser(_ context.Context, arg repository.UpdateClassroomByUserParams) (repository.Classroom, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpUpdateClassroomByUser); err != nil {
		return repository.Classroom{}, err
	}

	classroom, ok := r.classrooms[arg.ID]
	if !ok || classroom.UserID != arg.UserID {
		return repository.Classroom{}, pgx.ErrNoRows
	}

	if arg.Name.Valid {
		classroom.Name = arg.Name.String
	}
	if arg.Year.Valid {
		classroom.Year = arg.Year
	}
	if arg.MainTeacher.Valid {
		classroom.MainTeacher = arg.MainTeacher
	}

	classroom.UpdatedAt = time.Now()
	r.classrooms[arg.ID] = classroom

	return classroom, nil
}

func (r *Repository) DeleteClassroomByUser(_ context.Context, arg repository.DeleteClassroomByUserParams) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpDeleteClassroomByUser); err != nil {
		return 0, err
	}

	classroom, ok := r.classrooms[arg.ID]
	if !ok || classroom.UserID != arg.UserID {
		return 0, nil
	}

	delete(r.classrooms, arg.ID)

	// Simulate ON DELETE CASCADE for join table.
	for key, relation := range r.studentClassrooms {
		if relation.ClassroomID == arg.ID {
			delete(r.studentClassrooms, key)
		}
	}

	return 1, nil
}

func (r *Repository) AddStudentToClassroom(_ context.Context, arg repository.AddStudentToClassroomParams) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpAddStudentToClassroom); err != nil {
		return 0, err
	}

	student, studentExists := r.students[arg.StudentID]
	classroom, classroomExists := r.classrooms[arg.ClassroomID]
	if !studentExists || !classroomExists || student.UserID != arg.UserID || classroom.UserID != arg.UserID {
		return 0, nil
	}

	key := studentClassroomKey(arg.StudentID, arg.ClassroomID)
	if _, exists := r.studentClassrooms[key]; exists {
		return 0, &pgconn.PgError{Code: "23505"}
	}

	r.studentClassrooms[key] = repository.StudentClassroom{
		StudentID:   arg.StudentID,
		ClassroomID: arg.ClassroomID,
		CreatedAt:   time.Now(),
	}

	return 1, nil
}

func (r *Repository) RemoveStudentFromClassroom(_ context.Context, arg repository.RemoveStudentFromClassroomParams) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpRemoveStudentFromClassroom); err != nil {
		return 0, err
	}

	classroom, classroomExists := r.classrooms[arg.ClassroomID]
	if !classroomExists || classroom.UserID != arg.UserID {
		return 0, nil
	}

	key := studentClassroomKey(arg.StudentID, arg.ClassroomID)
	if _, exists := r.studentClassrooms[key]; !exists {
		return 0, nil
	}

	delete(r.studentClassrooms, key)
	return 1, nil
}

func (r *Repository) CountStudentsByClassroom(_ context.Context, arg repository.CountStudentsByClassroomParams) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpCountStudentsByClassroom); err != nil {
		return 0, err
	}

	classroom, classroomExists := r.classrooms[arg.ClassroomID]
	if !classroomExists || classroom.UserID != arg.UserID {
		return 0, nil
	}

	var count int64
	for _, relation := range r.studentClassrooms {
		if relation.ClassroomID == arg.ClassroomID {
			count++
		}
	}

	return count, nil
}

func (r *Repository) ListStudentsByClassroom(_ context.Context, arg repository.ListStudentsByClassroomParams) ([]repository.Student, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListStudentsByClassroom); err != nil {
		return nil, err
	}

	classroom, classroomExists := r.classrooms[arg.ClassroomID]
	if !classroomExists || classroom.UserID != arg.UserID {
		return []repository.Student{}, nil
	}

	items := make([]repository.Student, 0)
	for _, relation := range r.studentClassrooms {
		if relation.ClassroomID != arg.ClassroomID {
			continue
		}

		student, studentExists := r.students[relation.StudentID]
		if !studentExists {
			continue
		}

		items = append(items, student)
	}

	sortStudentsByCreatedAtDesc(items)
	return paginate(items, arg.QueryOffset, arg.QueryLimit), nil
}

func (r *Repository) CountClassroomsByStudent(_ context.Context, arg repository.CountClassroomsByStudentParams) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpCountClassroomsByStudent); err != nil {
		return 0, err
	}

	student, studentExists := r.students[arg.StudentID]
	if !studentExists || student.UserID != arg.UserID {
		return 0, nil
	}

	var count int64
	for _, relation := range r.studentClassrooms {
		if relation.StudentID != arg.StudentID {
			continue
		}

		classroom, classroomExists := r.classrooms[relation.ClassroomID]
		if !classroomExists || classroom.UserID != arg.UserID {
			continue
		}

		count++
	}

	return count, nil
}

func (r *Repository) ListClassroomsByStudent(_ context.Context, arg repository.ListClassroomsByStudentParams) ([]repository.Classroom, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListClassroomsByStudent); err != nil {
		return nil, err
	}

	student, studentExists := r.students[arg.StudentID]
	if !studentExists || student.UserID != arg.UserID {
		return []repository.Classroom{}, nil
	}

	items := make([]repository.Classroom, 0)
	for _, relation := range r.studentClassrooms {
		if relation.StudentID != arg.StudentID {
			continue
		}

		classroom, classroomExists := r.classrooms[relation.ClassroomID]
		if !classroomExists || classroom.UserID != arg.UserID {
			continue
		}

		items = append(items, classroom)
	}

	sortClassroomsByCreatedAtDesc(items)
	return paginate(items, arg.QueryOffset, arg.QueryLimit), nil
}
