package seeder

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func createClassroom(ctx context.Context, repo repository.Querier, userID uuid.UUID, index int) (repository.CreateClassroomRow, error) {
	year := "2025-2026"
	mainTeacher := fmt.Sprintf("%s %s", faker.FirstName(), faker.LastName())

	classroom, err := repo.CreateClassroom(ctx, repository.CreateClassroomParams{
		UserID:      userID,
		Name:        fmt.Sprintf("Classe %d", index),
		Year:        &year,
		MainTeacher: &mainTeacher,
	})
	if err != nil {
		return repository.CreateClassroomRow{}, fmt.Errorf("failed to create classroom: %w", err)
	}

	slog.Info("classroom created", "classroom_id", classroom.ID, "user_id", userID)

	return classroom, nil
}
