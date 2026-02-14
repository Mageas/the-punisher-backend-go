package seeder

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func createRandomBonusesForStudent(ctx context.Context, repo repository.Querier, userID uuid.UUID, studentID uuid.UUID, bonusTypeIDs []uuid.UUID, maxCount int) error {
	if len(bonusTypeIDs) == 0 {
		return fmt.Errorf("no bonus types available")
	}

	if maxCount <= 0 {
		return nil
	}

	bonusCount := rand.Intn(maxCount) + 1
	for range bonusCount {
		bonusTypeID := bonusTypeIDs[rand.Intn(len(bonusTypeIDs))]
		points := randomBonusPoints()
		_, err := repo.CreateBonus(ctx, repository.CreateBonusParams{
			UserID:      userID,
			StudentID:   studentID,
			BonusTypeID: bonusTypeID,
			Points:      points,
		})
		if err != nil {
			return fmt.Errorf("failed to create bonus: %w", err)
		}
	}

	return nil
}

func randomBonusPoints() float64 {
	steps := []float64{0.5, 1.0, 1.5, 2.0, 2.5, 3.0}
	return steps[rand.Intn(len(steps))]
}
