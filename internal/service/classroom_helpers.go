package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func ensureClassroomExists(ctx context.Context, repo repository.Querier, userID, classroomID uuid.UUID) error {
	if _, err := repo.GetClassroomByUser(ctx, repository.GetClassroomByUserParams{
		ID:     classroomID,
		UserID: userID,
	}); err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return api.ErrClassroomNotFound
		}

		return fmt.Errorf("failed to get classroom: %w", err)
	}

	return nil
}
