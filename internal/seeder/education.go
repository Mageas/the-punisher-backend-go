package seeder

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"

	"github.com/mageas/the-punisher-backend/internal/repository"
)

const (
	defaultClassCount       = 3
	defaultStudentsPerClass = 20
	bonusChance             = 0.30
	maxBonusesPerStudent    = 3
)

func EducationSeed(ctx context.Context, repo repository.Querier) error {
	admin, err := repo.GetUserCredentialsByEmailForAuth(ctx, adminEmail)
	if err != nil {
		return fmt.Errorf("failed to load admin user: %w", err)
	}

	bonusTypeIDs, err := ensureBonusTypes(ctx, repo, admin.ID)
	if err != nil {
		return err
	}

	if _, err := ensurePenaltyTypes(ctx, repo, admin.ID); err != nil {
		return err
	}

	for i := 1; i <= defaultClassCount; i++ {
		classroom, err := createClassroom(ctx, repo, admin.ID, i)
		if err != nil {
			return err
		}

		for j := 0; j < defaultStudentsPerClass; j++ {
			student, err := createStudent(ctx, repo, admin.ID)
			if err != nil {
				return err
			}

			if err := linkStudentToClassroom(ctx, repo, admin.ID, student.ID, classroom.ID); err != nil {
				return err
			}

			if rand.Float64() < bonusChance {
				if err := createRandomBonusesForStudent(ctx, repo, admin.ID, student.ID, bonusTypeIDs, maxBonusesPerStudent); err != nil {
					return err
				}
			}
		}
	}

	slog.Info("education seed completed", "user_id", admin.ID)
	return nil
}
