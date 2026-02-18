package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type PunishmentService interface {
	CreatePunishment(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, punishmentTypeID uuid.UUID, dueAt time.Time) (*dto.ReturnPunishmentDto, error)
	GetPunishment(ctx context.Context, userID uuid.UUID, punishmentID uuid.UUID) (*dto.ReturnPunishmentDto, error)
	ListPunishments(ctx context.Context, userID uuid.UUID, resolved *bool, search *string, limit, offset int32) ([]*dto.ReturnPunishmentDto, int64, error)
	ListPunishmentsByStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, resolved *bool, limit, offset int32) ([]*dto.ReturnPunishmentDto, int64, error)
	ResolvePunishment(ctx context.Context, userID uuid.UUID, punishmentID uuid.UUID) (*dto.ReturnPunishmentDto, error)
	DeletePunishment(ctx context.Context, userID uuid.UUID, punishmentID uuid.UUID) error
}

type punishmentService struct {
	repo repository.Querier
}

func NewPunishmentService(repo repository.Querier) PunishmentService {
	return &punishmentService{repo: repo}
}

func (s *punishmentService) CreatePunishment(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, punishmentTypeID uuid.UUID, dueAt time.Time) (*dto.ReturnPunishmentDto, error) {
	if _, err := s.repo.GetStudentByUser(ctx, repository.GetStudentByUserParams{ID: studentID, UserID: userID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrStudentNotFound
		}
		return nil, fmt.Errorf("failed to get student: %w", err)
	}

	if _, err := s.repo.GetPunishmentTypeByUser(ctx, repository.GetPunishmentTypeByUserParams{ID: punishmentTypeID, UserID: userID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrPunishmentTypeNotFound
		}
		return nil, fmt.Errorf("failed to get punishment type: %w", err)
	}

	punishment, err := s.repo.CreatePunishment(ctx, repository.CreatePunishmentParams{
		UserID:           userID,
		StudentID:        studentID,
		PunishmentTypeID: punishmentTypeID,
		DueAt:            dueAt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create punishment: %w", err)
	}

	slog.Info("punishment created", "punishment_id", punishment.ID, "user_id", userID, "student_id", studentID, "punishment_type_id", punishmentTypeID)

	return dto.PunishmentFromCreateRow(&punishment), nil
}

func (s *punishmentService) GetPunishment(ctx context.Context, userID uuid.UUID, punishmentID uuid.UUID) (*dto.ReturnPunishmentDto, error) {
	punishment, err := s.repo.GetPunishmentByUser(ctx, repository.GetPunishmentByUserParams{ID: punishmentID, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrPunishmentNotFound
		}
		return nil, fmt.Errorf("failed to get punishment: %w", err)
	}

	return dto.PunishmentFromGetRow(&punishment), nil
}

func (s *punishmentService) ListPunishments(ctx context.Context, userID uuid.UUID, resolved *bool, search *string, limit, offset int32) ([]*dto.ReturnPunishmentDto, int64, error) {
	resolvedParam := pgtype.Bool{}
	if resolved != nil {
		resolvedParam = pgtype.Bool{Bool: *resolved, Valid: true}
	}
	searchParam := pgtype.Text{}
	if search != nil {
		searchParam = pgtype.Text{String: *search, Valid: true}
	}

	totalCount, err := s.repo.CountPunishmentsByUser(ctx, repository.CountPunishmentsByUserParams{
		UserID:   userID,
		Resolved: resolvedParam,
		Search:   searchParam,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count punishments: %w", err)
	}

	punishments, err := s.repo.ListPunishmentsByUser(ctx, repository.ListPunishmentsByUserParams{
		UserID:      userID,
		Resolved:    resolvedParam,
		Search:      searchParam,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list punishments: %w", err)
	}

	return dto.PunishmentListFromListByUserRows(punishments), totalCount, nil
}

func (s *punishmentService) ListPunishmentsByStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, resolved *bool, limit, offset int32) ([]*dto.ReturnPunishmentDto, int64, error) {
	if _, err := s.repo.GetStudentByUser(ctx, repository.GetStudentByUserParams{ID: studentID, UserID: userID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, api.ErrStudentNotFound
		}
		return nil, 0, fmt.Errorf("failed to get student: %w", err)
	}

	resolvedParam := pgtype.Bool{}
	if resolved != nil {
		resolvedParam = pgtype.Bool{Bool: *resolved, Valid: true}
	}

	totalCount, err := s.repo.CountPunishmentsByStudent(ctx, repository.CountPunishmentsByStudentParams{
		StudentID: studentID,
		UserID:    userID,
		Resolved:  resolvedParam,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count punishments by student: %w", err)
	}

	punishments, err := s.repo.ListPunishmentsByStudent(ctx, repository.ListPunishmentsByStudentParams{
		StudentID:   studentID,
		UserID:      userID,
		Resolved:    resolvedParam,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list punishments by student: %w", err)
	}

	return dto.PunishmentListFromListByStudentRows(punishments), totalCount, nil
}

func (s *punishmentService) ResolvePunishment(ctx context.Context, userID uuid.UUID, punishmentID uuid.UUID) (*dto.ReturnPunishmentDto, error) {
	punishment, err := s.repo.ResolvePunishment(ctx, repository.ResolvePunishmentParams{ID: punishmentID, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			existing, getErr := s.repo.GetPunishmentByUser(ctx, repository.GetPunishmentByUserParams{ID: punishmentID, UserID: userID})
			if getErr != nil {
				if errors.Is(getErr, pgx.ErrNoRows) {
					return nil, api.ErrPunishmentNotFound
				}
				return nil, fmt.Errorf("failed to get punishment: %w", getErr)
			}
			if existing.ResolvedAt.Valid {
				return nil, api.ErrPunishmentAlreadyResolved
			}
			return nil, fmt.Errorf("failed to resolve punishment: %w", err)
		}
		return nil, fmt.Errorf("failed to resolve punishment: %w", err)
	}

	slog.Info("punishment resolved", "punishment_id", punishment.ID, "user_id", userID)

	return dto.PunishmentFromResolveRow(&punishment), nil
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
