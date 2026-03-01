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

type BonusTypeService interface {
	CreateBonusType(ctx context.Context, userID uuid.UUID, req dto.RequestBonusTypeDto) (*dto.ReturnBonusTypeDto, error)
	GetBonusType(ctx context.Context, userID, bonusTypeID uuid.UUID) (*dto.ReturnBonusTypeDto, error)
	ListBonusTypes(ctx context.Context, userID uuid.UUID, search *string, limit, offset int32) ([]*dto.ReturnBonusTypeDto, int64, error)
	UpdateBonusType(ctx context.Context, userID, bonusTypeID uuid.UUID, req dto.UpdateBonusTypeDto) (*dto.ReturnBonusTypeDto, error)
	DeleteBonusType(ctx context.Context, userID, bonusTypeID uuid.UUID) error
}

type bonusTypeService struct {
	repo repository.Querier
}

func NewBonusTypeService(repo repository.Querier) BonusTypeService {
	return &bonusTypeService{
		repo: repo,
	}
}

func (s *bonusTypeService) CreateBonusType(ctx context.Context, userID uuid.UUID, req dto.RequestBonusTypeDto) (*dto.ReturnBonusTypeDto, error) {
	entity, err := s.repo.CreateBonusType(ctx, repository.CreateBonusTypeParams{
		UserID: userID,
		Name:   req.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create bonus type: %w", err)
	}

	slog.Info("bonus type created", "bonus_type_id", entity.ID, "user_id", userID)

	return sqlcmapper.BonusTypeFromRepository(&entity), nil
}

func (s *bonusTypeService) GetBonusType(ctx context.Context, userID, bonusTypeID uuid.UUID) (*dto.ReturnBonusTypeDto, error) {
	entity, err := s.repo.GetBonusTypeByUser(ctx, repository.GetBonusTypeByUserParams{
		ID:     bonusTypeID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrBonusTypeNotFound
		}
		return nil, fmt.Errorf("failed to get bonus type: %w", err)
	}

	return sqlcmapper.BonusTypeFromRepository(&entity), nil
}

func (s *bonusTypeService) ListBonusTypes(ctx context.Context, userID uuid.UUID, search *string, limit, offset int32) ([]*dto.ReturnBonusTypeDto, int64, error) {
	totalCount, err := s.repo.CountBonusTypesByUser(ctx, repository.CountBonusTypesByUserParams{
		UserID: userID,
		Search: search,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count bonus types: %w", err)
	}

	entities, err := s.repo.ListBonusTypesByUser(ctx, repository.ListBonusTypesByUserParams{
		UserID:      userID,
		Search:      search,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list bonus types: %w", err)
	}

	mapped := make([]*dto.ReturnBonusTypeDto, 0, len(entities))
	for _, entity := range entities {
		if dto := sqlcmapper.BonusTypeFromRepository(&entity); dto != nil {
			mapped = append(mapped, dto)
		}
	}

	return mapped, totalCount, nil
}

func (s *bonusTypeService) UpdateBonusType(ctx context.Context, userID, bonusTypeID uuid.UUID, req dto.UpdateBonusTypeDto) (*dto.ReturnBonusTypeDto, error) {
	params := repository.UpdateBonusTypeByUserParams{
		ID:     bonusTypeID,
		UserID: userID,
	}
	if req.Name != nil {
		params.Name = req.Name
	}

	entity, err := s.repo.UpdateBonusTypeByUser(ctx, params)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrBonusTypeNotFound
		}
		return nil, fmt.Errorf("failed to update bonus type: %w", err)
	}

	return sqlcmapper.BonusTypeFromRepository(&entity), nil
}

func (s *bonusTypeService) DeleteBonusType(ctx context.Context, userID, bonusTypeID uuid.UUID) error {
	rowsAffected, err := s.repo.DeleteBonusTypeByUser(ctx, repository.DeleteBonusTypeByUserParams{
		ID:     bonusTypeID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete bonus type: %w", err)
	}

	if rowsAffected == 0 {
		return api.ErrBonusTypeNotFound
	}

	slog.Info("bonus type deleted", "bonus_type_id", bonusTypeID, "user_id", userID)

	return nil
}
