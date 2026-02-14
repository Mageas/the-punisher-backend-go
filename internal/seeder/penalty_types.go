package seeder

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func ensurePenaltyTypes(ctx context.Context, repo repository.Querier, userID uuid.UUID) ([]uuid.UUID, error) {
	count, err := repo.CountPenaltyTypesByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count penalty types: %w", err)
	}

	if count > 0 {
		penaltyTypes, err := repo.ListPenaltyTypesByUser(ctx, repository.ListPenaltyTypesByUserParams{
			UserID:      userID,
			QueryLimit:  50,
			QueryOffset: 0,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list penalty types: %w", err)
		}

		ids := make([]uuid.UUID, 0, len(penaltyTypes))
		for _, pt := range penaltyTypes {
			ids = append(ids, pt.ID)
		}

		if len(ids) > 0 {
			return ids, nil
		}
	}

	defaultTypes := []string{"Retard", "Bavardage", "Oubli materiel"}
	ids := make([]uuid.UUID, 0, len(defaultTypes))
	for _, name := range defaultTypes {
		pt, err := repo.CreatePenaltyType(ctx, repository.CreatePenaltyTypeParams{UserID: userID, Name: name})
		if err != nil {
			return nil, fmt.Errorf("failed to create penalty type: %w", err)
		}
		ids = append(ids, pt.ID)
	}

	return ids, nil
}
