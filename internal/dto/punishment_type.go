package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type RequestPunishmentTypeDto struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

type UpdatePunishmentTypeDto struct {
	Name *string `json:"name" validate:"omitempty,min=2,max=100"`
}

type ReturnPunishmentTypeDto struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func PunishmentTypeFromRepository(pt *repository.PunishmentType) *ReturnPunishmentTypeDto {
	if pt == nil {
		return nil
	}

	dto := &ReturnPunishmentTypeDto{
		ID:        pt.ID,
		Name:      pt.Name,
		CreatedAt: pt.CreatedAt,
		UpdatedAt: pt.UpdatedAt,
	}

	return dto
}

func PunishmentTypeListFromRepository(pts []repository.PunishmentType) []*ReturnPunishmentTypeDto {
	dtos := make([]*ReturnPunishmentTypeDto, 0, len(pts))

	for _, pt := range pts {
		if dto := PunishmentTypeFromRepository(&pt); dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
}
