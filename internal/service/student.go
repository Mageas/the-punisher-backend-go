package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/adapter/persistence/sqlcmapper"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type StudentService interface {
	CreateStudent(ctx context.Context, userID uuid.UUID, req dto.RequestStudentDto) (*dto.ReturnStudentDto, error)
	ImportStudents(ctx context.Context, userID uuid.UUID, file io.Reader, filename string) (*dto.StudentImportResultDto, error)
	GetStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID) (*dto.ReturnStudentDto, error)
	GetStudentKpis(ctx context.Context, userID uuid.UUID, studentID uuid.UUID) (*dto.StudentKpisDto, error)
	ListStudentHistory(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, limit int32, offset int32) ([]dto.StudentHistoryItemDto, int64, error)
	ListStudents(ctx context.Context, userID uuid.UUID, search *string, limit int32, offset int32) ([]*dto.ReturnStudentDto, int64, error)
	UpdateStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, req dto.UpdateStudentDto) (*dto.ReturnStudentDto, error)
	DeleteStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID) error
	DeleteAllStudents(ctx context.Context, userID uuid.UUID) error
}

type studentService struct {
	repo repository.Querier
}

func NewStudentService(repo repository.Querier) StudentService {
	return &studentService{repo: repo}
}

// --- CRUD Operations ---

func (s *studentService) CreateStudent(ctx context.Context, userID uuid.UUID, req dto.RequestStudentDto) (*dto.ReturnStudentDto, error) {
	student, err := s.repo.CreateStudent(ctx, repository.CreateStudentParams{
		UserID:    userID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create student: %w", err)
	}

	slog.Info("student created", "student_id", student.ID, "user_id", userID)

	response := sqlcmapper.StudentFromCreateRow(&student)
	if err := attachClassroomsToStudents(ctx, s.repo, userID, []*dto.ReturnStudentDto{response}); err != nil {
		return nil, fmt.Errorf("failed to list student classrooms: %w", err)
	}

	return response, nil
}

func (s *studentService) GetStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID) (*dto.ReturnStudentDto, error) {
	student, err := s.repo.GetStudentByUser(ctx, repository.GetStudentByUserParams{
		ID:     studentID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrStudentNotFound
		}
		return nil, fmt.Errorf("failed to get student: %w", err)
	}

	response := sqlcmapper.StudentFromGetRow(&student)
	if err := attachClassroomsToStudents(ctx, s.repo, userID, []*dto.ReturnStudentDto{response}); err != nil {
		return nil, fmt.Errorf("failed to list student classrooms: %w", err)
	}

	return response, nil
}

func (s *studentService) ListStudents(ctx context.Context, userID uuid.UUID, search *string, limit int32, offset int32) ([]*dto.ReturnStudentDto, int64, error) {
	totalCount, err := s.repo.CountStudentsByUser(ctx, repository.CountStudentsByUserParams{
		UserID: userID,
		Search: search,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count students: %w", err)
	}

	students, err := s.repo.ListStudentsByUser(ctx, repository.ListStudentsByUserParams{
		UserID:      userID,
		Search:      search,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list students: %w", err)
	}

	response := sqlcmapper.StudentListFromListByUserRows(students)
	if err := attachClassroomsToStudents(ctx, s.repo, userID, response); err != nil {
		return nil, 0, fmt.Errorf("failed to list student classrooms: %w", err)
	}

	return response, totalCount, nil
}

func (s *studentService) UpdateStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, req dto.UpdateStudentDto) (*dto.ReturnStudentDto, error) {
	params := repository.UpdateStudentByUserParams{
		ID:     studentID,
		UserID: userID,
	}

	if req.FirstName != nil {
		params.FirstName = req.FirstName
	}
	if req.LastName != nil {
		params.LastName = req.LastName
	}

	student, err := s.repo.UpdateStudentByUser(ctx, params)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrStudentNotFound
		}
		return nil, fmt.Errorf("failed to update student: %w", err)
	}

	response := sqlcmapper.StudentFromUpdateRow(&student)
	if err := attachClassroomsToStudents(ctx, s.repo, userID, []*dto.ReturnStudentDto{response}); err != nil {
		return nil, fmt.Errorf("failed to list student classrooms: %w", err)
	}

	return response, nil
}

func (s *studentService) DeleteStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID) error {
	rowsAffected, err := s.repo.DeleteStudentByUser(ctx, repository.DeleteStudentByUserParams{
		ID:     studentID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete student: %w", err)
	}

	if rowsAffected == 0 {
		return api.ErrStudentNotFound
	}

	slog.Info("student deleted", "student_id", studentID, "user_id", userID)

	return nil
}

func (s *studentService) DeleteAllStudents(ctx context.Context, userID uuid.UUID) error {
	rowsAffected, err := s.repo.DeleteAllStudentsByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to delete students in bulk: %w", err)
	}

	slog.Info("students deleted in bulk", "deleted_count", rowsAffected, "user_id", userID)

	return nil
}

// --- KPIs & History ---

func (s *studentService) GetStudentKpis(ctx context.Context, userID uuid.UUID, studentID uuid.UUID) (*dto.StudentKpisDto, error) {
	if err := s.ensureStudentExists(ctx, userID, studentID); err != nil {
		return nil, err
	}

	kpis, err := s.repo.GetStudentKpis(ctx, repository.GetStudentKpisParams{
		StudentID: studentID,
		UserID:    userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get student kpis: %w", err)
	}

	return sqlcmapper.StudentKpisFromRow(&kpis), nil
}

func (s *studentService) ListStudentHistory(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, limit int32, offset int32) ([]dto.StudentHistoryItemDto, int64, error) {
	if err := s.ensureStudentExists(ctx, userID, studentID); err != nil {
		return nil, 0, err
	}

	totalCount, err := s.repo.CountStudentHistory(ctx, repository.CountStudentHistoryParams{
		StudentID: studentID,
		UserID:    userID,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count student history: %w", err)
	}

	history, err := s.repo.ListStudentHistory(ctx, repository.ListStudentHistoryParams{
		StudentID:   studentID,
		UserID:      userID,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list student history: %w", err)
	}

	return sqlcmapper.StudentHistoryFromRows(history), totalCount, nil
}

// --- Helpers ---

func (s *studentService) ensureStudentExists(ctx context.Context, userID uuid.UUID, studentID uuid.UUID) error {
	_, err := s.repo.GetStudentByUser(ctx, repository.GetStudentByUserParams{
		ID:     studentID,
		UserID: userID,
	})
	if err == nil {
		return nil
	}

	if errors.Is(err, repository.ErrNoRows) {
		return api.ErrStudentNotFound
	}

	return fmt.Errorf("failed to get student: %w", err)
}

func attachClassroomsToStudents(ctx context.Context, repo repository.Querier, userID uuid.UUID, students []*dto.ReturnStudentDto) error {
	if len(students) == 0 {
		return nil
	}

	studentIDs := make([]uuid.UUID, 0, len(students))
	for _, student := range students {
		if student == nil {
			continue
		}
		studentIDs = append(studentIDs, student.ID)
	}

	if len(studentIDs) == 0 {
		return nil
	}

	rows, err := repo.ListClassroomRefsByStudentIDs(ctx, repository.ListClassroomRefsByStudentIDsParams{
		UserID:     userID,
		StudentIds: studentIDs,
	})
	if err != nil {
		return err
	}

	classroomsByStudent := sqlcmapper.StudentClassroomsByStudentFromRows(rows)
	for _, student := range students {
		if student == nil {
			continue
		}

		classrooms := classroomsByStudent[student.ID]
		if classrooms == nil {
			classrooms = []dto.StudentClassroomDto{}
		}
		student.Classrooms = classrooms
	}

	return nil
}
