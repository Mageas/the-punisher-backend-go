package sqlcmapper

import (
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func buildReturnClassroomDto(
	id uuid.UUID,
	name string,
	year *string,
	mainTeacher *string,
	studentCount int64,
	createdAt time.Time,
	updatedAt time.Time,
) *dto.ReturnClassroomDto {
	response := &dto.ReturnClassroomDto{
		ID:              id,
		Name:            name,
		StudentCount:    studentCount,
		StudentsPreview: []dto.ClassroomStudentPreviewDto{},
		CreatedAt:       normalizeAPITime(createdAt),
		UpdatedAt:       normalizeAPITime(updatedAt),
	}

	if convertedYear := classroomTextPtr(year); convertedYear != nil {
		response.Year = convertedYear
	}
	if convertedMainTeacher := classroomTextPtr(mainTeacher); convertedMainTeacher != nil {
		response.MainTeacher = convertedMainTeacher
	}

	return response
}

func ClassroomFromCreateRow(c *repository.CreateClassroomRow) *dto.ReturnClassroomDto {
	if c == nil {
		return nil
	}

	return buildReturnClassroomDto(
		c.ID,
		c.Name,
		c.Year,
		c.MainTeacher,
		c.StudentCount,
		c.CreatedAt,
		c.UpdatedAt,
	)
}

func ClassroomFromGetRow(c *repository.GetClassroomByUserRow) *dto.ReturnClassroomDto {
	if c == nil {
		return nil
	}

	return buildReturnClassroomDto(
		c.ID,
		c.Name,
		c.Year,
		c.MainTeacher,
		c.StudentCount,
		c.CreatedAt,
		c.UpdatedAt,
	)
}

func ClassroomFromUpdateRow(c *repository.UpdateClassroomByUserRow) *dto.ReturnClassroomDto {
	if c == nil {
		return nil
	}

	return buildReturnClassroomDto(
		c.ID,
		c.Name,
		c.Year,
		c.MainTeacher,
		c.StudentCount,
		c.CreatedAt,
		c.UpdatedAt,
	)
}

func ClassroomListFromListByUserRows(classrooms []repository.ListClassroomsByUserRow) []*dto.ReturnClassroomDto {
	responses := make([]*dto.ReturnClassroomDto, 0, len(classrooms))

	for _, classroom := range classrooms {
		response := buildReturnClassroomDto(
			classroom.ID,
			classroom.Name,
			classroom.Year,
			classroom.MainTeacher,
			classroom.StudentCount,
			classroom.CreatedAt,
			classroom.UpdatedAt,
		)
		if response != nil {
			responses = append(responses, response)
		}
	}

	return responses
}

func ClassroomListFromListByStudentRows(classrooms []repository.ListClassroomsByStudentRow) []*dto.ReturnClassroomDto {
	responses := make([]*dto.ReturnClassroomDto, 0, len(classrooms))

	for _, classroom := range classrooms {
		response := buildReturnClassroomDto(
			classroom.ID,
			classroom.Name,
			classroom.Year,
			classroom.MainTeacher,
			classroom.StudentCount,
			classroom.CreatedAt,
			classroom.UpdatedAt,
		)
		if response != nil {
			responses = append(responses, response)
		}
	}

	return responses
}

func ClassroomStudentsPreviewByClassroomFromRows(rows []repository.ListStudentsPreviewByClassroomIDsRow) map[uuid.UUID][]dto.ClassroomStudentPreviewDto {
	previewByClassroom := make(map[uuid.UUID][]dto.ClassroomStudentPreviewDto)

	for _, row := range rows {
		previewByClassroom[row.ClassroomID] = append(previewByClassroom[row.ClassroomID], dto.ClassroomStudentPreviewDto{
			ID:        row.StudentID,
			FirstName: row.FirstName,
			LastName:  row.LastName,
		})
	}

	return previewByClassroom
}

func classroomTextPtr(value *string) *string {
	if value == nil {
		return nil
	}

	converted := *value
	return &converted
}
