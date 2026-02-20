package service

import (
	"context"

	"github.com/google/uuid"
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
	managed *managedTypeService[dto.RequestBonusTypeDto, dto.UpdateBonusTypeDto, repository.BonusType, dto.ReturnBonusTypeDto]
}

func NewBonusTypeService(repo repository.Querier) BonusTypeService {
	return &bonusTypeService{
		managed: newManagedTypeService(
			repo,
			managedTypeMetadata[repository.BonusType]{
				label:       "bonus type",
				logIDKey:    "bonus_type_id",
				notFoundErr: api.ErrBonusTypeNotFound,
				entityID: func(entity repository.BonusType) uuid.UUID {
					return entity.ID
				},
			},
			managedTypeOperations[dto.RequestBonusTypeDto, dto.UpdateBonusTypeDto, repository.BonusType]{
				create: func(ctx context.Context, repo repository.Querier, userID uuid.UUID, req dto.RequestBonusTypeDto) (repository.BonusType, error) {
					return repo.CreateBonusType(ctx, repository.CreateBonusTypeParams{UserID: userID, Name: req.Name})
				},
				get: func(ctx context.Context, repo repository.Querier, userID, resourceID uuid.UUID) (repository.BonusType, error) {
					return repo.GetBonusTypeByUser(ctx, repository.GetBonusTypeByUserParams{ID: resourceID, UserID: userID})
				},
				count: func(ctx context.Context, repo repository.Querier, userID uuid.UUID) (int64, error) {
					return repo.CountBonusTypesByUser(ctx, userID)
				},
				list: func(ctx context.Context, repo repository.Querier, userID uuid.UUID, limit, offset int32) ([]repository.BonusType, error) {
					return repo.ListBonusTypesByUser(ctx, repository.ListBonusTypesByUserParams{
						UserID:      userID,
						QueryLimit:  limit,
						QueryOffset: offset,
					})
				},
				update: func(ctx context.Context, repo repository.Querier, userID, resourceID uuid.UUID, req dto.UpdateBonusTypeDto) (repository.BonusType, error) {
					params := repository.UpdateBonusTypeByUserParams{ID: resourceID, UserID: userID}
					if req.Name != nil {
						params.Name = pgtype.Text{String: *req.Name, Valid: true}
					}
					return repo.UpdateBonusTypeByUser(ctx, params)
				},
				delete: func(ctx context.Context, repo repository.Querier, userID, resourceID uuid.UUID) (int64, error) {
					return repo.DeleteBonusTypeByUser(ctx, repository.DeleteBonusTypeByUserParams{ID: resourceID, UserID: userID})
				},
			},
			dto.BonusTypeFromRepository,
		),
	}
}

func (s *bonusTypeService) CreateBonusType(ctx context.Context, userID uuid.UUID, req dto.RequestBonusTypeDto) (*dto.ReturnBonusTypeDto, error) {
	return s.managed.Create(ctx, userID, req)
}

func (s *bonusTypeService) GetBonusType(ctx context.Context, userID, bonusTypeID uuid.UUID) (*dto.ReturnBonusTypeDto, error) {
	return s.managed.Get(ctx, userID, bonusTypeID)
}

func (s *bonusTypeService) ListBonusTypes(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*dto.ReturnBonusTypeDto, int64, error) {
	return s.managed.List(ctx, userID, limit, offset)
}

func (s *bonusTypeService) UpdateBonusType(ctx context.Context, userID, bonusTypeID uuid.UUID, req dto.UpdateBonusTypeDto) (*dto.ReturnBonusTypeDto, error) {
	return s.managed.Update(ctx, userID, bonusTypeID, req)
}

func (s *bonusTypeService) DeleteBonusType(ctx context.Context, userID, bonusTypeID uuid.UUID) error {
	return s.managed.Delete(ctx, userID, bonusTypeID)
}
