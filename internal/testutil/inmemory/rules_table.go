package inmemory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

const (
	OpCreateRule                          = "CreateRule"
	OpGetRuleByUser                       = "GetRuleByUser"
	OpCountRulesByUser                    = "CountRulesByUser"
	OpListRulesByUser                     = "ListRulesByUser"
	OpUpdateRuleByUser                    = "UpdateRuleByUser"
	OpDeleteRuleByUser                    = "DeleteRuleByUser"
	OpListActiveRulesByUserAndPenaltyType = "ListActiveRulesByUserAndPenaltyType"
)

func (r *Repository) penaltyTypeNameLocked(penaltyTypeID uuid.UUID) string {
	if penaltyType, ok := r.penaltyTypes[penaltyTypeID]; ok {
		return penaltyType.Name
	}

	return ""
}

func (r *Repository) punishmentTypeNameLocked(punishmentTypeID uuid.UUID) string {
	if punishmentType, ok := r.punishmentTypes[punishmentTypeID]; ok {
		return punishmentType.Name
	}

	return ""
}

func (r *Repository) buildCreateRuleRowLocked(rule repository.Rule) repository.CreateRuleRow {
	return repository.CreateRuleRow{
		ID:                          rule.ID,
		UserID:                      rule.UserID,
		Name:                        rule.Name,
		ResultingPunishmentTypeID:   rule.ResultingPunishmentTypeID,
		PenaltyTypeID:               rule.PenaltyTypeID,
		Threshold:                   rule.Threshold,
		Mode:                        rule.Mode,
		IsActive:                    rule.IsActive,
		CreatedAt:                   rule.CreatedAt,
		UpdatedAt:                   rule.UpdatedAt,
		DueAtAfterDays:              rule.DueAtAfterDays,
		PenaltyTypeName:             r.penaltyTypeNameLocked(rule.PenaltyTypeID),
		ResultingPunishmentTypeName: r.punishmentTypeNameLocked(rule.ResultingPunishmentTypeID),
	}
}

func (r *Repository) buildGetRuleRowLocked(rule repository.Rule) repository.GetRuleByUserRow {
	return repository.GetRuleByUserRow{
		ID:                          rule.ID,
		UserID:                      rule.UserID,
		Name:                        rule.Name,
		ResultingPunishmentTypeID:   rule.ResultingPunishmentTypeID,
		PenaltyTypeID:               rule.PenaltyTypeID,
		Threshold:                   rule.Threshold,
		Mode:                        rule.Mode,
		IsActive:                    rule.IsActive,
		CreatedAt:                   rule.CreatedAt,
		UpdatedAt:                   rule.UpdatedAt,
		DueAtAfterDays:              rule.DueAtAfterDays,
		PenaltyTypeName:             r.penaltyTypeNameLocked(rule.PenaltyTypeID),
		ResultingPunishmentTypeName: r.punishmentTypeNameLocked(rule.ResultingPunishmentTypeID),
	}
}

func (r *Repository) buildListRuleRowLocked(rule repository.Rule) repository.ListRulesByUserRow {
	return repository.ListRulesByUserRow{
		ID:                          rule.ID,
		UserID:                      rule.UserID,
		Name:                        rule.Name,
		ResultingPunishmentTypeID:   rule.ResultingPunishmentTypeID,
		PenaltyTypeID:               rule.PenaltyTypeID,
		Threshold:                   rule.Threshold,
		Mode:                        rule.Mode,
		IsActive:                    rule.IsActive,
		CreatedAt:                   rule.CreatedAt,
		UpdatedAt:                   rule.UpdatedAt,
		DueAtAfterDays:              rule.DueAtAfterDays,
		PenaltyTypeName:             r.penaltyTypeNameLocked(rule.PenaltyTypeID),
		ResultingPunishmentTypeName: r.punishmentTypeNameLocked(rule.ResultingPunishmentTypeID),
	}
}

func (r *Repository) buildUpdateRuleRowLocked(rule repository.Rule) repository.UpdateRuleByUserRow {
	return repository.UpdateRuleByUserRow{
		ID:                          rule.ID,
		UserID:                      rule.UserID,
		Name:                        rule.Name,
		ResultingPunishmentTypeID:   rule.ResultingPunishmentTypeID,
		PenaltyTypeID:               rule.PenaltyTypeID,
		Threshold:                   rule.Threshold,
		Mode:                        rule.Mode,
		IsActive:                    rule.IsActive,
		CreatedAt:                   rule.CreatedAt,
		UpdatedAt:                   rule.UpdatedAt,
		DueAtAfterDays:              rule.DueAtAfterDays,
		PenaltyTypeName:             r.penaltyTypeNameLocked(rule.PenaltyTypeID),
		ResultingPunishmentTypeName: r.punishmentTypeNameLocked(rule.ResultingPunishmentTypeID),
	}
}

func (r *Repository) SeedRule(rule repository.Rule) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = now
	}
	if rule.UpdatedAt.IsZero() {
		rule.UpdatedAt = rule.CreatedAt
	}

	r.rules[rule.ID] = rule
}

func (r *Repository) CreateRule(_ context.Context, arg repository.CreateRuleParams) (repository.CreateRuleRow, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpCreateRule); err != nil {
		return repository.CreateRuleRow{}, err
	}

	now := time.Now()
	rule := repository.Rule{
		ID:                        uuid.New(),
		UserID:                    arg.UserID,
		Name:                      arg.Name,
		ResultingPunishmentTypeID: arg.ResultingPunishmentTypeID,
		PenaltyTypeID:             arg.PenaltyTypeID,
		Threshold:                 arg.Threshold,
		Mode:                      arg.Mode,
		IsActive:                  arg.IsActive,
		CreatedAt:                 now,
		UpdatedAt:                 now,
		DueAtAfterDays:            arg.DueAtAfterDays,
	}
	r.rules[rule.ID] = rule

	return r.buildCreateRuleRowLocked(rule), nil
}

func (r *Repository) GetRuleByUser(_ context.Context, arg repository.GetRuleByUserParams) (repository.GetRuleByUserRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpGetRuleByUser); err != nil {
		return repository.GetRuleByUserRow{}, err
	}

	rule, ok := r.rules[arg.ID]
	if !ok || rule.UserID != arg.UserID {
		return repository.GetRuleByUserRow{}, pgx.ErrNoRows
	}

	return r.buildGetRuleRowLocked(rule), nil
}

func (r *Repository) CountRulesByUser(_ context.Context, userID uuid.UUID) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpCountRulesByUser); err != nil {
		return 0, err
	}

	var count int64
	for _, rule := range r.rules {
		if rule.UserID == userID {
			count++
		}
	}

	return count, nil
}

func (r *Repository) ListRulesByUser(_ context.Context, arg repository.ListRulesByUserParams) ([]repository.ListRulesByUserRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListRulesByUser); err != nil {
		return nil, err
	}

	items := make([]repository.Rule, 0)
	for _, rule := range r.rules {
		if rule.UserID == arg.UserID {
			items = append(items, rule)
		}
	}

	sortRulesByCreatedAtDesc(items)
	paginated := paginate(items, arg.QueryOffset, arg.QueryLimit)

	rows := make([]repository.ListRulesByUserRow, 0, len(paginated))
	for _, rule := range paginated {
		rows = append(rows, r.buildListRuleRowLocked(rule))
	}

	return rows, nil
}

func (r *Repository) UpdateRuleByUser(_ context.Context, arg repository.UpdateRuleByUserParams) (repository.UpdateRuleByUserRow, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpUpdateRuleByUser); err != nil {
		return repository.UpdateRuleByUserRow{}, err
	}

	rule, ok := r.rules[arg.ID]
	if !ok || rule.UserID != arg.UserID {
		return repository.UpdateRuleByUserRow{}, pgx.ErrNoRows
	}

	if arg.Name.Valid {
		rule.Name = arg.Name.String
	}
	if arg.ResultingPunishmentTypeID.Valid {
		rule.ResultingPunishmentTypeID = uuid.UUID(arg.ResultingPunishmentTypeID.Bytes)
	}
	if arg.PenaltyTypeID.Valid {
		rule.PenaltyTypeID = uuid.UUID(arg.PenaltyTypeID.Bytes)
	}
	if arg.Threshold.Valid {
		rule.Threshold = arg.Threshold.Int32
	}
	if arg.Mode.Valid {
		rule.Mode = arg.Mode.String
	}
	if arg.IsActive.Valid {
		rule.IsActive = arg.IsActive.Bool
	}
	if arg.DueAtAfterDays.Valid {
		rule.DueAtAfterDays = arg.DueAtAfterDays.Int32
	}

	rule.UpdatedAt = time.Now()
	r.rules[arg.ID] = rule

	return r.buildUpdateRuleRowLocked(rule), nil
}

func (r *Repository) DeleteRuleByUser(_ context.Context, arg repository.DeleteRuleByUserParams) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpDeleteRuleByUser); err != nil {
		return 0, err
	}

	rule, ok := r.rules[arg.ID]
	if !ok || rule.UserID != arg.UserID {
		return 0, nil
	}

	delete(r.rules, arg.ID)
	return 1, nil
}

func (r *Repository) ListActiveRulesByUserAndPenaltyType(_ context.Context, arg repository.ListActiveRulesByUserAndPenaltyTypeParams) ([]repository.Rule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpListActiveRulesByUserAndPenaltyType); err != nil {
		return nil, err
	}

	items := make([]repository.Rule, 0)
	for _, rule := range r.rules {
		if rule.UserID == arg.UserID && rule.PenaltyTypeID == arg.PenaltyTypeID && rule.IsActive {
			items = append(items, rule)
		}
	}

	sortRulesByCreatedAtDesc(items)
	return items, nil
}
