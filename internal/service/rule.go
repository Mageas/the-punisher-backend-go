package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
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

type resolvedRulePayload struct {
	Name                      string
	ResultingPunishmentTypeID uuid.UUID
	PenaltyTypeID             uuid.UUID
	Threshold                 int32
	DueAtAfterDays            int32
	DueAtAfterDaysSet         bool
	DueAtMode                 string
	DueAtAfterLessons         *int32
	Mode                      string
	IsActive                  bool
}

func NewRuleService(repo repository.Querier) RuleService {
	return &ruleService{repo: repo}
}

func (s *ruleService) CreateRule(ctx context.Context, userID uuid.UUID, req dto.RequestRuleDto) (*dto.ReturnRuleDto, error) {
	payload, err := s.resolveCreateRulePayload(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	rule, err := s.repo.CreateRule(ctx, repository.CreateRuleParams{
		UserID:                    userID,
		Name:                      payload.Name,
		ResultingPunishmentTypeID: payload.ResultingPunishmentTypeID,
		PenaltyTypeID:             payload.PenaltyTypeID,
		Threshold:                 payload.Threshold,
		DueAtAfterDays:            payload.DueAtAfterDays,
		DueAtMode:                 payload.DueAtMode,
		DueAtAfterLessons:         payload.DueAtAfterLessons,
		Mode:                      payload.Mode,
		IsActive:                  payload.IsActive,
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
		if errors.Is(err, repository.ErrNoRows) {
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
	existingRule, err := s.repo.GetRuleByUser(ctx, repository.GetRuleByUserParams{
		ID:     ruleID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrRuleNotFound
		}
		return nil, fmt.Errorf("failed to get rule: %w", err)
	}

	payload, err := s.resolveUpdateRulePayload(ctx, userID, existingRule, req)
	if err != nil {
		return nil, err
	}

	rule, err := s.repo.UpdateRuleByUser(ctx, repository.UpdateRuleByUserParams{
		ID:                        ruleID,
		UserID:                    userID,
		Name:                      payload.Name,
		ResultingPunishmentTypeID: payload.ResultingPunishmentTypeID,
		PenaltyTypeID:             payload.PenaltyTypeID,
		Threshold:                 payload.Threshold,
		DueAtAfterDays:            payload.DueAtAfterDays,
		DueAtMode:                 payload.DueAtMode,
		DueAtAfterLessons:         payload.DueAtAfterLessons,
		Mode:                      payload.Mode,
		IsActive:                  payload.IsActive,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
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

func (s *ruleService) resolveCreateRulePayload(ctx context.Context, userID uuid.UUID, req dto.RequestRuleDto) (resolvedRulePayload, error) {
	resultingPunishmentTypeID, err := uuid.Parse(req.ResultingPunishmentTypeID)
	if err != nil {
		return resolvedRulePayload{}, api.ErrInvalidRequestBody
	}

	if err := ensurePunishmentTypeExists(ctx, s.repo, userID, resultingPunishmentTypeID); err != nil {
		return resolvedRulePayload{}, err
	}

	penaltyTypeID, err := uuid.Parse(req.PenaltyTypeID)
	if err != nil {
		return resolvedRulePayload{}, api.ErrInvalidRequestBody
	}

	if err := ensurePenaltyTypeExists(ctx, s.repo, userID, penaltyTypeID); err != nil {
		return resolvedRulePayload{}, err
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	payload := resolvedRulePayload{
		Name:                      req.Name,
		ResultingPunishmentTypeID: resultingPunishmentTypeID,
		PenaltyTypeID:             penaltyTypeID,
		Threshold:                 req.Threshold,
		DueAtMode:                 req.DueAtMode,
		DueAtAfterLessons:         req.DueAtAfterLessons,
		Mode:                      req.Mode,
		IsActive:                  isActive,
	}
	if req.DueAtAfterDays != nil {
		payload.DueAtAfterDays = *req.DueAtAfterDays
		payload.DueAtAfterDaysSet = true
	}

	if err := s.validateAndNormalizeRulePayload(&payload); err != nil {
		return resolvedRulePayload{}, err
	}

	return payload, nil
}

func (s *ruleService) resolveUpdateRulePayload(
	ctx context.Context,
	userID uuid.UUID,
	existingRule repository.GetRuleByUserRow,
	req dto.UpdateRuleDto,
) (resolvedRulePayload, error) {
	payload := resolvedRulePayload{
		Name:                      existingRule.Name,
		ResultingPunishmentTypeID: existingRule.ResultingPunishmentTypeID,
		PenaltyTypeID:             existingRule.PenaltyTypeID,
		Threshold:                 existingRule.Threshold,
		DueAtAfterDays:            existingRule.DueAtAfterDays,
		DueAtAfterDaysSet:         true,
		DueAtMode:                 existingRule.DueAtMode,
		DueAtAfterLessons:         existingRule.DueAtAfterLessons,
		Mode:                      existingRule.Mode,
		IsActive:                  existingRule.IsActive,
	}

	if req.Name != nil {
		payload.Name = *req.Name
	}

	if req.ResultingPunishmentTypeID != nil {
		resultingPunishmentTypeID, err := uuid.Parse(*req.ResultingPunishmentTypeID)
		if err != nil {
			return resolvedRulePayload{}, api.ErrInvalidRequestBody
		}
		if err := ensurePunishmentTypeExists(ctx, s.repo, userID, resultingPunishmentTypeID); err != nil {
			return resolvedRulePayload{}, err
		}
		payload.ResultingPunishmentTypeID = resultingPunishmentTypeID
	}

	if req.PenaltyTypeID != nil {
		penaltyTypeID, err := uuid.Parse(*req.PenaltyTypeID)
		if err != nil {
			return resolvedRulePayload{}, api.ErrInvalidRequestBody
		}
		if err := ensurePenaltyTypeExists(ctx, s.repo, userID, penaltyTypeID); err != nil {
			return resolvedRulePayload{}, err
		}
		payload.PenaltyTypeID = penaltyTypeID
	}

	if req.Threshold != nil {
		payload.Threshold = *req.Threshold
	}

	if req.DueAtAfterDays != nil {
		payload.DueAtAfterDays = *req.DueAtAfterDays
		payload.DueAtAfterDaysSet = true
	}

	if req.DueAtMode != nil {
		payload.DueAtMode = *req.DueAtMode
	}

	if req.DueAtAfterLessons != nil {
		payload.DueAtAfterLessons = req.DueAtAfterLessons
	}

	if req.Mode != nil {
		payload.Mode = *req.Mode
	}

	if req.IsActive != nil {
		payload.IsActive = *req.IsActive
	}

	if req.DueAtMode != nil && *req.DueAtMode == ruleDueAtModeDays {
		if req.DueAtAfterLessons != nil {
			return resolvedRulePayload{}, newRuleValidationError("due_at_after_lessons", "rule_due_at_after_lessons_forbidden_for_due_at_mode_days")
		}

		payload.DueAtAfterLessons = nil
	}

	if err := s.validateAndNormalizeRulePayload(&payload); err != nil {
		return resolvedRulePayload{}, err
	}

	return payload, nil
}

func (s *ruleService) validateAndNormalizeRulePayload(payload *resolvedRulePayload) error {
	if payload.DueAtMode == "" {
		return newRuleValidationError("due_at_mode", api.KeyValidationFieldRequired)
	}

	switch payload.DueAtMode {
	case ruleDueAtModeDays:
		if payload.DueAtAfterLessons != nil {
			return newRuleValidationError("due_at_after_lessons", "rule_due_at_after_lessons_forbidden_for_due_at_mode_days")
		}
		if !payload.DueAtAfterDaysSet {
			return newRuleValidationError("due_at_after_days", api.KeyValidationFieldRequired)
		}

		payload.DueAtAfterLessons = nil
		return nil
	case ruleDueAtModeNextLessons:
		if payload.DueAtAfterDays != 0 {
			return newRuleValidationError("due_at_after_days", "rule_due_at_after_days_must_be_zero_for_due_at_mode_next_lessons")
		}
		if payload.DueAtAfterLessons == nil {
			return newRuleValidationError("due_at_after_lessons", api.KeyValidationFieldRequired)
		}
		if *payload.DueAtAfterLessons < ruleDueAtAfterLessonsMin {
			return newRuleValidationError("due_at_after_lessons", fmt.Sprintf(api.KeyValidationMinLength, fmt.Sprintf("%d", ruleDueAtAfterLessonsMin)))
		}
		if *payload.DueAtAfterLessons > ruleDueAtAfterLessonsMax {
			return newRuleValidationError("due_at_after_lessons", fmt.Sprintf(api.KeyValidationMaxLength, fmt.Sprintf("%d", ruleDueAtAfterLessonsMax)))
		}

		payload.DueAtAfterDays = 0
		return nil
	default:
		return newRuleValidationError("due_at_mode", fmt.Sprintf(api.KeyValidationOneOf, fmt.Sprintf("%s|%s", ruleDueAtModeDays, ruleDueAtModeNextLessons)))
	}
}

func ensurePunishmentTypeExists(ctx context.Context, repo repository.Querier, userID, punishmentTypeID uuid.UUID) error {
	if _, err := repo.GetPunishmentTypeByUser(ctx, repository.GetPunishmentTypeByUserParams{
		ID:     punishmentTypeID,
		UserID: userID,
	}); err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return api.ErrPunishmentTypeNotFound
		}
		return fmt.Errorf("failed to get punishment type: %w", err)
	}

	return nil
}

func ensurePenaltyTypeExists(ctx context.Context, repo repository.Querier, userID, penaltyTypeID uuid.UUID) error {
	if _, err := repo.GetPenaltyTypeByUser(ctx, repository.GetPenaltyTypeByUserParams{
		ID:     penaltyTypeID,
		UserID: userID,
	}); err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return api.ErrPenaltyTypeNotFound
		}
		return fmt.Errorf("failed to get penalty type: %w", err)
	}

	return nil
}

func newRuleValidationError(field, errorCode string) error {
	return api.NewAPIError(http.StatusBadRequest, "validation_failed", api.ErrorDetail{
		Field: field,
		Error: errorCode,
	})
}
