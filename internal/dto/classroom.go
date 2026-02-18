package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type RequestClassroomDto struct {
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Year        *string `json:"year" validate:"omitempty,max=20"`
	MainTeacher *string `json:"main_teacher" validate:"omitempty,max=100"`
}

type UpdateClassroomDto struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=100"`
	Year        *string `json:"year" validate:"omitempty,max=20"`
	MainTeacher *string `json:"main_teacher" validate:"omitempty,max=100"`
}

type ReturnClassroomDto struct {
	ID                uuid.UUID                    `json:"id"`
	Name              string                       `json:"name"`
	Year              *string                      `json:"year"`
	MainTeacher       *string                      `json:"main_teacher"`
	StudentCount      int64                        `json:"student_count"`
	StudentsPreview   []ClassroomStudentPreviewDto `json:"students_preview"`
	TotalBonusPoints  float64                      `json:"total_bonus_points"`
	TotalPenaltyCount int64                        `json:"total_penalty_count"`
	CreatedAt         time.Time                    `json:"created_at"`
	UpdatedAt         time.Time                    `json:"updated_at"`
}

type StudentClassroomRequestDto struct {
	StudentID string `json:"student_id" validate:"required,uuid"`
}

type ClassroomStudentPreviewDto struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
}

func buildReturnClassroomDto(
	id uuid.UUID,
	name string,
	year pgtype.Text,
	mainTeacher pgtype.Text,
	studentCount int64,
	totalBonusPoints float64,
	totalPenaltyCount int64,
	createdAt time.Time,
	updatedAt time.Time,
) *ReturnClassroomDto {
	dto := &ReturnClassroomDto{
		ID:                id,
		Name:              name,
		StudentCount:      studentCount,
		StudentsPreview:   []ClassroomStudentPreviewDto{},
		TotalBonusPoints:  totalBonusPoints,
		TotalPenaltyCount: totalPenaltyCount,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
	}

	if convertedYear := classroomTextPtr(year); convertedYear != nil {
		dto.Year = convertedYear
	}
	if convertedMainTeacher := classroomTextPtr(mainTeacher); convertedMainTeacher != nil {
		dto.MainTeacher = convertedMainTeacher
	}

	return dto
}

func ClassroomFromCreateRow(c *repository.CreateClassroomRow) *ReturnClassroomDto {
	if c == nil {
		return nil
	}

	return buildReturnClassroomDto(
		c.ID,
		c.Name,
		c.Year,
		c.MainTeacher,
		c.StudentCount,
		c.TotalBonusPoints,
		c.TotalPenaltyCount,
		c.CreatedAt,
		c.UpdatedAt,
	)
}

func ClassroomFromGetRow(c *repository.GetClassroomByUserRow) *ReturnClassroomDto {
	if c == nil {
		return nil
	}

	return buildReturnClassroomDto(
		c.ID,
		c.Name,
		c.Year,
		c.MainTeacher,
		c.StudentCount,
		c.TotalBonusPoints,
		c.TotalPenaltyCount,
		c.CreatedAt,
		c.UpdatedAt,
	)
}

func ClassroomFromUpdateRow(c *repository.UpdateClassroomByUserRow) *ReturnClassroomDto {
	if c == nil {
		return nil
	}

	return buildReturnClassroomDto(
		c.ID,
		c.Name,
		c.Year,
		c.MainTeacher,
		c.StudentCount,
		c.TotalBonusPoints,
		c.TotalPenaltyCount,
		c.CreatedAt,
		c.UpdatedAt,
	)
}

func ClassroomListFromListByUserRows(classrooms []repository.ListClassroomsByUserRow) []*ReturnClassroomDto {
	dtos := make([]*ReturnClassroomDto, 0, len(classrooms))

	for _, c := range classrooms {
		dto := buildReturnClassroomDto(
			c.ID,
			c.Name,
			c.Year,
			c.MainTeacher,
			c.StudentCount,
			c.TotalBonusPoints,
			c.TotalPenaltyCount,
			c.CreatedAt,
			c.UpdatedAt,
		)
		if dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
}

func ClassroomListFromListByStudentRows(classrooms []repository.ListClassroomsByStudentRow) []*ReturnClassroomDto {
	dtos := make([]*ReturnClassroomDto, 0, len(classrooms))

	for _, c := range classrooms {
		dto := buildReturnClassroomDto(
			c.ID,
			c.Name,
			c.Year,
			c.MainTeacher,
			c.StudentCount,
			c.TotalBonusPoints,
			c.TotalPenaltyCount,
			c.CreatedAt,
			c.UpdatedAt,
		)
		if dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
}

func ClassroomStudentsPreviewByClassroomFromRows(rows []repository.ListStudentsPreviewByClassroomIDsRow) map[uuid.UUID][]ClassroomStudentPreviewDto {
	previewByClassroom := make(map[uuid.UUID][]ClassroomStudentPreviewDto)

	for _, row := range rows {
		previewByClassroom[row.ClassroomID] = append(previewByClassroom[row.ClassroomID], ClassroomStudentPreviewDto{
			ID:        row.StudentID,
			FirstName: row.FirstName,
			LastName:  row.LastName,
		})
	}

	return previewByClassroom
}

func classroomTextPtr(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}

	converted := value.String
	return &converted
}
