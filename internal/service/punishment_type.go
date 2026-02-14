package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type PunishmentTypeService interface {
	CreatePunishmentType(ctx context.Context, userID uuid.UUID, req dto.RequestPunishmentTypeDto) (*dto.ReturnPunishmentTypeDto, error)
	GetPunishmentType(ctx context.Context, userID, punishmentTypeID uuid.UUID) (*dto.ReturnPunishmentTypeDto, error)
	ListPunishmentTypes(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*dto.ReturnPunishmentTypeDto, int64, error)
	UpdatePunishmentType(ctx context.Context, userID, punishmentTypeID uuid.UUID, req dto.UpdatePunishmentTypeDto) (*dto.ReturnPunishmentTypeDto, error)
	DeletePunishmentType(ctx context.Context, userID, punishmentTypeID uuid.UUID) error
}

type punishmentTypeService struct {
	repo repository.Querier
}

func NewPunishmentTypeService(repo repository.Querier) PunishmentTypeService {
	return &punishmentTypeService{
		repo: repo,
	}
}

func (s *punishmentTypeService) CreatePunishmentType(ctx context.Context, userID uuid.UUID, req dto.RequestPunishmentTypeDto) (*dto.ReturnPunishmentTypeDto, error) {
	pt, err := s.repo.CreatePunishmentType(ctx, repository.CreatePunishmentTypeParams{
		UserID: userID,
		Name:   req.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create punishment type: %w", err)
	}

	slog.Info("punishment type created", "punishment_type_id", pt.ID, "user_id", userID)

	return dto.PunishmentTypeFromRepository(&pt), nil
}

func (s *punishmentTypeService) GetPunishmentType(ctx context.Context, userID, punishmentTypeID uuid.UUID) (*dto.ReturnPunishmentTypeDto, error) {
	pt, err := s.repo.GetPunishmentTypeByUser(ctx, repository.GetPunishmentTypeByUserParams{
		ID:     punishmentTypeID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrPunishmentTypeNotFound
		}
		return nil, fmt.Errorf("failed to get punishment type: %w", err)
	}

	return dto.PunishmentTypeFromRepository(&pt), nil
}

func (s *punishmentTypeService) ListPunishmentTypes(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*dto.ReturnPunishmentTypeDto, int64, error) {
	totalCount, err := s.repo.CountPunishmentTypesByUser(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count punishment types: %w", err)
	}

	pts, err := s.repo.ListPunishmentTypesByUser(ctx, repository.ListPunishmentTypesByUserParams{
		UserID:      userID,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list punishment types: %w", err)
	}

	return dto.PunishmentTypeListFromRepository(pts), totalCount, nil
}

func (s *punishmentTypeService) UpdatePunishmentType(ctx context.Context, userID, punishmentTypeID uuid.UUID, req dto.UpdatePunishmentTypeDto) (*dto.ReturnPunishmentTypeDto, error) {
	arg := repository.UpdatePunishmentTypeByUserParams{
		ID:     punishmentTypeID,
		UserID: userID,
	}

	if req.Name != nil {
		arg.Name = pgtype.Text{String: *req.Name, Valid: true}
	}

	pt, err := s.repo.UpdatePunishmentTypeByUser(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrPunishmentTypeNotFound
		}
		return nil, fmt.Errorf("failed to update punishment type: %w", err)
	}

	return dto.PunishmentTypeFromRepository(&pt), nil
}

func (s *punishmentTypeService) DeletePunishmentType(ctx context.Context, userID, punishmentTypeID uuid.UUID) error {
	rowsAffected, err := s.repo.DeletePunishmentTypeByUser(ctx, repository.DeletePunishmentTypeByUserParams{
		ID:     punishmentTypeID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete punishment type: %w", err)
	}

	if rowsAffected == 0 {
		return api.ErrPunishmentTypeNotFound
	}

	slog.Info("punishment type deleted", "punishment_type_id", punishmentTypeID, "user_id", userID)

	return nil
}
