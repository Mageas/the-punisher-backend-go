package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type RequestStudentDto struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=70"`
	LastName  string `json:"last_name" validate:"required,min=2,max=70"`
}

type UpdateStudentDto struct {
	FirstName *string `json:"first_name" validate:"omitempty,min=2,max=70"`
	LastName  *string `json:"last_name" validate:"omitempty,min=2,max=70"`
}

type StudentClassroomDto struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type ReturnStudentDto struct {
	ID                   uuid.UUID             `json:"id"`
	FirstName            string                `json:"first_name"`
	LastName             string                `json:"last_name"`
	Classrooms           []StudentClassroomDto `json:"classrooms"`
	AvailableBonusPoints float64               `json:"available_bonus_points"`
	PenaltyCount         int64                 `json:"penalty_count"`
	CreatedAt            time.Time             `json:"created_at"`
	UpdatedAt            time.Time             `json:"updated_at"`
}

func buildReturnStudentDto(
	id uuid.UUID,
	firstName string,
	lastName string,
	availableBonusPoints float64,
	penaltyCount int64,
	createdAt time.Time,
	updatedAt time.Time,
) *ReturnStudentDto {
	return &ReturnStudentDto{
		ID:                   id,
		FirstName:            firstName,
		LastName:             lastName,
		Classrooms:           []StudentClassroomDto{},
		AvailableBonusPoints: availableBonusPoints,
		PenaltyCount:         penaltyCount,
		CreatedAt:            createdAt,
		UpdatedAt:            updatedAt,
	}
}

func StudentFromCreateRow(s *repository.CreateStudentRow) *ReturnStudentDto {
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

func StudentFromGetRow(s *repository.GetStudentByUserRow) *ReturnStudentDto {
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

func StudentFromUpdateRow(s *repository.UpdateStudentByUserRow) *ReturnStudentDto {
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

func StudentListFromListByUserRows(students []repository.ListStudentsByUserRow) []*ReturnStudentDto {
	dtos := make([]*ReturnStudentDto, 0, len(students))

	for _, s := range students {
		dto := buildReturnStudentDto(
			s.ID,
			s.FirstName,
			s.LastName,
			s.AvailableBonusPoints,
			s.PenaltyCount,
			s.CreatedAt,
			s.UpdatedAt,
		)
		if dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
}

func StudentListFromListByClassroomRows(students []repository.ListStudentsByClassroomRow) []*ReturnStudentDto {
	dtos := make([]*ReturnStudentDto, 0, len(students))

	for _, s := range students {
		dto := buildReturnStudentDto(
			s.ID,
			s.FirstName,
			s.LastName,
			s.AvailableBonusPoints,
			s.PenaltyCount,
			s.CreatedAt,
			s.UpdatedAt,
		)
		if dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
}

func StudentClassroomsByStudentFromRows(rows []repository.ListClassroomRefsByStudentIDsRow) map[uuid.UUID][]StudentClassroomDto {
	classroomsByStudent := make(map[uuid.UUID][]StudentClassroomDto)

	for _, row := range rows {
		classroomsByStudent[row.StudentID] = append(classroomsByStudent[row.StudentID], StudentClassroomDto{
			ID:   row.ClassroomID,
			Name: row.ClassroomName,
		})
	}

	return classroomsByStudent
}
