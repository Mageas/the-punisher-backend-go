package seeder

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

	for i := 1; i <= defaultClassCount; i++ {
		classroom, err := repo.CreateClassroom(ctx, repository.CreateClassroomParams{
			UserID:      admin.ID,
			Name:        fmt.Sprintf("Classe %d", i),
			Year:        pgtype.Text{String: "2025-2026", Valid: true},
			MainTeacher: pgtype.Text{String: fmt.Sprintf("%s %s", faker.FirstName(), faker.LastName()), Valid: true},
		})
		if err != nil {
			return fmt.Errorf("failed to create classroom: %w", err)
		}

		slog.Info("classroom created", "classroom_id", classroom.ID, "user_id", admin.ID)

		for j := 0; j < defaultStudentsPerClass; j++ {
			student, err := repo.CreateStudent(ctx, repository.CreateStudentParams{
				UserID:    admin.ID,
				FirstName: faker.FirstName(),
				LastName:  faker.LastName(),
			})
			if err != nil {
				return fmt.Errorf("failed to create student: %w", err)
			}

			rowsAffected, err := repo.AddStudentToClassroom(ctx, repository.AddStudentToClassroomParams{
				StudentID:   student.ID,
				ClassroomID: classroom.ID,
				UserID:      admin.ID,
			})
			if err != nil {
				return fmt.Errorf("failed to add student to classroom: %w", err)
			}
			if rowsAffected == 0 {
				return fmt.Errorf("failed to link student to classroom")
			}

			if rand.Float64() < bonusChance {
				bonusCount := rand.Intn(maxBonusesPerStudent) + 1
				for range bonusCount {
					bonusTypeID := bonusTypeIDs[rand.Intn(len(bonusTypeIDs))]
					points := randomBonusPoints()
					_, err := repo.CreateBonus(ctx, repository.CreateBonusParams{
						UserID:      admin.ID,
						StudentID:   student.ID,
						BonusTypeID: bonusTypeID,
						Points:      points,
					})
					if err != nil {
						return fmt.Errorf("failed to create bonus: %w", err)
					}
				}
			}
		}
	}

	slog.Info("education seed completed", "user_id", admin.ID)
	return nil
}

func ensureBonusTypes(ctx context.Context, repo repository.Querier, userID uuid.UUID) ([]uuid.UUID, error) {
	count, err := repo.CountBonusTypesByUser(ctx, userID)
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

func randomBonusPoints() float64 {
	steps := []float64{0.5, 1.0, 1.5, 2.0, 2.5, 3.0}
	return steps[rand.Intn(len(steps))]
}
