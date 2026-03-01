package seeder

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func ensureBonusTypes(ctx context.Context, repo repository.Querier, userID uuid.UUID) ([]uuid.UUID, error) {
	count, err := repo.CountBonusTypesByUser(ctx, repository.CountBonusTypesByUserParams{
		UserID: userID,
		Search: nil,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count bonus types: %w", err)
	}

	if count > 0 {
		bonusTypes, err := repo.ListBonusTypesByUser(ctx, repository.ListBonusTypesByUserParams{
			UserID:      userID,
			QueryLimit:  50,
			QueryOffset: 0,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list bonus types: %w", err)
		}

		ids := make([]uuid.UUID, 0, len(bonusTypes))
		for _, bt := range bonusTypes {
			ids = append(ids, bt.ID)
		}

		if len(ids) > 0 {
			return ids, nil
		}
	}

	defaultTypes := []string{"Participation", "Devoir rendu", "Aide aux camarades"}
	ids := make([]uuid.UUID, 0, len(defaultTypes))
	for _, name := range defaultTypes {
		bt, err := repo.CreateBonusType(ctx, repository.CreateBonusTypeParams{UserID: userID, Name: name})
		if err != nil {
			return nil, fmt.Errorf("failed to create bonus type: %w", err)
		}
		ids = append(ids, bt.ID)
	}

	return ids, nil
}
