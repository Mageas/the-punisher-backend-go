package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/adapter/persistence/sqlcmapper"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type ClassroomService interface {
	CreateClassroom(ctx context.Context, userID uuid.UUID, req dto.RequestClassroomDto) (*dto.ReturnClassroomDto, error)
	GetClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID) (*dto.ReturnClassroomDto, error)
	GetClassroomKpis(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID) (*dto.DashboardKpisDto, error)
	ListClassrooms(ctx context.Context, userID uuid.UUID, limit int32, offset int32) ([]*dto.ReturnClassroomDto, int64, error)
	UpdateClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID, req dto.UpdateClassroomDto) (*dto.ReturnClassroomDto, error)
	DeleteClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID) error

	AddStudentToClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID, studentID uuid.UUID) error
	RemoveStudentFromClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID, studentID uuid.UUID) error
	ListStudentsByClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID, limit int32, offset int32) ([]*dto.ReturnStudentDto, int64, error)
	ListClassroomsByStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, limit int32, offset int32) ([]*dto.ReturnClassroomDto, int64, error)
}

type classroomService struct {
	repo repository.Querier
}

func NewClassroomService(repo repository.Querier) ClassroomService {
	return &classroomService{repo: repo}
}

// --- CRUD Operations ---

func (s *classroomService) CreateClassroom(ctx context.Context, userID uuid.UUID, req dto.RequestClassroomDto) (*dto.ReturnClassroomDto, error) {
	params := repository.CreateClassroomParams{
		UserID: userID,
		Name:   req.Name,
	}

	if req.Year != nil {
		params.Year = req.Year
	}
	if req.MainTeacher != nil {
		params.MainTeacher = req.MainTeacher
	}

	classroom, err := s.repo.CreateClassroom(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create classroom: %w", err)
	}

	slog.Info("classroom created", "classroom_id", classroom.ID, "user_id", userID)

	response := sqlcmapper.ClassroomFromCreateRow(&classroom)
	if err := attachStudentsPreviewToClassrooms(ctx, s.repo, userID, []*dto.ReturnClassroomDto{response}); err != nil {
		return nil, fmt.Errorf("failed to list classroom students preview: %w", err)
	}

	return response, nil
}

func (s *classroomService) GetClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID) (*dto.ReturnClassroomDto, error) {
	classroom, err := s.repo.GetClassroomByUser(ctx, repository.GetClassroomByUserParams{
		ID:     classroomID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrClassroomNotFound
		}
		return nil, fmt.Errorf("failed to get classroom: %w", err)
	}

	response := sqlcmapper.ClassroomFromGetRow(&classroom)
	if err := attachStudentsPreviewToClassrooms(ctx, s.repo, userID, []*dto.ReturnClassroomDto{response}); err != nil {
		return nil, fmt.Errorf("failed to list classroom students preview: %w", err)
	}

	return response, nil
}

func (s *classroomService) GetClassroomKpis(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID) (*dto.DashboardKpisDto, error) {
	if _, err := s.repo.GetClassroomByUser(ctx, repository.GetClassroomByUserParams{
		ID:     classroomID,
		UserID: userID,
	}); err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrClassroomNotFound
		}
		return nil, fmt.Errorf("failed to get classroom: %w", err)
	}

	kpis, err := s.repo.GetDashboardKpis(ctx, repository.GetDashboardKpisParams{
		UserID:      userID,
		ClassroomID: &classroomID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get classroom kpis: %w", err)
	}

	return sqlcmapper.DashboardKpisFromRow(&kpis), nil
}

func (s *classroomService) ListClassrooms(ctx context.Context, userID uuid.UUID, limit int32, offset int32) ([]*dto.ReturnClassroomDto, int64, error) {
	totalCount, err := s.repo.CountClassroomsByUser(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count classrooms: %w", err)
	}

	classrooms, err := s.repo.ListClassroomsByUser(ctx, repository.ListClassroomsByUserParams{
		UserID:      userID,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list classrooms: %w", err)
	}

	response := sqlcmapper.ClassroomListFromListByUserRows(classrooms)
	if err := attachStudentsPreviewToClassrooms(ctx, s.repo, userID, response); err != nil {
		return nil, 0, fmt.Errorf("failed to list classroom students preview: %w", err)
	}

	return response, totalCount, nil
}

func (s *classroomService) UpdateClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID, req dto.UpdateClassroomDto) (*dto.ReturnClassroomDto, error) {
	params := repository.UpdateClassroomByUserParams{
		ID:     classroomID,
		UserID: userID,
	}

	if req.Name != nil {
		params.Name = req.Name
	}
	if req.Year != nil {
		params.Year = req.Year
	}
	if req.MainTeacher != nil {
		params.MainTeacher = req.MainTeacher
	}

	classroom, err := s.repo.UpdateClassroomByUser(ctx, params)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrClassroomNotFound
		}
		return nil, fmt.Errorf("failed to update classroom: %w", err)
	}

	response := sqlcmapper.ClassroomFromUpdateRow(&classroom)
	if err := attachStudentsPreviewToClassrooms(ctx, s.repo, userID, []*dto.ReturnClassroomDto{response}); err != nil {
		return nil, fmt.Errorf("failed to list classroom students preview: %w", err)
	}

	return response, nil
}

func (s *classroomService) DeleteClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID) error {
	rowsAffected, err := s.repo.DeleteClassroomByUser(ctx, repository.DeleteClassroomByUserParams{
		ID:     classroomID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete classroom: %w", err)
	}

	if rowsAffected == 0 {
		return api.ErrClassroomNotFound
	}

	slog.Info("classroom deleted", "classroom_id", classroomID, "user_id", userID)

	return nil
}

// --- Membership Operations ---

func (s *classroomService) AddStudentToClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID, studentID uuid.UUID) error {
	rowsAffected, err := s.repo.AddStudentToClassroom(ctx, repository.AddStudentToClassroomParams{
		StudentID:   studentID,
		ClassroomID: classroomID,
		UserID:      userID,
	})
	if err != nil {
		if repository.IsUniqueViolation(err) {
			return api.ErrStudentClassroomRelationExists
		}
		return fmt.Errorf("failed to add student to classroom: %w", err)
	}

	if rowsAffected == 0 {
		return api.ErrStudentOrClassroomNotFound
	}

	slog.Info("student added to classroom", "student_id", studentID, "classroom_id", classroomID, "user_id", userID)

	return nil
}

func (s *classroomService) RemoveStudentFromClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID, studentID uuid.UUID) error {
	rowsAffected, err := s.repo.RemoveStudentFromClassroom(ctx, repository.RemoveStudentFromClassroomParams{
		StudentID:   studentID,
		ClassroomID: classroomID,
		UserID:      userID,
	})
	if err != nil {
		return fmt.Errorf("failed to remove student from classroom: %w", err)
	}

	if rowsAffected == 0 {
		return api.ErrStudentOrClassroomNotFound
	}

	slog.Info("student removed from classroom", "student_id", studentID, "classroom_id", classroomID, "user_id", userID)

	return nil
}

func (s *classroomService) ListStudentsByClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID, limit int32, offset int32) ([]*dto.ReturnStudentDto, int64, error) {
	_, err := s.repo.GetClassroomByUser(ctx, repository.GetClassroomByUserParams{
		ID:     classroomID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, 0, api.ErrClassroomNotFound
		}
		return nil, 0, fmt.Errorf("failed to get classroom: %w", err)
	}

	totalCount, err := s.repo.CountStudentsByClassroom(ctx, repository.CountStudentsByClassroomParams{
		ClassroomID: classroomID,
		UserID:      userID,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count students by classroom: %w", err)
	}

	students, err := s.repo.ListStudentsByClassroom(ctx, repository.ListStudentsByClassroomParams{
		ClassroomID: classroomID,
		UserID:      userID,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list students by classroom: %w", err)
	}

	response := sqlcmapper.StudentListFromListByClassroomRows(students)
	if err := attachClassroomsToStudents(ctx, s.repo, userID, response); err != nil {
		return nil, 0, fmt.Errorf("failed to list student classrooms: %w", err)
	}

	return response, totalCount, nil
}

func (s *classroomService) ListClassroomsByStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, limit int32, offset int32) ([]*dto.ReturnClassroomDto, int64, error) {
	_, err := s.repo.GetStudentByUser(ctx, repository.GetStudentByUserParams{
		ID:     studentID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, 0, api.ErrStudentNotFound
		}
		return nil, 0, fmt.Errorf("failed to get student: %w", err)
	}

	totalCount, err := s.repo.CountClassroomsByStudent(ctx, repository.CountClassroomsByStudentParams{
		StudentID: studentID,
		UserID:    userID,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count classrooms by student: %w", err)
	}

	classrooms, err := s.repo.ListClassroomsByStudent(ctx, repository.ListClassroomsByStudentParams{
		StudentID:   studentID,
		UserID:      userID,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list classrooms by student: %w", err)
	}

	response := sqlcmapper.ClassroomListFromListByStudentRows(classrooms)
	if err := attachStudentsPreviewToClassrooms(ctx, s.repo, userID, response); err != nil {
		return nil, 0, fmt.Errorf("failed to list classroom students preview: %w", err)
	}

	return response, totalCount, nil
}

// --- Helpers ---

const classroomStudentsPreviewLimit int32 = 5

func attachStudentsPreviewToClassrooms(ctx context.Context, repo repository.Querier, userID uuid.UUID, classrooms []*dto.ReturnClassroomDto) error {
	if len(classrooms) == 0 {
		return nil
	}

	classroomIDs := make([]uuid.UUID, 0, len(classrooms))
	for _, classroom := range classrooms {
		if classroom == nil {
			continue
		}
		classroomIDs = append(classroomIDs, classroom.ID)
	}

	if len(classroomIDs) == 0 {
		return nil
	}

	rows, err := repo.ListStudentsPreviewByClassroomIDs(ctx, repository.ListStudentsPreviewByClassroomIDsParams{
		PreviewLimit: classroomStudentsPreviewLimit,
		UserID:       userID,
		ClassroomIds: classroomIDs,
	})
	if err != nil {
		return err
	}

	previewByClassroom := sqlcmapper.ClassroomStudentsPreviewByClassroomFromRows(rows)
	for _, classroom := range classrooms {
		if classroom == nil {
			continue
		}

		preview := previewByClassroom[classroom.ID]
		if preview == nil {
			preview = []dto.ClassroomStudentPreviewDto{}
		}
		classroom.StudentsPreview = preview
	}

	return nil
}
