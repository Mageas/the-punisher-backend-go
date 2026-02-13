package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type ClassroomService interface {
	CreateClassroom(ctx context.Context, userID uuid.UUID, req dto.RequestClassroomDto) (*dto.ReturnClassroomDto, error)
	GetClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID) (*dto.ReturnClassroomDto, error)
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

func (s *classroomService) CreateClassroom(ctx context.Context, userID uuid.UUID, req dto.RequestClassroomDto) (*dto.ReturnClassroomDto, error) {
	params := repository.CreateClassroomParams{
		UserID: userID,
		Name:   req.Name,
	}

	if req.Year != nil {
		params.Year = pgtype.Text{String: *req.Year, Valid: true}
	}
	if req.MainTeacher != nil {
		params.MainTeacher = pgtype.Text{String: *req.MainTeacher, Valid: true}
	}

	classroom, err := s.repo.CreateClassroom(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create classroom: %w", err)
	}

	slog.Info("classroom created", "classroom_id", classroom.ID, "user_id", userID)

	return dto.ClassroomFromRepository(&classroom), nil
}

func (s *classroomService) GetClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID) (*dto.ReturnClassroomDto, error) {
	classroom, err := s.repo.GetClassroomByUser(ctx, repository.GetClassroomByUserParams{
		ID:     classroomID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrClassroomNotFound
		}
		return nil, fmt.Errorf("failed to get classroom: %w", err)
	}

	return dto.ClassroomFromRepository(&classroom), nil
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

	return dto.ClassroomListFromRepository(classrooms), totalCount, nil
}

func (s *classroomService) UpdateClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID, req dto.UpdateClassroomDto) (*dto.ReturnClassroomDto, error) {
	params := repository.UpdateClassroomByUserParams{
		ID:     classroomID,
		UserID: userID,
	}

	if req.Name != nil {
		params.Name = pgtype.Text{String: *req.Name, Valid: true}
	}
	if req.Year != nil {
		params.Year = pgtype.Text{String: *req.Year, Valid: true}
	}
	if req.MainTeacher != nil {
		params.MainTeacher = pgtype.Text{String: *req.MainTeacher, Valid: true}
	}

	classroom, err := s.repo.UpdateClassroomByUser(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrClassroomNotFound
		}
		return nil, fmt.Errorf("failed to update classroom: %w", err)
	}

	return dto.ClassroomFromRepository(&classroom), nil
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

func (s *classroomService) AddStudentToClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID, studentID uuid.UUID) error {
	rowsAffected, err := s.repo.AddStudentToClassroom(ctx, repository.AddStudentToClassroomParams{
		StudentID:   studentID,
		ClassroomID: classroomID,
		UserID:      userID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
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
		if errors.Is(err, pgx.ErrNoRows) {
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

	return dto.StudentListFromRepository(students), totalCount, nil
}

func (s *classroomService) ListClassroomsByStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, limit int32, offset int32) ([]*dto.ReturnClassroomDto, int64, error) {
	_, err := s.repo.GetStudentByUser(ctx, repository.GetStudentByUserParams{
		ID:     studentID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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

	return dto.ClassroomListFromRepository(classrooms), totalCount, nil
}
