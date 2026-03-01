package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/adapter/persistence/sqlcmapper"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type PunishmentTypeService interface {
	CreatePunishmentType(ctx context.Context, userID uuid.UUID, req dto.RequestPunishmentTypeDto) (*dto.ReturnPunishmentTypeDto, error)
	GetPunishmentType(ctx context.Context, userID, punishmentTypeID uuid.UUID) (*dto.ReturnPunishmentTypeDto, error)
	ListPunishmentTypes(ctx context.Context, userID uuid.UUID, search *string, limit, offset int32) ([]*dto.ReturnPunishmentTypeDto, int64, error)
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
	entity, err := s.repo.CreatePunishmentType(ctx, repository.CreatePunishmentTypeParams{
		UserID: userID,
		Name:   req.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create punishment type: %w", err)
	}

	slog.Info("punishment type created", "punishment_type_id", entity.ID, "user_id", userID)

	return sqlcmapper.PunishmentTypeFromRepository(&entity), nil
}

func (s *punishmentTypeService) GetPunishmentType(ctx context.Context, userID, punishmentTypeID uuid.UUID) (*dto.ReturnPunishmentTypeDto, error) {
	entity, err := s.repo.GetPunishmentTypeByUser(ctx, repository.GetPunishmentTypeByUserParams{
		ID:     punishmentTypeID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrPunishmentTypeNotFound
		}
		return nil, fmt.Errorf("failed to get punishment type: %w", err)
	}

	return sqlcmapper.PunishmentTypeFromRepository(&entity), nil
}

func (s *punishmentTypeService) ListPunishmentTypes(ctx context.Context, userID uuid.UUID, search *string, limit, offset int32) ([]*dto.ReturnPunishmentTypeDto, int64, error) {
	totalCount, err := s.repo.CountPunishmentTypesByUser(ctx, repository.CountPunishmentTypesByUserParams{
		UserID: userID,
		Search: search,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count punishment types: %w", err)
	}

	entities, err := s.repo.ListPunishmentTypesByUser(ctx, repository.ListPunishmentTypesByUserParams{
		UserID:      userID,
		Search:      search,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list punishment types: %w", err)
	}

	mapped := make([]*dto.ReturnPunishmentTypeDto, 0, len(entities))
	for _, entity := range entities {
		if dto := sqlcmapper.PunishmentTypeFromRepository(&entity); dto != nil {
			mapped = append(mapped, dto)
		}
	}

	return mapped, totalCount, nil
}

func (s *punishmentTypeService) UpdatePunishmentType(ctx context.Context, userID, punishmentTypeID uuid.UUID, req dto.UpdatePunishmentTypeDto) (*dto.ReturnPunishmentTypeDto, error) {
	params := repository.UpdatePunishmentTypeByUserParams{
		ID:     punishmentTypeID,
		UserID: userID,
	}
	if req.Name != nil {
		params.Name = req.Name
	}

	entity, err := s.repo.UpdatePunishmentTypeByUser(ctx, params)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrPunishmentTypeNotFound
		}
		return nil, fmt.Errorf("failed to update punishment type: %w", err)
	}

	return sqlcmapper.PunishmentTypeFromRepository(&entity), nil
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
