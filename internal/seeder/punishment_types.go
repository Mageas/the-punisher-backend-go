package seeder

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func ensurePunishmentTypes(ctx context.Context, repo repository.Querier, userID uuid.UUID) ([]uuid.UUID, error) {
	count, err := repo.CountPunishmentTypesByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count punishment types: %w", err)
	}

	if count > 0 {
		punishmentTypes, err := repo.ListPunishmentTypesByUser(ctx, repository.ListPunishmentTypesByUserParams{
			UserID:      userID,
			QueryLimit:  50,
			QueryOffset: 0,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list punishment types: %w", err)
		}

		ids := make([]uuid.UUID, 0, len(punishmentTypes))
		for _, pt := range punishmentTypes {
			ids = append(ids, pt.ID)
		}

		if len(ids) > 0 {
			return ids, nil
		}
	}

	defaultTypes := []string{"Retenue", "Mot aux parents", "Exclusion"}
	ids := make([]uuid.UUID, 0, len(defaultTypes))
	for _, name := range defaultTypes {
		pt, err := repo.CreatePunishmentType(ctx, repository.CreatePunishmentTypeParams{UserID: userID, Name: name})
		if err != nil {
			return nil, fmt.Errorf("failed to create punishment type: %w", err)
		}
		ids = append(ids, pt.ID)
	}

	return ids, nil
}
