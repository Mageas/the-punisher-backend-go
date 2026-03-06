package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/adapter/persistence/sqlcmapper"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type BonusService interface {
	CreateBonus(
		ctx context.Context,
		userID uuid.UUID,
		studentID uuid.UUID,
		bonusTypeID uuid.UUID,
		points float64,
		occurredAt *time.Time,
		evaluationLabel *string,
	) (*dto.ReturnBonusDto, error)
	GetBonus(ctx context.Context, userID uuid.UUID, bonusID uuid.UUID) (*dto.ReturnBonusDto, error)
	ListBonuses(ctx context.Context, userID uuid.UUID, filters ListBonusesFilters) ([]*dto.ReturnBonusDto, int64, error)
	ListBonusesByStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, used *bool, limit, offset int32) ([]*dto.ReturnBonusDto, int64, error)
	UpdateBonus(
		ctx context.Context,
		userID uuid.UUID,
		bonusID uuid.UUID,
		occurredAt *time.Time,
		evaluationLabelSet bool,
		evaluationLabel *string,
	) (*dto.ReturnBonusDto, error)
	UseBonus(ctx context.Context, userID uuid.UUID, bonusID uuid.UUID) (*dto.ReturnBonusDto, error)
	DeleteBonus(ctx context.Context, userID uuid.UUID, bonusID uuid.UUID) error
}

type bonusService struct {
	repo repository.Querier
}

func NewBonusService(repo repository.Querier) BonusService {
	return &bonusService{repo: repo}
}

func (s *bonusService) CreateBonus(
	ctx context.Context,
	userID uuid.UUID,
	studentID uuid.UUID,
	bonusTypeID uuid.UUID,
	points float64,
	occurredAt *time.Time,
	evaluationLabel *string,
) (*dto.ReturnBonusDto, error) {
	if _, err := s.repo.GetStudentByUser(ctx, repository.GetStudentByUserParams{ID: studentID, UserID: userID}); err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrStudentNotFound
		}
		return nil, fmt.Errorf("failed to get student: %w", err)
	}

	if _, err := s.repo.GetBonusTypeByUser(ctx, repository.GetBonusTypeByUserParams{ID: bonusTypeID, UserID: userID}); err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrBonusTypeNotFound
		}
		return nil, fmt.Errorf("failed to get bonus type: %w", err)
	}

	bonus, err := s.repo.CreateBonus(ctx, repository.CreateBonusParams{
		UserID:          userID,
		StudentID:       studentID,
		BonusTypeID:     bonusTypeID,
		Points:          points,
		OccurredAt:      occurredAt,
		EvaluationLabel: evaluationLabel,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create bonus: %w", err)
	}

	slog.Info("bonus created", "bonus_id", bonus.ID, "user_id", userID, "student_id", studentID, "bonus_type_id", bonusTypeID)

	return sqlcmapper.BonusFromCreateRow(&bonus), nil
}

func (s *bonusService) GetBonus(ctx context.Context, userID uuid.UUID, bonusID uuid.UUID) (*dto.ReturnBonusDto, error) {
	bonus, err := s.repo.GetBonusByUser(ctx, repository.GetBonusByUserParams{ID: bonusID, UserID: userID})
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrBonusNotFound
		}
		return nil, fmt.Errorf("failed to get bonus: %w", err)
	}

	return sqlcmapper.BonusFromGetRow(&bonus), nil
}

func (s *bonusService) ListBonuses(ctx context.Context, userID uuid.UUID, filters ListBonusesFilters) ([]*dto.ReturnBonusDto, int64, error) {
	var used *bool
	if filters.State != nil {
		usedValue := filters.State.Used()
		used = &usedValue
	}

	totalCount, err := s.repo.CountBonusesByUser(ctx, repository.CountBonusesByUserParams{
		UserID:      userID,
		StudentID:   filters.StudentID,
		BonusTypeID: filters.BonusTypeID,
		Used:        used,
		CreatedFrom: filters.CreatedFrom,
		CreatedTo:   filters.CreatedTo,
		ClassroomID: filters.ClassroomID,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count bonuses: %w", err)
	}

	bonuses, err := s.repo.ListBonusesByUser(ctx, repository.ListBonusesByUserParams{
		UserID:      userID,
		StudentID:   filters.StudentID,
		BonusTypeID: filters.BonusTypeID,
		Used:        used,
		CreatedFrom: filters.CreatedFrom,
		CreatedTo:   filters.CreatedTo,
		ClassroomID: filters.ClassroomID,
		QueryOffset: filters.Offset,
		QueryLimit:  filters.Limit,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list bonuses: %w", err)
	}

	return sqlcmapper.BonusListFromListByUserRows(bonuses), totalCount, nil
}

func (s *bonusService) ListBonusesByStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, used *bool, limit, offset int32) ([]*dto.ReturnBonusDto, int64, error) {
	if _, err := s.repo.GetStudentByUser(ctx, repository.GetStudentByUserParams{ID: studentID, UserID: userID}); err != nil {
		if errors.Is(err, repository.ErrNoRows) {
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

	return sqlcmapper.BonusListFromListByStudentRows(bonuses), totalCount, nil
}

func (s *bonusService) UpdateBonus(
	ctx context.Context,
	userID uuid.UUID,
	bonusID uuid.UUID,
	occurredAt *time.Time,
	evaluationLabelSet bool,
	evaluationLabel *string,
) (*dto.ReturnBonusDto, error) {
	bonus, err := s.repo.UpdateBonusByUser(ctx, repository.UpdateBonusByUserParams{
		OccurredAt:         occurredAt,
		EvaluationLabelSet: evaluationLabelSet,
		EvaluationLabel:    evaluationLabel,
		ID:                 bonusID,
		UserID:             userID,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrBonusNotFound
		}
		return nil, fmt.Errorf("failed to update bonus: %w", err)
	}

	slog.Info("bonus updated", "bonus_id", bonus.ID, "user_id", userID)

	return sqlcmapper.BonusFromUpdateRow(&bonus), nil
}

func (s *bonusService) UseBonus(ctx context.Context, userID uuid.UUID, bonusID uuid.UUID) (*dto.ReturnBonusDto, error) {
	bonus, err := s.repo.UseBonus(ctx, repository.UseBonusParams{ID: bonusID, UserID: userID})
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			if _, getErr := s.repo.GetBonusByUser(ctx, repository.GetBonusByUserParams{ID: bonusID, UserID: userID}); getErr != nil {
				if errors.Is(getErr, repository.ErrNoRows) {
					return nil, api.ErrBonusNotFound
				}
				return nil, fmt.Errorf("failed to get bonus: %w", getErr)
			}
			return nil, api.ErrBonusAlreadyUsed
		}
		return nil, fmt.Errorf("failed to use bonus: %w", err)
	}

	slog.Info("bonus used", "bonus_id", bonus.ID, "user_id", userID)

	return sqlcmapper.BonusFromUseRow(&bonus), nil
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
