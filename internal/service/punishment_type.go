package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/adapter/persistence/sqlcmapper"
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
	managed *managedTypeService[dto.RequestPunishmentTypeDto, dto.UpdatePunishmentTypeDto, repository.PunishmentType, dto.ReturnPunishmentTypeDto]
}

func NewPunishmentTypeService(repo repository.Querier) PunishmentTypeService {
	return &punishmentTypeService{
		managed: newManagedTypeService(
			repo,
			managedTypeMetadata[repository.PunishmentType]{
				label:       "punishment type",
				logIDKey:    "punishment_type_id",
				notFoundErr: api.ErrPunishmentTypeNotFound,
				entityID: func(entity repository.PunishmentType) uuid.UUID {
					return entity.ID
				},
			},
			managedTypeOperations[dto.RequestPunishmentTypeDto, dto.UpdatePunishmentTypeDto, repository.PunishmentType]{
				create: func(ctx context.Context, repo repository.Querier, userID uuid.UUID, req dto.RequestPunishmentTypeDto) (repository.PunishmentType, error) {
					return repo.CreatePunishmentType(ctx, repository.CreatePunishmentTypeParams{UserID: userID, Name: req.Name})
				},
				get: func(ctx context.Context, repo repository.Querier, userID, resourceID uuid.UUID) (repository.PunishmentType, error) {
					return repo.GetPunishmentTypeByUser(ctx, repository.GetPunishmentTypeByUserParams{ID: resourceID, UserID: userID})
				},
				count: func(ctx context.Context, repo repository.Querier, userID uuid.UUID) (int64, error) {
					return repo.CountPunishmentTypesByUser(ctx, userID)
				},
				list: func(ctx context.Context, repo repository.Querier, userID uuid.UUID, limit, offset int32) ([]repository.PunishmentType, error) {
					return repo.ListPunishmentTypesByUser(ctx, repository.ListPunishmentTypesByUserParams{
						UserID:      userID,
						QueryLimit:  limit,
						QueryOffset: offset,
					})
				},
				update: func(ctx context.Context, repo repository.Querier, userID, resourceID uuid.UUID, req dto.UpdatePunishmentTypeDto) (repository.PunishmentType, error) {
					params := repository.UpdatePunishmentTypeByUserParams{ID: resourceID, UserID: userID}
					if req.Name != nil {
						params.Name = req.Name
					}
					return repo.UpdatePunishmentTypeByUser(ctx, params)
				},
				delete: func(ctx context.Context, repo repository.Querier, userID, resourceID uuid.UUID) (int64, error) {
					return repo.DeletePunishmentTypeByUser(ctx, repository.DeletePunishmentTypeByUserParams{ID: resourceID, UserID: userID})
				},
			},
			sqlcmapper.PunishmentTypeFromRepository,
		),
	}
}

func (s *punishmentTypeService) CreatePunishmentType(ctx context.Context, userID uuid.UUID, req dto.RequestPunishmentTypeDto) (*dto.ReturnPunishmentTypeDto, error) {
	return s.managed.Create(ctx, userID, req)
}

func (s *punishmentTypeService) GetPunishmentType(ctx context.Context, userID, punishmentTypeID uuid.UUID) (*dto.ReturnPunishmentTypeDto, error) {
	return s.managed.Get(ctx, userID, punishmentTypeID)
}

func (s *punishmentTypeService) ListPunishmentTypes(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*dto.ReturnPunishmentTypeDto, int64, error) {
	return s.managed.List(ctx, userID, limit, offset)
}

func (s *punishmentTypeService) UpdatePunishmentType(ctx context.Context, userID, punishmentTypeID uuid.UUID, req dto.UpdatePunishmentTypeDto) (*dto.ReturnPunishmentTypeDto, error) {
	return s.managed.Update(ctx, userID, punishmentTypeID, req)
}

func (s *punishmentTypeService) DeletePunishmentType(ctx context.Context, userID, punishmentTypeID uuid.UUID) error {
	return s.managed.Delete(ctx, userID, punishmentTypeID)
}
