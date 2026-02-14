package seeder

import (
	"context"
	"fmt"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func createStudent(ctx context.Context, repo repository.Querier, userID uuid.UUID) (repository.Student, error) {
	student, err := repo.CreateStudent(ctx, repository.CreateStudentParams{
		UserID:    userID,
		FirstName: faker.FirstName(),
		LastName:  faker.LastName(),
	})
	if err != nil {
		return repository.Student{}, fmt.Errorf("failed to create student: %w", err)
	}

	return student, nil
}
