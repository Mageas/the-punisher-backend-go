package seeder

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func createRandomPenaltiesForStudent(ctx context.Context, repo repository.Querier, userID uuid.UUID, studentID uuid.UUID, penaltyTypeIDs []uuid.UUID, maxCount int) error {
	if len(penaltyTypeIDs) == 0 {
		return fmt.Errorf("no penalty types available")
	}

	if maxCount <= 0 {
		return nil
	}

	penaltyCount := rand.Intn(maxCount) + 1
	for range penaltyCount {
		penaltyTypeID := penaltyTypeIDs[rand.Intn(len(penaltyTypeIDs))]
		_, err := repo.CreatePenalty(ctx, repository.CreatePenaltyParams{
			UserID:        userID,
			StudentID:     studentID,
			PenaltyTypeID: penaltyTypeID,
		})
		if err != nil {
			return fmt.Errorf("failed to create penalty: %w", err)
		}
	}

	return nil
}
