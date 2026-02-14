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
	pt, err := s.repo.CreatePenaltyType(ctx, repository.CreatePenaltyTypeParams{
		UserID: userID,
		Name:   req.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create penalty type: %w", err)
	}

	slog.Info("penalty type created", "penalty_type_id", pt.ID, "user_id", userID)

	return dto.PenaltyTypeFromRepository(&pt), nil
}

func (s *penaltyTypeService) GetPenaltyType(ctx context.Context, userID, penaltyTypeID uuid.UUID) (*dto.ReturnPenaltyTypeDto, error) {
	pt, err := s.repo.GetPenaltyTypeByUser(ctx, repository.GetPenaltyTypeByUserParams{
		ID:     penaltyTypeID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrPenaltyTypeNotFound
		}
		return nil, fmt.Errorf("failed to get penalty type: %w", err)
	}

	return dto.PenaltyTypeFromRepository(&pt), nil
}

func (s *penaltyTypeService) ListPenaltyTypes(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*dto.ReturnPenaltyTypeDto, int64, error) {
	totalCount, err := s.repo.CountPenaltyTypesByUser(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count penalty types: %w", err)
	}

	pts, err := s.repo.ListPenaltyTypesByUser(ctx, repository.ListPenaltyTypesByUserParams{
		UserID:      userID,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list penalty types: %w", err)
	}

	return dto.PenaltyTypeListFromRepository(pts), totalCount, nil
}

func (s *penaltyTypeService) UpdatePenaltyType(ctx context.Context, userID, penaltyTypeID uuid.UUID, req dto.UpdatePenaltyTypeDto) (*dto.ReturnPenaltyTypeDto, error) {
	arg := repository.UpdatePenaltyTypeByUserParams{
		ID:     penaltyTypeID,
		UserID: userID,
	}

	if req.Name != nil {
		arg.Name = pgtype.Text{String: *req.Name, Valid: true}
	}

	pt, err := s.repo.UpdatePenaltyTypeByUser(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrPenaltyTypeNotFound
		}
		return nil, fmt.Errorf("failed to update penalty type: %w", err)
	}

	return dto.PenaltyTypeFromRepository(&pt), nil
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
