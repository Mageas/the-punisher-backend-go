package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

const classroomStudentsPreviewLimit int32 = 5

func attachStudentsPreviewToClassrooms(ctx context.Context, repo repository.Querier, userID uuid.UUID, classrooms []*dto.ReturnClassroomDto) error {
	if len(classrooms) == 0 {
		return nil
	}

	classroomIDs := make([]uuid.UUID, 0, len(classrooms))
	for _, classroom := range classrooms {
		if classroom == nil {
			continue
		}
		classroomIDs = append(classroomIDs, classroom.ID)
	}

	if len(classroomIDs) == 0 {
		return nil
	}

	rows, err := repo.ListStudentsPreviewByClassroomIDs(ctx, repository.ListStudentsPreviewByClassroomIDsParams{
		PreviewLimit: classroomStudentsPreviewLimit,
		UserID:       userID,
		ClassroomIds: classroomIDs,
	})
	if err != nil {
		return err
	}

	previewByClassroom := dto.ClassroomStudentsPreviewByClassroomFromRows(rows)
	for _, classroom := range classrooms {
		if classroom == nil {
			continue
		}

		preview := previewByClassroom[classroom.ID]
		if preview == nil {
			preview = []dto.ClassroomStudentPreviewDto{}
		}
		classroom.StudentsPreview = preview
	}

	return nil
}
