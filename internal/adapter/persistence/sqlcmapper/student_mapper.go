package sqlcmapper

import (
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func buildReturnStudentDto(
	id uuid.UUID,
	firstName string,
	lastName string,
	availableBonusPoints float64,
	penaltyCount int64,
	createdAt time.Time,
	updatedAt time.Time,
) *dto.ReturnStudentDto {
	return &dto.ReturnStudentDto{
		ID:                   id,
		FirstName:            firstName,
		LastName:             lastName,
		Classrooms:           []dto.StudentClassroomDto{},
		AvailableBonusPoints: availableBonusPoints,
		PenaltyCount:         penaltyCount,
		CreatedAt:            createdAt,
		UpdatedAt:            updatedAt,
	}
}

func StudentFromCreateRow(s *repository.CreateStudentRow) *dto.ReturnStudentDto {
	if s == nil {
		return nil
	}

	return buildReturnStudentDto(
		s.ID,
		s.FirstName,
		s.LastName,
		s.AvailableBonusPoints,
		s.PenaltyCount,
		s.CreatedAt,
		s.UpdatedAt,
	)
}

func StudentFromGetRow(s *repository.GetStudentByUserRow) *dto.ReturnStudentDto {
	if s == nil {
		return nil
	}

	return buildReturnStudentDto(
		s.ID,
		s.FirstName,
		s.LastName,
		s.AvailableBonusPoints,
		s.PenaltyCount,
		s.CreatedAt,
		s.UpdatedAt,
	)
}

func StudentFromUpdateRow(s *repository.UpdateStudentByUserRow) *dto.ReturnStudentDto {
	if s == nil {
		return nil
	}

	return buildReturnStudentDto(
		s.ID,
		s.FirstName,
		s.LastName,
		s.AvailableBonusPoints,
		s.PenaltyCount,
		s.CreatedAt,
		s.UpdatedAt,
	)
}

func StudentListFromListByUserRows(students []repository.ListStudentsByUserRow) []*dto.ReturnStudentDto {
	responses := make([]*dto.ReturnStudentDto, 0, len(students))

	for _, student := range students {
		response := buildReturnStudentDto(
			student.ID,
			student.FirstName,
			student.LastName,
			student.AvailableBonusPoints,
			student.PenaltyCount,
			student.CreatedAt,
			student.UpdatedAt,
		)
		if response != nil {
			responses = append(responses, response)
		}
	}

	return responses
}

func StudentListFromListByClassroomRows(students []repository.ListStudentsByClassroomRow) []*dto.ReturnStudentDto {
	responses := make([]*dto.ReturnStudentDto, 0, len(students))

	for _, student := range students {
		response := buildReturnStudentDto(
			student.ID,
			student.FirstName,
			student.LastName,
			student.AvailableBonusPoints,
			student.PenaltyCount,
			student.CreatedAt,
			student.UpdatedAt,
		)
		if response != nil {
			responses = append(responses, response)
		}
	}

	return responses
}

func StudentClassroomsByStudentFromRows(rows []repository.ListClassroomRefsByStudentIDsRow) map[uuid.UUID][]dto.StudentClassroomDto {
	classroomsByStudent := make(map[uuid.UUID][]dto.StudentClassroomDto)

	for _, row := range rows {
		classroomsByStudent[row.StudentID] = append(classroomsByStudent[row.StudentID], dto.StudentClassroomDto{
			ID:   row.ClassroomID,
			Name: row.ClassroomName,
		})
	}

	return classroomsByStudent
}
