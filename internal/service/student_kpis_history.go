package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mageas/the-punisher-backend/internal/adapter/persistence/sqlcmapper"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

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

func (s *studentService) ListStudentHistory(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, limit int32, offset int32) ([]dto.StudentHistoryItemDto, error) {
	if err := s.ensureStudentExists(ctx, userID, studentID); err != nil {
		return nil, err
	}

	history, err := s.repo.ListStudentHistory(ctx, repository.ListStudentHistoryParams{
		StudentID:   studentID,
		UserID:      userID,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list student history: %w", err)
	}

	return sqlcmapper.StudentHistoryFromRows(history), nil
}

func (s *studentService) ensureStudentExists(ctx context.Context, userID uuid.UUID, studentID uuid.UUID) error {
	_, err := s.repo.GetStudentByUser(ctx, repository.GetStudentByUserParams{
		ID:     studentID,
		UserID: userID,
	})
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return api.ErrStudentNotFound
	}

	return fmt.Errorf("failed to get student: %w", err)
}
