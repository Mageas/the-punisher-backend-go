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
	ID            uuid.UUID `json:"id"`
	StudentID     uuid.UUID `json:"student_id"`
	PenaltyTypeID uuid.UUID `json:"penalty_type_id"`
	CreatedAt     time.Time `json:"created_at"`
}

func PenaltyFromRepository(p *repository.Penalty) *ReturnPenaltyDto {
	if p == nil {
		return nil
	}

	dto := &ReturnPenaltyDto{
		ID:            p.ID,
		StudentID:     p.StudentID,
		PenaltyTypeID: p.PenaltyTypeID,
		CreatedAt:     p.CreatedAt,
	}

	return dto
}

func PenaltyListFromRepository(penalties []repository.Penalty) []*ReturnPenaltyDto {
	dtos := make([]*ReturnPenaltyDto, 0, len(penalties))

	for _, penalty := range penalties {
		if dto := PenaltyFromRepository(&penalty); dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
}
