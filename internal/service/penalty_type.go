package service

import (
	"context"

	"github.com/google/uuid"
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
	managed *managedTypeService[dto.RequestPenaltyTypeDto, dto.UpdatePenaltyTypeDto, repository.PenaltyType, dto.ReturnPenaltyTypeDto]
}

func NewPenaltyTypeService(repo repository.Querier) PenaltyTypeService {
	return &penaltyTypeService{
		managed: newManagedTypeService(
			repo,
			managedTypeMetadata[repository.PenaltyType]{
				label:       "penalty type",
				logIDKey:    "penalty_type_id",
				notFoundErr: api.ErrPenaltyTypeNotFound,
				entityID: func(entity repository.PenaltyType) uuid.UUID {
					return entity.ID
				},
			},
			managedTypeOperations[dto.RequestPenaltyTypeDto, dto.UpdatePenaltyTypeDto, repository.PenaltyType]{
				create: func(ctx context.Context, repo repository.Querier, userID uuid.UUID, req dto.RequestPenaltyTypeDto) (repository.PenaltyType, error) {
					return repo.CreatePenaltyType(ctx, repository.CreatePenaltyTypeParams{UserID: userID, Name: req.Name})
				},
				get: func(ctx context.Context, repo repository.Querier, userID, resourceID uuid.UUID) (repository.PenaltyType, error) {
					return repo.GetPenaltyTypeByUser(ctx, repository.GetPenaltyTypeByUserParams{ID: resourceID, UserID: userID})
				},
				count: func(ctx context.Context, repo repository.Querier, userID uuid.UUID) (int64, error) {
					return repo.CountPenaltyTypesByUser(ctx, userID)
				},
				list: func(ctx context.Context, repo repository.Querier, userID uuid.UUID, limit, offset int32) ([]repository.PenaltyType, error) {
					return repo.ListPenaltyTypesByUser(ctx, repository.ListPenaltyTypesByUserParams{
						UserID:      userID,
						QueryLimit:  limit,
						QueryOffset: offset,
					})
				},
				update: func(ctx context.Context, repo repository.Querier, userID, resourceID uuid.UUID, req dto.UpdatePenaltyTypeDto) (repository.PenaltyType, error) {
					params := repository.UpdatePenaltyTypeByUserParams{ID: resourceID, UserID: userID}
					if req.Name != nil {
						params.Name = req.Name
					}
					return repo.UpdatePenaltyTypeByUser(ctx, params)
				},
				delete: func(ctx context.Context, repo repository.Querier, userID, resourceID uuid.UUID) (int64, error) {
					return repo.DeletePenaltyTypeByUser(ctx, repository.DeletePenaltyTypeByUserParams{ID: resourceID, UserID: userID})
				},
			},
			dto.PenaltyTypeFromRepository,
		),
	}
}

func (s *penaltyTypeService) CreatePenaltyType(ctx context.Context, userID uuid.UUID, req dto.RequestPenaltyTypeDto) (*dto.ReturnPenaltyTypeDto, error) {
	return s.managed.Create(ctx, userID, req)
}

func (s *penaltyTypeService) GetPenaltyType(ctx context.Context, userID, penaltyTypeID uuid.UUID) (*dto.ReturnPenaltyTypeDto, error) {
	return s.managed.Get(ctx, userID, penaltyTypeID)
}

func (s *penaltyTypeService) ListPenaltyTypes(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*dto.ReturnPenaltyTypeDto, int64, error) {
	return s.managed.List(ctx, userID, limit, offset)
}

func (s *penaltyTypeService) UpdatePenaltyType(ctx context.Context, userID, penaltyTypeID uuid.UUID, req dto.UpdatePenaltyTypeDto) (*dto.ReturnPenaltyTypeDto, error) {
	return s.managed.Update(ctx, userID, penaltyTypeID, req)
}

func (s *penaltyTypeService) DeletePenaltyType(ctx context.Context, userID, penaltyTypeID uuid.UUID) error {
	return s.managed.Delete(ctx, userID, penaltyTypeID)
}
