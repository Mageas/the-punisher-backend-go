package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mageas/the-punisher-backend/internal/adapter/persistence/sqlcmapper"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type RuleService interface {
	CreateRule(ctx context.Context, userID uuid.UUID, req dto.RequestRuleDto) (*dto.ReturnRuleDto, error)
	GetRule(ctx context.Context, userID, ruleID uuid.UUID) (*dto.ReturnRuleDto, error)
	ListRules(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*dto.ReturnRuleDto, int64, error)
	UpdateRule(ctx context.Context, userID, ruleID uuid.UUID, req dto.UpdateRuleDto) (*dto.ReturnRuleDto, error)
	DeleteRule(ctx context.Context, userID, ruleID uuid.UUID) error
}

type ruleService struct {
	repo repository.Querier
}

func NewRuleService(repo repository.Querier) RuleService {
	return &ruleService{
		repo: repo,
	}
}

func (s *ruleService) CreateRule(ctx context.Context, userID uuid.UUID, req dto.RequestRuleDto) (*dto.ReturnRuleDto, error) {
	resultingPunishmentTypeID, err := uuid.Parse(req.ResultingPunishmentTypeID)
	if err != nil {
		return nil, api.ErrInvalidRequestBody
	}

	penaltyTypeID, err := uuid.Parse(req.PenaltyTypeID)
	if err != nil {
		return nil, api.ErrInvalidRequestBody
	}

	if _, err := s.repo.GetPunishmentTypeByUser(ctx, repository.GetPunishmentTypeByUserParams{
		ID:     resultingPunishmentTypeID,
		UserID: userID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrPunishmentTypeNotFound
		}
		return nil, fmt.Errorf("failed to get punishment type: %w", err)
	}

	if _, err := s.repo.GetPenaltyTypeByUser(ctx, repository.GetPenaltyTypeByUserParams{
		ID:     penaltyTypeID,
		UserID: userID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrPenaltyTypeNotFound
		}
		return nil, fmt.Errorf("failed to get penalty type: %w", err)
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	rule, err := s.repo.CreateRule(ctx, repository.CreateRuleParams{
		UserID:                    userID,
		Name:                      req.Name,
		ResultingPunishmentTypeID: resultingPunishmentTypeID,
		PenaltyTypeID:             penaltyTypeID,
		Threshold:                 req.Threshold,
		DueAtAfterDays:            req.DueAtAfterDays,
		Mode:                      req.Mode,
		IsActive:                  isActive,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create rule: %w", err)
	}

	slog.Info("rule created", "rule_id", rule.ID, "user_id", userID)

	return sqlcmapper.RuleFromCreateRow(&rule), nil
}

func (s *ruleService) GetRule(ctx context.Context, userID, ruleID uuid.UUID) (*dto.ReturnRuleDto, error) {
	rule, err := s.repo.GetRuleByUser(ctx, repository.GetRuleByUserParams{
		ID:     ruleID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrRuleNotFound
		}
		return nil, fmt.Errorf("failed to get rule: %w", err)
	}

	return sqlcmapper.RuleFromGetRow(&rule), nil
}

func (s *ruleService) ListRules(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*dto.ReturnRuleDto, int64, error) {
	totalCount, err := s.repo.CountRulesByUser(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count rules: %w", err)
	}

	rules, err := s.repo.ListRulesByUser(ctx, repository.ListRulesByUserParams{
		UserID:      userID,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list rules: %w", err)
	}

	return sqlcmapper.RuleListFromListByUserRows(rules), totalCount, nil
}

func (s *ruleService) UpdateRule(ctx context.Context, userID, ruleID uuid.UUID, req dto.UpdateRuleDto) (*dto.ReturnRuleDto, error) {
	arg := repository.UpdateRuleByUserParams{
		ID:     ruleID,
		UserID: userID,
	}

	if req.Name != nil {
		arg.Name = req.Name
	}

	if req.ResultingPunishmentTypeID != nil {
		resultingPunishmentTypeID, err := uuid.Parse(*req.ResultingPunishmentTypeID)
		if err != nil {
			return nil, api.ErrInvalidRequestBody
		}

		if _, err := s.repo.GetPunishmentTypeByUser(ctx, repository.GetPunishmentTypeByUserParams{
			ID:     resultingPunishmentTypeID,
			UserID: userID,
		}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, api.ErrPunishmentTypeNotFound
			}
			return nil, fmt.Errorf("failed to get punishment type: %w", err)
		}

		arg.ResultingPunishmentTypeID = &resultingPunishmentTypeID
	}

	if req.PenaltyTypeID != nil {
		penaltyTypeID, err := uuid.Parse(*req.PenaltyTypeID)
		if err != nil {
			return nil, api.ErrInvalidRequestBody
		}

		if _, err := s.repo.GetPenaltyTypeByUser(ctx, repository.GetPenaltyTypeByUserParams{
			ID:     penaltyTypeID,
			UserID: userID,
		}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, api.ErrPenaltyTypeNotFound
			}
			return nil, fmt.Errorf("failed to get penalty type: %w", err)
		}

		arg.PenaltyTypeID = &penaltyTypeID
	}

	if req.Threshold != nil {
		arg.Threshold = req.Threshold
	}

	if req.DueAtAfterDays != nil {
		arg.DueAtAfterDays = req.DueAtAfterDays
	}

	if req.Mode != nil {
		arg.Mode = req.Mode
	}

	if req.IsActive != nil {
		arg.IsActive = req.IsActive
	}

	rule, err := s.repo.UpdateRuleByUser(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrRuleNotFound
		}
		return nil, fmt.Errorf("failed to update rule: %w", err)
	}

	return sqlcmapper.RuleFromUpdateRow(&rule), nil
}

func (s *ruleService) DeleteRule(ctx context.Context, userID, ruleID uuid.UUID) error {
	rowsAffected, err := s.repo.DeleteRuleByUser(ctx, repository.DeleteRuleByUserParams{
		ID:     ruleID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}

	if rowsAffected == 0 {
		return api.ErrRuleNotFound
	}

	slog.Info("rule deleted", "rule_id", ruleID, "user_id", userID)

	return nil
}
