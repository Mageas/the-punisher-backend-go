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

type PenaltyTypeService interface {
	CreatePenaltyType(ctx context.Context, userID uuid.UUID, req dto.RequestPenaltyTypeDto) (*dto.ReturnPenaltyTypeDto, error)
	GetPenaltyType(ctx context.Context, userID, penaltyTypeID uuid.UUID) (*dto.ReturnPenaltyTypeDto, error)
	ListPenaltyTypes(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*dto.ReturnPenaltyTypeDto, int64, error)
	UpdatePenaltyType(ctx context.Context, userID, penaltyTypeID uuid.UUID, req dto.UpdatePenaltyTypeDto) (*dto.ReturnPenaltyTypeDto, error)
	DeletePenaltyType(ctx context.Context, userID, penaltyTypeID uuid.UUID) error
}

type penaltyTypeService struct {
	repo repository.Querier
}

func NewPenaltyTypeService(repo repository.Querier) PenaltyTypeService {
	return &penaltyTypeService{
		repo: repo,
	}
}

func (s *penaltyTypeService) CreatePenaltyType(ctx context.Context, userID uuid.UUID, req dto.RequestPenaltyTypeDto) (*dto.ReturnPenaltyTypeDto, error) {
	entity, err := s.repo.CreatePenaltyType(ctx, repository.CreatePenaltyTypeParams{
		UserID: userID,
		Name:   req.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create penalty type: %w", err)
	}

	slog.Info("penalty type created", "penalty_type_id", entity.ID, "user_id", userID)

	return sqlcmapper.PenaltyTypeFromRepository(&entity), nil
}

func (s *penaltyTypeService) GetPenaltyType(ctx context.Context, userID, penaltyTypeID uuid.UUID) (*dto.ReturnPenaltyTypeDto, error) {
	entity, err := s.repo.GetPenaltyTypeByUser(ctx, repository.GetPenaltyTypeByUserParams{
		ID:     penaltyTypeID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrPenaltyTypeNotFound
		}
		return nil, fmt.Errorf("failed to get penalty type: %w", err)
	}

	return sqlcmapper.PenaltyTypeFromRepository(&entity), nil
}

func (s *penaltyTypeService) ListPenaltyTypes(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*dto.ReturnPenaltyTypeDto, int64, error) {
	totalCount, err := s.repo.CountPenaltyTypesByUser(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count penalty types: %w", err)
	}

	entities, err := s.repo.ListPenaltyTypesByUser(ctx, repository.ListPenaltyTypesByUserParams{
		UserID:      userID,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list penalty types: %w", err)
	}

	mapped := make([]*dto.ReturnPenaltyTypeDto, 0, len(entities))
	for _, entity := range entities {
		if dto := sqlcmapper.PenaltyTypeFromRepository(&entity); dto != nil {
			mapped = append(mapped, dto)
		}
	}

	return mapped, totalCount, nil
}

func (s *penaltyTypeService) UpdatePenaltyType(ctx context.Context, userID, penaltyTypeID uuid.UUID, req dto.UpdatePenaltyTypeDto) (*dto.ReturnPenaltyTypeDto, error) {
	params := repository.UpdatePenaltyTypeByUserParams{
		ID:     penaltyTypeID,
		UserID: userID,
	}
	if req.Name != nil {
		params.Name = req.Name
	}

	entity, err := s.repo.UpdatePenaltyTypeByUser(ctx, params)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrPenaltyTypeNotFound
		}
		return nil, fmt.Errorf("failed to update penalty type: %w", err)
	}

	return sqlcmapper.PenaltyTypeFromRepository(&entity), nil
}

func (s *penaltyTypeService) DeletePenaltyType(ctx context.Context, userID, penaltyTypeID uuid.UUID) error {
	rowsAffected, err := s.repo.DeletePenaltyTypeByUser(ctx, repository.DeletePenaltyTypeByUserParams{
		ID:     penaltyTypeID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete penalty type: %w", err)
	}

	if rowsAffected == 0 {
		return api.ErrPenaltyTypeNotFound
	}

	slog.Info("penalty type deleted", "penalty_type_id", penaltyTypeID, "user_id", userID)

	return nil
}
