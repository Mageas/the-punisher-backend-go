package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type RequestPenaltyTypeDto struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

type UpdatePenaltyTypeDto struct {
	Name *string `json:"name" validate:"omitempty,min=2,max=100"`
}

type ReturnPenaltyTypeDto struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func PenaltyTypeFromRepository(pt *repository.PenaltyType) *ReturnPenaltyTypeDto {
	if pt == nil {
		return nil
	}

	dto := &ReturnPenaltyTypeDto{
		ID:        pt.ID,
		Name:      pt.Name,
		CreatedAt: pt.CreatedAt,
		UpdatedAt: pt.UpdatedAt,
	}

	return dto
}

func PenaltyTypeListFromRepository(pts []repository.PenaltyType) []*ReturnPenaltyTypeDto {
	dtos := make([]*ReturnPenaltyTypeDto, 0, len(pts))

	for _, pt := range pts {
		if dto := PenaltyTypeFromRepository(&pt); dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
}
