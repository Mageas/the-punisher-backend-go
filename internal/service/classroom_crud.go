package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mageas/the-punisher-backend/internal/adapter/persistence/sqlcmapper"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func (s *classroomService) CreateClassroom(ctx context.Context, userID uuid.UUID, req dto.RequestClassroomDto) (*dto.ReturnClassroomDto, error) {
	params := repository.CreateClassroomParams{
		UserID: userID,
		Name:   req.Name,
	}

	if req.Year != nil {
		params.Year = req.Year
	}
	if req.MainTeacher != nil {
		params.MainTeacher = req.MainTeacher
	}

	classroom, err := s.repo.CreateClassroom(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create classroom: %w", err)
	}

	slog.Info("classroom created", "classroom_id", classroom.ID, "user_id", userID)

	response := sqlcmapper.ClassroomFromCreateRow(&classroom)
	if err := attachStudentsPreviewToClassrooms(ctx, s.repo, userID, []*dto.ReturnClassroomDto{response}); err != nil {
		return nil, fmt.Errorf("failed to list classroom students preview: %w", err)
	}

	return response, nil
}

func (s *classroomService) GetClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID) (*dto.ReturnClassroomDto, error) {
	classroom, err := s.repo.GetClassroomByUser(ctx, repository.GetClassroomByUserParams{
		ID:     classroomID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrClassroomNotFound
		}
		return nil, fmt.Errorf("failed to get classroom: %w", err)
	}

	response := sqlcmapper.ClassroomFromGetRow(&classroom)
	if err := attachStudentsPreviewToClassrooms(ctx, s.repo, userID, []*dto.ReturnClassroomDto{response}); err != nil {
		return nil, fmt.Errorf("failed to list classroom students preview: %w", err)
	}

	return response, nil
}

func (s *classroomService) ListClassrooms(ctx context.Context, userID uuid.UUID, limit int32, offset int32) ([]*dto.ReturnClassroomDto, int64, error) {
	totalCount, err := s.repo.CountClassroomsByUser(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count classrooms: %w", err)
	}

	classrooms, err := s.repo.ListClassroomsByUser(ctx, repository.ListClassroomsByUserParams{
		UserID:      userID,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list classrooms: %w", err)
	}

	response := sqlcmapper.ClassroomListFromListByUserRows(classrooms)
	if err := attachStudentsPreviewToClassrooms(ctx, s.repo, userID, response); err != nil {
		return nil, 0, fmt.Errorf("failed to list classroom students preview: %w", err)
	}

	return response, totalCount, nil
}

func (s *classroomService) UpdateClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID, req dto.UpdateClassroomDto) (*dto.ReturnClassroomDto, error) {
	params := repository.UpdateClassroomByUserParams{
		ID:     classroomID,
		UserID: userID,
	}

	if req.Name != nil {
		params.Name = req.Name
	}
	if req.Year != nil {
		params.Year = req.Year
	}
	if req.MainTeacher != nil {
		params.MainTeacher = req.MainTeacher
	}

	classroom, err := s.repo.UpdateClassroomByUser(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrClassroomNotFound
		}
		return nil, fmt.Errorf("failed to update classroom: %w", err)
	}

	response := sqlcmapper.ClassroomFromUpdateRow(&classroom)
	if err := attachStudentsPreviewToClassrooms(ctx, s.repo, userID, []*dto.ReturnClassroomDto{response}); err != nil {
		return nil, fmt.Errorf("failed to list classroom students preview: %w", err)
	}

	return response, nil
}

func (s *classroomService) DeleteClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID) error {
	rowsAffected, err := s.repo.DeleteClassroomByUser(ctx, repository.DeleteClassroomByUserParams{
		ID:     classroomID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete classroom: %w", err)
	}

	if rowsAffected == 0 {
		return api.ErrClassroomNotFound
	}

	slog.Info("classroom deleted", "classroom_id", classroomID, "user_id", userID)

	return nil
}
