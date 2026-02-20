package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type managedTypeMetadata[TEntity any] struct {
	label       string
	logIDKey    string
	notFoundErr error
	entityID    func(entity TEntity) uuid.UUID
}

type managedTypeOperations[TCreateReq any, TUpdateReq any, TEntity any] struct {
	create func(ctx context.Context, repo repository.Querier, userID uuid.UUID, req TCreateReq) (TEntity, error)
	get    func(ctx context.Context, repo repository.Querier, userID, resourceID uuid.UUID) (TEntity, error)
	count  func(ctx context.Context, repo repository.Querier, userID uuid.UUID) (int64, error)
	list   func(ctx context.Context, repo repository.Querier, userID uuid.UUID, limit, offset int32) ([]TEntity, error)
	update func(ctx context.Context, repo repository.Querier, userID, resourceID uuid.UUID, req TUpdateReq) (TEntity, error)
	delete func(ctx context.Context, repo repository.Querier, userID, resourceID uuid.UUID) (int64, error)
}

type managedTypeService[TCreateReq any, TUpdateReq any, TEntity any, TReturn any] struct {
	repo     repository.Querier
	meta     managedTypeMetadata[TEntity]
	ops      managedTypeOperations[TCreateReq, TUpdateReq, TEntity]
	toReturn func(*TEntity) *TReturn
}

func newManagedTypeService[TCreateReq any, TUpdateReq any, TEntity any, TReturn any](
	repo repository.Querier,
	meta managedTypeMetadata[TEntity],
	ops managedTypeOperations[TCreateReq, TUpdateReq, TEntity],
	toReturn func(*TEntity) *TReturn,
) *managedTypeService[TCreateReq, TUpdateReq, TEntity, TReturn] {
	return &managedTypeService[TCreateReq, TUpdateReq, TEntity, TReturn]{
		repo:     repo,
		meta:     meta,
		ops:      ops,
		toReturn: toReturn,
	}
}

func (s *managedTypeService[TCreateReq, TUpdateReq, TEntity, TReturn]) Create(ctx context.Context, userID uuid.UUID, req TCreateReq) (*TReturn, error) {
	entity, err := s.ops.create(ctx, s.repo, userID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s: %w", s.meta.label, err)
	}

	slog.Info(s.meta.label+" created", s.meta.logIDKey, s.meta.entityID(entity), "user_id", userID)

	return s.toReturn(&entity), nil
}

func (s *managedTypeService[TCreateReq, TUpdateReq, TEntity, TReturn]) Get(ctx context.Context, userID, resourceID uuid.UUID) (*TReturn, error) {
	entity, err := s.ops.get(ctx, s.repo, userID, resourceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, s.meta.notFoundErr
		}
		return nil, fmt.Errorf("failed to get %s: %w", s.meta.label, err)
	}

	return s.toReturn(&entity), nil
}

func (s *managedTypeService[TCreateReq, TUpdateReq, TEntity, TReturn]) List(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*TReturn, int64, error) {
	totalCount, err := s.ops.count(ctx, s.repo, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count %ss: %w", s.meta.label, err)
	}

	entities, err := s.ops.list(ctx, s.repo, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list %ss: %w", s.meta.label, err)
	}

	return mapManagedTypeList(entities, s.toReturn), totalCount, nil
}

func (s *managedTypeService[TCreateReq, TUpdateReq, TEntity, TReturn]) Update(ctx context.Context, userID, resourceID uuid.UUID, req TUpdateReq) (*TReturn, error) {
	entity, err := s.ops.update(ctx, s.repo, userID, resourceID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, s.meta.notFoundErr
		}
		return nil, fmt.Errorf("failed to update %s: %w", s.meta.label, err)
	}

	return s.toReturn(&entity), nil
}

func (s *managedTypeService[TCreateReq, TUpdateReq, TEntity, TReturn]) Delete(ctx context.Context, userID, resourceID uuid.UUID) error {
	rowsAffected, err := s.ops.delete(ctx, s.repo, userID, resourceID)
	if err != nil {
		return fmt.Errorf("failed to delete %s: %w", s.meta.label, err)
	}

	if rowsAffected == 0 {
		return s.meta.notFoundErr
	}

	slog.Info(s.meta.label+" deleted", s.meta.logIDKey, resourceID, "user_id", userID)

	return nil
}

func mapManagedTypeList[TEntity any, TReturn any](entities []TEntity, toReturn func(*TEntity) *TReturn) []*TReturn {
	mapped := make([]*TReturn, 0, len(entities))
	for _, entity := range entities {
		if dto := toReturn(&entity); dto != nil {
			mapped = append(mapped, dto)
		}
	}

	return mapped
}
