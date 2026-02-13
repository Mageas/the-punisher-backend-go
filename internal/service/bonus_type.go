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

type BonusTypeService interface {
	CreateBonusType(ctx context.Context, userID uuid.UUID, req dto.RequestBonusTypeDto) (*dto.ReturnBonusTypeDto, error)
	GetBonusType(ctx context.Context, userID, bonusTypeID uuid.UUID) (*dto.ReturnBonusTypeDto, error)
	ListBonusTypes(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*dto.ReturnBonusTypeDto, int64, error)
	UpdateBonusType(ctx context.Context, userID, bonusTypeID uuid.UUID, req dto.UpdateBonusTypeDto) (*dto.ReturnBonusTypeDto, error)
	DeleteBonusType(ctx context.Context, userID, bonusTypeID uuid.UUID) error
}

type bonusTypeService struct {
	repo *repository.Queries
}

func NewBonusTypeService(repo *repository.Queries) BonusTypeService {
	return &bonusTypeService{
		repo: repo,
	}
}

func (s *bonusTypeService) CreateBonusType(ctx context.Context, userID uuid.UUID, req dto.RequestBonusTypeDto) (*dto.ReturnBonusTypeDto, error) {
	bt, err := s.repo.CreateBonusType(ctx, repository.CreateBonusTypeParams{
		UserID: userID,
		Name:   req.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create bonus type: %w", err)
	}

	slog.Info("bonus type created", "bonus_type_id", bt.ID, "user_id", userID)

	return dto.BonusTypeFromRepository(&bt), nil
}

func (s *bonusTypeService) GetBonusType(ctx context.Context, userID, bonusTypeID uuid.UUID) (*dto.ReturnBonusTypeDto, error) {
	bt, err := s.repo.GetBonusType(ctx, repository.GetBonusTypeParams{
		ID:     bonusTypeID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrBonusTypeNotFound
		}
		return nil, fmt.Errorf("failed to get bonus type: %w", err)
	}

	return dto.BonusTypeFromRepository(&bt), nil
}

func (s *bonusTypeService) ListBonusTypes(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*dto.ReturnBonusTypeDto, int64, error) {
	totalCount, err := s.repo.CountBonusTypesByUser(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count bonus types: %w", err)
	}

	bts, err := s.repo.ListBonusTypesByUser(ctx, repository.ListBonusTypesByUserParams{
		UserID:      userID,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list bonus types: %w", err)
	}

	return dto.BonusTypeListFromRepository(bts), totalCount, nil
}

func (s *bonusTypeService) UpdateBonusType(ctx context.Context, userID, bonusTypeID uuid.UUID, req dto.UpdateBonusTypeDto) (*dto.ReturnBonusTypeDto, error) {
	arg := repository.UpdateBonusTypeParams{
		ID:     bonusTypeID,
		UserID: userID,
	}

	if req.Name != nil {
		arg.Name = pgtype.Text{String: *req.Name, Valid: true}
	}

	bt, err := s.repo.UpdateBonusType(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrBonusTypeNotFound
		}
		return nil, fmt.Errorf("failed to update bonus type: %w", err)
	}

	return dto.BonusTypeFromRepository(&bt), nil
}

func (s *bonusTypeService) DeleteBonusType(ctx context.Context, userID, bonusTypeID uuid.UUID) error {
	rowsAffected, err := s.repo.DeleteBonusType(ctx, repository.DeleteBonusTypeParams{
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
