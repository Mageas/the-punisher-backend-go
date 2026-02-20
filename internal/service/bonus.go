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

type BonusService interface {
	CreateBonus(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, bonusTypeID uuid.UUID, points float64) (*dto.ReturnBonusDto, error)
	GetBonus(ctx context.Context, userID uuid.UUID, bonusID uuid.UUID) (*dto.ReturnBonusDto, error)
	ListBonuses(ctx context.Context, userID uuid.UUID, used *bool, search *string, limit, offset int32) ([]*dto.ReturnBonusDto, int64, error)
	ListBonusesByStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, used *bool, limit, offset int32) ([]*dto.ReturnBonusDto, int64, error)
	UseBonus(ctx context.Context, userID uuid.UUID, bonusID uuid.UUID) (*dto.ReturnBonusDto, error)
	DeleteBonus(ctx context.Context, userID uuid.UUID, bonusID uuid.UUID) error
}

type bonusService struct {
	repo repository.Querier
}

func NewBonusService(repo repository.Querier) BonusService {
	return &bonusService{repo: repo}
}

func (s *bonusService) CreateBonus(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, bonusTypeID uuid.UUID, points float64) (*dto.ReturnBonusDto, error) {
	if _, err := s.repo.GetStudentByUser(ctx, repository.GetStudentByUserParams{ID: studentID, UserID: userID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrStudentNotFound
		}
		return nil, fmt.Errorf("failed to get student: %w", err)
	}

	if _, err := s.repo.GetBonusTypeByUser(ctx, repository.GetBonusTypeByUserParams{ID: bonusTypeID, UserID: userID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrBonusTypeNotFound
		}
		return nil, fmt.Errorf("failed to get bonus type: %w", err)
	}

	bonus, err := s.repo.CreateBonus(ctx, repository.CreateBonusParams{
		UserID:      userID,
		StudentID:   studentID,
		BonusTypeID: bonusTypeID,
		Points:      points,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create bonus: %w", err)
	}

	slog.Info("bonus created", "bonus_id", bonus.ID, "user_id", userID, "student_id", studentID, "bonus_type_id", bonusTypeID)

	return dto.BonusFromCreateRow(&bonus), nil
}

func (s *bonusService) GetBonus(ctx context.Context, userID uuid.UUID, bonusID uuid.UUID) (*dto.ReturnBonusDto, error) {
	bonus, err := s.repo.GetBonusByUser(ctx, repository.GetBonusByUserParams{ID: bonusID, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrBonusNotFound
		}
		return nil, fmt.Errorf("failed to get bonus: %w", err)
	}

	return dto.BonusFromGetRow(&bonus), nil
}

func (s *bonusService) ListBonuses(ctx context.Context, userID uuid.UUID, used *bool, search *string, limit, offset int32) ([]*dto.ReturnBonusDto, int64, error) {
	totalCount, err := s.repo.CountBonusesByUser(ctx, repository.CountBonusesByUserParams{
		UserID: userID,
		Used:   used,
		Search: search,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count bonuses: %w", err)
	}

	bonuses, err := s.repo.ListBonusesByUser(ctx, repository.ListBonusesByUserParams{
		UserID:      userID,
		Used:        used,
		Search:      search,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list bonuses: %w", err)
	}

	return dto.BonusListFromListByUserRows(bonuses), totalCount, nil
}

func (s *bonusService) ListBonusesByStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, used *bool, limit, offset int32) ([]*dto.ReturnBonusDto, int64, error) {
	if _, err := s.repo.GetStudentByUser(ctx, repository.GetStudentByUserParams{ID: studentID, UserID: userID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, api.ErrStudentNotFound
		}
		return nil, 0, fmt.Errorf("failed to get student: %w", err)
	}

	totalCount, err := s.repo.CountBonusesByStudent(ctx, repository.CountBonusesByStudentParams{
		StudentID: studentID,
		UserID:    userID,
		Used:      used,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count bonuses by student: %w", err)
	}

	bonuses, err := s.repo.ListBonusesByStudent(ctx, repository.ListBonusesByStudentParams{
		StudentID:   studentID,
		UserID:      userID,
		Used:        used,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list bonuses by student: %w", err)
	}

	return dto.BonusListFromListByStudentRows(bonuses), totalCount, nil
}

func (s *bonusService) UseBonus(ctx context.Context, userID uuid.UUID, bonusID uuid.UUID) (*dto.ReturnBonusDto, error) {
	bonus, err := s.repo.UseBonus(ctx, repository.UseBonusParams{ID: bonusID, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if _, getErr := s.repo.GetBonusByUser(ctx, repository.GetBonusByUserParams{ID: bonusID, UserID: userID}); getErr != nil {
				if errors.Is(getErr, pgx.ErrNoRows) {
					return nil, api.ErrBonusNotFound
				}
				return nil, fmt.Errorf("failed to get bonus: %w", getErr)
			}
			return nil, api.ErrBonusAlreadyUsed
		}
		return nil, fmt.Errorf("failed to use bonus: %w", err)
	}

	slog.Info("bonus used", "bonus_id", bonus.ID, "user_id", userID)

	return dto.BonusFromUseRow(&bonus), nil
}

func (s *bonusService) DeleteBonus(ctx context.Context, userID uuid.UUID, bonusID uuid.UUID) error {
	rowsAffected, err := s.repo.DeleteBonusByUser(ctx, repository.DeleteBonusByUserParams{ID: bonusID, UserID: userID})
	if err != nil {
		return fmt.Errorf("failed to delete bonus: %w", err)
	}

	if rowsAffected == 0 {
		return api.ErrBonusNotFound
	}

	slog.Info("bonus deleted", "bonus_id", bonusID, "user_id", userID)

	return nil
}
