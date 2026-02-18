package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func (s *studentService) GetStudentProfile(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, historyLimit int32, historyOffset int32) (*dto.ReturnStudentProfileDto, error) {
	student, err := s.repo.GetStudentByUser(ctx, repository.GetStudentByUserParams{
		ID:     studentID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrStudentNotFound
		}
		return nil, fmt.Errorf("failed to get student: %w", err)
	}

	kpis, err := s.repo.GetStudentProfileKpis(ctx, repository.GetStudentProfileKpisParams{
		StudentID: studentID,
		UserID:    userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get student profile kpis: %w", err)
	}

	classrooms, err := s.repo.ListStudentProfileClassrooms(ctx, repository.ListStudentProfileClassroomsParams{
		StudentID: studentID,
		UserID:    userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list student profile classrooms: %w", err)
	}

	pendingPunishments, err := s.repo.ListStudentProfilePendingPunishments(ctx, repository.ListStudentProfilePendingPunishmentsParams{
		StudentID: studentID,
		UserID:    userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list student profile pending punishments: %w", err)
	}

	availableBonuses, err := s.repo.ListStudentProfileAvailableBonuses(ctx, repository.ListStudentProfileAvailableBonusesParams{
		StudentID: studentID,
		UserID:    userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list student profile available bonuses: %w", err)
	}

	history, err := s.repo.ListStudentProfileHistory(ctx, repository.ListStudentProfileHistoryParams{
		StudentID:   studentID,
		UserID:      userID,
		QueryLimit:  historyLimit,
		QueryOffset: historyOffset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list student profile history: %w", err)
	}

	return dto.StudentProfileFromRows(&student, &kpis, classrooms, pendingPunishments, availableBonuses, history), nil
}
