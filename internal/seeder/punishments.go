package seeder

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func createRandomPunishmentsForStudent(ctx context.Context, repo repository.Querier, userID uuid.UUID, studentID uuid.UUID, punishmentTypeIDs []uuid.UUID, maxCount int) error {
	if len(punishmentTypeIDs) == 0 {
		return fmt.Errorf("no punishment types available")
	}

	if maxCount <= 0 {
		return nil
	}

	punishmentCount := rand.Intn(maxCount) + 1
	for range punishmentCount {
		punishmentTypeID := punishmentTypeIDs[rand.Intn(len(punishmentTypeIDs))]
		dueAt := randomFutureDueAt()
		_, err := repo.CreatePunishment(ctx, repository.CreatePunishmentParams{
			UserID:           userID,
			StudentID:        studentID,
			PunishmentTypeID: punishmentTypeID,
			DueAt:            dueAt,
		})
		if err != nil {
			return fmt.Errorf("failed to create punishment: %w", err)
		}
	}

	return nil
}

func randomFutureDueAt() time.Time {
	days := rand.Intn(14) + 1
	hours := rand.Intn(9) + 8
	return time.Now().AddDate(0, 0, days).Truncate(time.Hour).Add(time.Duration(hours) * time.Hour)
}
