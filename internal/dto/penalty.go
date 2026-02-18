package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type RequestPenaltyDto struct {
	StudentID     string `json:"student_id" validate:"required,uuid"`
	PenaltyTypeID string `json:"penalty_type_id" validate:"required,uuid"`
}

type ReturnPenaltyDto struct {
	ID               uuid.UUID `json:"id"`
	StudentID        uuid.UUID `json:"student_id"`
	StudentFirstName string    `json:"student_first_name"`
	StudentLastName  string    `json:"student_last_name"`
	PenaltyTypeID    uuid.UUID `json:"penalty_type_id"`
	PenaltyTypeName  string    `json:"penalty_type_name"`
	CreatedAt        time.Time `json:"created_at"`
}

func buildReturnPenaltyDto(
	id uuid.UUID,
	studentID uuid.UUID,
	studentFirstName string,
	studentLastName string,
	penaltyTypeID uuid.UUID,
	penaltyTypeName string,
	createdAt time.Time,
) *ReturnPenaltyDto {
	return &ReturnPenaltyDto{
		ID:               id,
		StudentID:        studentID,
		StudentFirstName: studentFirstName,
		StudentLastName:  studentLastName,
		PenaltyTypeID:    penaltyTypeID,
		PenaltyTypeName:  penaltyTypeName,
		CreatedAt:        createdAt,
	}
}

func PenaltyFromCreateRow(p *repository.CreatePenaltyRow) *ReturnPenaltyDto {
	if p == nil {
		return nil
	}

	return buildReturnPenaltyDto(
		p.ID,
		p.StudentID,
		p.StudentFirstName,
		p.StudentLastName,
		p.PenaltyTypeID,
		p.PenaltyTypeName,
		p.CreatedAt,
	)
}

func PenaltyFromGetRow(p *repository.GetPenaltyByUserRow) *ReturnPenaltyDto {
	if p == nil {
		return nil
	}

	return buildReturnPenaltyDto(
		p.ID,
		p.StudentID,
		p.StudentFirstName,
		p.StudentLastName,
		p.PenaltyTypeID,
		p.PenaltyTypeName,
		p.CreatedAt,
	)
}

func PenaltyListFromListByUserRows(penalties []repository.ListPenaltiesByUserRow) []*ReturnPenaltyDto {
	dtos := make([]*ReturnPenaltyDto, 0, len(penalties))

	for _, penalty := range penalties {
		dto := buildReturnPenaltyDto(
			penalty.ID,
			penalty.StudentID,
			penalty.StudentFirstName,
			penalty.StudentLastName,
			penalty.PenaltyTypeID,
			penalty.PenaltyTypeName,
			penalty.CreatedAt,
		)
		if dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
}

func PenaltyListFromListByStudentRows(penalties []repository.ListPenaltiesByStudentRow) []*ReturnPenaltyDto {
	dtos := make([]*ReturnPenaltyDto, 0, len(penalties))

	for _, penalty := range penalties {
		dto := buildReturnPenaltyDto(
			penalty.ID,
			penalty.StudentID,
			penalty.StudentFirstName,
			penalty.StudentLastName,
			penalty.PenaltyTypeID,
			penalty.PenaltyTypeName,
			penalty.CreatedAt,
		)
		if dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
}
