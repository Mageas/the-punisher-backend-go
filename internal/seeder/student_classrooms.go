package seeder

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func linkStudentToClassroom(ctx context.Context, repo repository.Querier, userID uuid.UUID, studentID uuid.UUID, classroomID uuid.UUID) error {
	rowsAffected, err := repo.AddStudentToClassroom(ctx, repository.AddStudentToClassroomParams{
		StudentID:   studentID,
		ClassroomID: classroomID,
		UserID:      userID,
	})
	if err != nil {
		return fmt.Errorf("failed to add student to classroom: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("failed to link student to classroom")
	}

	return nil
}
