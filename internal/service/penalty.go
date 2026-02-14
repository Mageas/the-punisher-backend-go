package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type PenaltyService interface {
	CreatePenalty(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, penaltyTypeID uuid.UUID) (*dto.ReturnPenaltyDto, error)
	GetPenalty(ctx context.Context, userID uuid.UUID, penaltyID uuid.UUID) (*dto.ReturnPenaltyDto, error)
	ListPenalties(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*dto.ReturnPenaltyDto, int64, error)
	ListPenaltiesByStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, limit, offset int32) ([]*dto.ReturnPenaltyDto, int64, error)
	DeletePenalty(ctx context.Context, userID uuid.UUID, penaltyID uuid.UUID) error
}

type penaltyService struct {
	repo repository.Querier
}

func NewPenaltyService(repo repository.Querier) PenaltyService {
	return &penaltyService{repo: repo}
}

func (s *penaltyService) CreatePenalty(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, penaltyTypeID uuid.UUID) (*dto.ReturnPenaltyDto, error) {
	if _, err := s.repo.GetStudentByUser(ctx, repository.GetStudentByUserParams{ID: studentID, UserID: userID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrStudentNotFound
		}
		return nil, fmt.Errorf("failed to get student: %w", err)
	}

	if _, err := s.repo.GetPenaltyTypeByUser(ctx, repository.GetPenaltyTypeByUserParams{ID: penaltyTypeID, UserID: userID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrPenaltyTypeNotFound
		}
		return nil, fmt.Errorf("failed to get penalty type: %w", err)
	}

	penalty, err := s.repo.CreatePenalty(ctx, repository.CreatePenaltyParams{
		UserID:        userID,
		StudentID:     studentID,
		PenaltyTypeID: penaltyTypeID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create penalty: %w", err)
	}

	slog.Info("penalty created", "penalty_id", penalty.ID, "user_id", userID, "student_id", studentID, "penalty_type_id", penaltyTypeID)

	return dto.PenaltyFromRepository(&penalty), nil
}

func (s *penaltyService) GetPenalty(ctx context.Context, userID uuid.UUID, penaltyID uuid.UUID) (*dto.ReturnPenaltyDto, error) {
	penalty, err := s.repo.GetPenaltyByUser(ctx, repository.GetPenaltyByUserParams{ID: penaltyID, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrPenaltyNotFound
		}
		return nil, fmt.Errorf("failed to get penalty: %w", err)
	}

	return dto.PenaltyFromRepository(&penalty), nil
}

func (s *penaltyService) ListPenalties(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*dto.ReturnPenaltyDto, int64, error) {
	totalCount, err := s.repo.CountPenaltiesByUser(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count penalties: %w", err)
	}

	penalties, err := s.repo.ListPenaltiesByUser(ctx, repository.ListPenaltiesByUserParams{
		UserID:      userID,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list penalties: %w", err)
	}

	return dto.PenaltyListFromRepository(penalties), totalCount, nil
}

func (s *penaltyService) ListPenaltiesByStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, limit, offset int32) ([]*dto.ReturnPenaltyDto, int64, error) {
	if _, err := s.repo.GetStudentByUser(ctx, repository.GetStudentByUserParams{ID: studentID, UserID: userID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, api.ErrStudentNotFound
		}
		return nil, 0, fmt.Errorf("failed to get student: %w", err)
	}

	totalCount, err := s.repo.CountPenaltiesByStudent(ctx, repository.CountPenaltiesByStudentParams{
		StudentID: studentID,
		UserID:    userID,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count penalties by student: %w", err)
	}

	penalties, err := s.repo.ListPenaltiesByStudent(ctx, repository.ListPenaltiesByStudentParams{
		StudentID:   studentID,
		UserID:      userID,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list penalties by student: %w", err)
	}

	return dto.PenaltyListFromRepository(penalties), totalCount, nil
}

func (s *penaltyService) DeletePenalty(ctx context.Context, userID uuid.UUID, penaltyID uuid.UUID) error {
	rowsAffected, err := s.repo.DeletePenaltyByUser(ctx, repository.DeletePenaltyByUserParams{ID: penaltyID, UserID: userID})
	if err != nil {
		return fmt.Errorf("failed to delete penalty: %w", err)
	}

	if rowsAffected == 0 {
		return api.ErrPenaltyNotFound
	}

	slog.Info("penalty deleted", "penalty_id", penaltyID, "user_id", userID)

	return nil
}
