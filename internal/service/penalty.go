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

type PenaltyService interface {
	CreatePenalty(
		ctx context.Context,
		userID uuid.UUID,
		studentID uuid.UUID,
		penaltyTypeID uuid.UUID,
		classroomID *uuid.UUID,
		occurredAt *time.Time,
		evaluationLabel *string,
	) (*dto.ReturnPenaltyDto, error)
	CreatePenaltiesInClassroom(
		ctx context.Context,
		userID, classroomID uuid.UUID,
		studentIDs []uuid.UUID,
		penaltyTypeID uuid.UUID,
		occurredAt *time.Time,
		evaluationLabel *string,
	) ([]*dto.ReturnPenaltyDto, error)
	GetPenalty(ctx context.Context, userID uuid.UUID, penaltyID uuid.UUID) (*dto.ReturnPenaltyDto, error)
	ListPenalties(ctx context.Context, userID uuid.UUID, filters ListPenaltiesFilters) ([]*dto.ReturnPenaltyDto, int64, error)
	ListPenaltiesByStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, limit, offset int32) ([]*dto.ReturnPenaltyDto, int64, error)
	UpdatePenalty(
		ctx context.Context,
		userID uuid.UUID,
		penaltyID uuid.UUID,
		occurredAt *time.Time,
		evaluationLabel *string,
	) (*dto.ReturnPenaltyDto, error)
	DeletePenalty(ctx context.Context, userID uuid.UUID, penaltyID uuid.UUID) error
}

type penaltyService struct {
	repo repository.Querier
	now  func() time.Time
}

type transactionalPenaltyRepo interface {
	repository.Querier
	WithinTransaction(ctx context.Context, fn func(repository.Querier) error) error
}

func NewPenaltyService(repo repository.Querier) PenaltyService {
	return &penaltyService{
		repo: repo,
		now:  time.Now,
	}
}

func (s *penaltyService) CreatePenalty(
	ctx context.Context,
	userID uuid.UUID,
	studentID uuid.UUID,
	penaltyTypeID uuid.UUID,
	classroomID *uuid.UUID,
	occurredAt *time.Time,
	evaluationLabel *string,
) (*dto.ReturnPenaltyDto, error) {
	txRepo, ok := s.repo.(transactionalPenaltyRepo)
	if !ok {
		return nil, fmt.Errorf("penalty repository does not support transactions")
	}

	var penalty repository.CreatePenaltyRow
	err := txRepo.WithinTransaction(ctx, func(txQuerier repository.Querier) error {
		createdPenalty, createErr := s.createPenaltyWithRepo(ctx, txQuerier, userID, studentID, penaltyTypeID, classroomID, occurredAt, evaluationLabel)
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

func (s *penaltyService) CreatePenaltiesInClassroom(
	ctx context.Context,
	userID, classroomID uuid.UUID,
	studentIDs []uuid.UUID,
	penaltyTypeID uuid.UUID,
	occurredAt *time.Time,
	evaluationLabel *string,
) ([]*dto.ReturnPenaltyDto, error) {
	txRepo, ok := s.repo.(transactionalPenaltyRepo)
	if !ok {
		return nil, fmt.Errorf("penalty repository does not support transactions")
	}

	createdPenalties := make([]*dto.ReturnPenaltyDto, 0, len(studentIDs))
	err := txRepo.WithinTransaction(ctx, func(txQuerier repository.Querier) error {
		if err := ensureClassroomExists(ctx, txQuerier, userID, classroomID); err != nil {
			return err
		}

		createdPenalties = make([]*dto.ReturnPenaltyDto, 0, len(studentIDs))
		for _, studentID := range studentIDs {
			penalty, err := s.createPenaltyWithRepo(
				ctx,
				txQuerier,
				userID,
				studentID,
				penaltyTypeID,
				&classroomID,
				occurredAt,
				evaluationLabel,
			)
			if err != nil {
				return err
			}

			createdPenalties = append(createdPenalties, sqlcmapper.PenaltyFromCreateRow(&penalty))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	slog.Info(
		"penalties created in classroom",
		"classroom_id", classroomID,
		"student_count", len(createdPenalties),
		"user_id", userID,
		"penalty_type_id", penaltyTypeID,
	)

	return createdPenalties, nil
}

func (s *penaltyService) createPenaltyWithRepo(
	ctx context.Context,
	repo repository.Querier,
	userID uuid.UUID,
	studentID uuid.UUID,
	penaltyTypeID uuid.UUID,
	classroomID *uuid.UUID,
	occurredAt *time.Time,
	evaluationLabel *string,
) (repository.CreatePenaltyRow, error) {
	if _, err := repo.GetStudentByUser(ctx, repository.GetStudentByUserParams{ID: studentID, UserID: userID}); err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return repository.CreatePenaltyRow{}, api.ErrStudentNotFound
		}
		return repository.CreatePenaltyRow{}, fmt.Errorf("failed to get student: %w", err)
	}

	if _, err := repo.GetPenaltyTypeByUser(ctx, repository.GetPenaltyTypeByUserParams{ID: penaltyTypeID, UserID: userID}); err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return repository.CreatePenaltyRow{}, api.ErrPenaltyTypeNotFound
		}
		return repository.CreatePenaltyRow{}, fmt.Errorf("failed to get penalty type: %w", err)
	}

	if classroomID != nil {
		if _, err := resolvePunishmentClassroomID(ctx, repo, userID, studentID, classroomID); err != nil {
			if errors.Is(err, api.ErrClassroomNotFound) ||
				errors.Is(err, api.ErrPunishmentClassroomNotResolved) ||
				errors.Is(err, api.ErrPunishmentStudentNotInClassroom) {
				return repository.CreatePenaltyRow{}, err
			}
			return repository.CreatePenaltyRow{}, fmt.Errorf("failed to validate punishment classroom: %w", err)
		}
	}

	penalty, err := repo.CreatePenalty(ctx, repository.CreatePenaltyParams{
		UserID:          userID,
		StudentID:       studentID,
		PenaltyTypeID:   penaltyTypeID,
		OccurredAt:      occurredAt,
		EvaluationLabel: evaluationLabel,
	})
	if err != nil {
		return repository.CreatePenaltyRow{}, fmt.Errorf("failed to create penalty: %w", err)
	}

	if err := s.evaluateRulesForPenalty(ctx, repo, userID, studentID, penaltyTypeID, classroomID); err != nil {
		return repository.CreatePenaltyRow{}, err
	}

	return penalty, nil
}

func (s *penaltyService) evaluateRulesForPenalty(
	ctx context.Context,
	repo repository.Querier,
	userID uuid.UUID,
	studentID uuid.UUID,
	penaltyTypeID uuid.UUID,
	requestedClassroomID *uuid.UUID,
) error {
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

	referenceNow := s.now()
	location, err := resolveUserLocation(ctx, repo, userID)
	if err != nil {
		return err
	}
	var resolvedClassroomID *uuid.UUID
	classroomResolved := false
	for _, rule := range rules {
		if !shouldTriggerRule(rule.Mode, rule.Threshold, penaltyCount) {
			continue
		}

		if rule.DueAtMode == ruleDueAtModeNextLessons && !classroomResolved {
			resolvedClassroomID, err = resolvePunishmentClassroomID(ctx, repo, userID, studentID, requestedClassroomID)
			if err != nil {
				if errors.Is(err, api.ErrClassroomNotFound) ||
					errors.Is(err, api.ErrPunishmentClassroomNotResolved) ||
					errors.Is(err, api.ErrPunishmentStudentNotInClassroom) {
					return err
				}
				return fmt.Errorf("failed to resolve punishment classroom: %w", err)
			}
			classroomResolved = true
		}

		dueAt, err := computeRuleDueAt(ctx, repo, userID, rule, resolvedClassroomID, referenceNow, location)
		if err != nil {
			if errors.Is(err, api.ErrRuleDueAtNotComputable) || errors.Is(err, api.ErrPunishmentClassroomNotResolved) {
				return err
			}
			return fmt.Errorf("failed to compute rule due_at: %w", err)
		}
		triggeringRuleID := rule.ID

		_, err = repo.CreatePunishmentFromRule(ctx, repository.CreatePunishmentFromRuleParams{
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
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrPenaltyNotFound
		}
		return nil, fmt.Errorf("failed to get penalty: %w", err)
	}

	return sqlcmapper.PenaltyFromGetRow(&penalty), nil
}

func (s *penaltyService) ListPenalties(ctx context.Context, userID uuid.UUID, filters ListPenaltiesFilters) ([]*dto.ReturnPenaltyDto, int64, error) {
	createdFrom := filters.CreatedFrom
	createdTo := filters.CreatedTo
	if filters.CreatedFrom != nil || filters.CreatedTo != nil {
		location, err := resolveUserLocation(ctx, s.repo, userID)
		if err != nil {
			return nil, 0, err
		}
		createdFrom, createdTo = localDateBoundsToUTC(filters.CreatedFrom, filters.CreatedTo, location)
	}

	totalCount, err := s.repo.CountPenaltiesByUser(ctx, repository.CountPenaltiesByUserParams{
		UserID:        userID,
		StudentID:     filters.StudentID,
		PenaltyTypeID: filters.PenaltyTypeID,
		CreatedFrom:   createdFrom,
		CreatedTo:     createdTo,
		ClassroomID:   filters.ClassroomID,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count penalties: %w", err)
	}

	penalties, err := s.repo.ListPenaltiesByUser(ctx, repository.ListPenaltiesByUserParams{
		UserID:        userID,
		StudentID:     filters.StudentID,
		PenaltyTypeID: filters.PenaltyTypeID,
		CreatedFrom:   createdFrom,
		CreatedTo:     createdTo,
		ClassroomID:   filters.ClassroomID,
		QueryOffset:   filters.Offset,
		QueryLimit:    filters.Limit,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list penalties: %w", err)
	}

	return sqlcmapper.PenaltyListFromListByUserRows(penalties), totalCount, nil
}

func (s *penaltyService) ListPenaltiesByStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, limit, offset int32) ([]*dto.ReturnPenaltyDto, int64, error) {
	if _, err := s.repo.GetStudentByUser(ctx, repository.GetStudentByUserParams{ID: studentID, UserID: userID}); err != nil {
		if errors.Is(err, repository.ErrNoRows) {
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

func (s *penaltyService) UpdatePenalty(
	ctx context.Context,
	userID uuid.UUID,
	penaltyID uuid.UUID,
	occurredAt *time.Time,
	evaluationLabel *string,
) (*dto.ReturnPenaltyDto, error) {
	penalty, err := s.repo.UpdatePenaltyByUser(ctx, repository.UpdatePenaltyByUserParams{
		OccurredAt:      occurredAt,
		EvaluationLabel: evaluationLabel,
		ID:              penaltyID,
		UserID:          userID,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrPenaltyNotFound
		}
		return nil, fmt.Errorf("failed to update penalty: %w", err)
	}

	slog.Info("penalty updated", "penalty_id", penalty.ID, "user_id", userID)

	return sqlcmapper.PenaltyFromUpdateRow(&penalty), nil
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
