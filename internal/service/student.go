package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type StudentService interface {
	CreateStudent(ctx context.Context, userID uuid.UUID, req dto.RequestStudentDto) (*dto.ReturnStudentDto, error)
	GetStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID) (*dto.ReturnStudentDto, error)
	ListStudents(ctx context.Context, userID uuid.UUID, limit int32, offset int32) ([]*dto.ReturnStudentDto, int64, error)
	UpdateStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, req dto.UpdateStudentDto) (*dto.ReturnStudentDto, error)
	DeleteStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID) error
}

type studentService struct {
	repo repository.Querier
}

func NewStudentService(repo repository.Querier) StudentService {
	return &studentService{repo: repo}
}

func (s *studentService) CreateStudent(ctx context.Context, userID uuid.UUID, req dto.RequestStudentDto) (*dto.ReturnStudentDto, error) {
	student, err := s.repo.CreateStudent(ctx, repository.CreateStudentParams{
		UserID:    userID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create student: %w", err)
	}

	return dto.StudentFromRepository(&student), nil
}

func (s *studentService) GetStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID) (*dto.ReturnStudentDto, error) {
	student, err := s.repo.GetStudentByUser(ctx, repository.GetStudentByUserParams{
		ID:     studentID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrStudentNotFound
		}
		return nil, fmt.Errorf("failed to get student: %w", err)
	}

	return dto.StudentFromRepository(&student), nil
}

func (s *studentService) ListStudents(ctx context.Context, userID uuid.UUID, limit int32, offset int32) ([]*dto.ReturnStudentDto, int64, error) {
	totalCount, err := s.repo.CountStudentsByUser(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count students: %w", err)
	}

	students, err := s.repo.ListStudentsByUser(ctx, repository.ListStudentsByUserParams{
		UserID:      userID,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list students: %w", err)
	}

	return dto.StudentListFromRepository(students), totalCount, nil
}

func (s *studentService) UpdateStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, req dto.UpdateStudentDto) (*dto.ReturnStudentDto, error) {
	params := repository.UpdateStudentByUserParams{
		ID:     studentID,
		UserID: userID,
	}

	if req.FirstName != nil {
		params.FirstName = pgtype.Text{String: *req.FirstName, Valid: true}
	}
	if req.LastName != nil {
		params.LastName = pgtype.Text{String: *req.LastName, Valid: true}
	}

	student, err := s.repo.UpdateStudentByUser(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrStudentNotFound
		}
		return nil, fmt.Errorf("failed to update student: %w", err)
	}

	return dto.StudentFromRepository(&student), nil
}

func (s *studentService) DeleteStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID) error {
	err := s.repo.DeleteStudentByUser(ctx, repository.DeleteStudentByUserParams{
		ID:     studentID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete student: %w", err)
	}

	return nil
}
