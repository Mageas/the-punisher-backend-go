package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mageas/the-punisher-backend/internal/adapter/persistence/sqlcmapper"
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

type transactionalPenaltyRepo interface {
	repository.Querier
	WithinTransaction(ctx context.Context, fn func(repository.Querier) error) error
}

func NewPenaltyService(repo repository.Querier) PenaltyService {
	return &penaltyService{repo: repo}
}

func (s *penaltyService) CreatePenalty(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, penaltyTypeID uuid.UUID) (*dto.ReturnPenaltyDto, error) {
	txRepo, ok := s.repo.(transactionalPenaltyRepo)
	if !ok {
		return nil, fmt.Errorf("penalty repository does not support transactions")
	}

	var penalty repository.CreatePenaltyRow
	err := txRepo.WithinTransaction(ctx, func(txQuerier repository.Querier) error {
		createdPenalty, createErr := s.createPenaltyWithRepo(ctx, txQuerier, userID, studentID, penaltyTypeID)
		if createErr != nil {
			return createErr
		}
		penalty = createdPenalty
		return nil
	})
	if err != nil {
		return nil, err
	}

	slog.Info("penalty created", "penalty_id", penalty.ID, "user_id", userID, "student_id", studentID, "penalty_type_id", penaltyTypeID)

	return sqlcmapper.PenaltyFromCreateRow(&penalty), nil
}

func (s *penaltyService) createPenaltyWithRepo(ctx context.Context, repo repository.Querier, userID uuid.UUID, studentID uuid.UUID, penaltyTypeID uuid.UUID) (repository.CreatePenaltyRow, error) {
	if _, err := repo.GetStudentByUser(ctx, repository.GetStudentByUserParams{ID: studentID, UserID: userID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.CreatePenaltyRow{}, api.ErrStudentNotFound
		}
		return repository.CreatePenaltyRow{}, fmt.Errorf("failed to get student: %w", err)
	}

	if _, err := repo.GetPenaltyTypeByUser(ctx, repository.GetPenaltyTypeByUserParams{ID: penaltyTypeID, UserID: userID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.CreatePenaltyRow{}, api.ErrPenaltyTypeNotFound
		}
		return repository.CreatePenaltyRow{}, fmt.Errorf("failed to get penalty type: %w", err)
	}

	penalty, err := repo.CreatePenalty(ctx, repository.CreatePenaltyParams{
		UserID:        userID,
		StudentID:     studentID,
		PenaltyTypeID: penaltyTypeID,
	})
	if err != nil {
		return repository.CreatePenaltyRow{}, fmt.Errorf("failed to create penalty: %w", err)
	}

	if err := s.evaluateRulesForPenalty(ctx, repo, userID, studentID, penaltyTypeID); err != nil {
		return repository.CreatePenaltyRow{}, err
	}

	return penalty, nil
}

func (s *penaltyService) evaluateRulesForPenalty(ctx context.Context, repo repository.Querier, userID uuid.UUID, studentID uuid.UUID, penaltyTypeID uuid.UUID) error {
	rules, err := repo.ListActiveRulesByUserAndPenaltyType(ctx, repository.ListActiveRulesByUserAndPenaltyTypeParams{
		UserID:        userID,
		PenaltyTypeID: penaltyTypeID,
	})
	if err != nil {
		return fmt.Errorf("failed to list active rules: %w", err)
	}

	if len(rules) == 0 {
		return nil
	}

	penaltyCount, err := repo.CountPenaltiesByStudentAndType(ctx, repository.CountPenaltiesByStudentAndTypeParams{
		StudentID:     studentID,
		UserID:        userID,
		PenaltyTypeID: penaltyTypeID,
	})
	if err != nil {
		return fmt.Errorf("failed to count penalties for rule evaluation: %w", err)
	}

	for _, rule := range rules {
		if !shouldTriggerRule(rule.Mode, rule.Threshold, penaltyCount) {
			continue
		}

		dueAt := time.Now().UTC().Add(time.Duration(rule.DueAtAfterDays) * 24 * time.Hour)
		triggeringRuleID := rule.ID

		_, err := repo.CreatePunishmentFromRule(ctx, repository.CreatePunishmentFromRuleParams{
			UserID:           userID,
			StudentID:        studentID,
			PunishmentTypeID: rule.ResultingPunishmentTypeID,
			TriggeringRuleID: &triggeringRuleID,
			Automated:        true,
			DueAt:            dueAt,
		})
		if err != nil {
			return fmt.Errorf("failed to create punishment from rule: %w", err)
		}

		slog.Info(
			"punishment created from rule",
			"rule_id", rule.ID,
			"user_id", userID,
			"student_id", studentID,
			"penalty_type_id", penaltyTypeID,
			"penalty_count", penaltyCount,
		)
	}

	return nil
}

func shouldTriggerRule(mode string, threshold int32, count int64) bool {
	if threshold <= 0 {
		return false
	}

	thresholdAsInt64 := int64(threshold)

	switch mode {
	case "at":
		return count == thresholdAsInt64
	case "every":
		return count > 0 && count%thresholdAsInt64 == 0
	case "after":
		return count > thresholdAsInt64
	default:
		return false
	}
}

func (s *penaltyService) GetPenalty(ctx context.Context, userID uuid.UUID, penaltyID uuid.UUID) (*dto.ReturnPenaltyDto, error) {
	penalty, err := s.repo.GetPenaltyByUser(ctx, repository.GetPenaltyByUserParams{ID: penaltyID, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrPenaltyNotFound
		}
		return nil, fmt.Errorf("failed to get penalty: %w", err)
	}

	return sqlcmapper.PenaltyFromGetRow(&penalty), nil
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

	return sqlcmapper.PenaltyListFromListByUserRows(penalties), totalCount, nil
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

	return sqlcmapper.PenaltyListFromListByStudentRows(penalties), totalCount, nil
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
