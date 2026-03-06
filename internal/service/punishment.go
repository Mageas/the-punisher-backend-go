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

type PunishmentService interface {
	CreatePunishment(
		ctx context.Context,
		userID uuid.UUID,
		studentID uuid.UUID,
		punishmentTypeID uuid.UUID,
		dueAt time.Time,
		occurredAt *time.Time,
		evaluationLabel *string,
	) (*dto.ReturnPunishmentDto, error)
	GetPunishment(ctx context.Context, userID uuid.UUID, punishmentID uuid.UUID) (*dto.ReturnPunishmentDto, error)
	ListPunishments(ctx context.Context, userID uuid.UUID, filters ListPunishmentsFilters) ([]*dto.ReturnPunishmentDto, int64, error)
	ListPunishmentsByStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, resolved *bool, limit, offset int32) ([]*dto.ReturnPunishmentDto, int64, error)
	UpdatePunishment(
		ctx context.Context,
		userID uuid.UUID,
		punishmentID uuid.UUID,
		occurredAt *time.Time,
		evaluationLabel *string,
	) (*dto.ReturnPunishmentDto, error)
	ResolvePunishment(ctx context.Context, userID uuid.UUID, punishmentID uuid.UUID) (*dto.ReturnPunishmentDto, error)
	DeletePunishment(ctx context.Context, userID uuid.UUID, punishmentID uuid.UUID) error
}

type punishmentService struct {
	repo repository.Querier
}

func NewPunishmentService(repo repository.Querier) PunishmentService {
	return &punishmentService{repo: repo}
}

func (s *punishmentService) CreatePunishment(
	ctx context.Context,
	userID uuid.UUID,
	studentID uuid.UUID,
	punishmentTypeID uuid.UUID,
	dueAt time.Time,
	occurredAt *time.Time,
	evaluationLabel *string,
) (*dto.ReturnPunishmentDto, error) {
	if _, err := s.repo.GetStudentByUser(ctx, repository.GetStudentByUserParams{ID: studentID, UserID: userID}); err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrStudentNotFound
		}
		return nil, fmt.Errorf("failed to get student: %w", err)
	}

	if _, err := s.repo.GetPunishmentTypeByUser(ctx, repository.GetPunishmentTypeByUserParams{ID: punishmentTypeID, UserID: userID}); err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrPunishmentTypeNotFound
		}
		return nil, fmt.Errorf("failed to get punishment type: %w", err)
	}

	punishment, err := s.repo.CreatePunishment(ctx, repository.CreatePunishmentParams{
		UserID:           userID,
		StudentID:        studentID,
		PunishmentTypeID: punishmentTypeID,
		DueAt:            dueAt,
		OccurredAt:       occurredAt,
		EvaluationLabel:  evaluationLabel,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create punishment: %w", err)
	}

	slog.Info("punishment created", "punishment_id", punishment.ID, "user_id", userID, "student_id", studentID, "punishment_type_id", punishmentTypeID)

	return sqlcmapper.PunishmentFromCreateRow(&punishment), nil
}

func (s *punishmentService) GetPunishment(ctx context.Context, userID uuid.UUID, punishmentID uuid.UUID) (*dto.ReturnPunishmentDto, error) {
	punishment, err := s.repo.GetPunishmentByUser(ctx, repository.GetPunishmentByUserParams{ID: punishmentID, UserID: userID})
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrPunishmentNotFound
		}
		return nil, fmt.Errorf("failed to get punishment: %w", err)
	}

	return sqlcmapper.PunishmentFromGetRow(&punishment), nil
}

func (s *punishmentService) ListPunishments(ctx context.Context, userID uuid.UUID, filters ListPunishmentsFilters) ([]*dto.ReturnPunishmentDto, int64, error) {
	var resolved *bool
	if filters.State != nil {
		resolvedValue := filters.State.Resolved()
		resolved = &resolvedValue
	}

	totalCount, err := s.repo.CountPunishmentsByUser(ctx, repository.CountPunishmentsByUserParams{
		UserID:           userID,
		StudentID:        filters.StudentID,
		PunishmentTypeID: filters.PunishmentTypeID,
		Resolved:         resolved,
		Automated:        filters.Automated,
		Overdue:          filters.Overdue,
		CreatedFrom:      filters.CreatedFrom,
		CreatedTo:        filters.CreatedTo,
		DueFrom:          filters.DueFrom,
		DueTo:            filters.DueTo,
		ClassroomID:      filters.ClassroomID,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count punishments: %w", err)
	}

	punishments, err := s.repo.ListPunishmentsByUser(ctx, repository.ListPunishmentsByUserParams{
		UserID:           userID,
		StudentID:        filters.StudentID,
		PunishmentTypeID: filters.PunishmentTypeID,
		Resolved:         resolved,
		Automated:        filters.Automated,
		Overdue:          filters.Overdue,
		CreatedFrom:      filters.CreatedFrom,
		CreatedTo:        filters.CreatedTo,
		DueFrom:          filters.DueFrom,
		DueTo:            filters.DueTo,
		ClassroomID:      filters.ClassroomID,
		QueryOffset:      filters.Offset,
		QueryLimit:       filters.Limit,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list punishments: %w", err)
	}

	return sqlcmapper.PunishmentListFromListByUserRows(punishments), totalCount, nil
}

func (s *punishmentService) ListPunishmentsByStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, resolved *bool, limit, offset int32) ([]*dto.ReturnPunishmentDto, int64, error) {
	if _, err := s.repo.GetStudentByUser(ctx, repository.GetStudentByUserParams{ID: studentID, UserID: userID}); err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, 0, api.ErrStudentNotFound
		}
		return nil, 0, fmt.Errorf("failed to get student: %w", err)
	}

	totalCount, err := s.repo.CountPunishmentsByStudent(ctx, repository.CountPunishmentsByStudentParams{
		StudentID: studentID,
		UserID:    userID,
		Resolved:  resolved,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count punishments by student: %w", err)
	}

	punishments, err := s.repo.ListPunishmentsByStudent(ctx, repository.ListPunishmentsByStudentParams{
		StudentID:   studentID,
		UserID:      userID,
		Resolved:    resolved,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list punishments by student: %w", err)
	}

	return sqlcmapper.PunishmentListFromListByStudentRows(punishments), totalCount, nil
}

func (s *punishmentService) UpdatePunishment(
	ctx context.Context,
	userID uuid.UUID,
	punishmentID uuid.UUID,
	occurredAt *time.Time,
	evaluationLabel *string,
) (*dto.ReturnPunishmentDto, error) {
	punishment, err := s.repo.UpdatePunishmentByUser(ctx, repository.UpdatePunishmentByUserParams{
		OccurredAt:      occurredAt,
		EvaluationLabel: evaluationLabel,
		ID:              punishmentID,
		UserID:          userID,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrPunishmentNotFound
		}
		return nil, fmt.Errorf("failed to update punishment: %w", err)
	}

	slog.Info("punishment updated", "punishment_id", punishment.ID, "user_id", userID)

	return sqlcmapper.PunishmentFromUpdateRow(&punishment), nil
}

func (s *punishmentService) ResolvePunishment(ctx context.Context, userID uuid.UUID, punishmentID uuid.UUID) (*dto.ReturnPunishmentDto, error) {
	punishment, err := s.repo.ResolvePunishment(ctx, repository.ResolvePunishmentParams{ID: punishmentID, UserID: userID})
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			existing, getErr := s.repo.GetPunishmentByUser(ctx, repository.GetPunishmentByUserParams{ID: punishmentID, UserID: userID})
			if getErr != nil {
				if errors.Is(getErr, repository.ErrNoRows) {
					return nil, api.ErrPunishmentNotFound
				}
				return nil, fmt.Errorf("failed to get punishment: %w", getErr)
			}
			if existing.ResolvedAt != nil {
				return nil, api.ErrPunishmentAlreadyResolved
			}
			return nil, fmt.Errorf("failed to resolve punishment: %w", err)
		}
		return nil, fmt.Errorf("failed to resolve punishment: %w", err)
	}

	slog.Info("punishment resolved", "punishment_id", punishment.ID, "user_id", userID)

	return sqlcmapper.PunishmentFromResolveRow(&punishment), nil
}

func (s *punishmentService) DeletePunishment(ctx context.Context, userID uuid.UUID, punishmentID uuid.UUID) error {
	rowsAffected, err := s.repo.DeletePunishmentByUser(ctx, repository.DeletePunishmentByUserParams{ID: punishmentID, UserID: userID})
	if err != nil {
		return fmt.Errorf("failed to delete punishment: %w", err)
	}

	if rowsAffected == 0 {
		return api.ErrPunishmentNotFound
	}

	slog.Info("punishment deleted", "punishment_id", punishmentID, "user_id", userID)

	return nil
}
