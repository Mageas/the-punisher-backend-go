package seeder

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func createClassroom(ctx context.Context, repo repository.Querier, userID uuid.UUID, index int) (repository.Classroom, error) {
	classroom, err := repo.CreateClassroom(ctx, repository.CreateClassroomParams{
		UserID:      userID,
		Name:        fmt.Sprintf("Classe %d", index),
		Year:        pgtype.Text{String: "2025-2026", Valid: true},
		MainTeacher: pgtype.Text{String: fmt.Sprintf("%s %s", faker.FirstName(), faker.LastName()), Valid: true},
	})
	if err != nil {
		return repository.Classroom{}, fmt.Errorf("failed to create classroom: %w", err)
	}

	slog.Info("classroom created", "classroom_id", classroom.ID, "user_id", userID)

	return classroom, nil
}
