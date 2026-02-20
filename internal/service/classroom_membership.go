package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func (s *classroomService) AddStudentToClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID, studentID uuid.UUID) error {
	rowsAffected, err := s.repo.AddStudentToClassroom(ctx, repository.AddStudentToClassroomParams{
		StudentID:   studentID,
		ClassroomID: classroomID,
		UserID:      userID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return api.ErrStudentClassroomRelationExists
		}
		return fmt.Errorf("failed to add student to classroom: %w", err)
	}

	if rowsAffected == 0 {
		return api.ErrStudentOrClassroomNotFound
	}

	slog.Info("student added to classroom", "student_id", studentID, "classroom_id", classroomID, "user_id", userID)

	return nil
}

func (s *classroomService) RemoveStudentFromClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID, studentID uuid.UUID) error {
	rowsAffected, err := s.repo.RemoveStudentFromClassroom(ctx, repository.RemoveStudentFromClassroomParams{
		StudentID:   studentID,
		ClassroomID: classroomID,
		UserID:      userID,
	})
	if err != nil {
		return fmt.Errorf("failed to remove student from classroom: %w", err)
	}

	if rowsAffected == 0 {
		return api.ErrStudentOrClassroomNotFound
	}

	slog.Info("student removed from classroom", "student_id", studentID, "classroom_id", classroomID, "user_id", userID)

	return nil
}

func (s *classroomService) ListStudentsByClassroom(ctx context.Context, userID uuid.UUID, classroomID uuid.UUID, limit int32, offset int32) ([]*dto.ReturnStudentDto, int64, error) {
	_, err := s.repo.GetClassroomByUser(ctx, repository.GetClassroomByUserParams{
		ID:     classroomID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, api.ErrClassroomNotFound
		}
		return nil, 0, fmt.Errorf("failed to get classroom: %w", err)
	}

	totalCount, err := s.repo.CountStudentsByClassroom(ctx, repository.CountStudentsByClassroomParams{
		ClassroomID: classroomID,
		UserID:      userID,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count students by classroom: %w", err)
	}

	students, err := s.repo.ListStudentsByClassroom(ctx, repository.ListStudentsByClassroomParams{
		ClassroomID: classroomID,
		UserID:      userID,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list students by classroom: %w", err)
	}

	response := dto.StudentListFromListByClassroomRows(students)
	if err := attachClassroomsToStudents(ctx, s.repo, userID, response); err != nil {
		return nil, 0, fmt.Errorf("failed to list student classrooms: %w", err)
	}

	return response, totalCount, nil
}

func (s *classroomService) ListClassroomsByStudent(ctx context.Context, userID uuid.UUID, studentID uuid.UUID, limit int32, offset int32) ([]*dto.ReturnClassroomDto, int64, error) {
	_, err := s.repo.GetStudentByUser(ctx, repository.GetStudentByUserParams{
		ID:     studentID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, api.ErrStudentNotFound
		}
		return nil, 0, fmt.Errorf("failed to get student: %w", err)
	}

	totalCount, err := s.repo.CountClassroomsByStudent(ctx, repository.CountClassroomsByStudentParams{
		StudentID: studentID,
		UserID:    userID,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count classrooms by student: %w", err)
	}

	classrooms, err := s.repo.ListClassroomsByStudent(ctx, repository.ListClassroomsByStudentParams{
		StudentID:   studentID,
		UserID:      userID,
		QueryLimit:  limit,
		QueryOffset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list classrooms by student: %w", err)
	}

	response := dto.ClassroomListFromListByStudentRows(classrooms)
	if err := attachStudentsPreviewToClassrooms(ctx, s.repo, userID, response); err != nil {
		return nil, 0, fmt.Errorf("failed to list classroom students preview: %w", err)
	}

	return response, totalCount, nil
}
