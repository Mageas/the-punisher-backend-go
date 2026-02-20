package inmemory

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

const (
	OpCreateClassroom                   = "CreateClassroom"
	OpGetClassroomByUser                = "GetClassroomByUser"
	OpCountClassroomsByUser             = "CountClassroomsByUser"
	OpListClassroomsByUser              = "ListClassroomsByUser"
	OpUpdateClassroomByUser             = "UpdateClassroomByUser"
	OpDeleteClassroomByUser             = "DeleteClassroomByUser"
	OpAddStudentToClassroom             = "AddStudentToClassroom"
	OpRemoveStudentFromClassroom        = "RemoveStudentFromClassroom"
	OpCountStudentsByClassroom          = "CountStudentsByClassroom"
	OpListStudentsByClassroom           = "ListStudentsByClassroom"
	OpCountClassroomsByStudent          = "CountClassroomsByStudent"
	OpListClassroomsByStudent           = "ListClassroomsByStudent"
	OpListStudentsPreviewByClassroomIDs = "ListStudentsPreviewByClassroomIDs"
)

func (r *Repository) classroomAggregateFieldsLocked(classroom repository.Classroom) (int64, float64, int64) {
	studentIDs := make(map[uuid.UUID]struct{})
	for _, relation := range r.studentClassrooms {
		if relation.ClassroomID != classroom.ID {
			continue
		}
		studentIDs[relation.StudentID] = struct{}{}
	}

	studentCount := int64(len(studentIDs))

	var totalBonusPoints float64
	for _, bonus := range r.bonuses {
		if bonus.UserID != classroom.UserID || hasTime(bonus.UsedAt) {
			continue
		}
		if _, ok := studentIDs[bonus.StudentID]; ok {
			totalBonusPoints += bonus.Points
		}
	}

	var totalPenaltyCount int64
	for _, penalty := range r.penalties {
		if penalty.UserID != classroom.UserID {
			continue
		}
		if _, ok := studentIDs[penalty.StudentID]; ok {
			totalPenaltyCount++
		}
	}

	return studentCount, totalBonusPoints, totalPenaltyCount
}

func (r *Repository) buildCreateClassroomRowLocked(classroom repository.Classroom) repository.CreateClassroomRow {
	studentCount, totalBonusPoints, totalPenaltyCount := r.classroomAggregateFieldsLocked(classroom)

	return repository.CreateClassroomRow{
		ID:                classroom.ID,
		UserID:            classroom.UserID,
		Name:              classroom.Name,
		Year:              classroom.Year,
		MainTeacher:       classroom.MainTeacher,
		CreatedAt:         classroom.CreatedAt,
		UpdatedAt:         classroom.UpdatedAt,
		StudentCount:      studentCount,
		TotalBonusPoints:  totalBonusPoints,
		TotalPenaltyCount: totalPenaltyCount,
	}
}

func (r *Repository) buildGetClassroomRowLocked(classroom repository.Classroom) repository.GetClassroomByUserRow {
	studentCount, totalBonusPoints, totalPenaltyCount := r.classroomAggregateFieldsLocked(classroom)

	return repository.GetClassroomByUserRow{
		ID:                classroom.ID,
		UserID:            classroom.UserID,
		Name:              classroom.Name,
		Year:              classroom.Year,
		MainTeacher:       classroom.MainTeacher,
		CreatedAt:         classroom.CreatedAt,
		UpdatedAt:         classroom.UpdatedAt,
		StudentCount:      studentCount,
		TotalBonusPoints:  totalBonusPoints,
		TotalPenaltyCount: totalPenaltyCount,
	}
}

func (r *Repository) buildListClassroomByUserRowLocked(classroom repository.Classroom) repository.ListClassroomsByUserRow {
	studentCount, totalBonusPoints, totalPenaltyCount := r.classroomAggregateFieldsLocked(classroom)

	return repository.ListClassroomsByUserRow{
		ID:                classroom.ID,
		UserID:            classroom.UserID,
		Name:              classroom.Name,
		Year:              classroom.Year,
		MainTeacher:       classroom.MainTeacher,
		CreatedAt:         classroom.CreatedAt,
		UpdatedAt:         classroom.UpdatedAt,
		StudentCount:      studentCount,
		TotalBonusPoints:  totalBonusPoints,
		TotalPenaltyCount: totalPenaltyCount,
	}
}

func (r *Repository) buildUpdateClassroomRowLocked(classroom repository.Classroom) repository.UpdateClassroomByUserRow {
	studentCount, totalBonusPoints, totalPenaltyCount := r.classroomAggregateFieldsLocked(classroom)

	return repository.UpdateClassroomByUserRow{
		ID:                classroom.ID,
		UserID:            classroom.UserID,
		Name:              classroom.Name,
		Year:              classroom.Year,
		MainTeacher:       classroom.MainTeacher,
		CreatedAt:         classroom.CreatedAt,
		UpdatedAt:         classroom.UpdatedAt,
		StudentCount:      studentCount,
		TotalBonusPoints:  totalBonusPoints,
		TotalPenaltyCount: totalPenaltyCount,
	}
}

func (r *Repository) buildListClassroomByStudentRowLocked(classroom repository.Classroom) repository.ListClassroomsByStudentRow {
	studentCount, totalBonusPoints, totalPenaltyCount := r.classroomAggregateFieldsLocked(classroom)

	return repository.ListClassroomsByStudentRow{
		ID:                classroom.ID,
		UserID:            classroom.UserID,
		Name:              classroom.Name,
		Year:              classroom.Year,
		MainTeacher:       classroom.MainTeacher,
		CreatedAt:         classroom.CreatedAt,
		UpdatedAt:         classroom.UpdatedAt,
		StudentCount:      studentCount,
		TotalBonusPoints:  totalBonusPoints,
		TotalPenaltyCount: totalPenaltyCount,
	}
}

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

func (r *Repository) CreateClassroom(_ context.Context, arg repository.CreateClassroomParams) (repository.CreateClassroomRow, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpCreateClassroom); err != nil {
		return repository.CreateClassroomRow{}, err
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

	return r.buildCreateClassroomRowLocked(classroom), nil
}

func (r *Repository) GetClassroomByUser(_ context.Context, arg repository.GetClassroomByUserParams) (repository.GetClassroomByUserRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpGetClassroomByUser); err != nil {
		return repository.GetClassroomByUserRow{}, err
	}

	classroom, ok := r.classrooms[arg.ID]
	if !ok || classroom.UserID != arg.UserID {
		return repository.GetClassroomByUserRow{}, pgx.ErrNoRows
	}

	return r.buildGetClassroomRowLocked(classroom), nil
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

func (r *Repository) ListClassroomsByUser(_ context.Context, arg repository.ListClassroomsByUserParams) ([]repository.ListClassroomsByUserRow, error) {
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
	paginated := paginate(items, arg.QueryOffset, arg.QueryLimit)

	rows := make([]repository.ListClassroomsByUserRow, 0, len(paginated))
	for _, classroom := range paginated {
		rows = append(rows, r.buildListClassroomByUserRowLocked(classroom))
	}

	return rows, nil
}

func (r *Repository) UpdateClassroomByUser(_ context.Context, arg repository.UpdateClassroomByUserParams) (repository.UpdateClassroomByUserRow, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpUpdateClassroomByUser); err != nil {
		return repository.UpdateClassroomByUserRow{}, err
	}

	classroom, ok := r.classrooms[arg.ID]
	if !ok || classroom.UserID != arg.UserID {
		return repository.UpdateClassroomByUserRow{}, pgx.ErrNoRows
	}

	if arg.Name != nil {
		classroom.Name = *arg.Name
	}
	if arg.Year != nil {
		classroom.Year = arg.Year
	}
	if arg.MainTeacher != nil {
		classroom.MainTeacher = arg.MainTeacher
	}

	classroom.UpdatedAt = time.Now()
	r.classrooms[arg.ID] = classroom

	return r.buildUpdateClassroomRowLocked(classroom), nil
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

func (r *Repository) ListStudentsByClassroom(_ context.Context, arg repository.ListStudentsByClassroomParams) ([]repository.ListStudentsByClassroomRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListStudentsByClassroom); err != nil {
		return nil, err
	}

	classroom, classroomExists := r.classrooms[arg.ClassroomID]
	if !classroomExists || classroom.UserID != arg.UserID {
		return []repository.ListStudentsByClassroomRow{}, nil
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
	paginated := paginate(items, arg.QueryOffset, arg.QueryLimit)

	rows := make([]repository.ListStudentsByClassroomRow, 0, len(paginated))
	for _, student := range paginated {
		rows = append(rows, r.buildListStudentByClassroomRowLocked(student))
	}

	return rows, nil
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

func (r *Repository) ListClassroomsByStudent(_ context.Context, arg repository.ListClassroomsByStudentParams) ([]repository.ListClassroomsByStudentRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListClassroomsByStudent); err != nil {
		return nil, err
	}

	student, studentExists := r.students[arg.StudentID]
	if !studentExists || student.UserID != arg.UserID {
		return []repository.ListClassroomsByStudentRow{}, nil
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
	paginated := paginate(items, arg.QueryOffset, arg.QueryLimit)

	rows := make([]repository.ListClassroomsByStudentRow, 0, len(paginated))
	for _, classroom := range paginated {
		rows = append(rows, r.buildListClassroomByStudentRowLocked(classroom))
	}

	return rows, nil
}

func (r *Repository) ListStudentsPreviewByClassroomIDs(_ context.Context, arg repository.ListStudentsPreviewByClassroomIDsParams) ([]repository.ListStudentsPreviewByClassroomIDsRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListStudentsPreviewByClassroomIDs); err != nil {
		return nil, err
	}
	if arg.PreviewLimit <= 0 || len(arg.ClassroomIds) == 0 {
		return []repository.ListStudentsPreviewByClassroomIDsRow{}, nil
	}

	classroomIDs := make([]uuid.UUID, 0, len(arg.ClassroomIds))
	seenClassrooms := make(map[uuid.UUID]struct{}, len(arg.ClassroomIds))
	for _, classroomID := range arg.ClassroomIds {
		if _, seen := seenClassrooms[classroomID]; seen {
			continue
		}

		classroom, classroomExists := r.classrooms[classroomID]
		if !classroomExists || classroom.UserID != arg.UserID {
			continue
		}

		seenClassrooms[classroomID] = struct{}{}
		classroomIDs = append(classroomIDs, classroomID)
	}

	sort.Slice(classroomIDs, func(i, j int) bool {
		return classroomIDs[i].String() < classroomIDs[j].String()
	})

	rows := make([]repository.ListStudentsPreviewByClassroomIDsRow, 0)
	for _, classroomID := range classroomIDs {
		students := make([]repository.Student, 0)
		for _, relation := range r.studentClassrooms {
			if relation.ClassroomID != classroomID {
				continue
			}

			student, studentExists := r.students[relation.StudentID]
			if !studentExists {
				continue
			}

			students = append(students, student)
		}

		sortStudentsByCreatedAtDesc(students)
		paginatedStudents := paginate(students, 0, arg.PreviewLimit)

		for _, student := range paginatedStudents {
			rows = append(rows, repository.ListStudentsPreviewByClassroomIDsRow{
				ClassroomID: classroomID,
				StudentID:   student.ID,
				FirstName:   student.FirstName,
				LastName:    student.LastName,
			})
		}
	}

	return rows, nil
}
