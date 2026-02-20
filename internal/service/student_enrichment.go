package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/adapter/persistence/sqlcmapper"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func attachClassroomsToStudents(ctx context.Context, repo repository.Querier, userID uuid.UUID, students []*dto.ReturnStudentDto) error {
	if len(students) == 0 {
		return nil
	}

	studentIDs := make([]uuid.UUID, 0, len(students))
	for _, student := range students {
		if student == nil {
			continue
		}
		studentIDs = append(studentIDs, student.ID)
	}

	if len(studentIDs) == 0 {
		return nil
	}

	rows, err := repo.ListClassroomRefsByStudentIDs(ctx, repository.ListClassroomRefsByStudentIDsParams{
		UserID:     userID,
		StudentIds: studentIDs,
	})
	if err != nil {
		return err
	}

	classroomsByStudent := sqlcmapper.StudentClassroomsByStudentFromRows(rows)
	for _, student := range students {
		if student == nil {
			continue
		}

		classrooms := classroomsByStudent[student.ID]
		if classrooms == nil {
			classrooms = []dto.StudentClassroomDto{}
		}
		student.Classrooms = classrooms
	}

	return nil
}
