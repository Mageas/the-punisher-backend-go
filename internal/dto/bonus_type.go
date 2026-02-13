package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type RequestBonusTypeDto struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

type UpdateBonusTypeDto struct {
	Name *string `json:"name" validate:"omitempty,min=2,max=100"`
}

type ReturnBonusTypeDto struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func BonusTypeFromRepository(bt *repository.BonusType) *ReturnBonusTypeDto {
	if bt == nil {
		return nil
	}

	dto := &ReturnBonusTypeDto{
		ID:        bt.ID,
		Name:      bt.Name,
		CreatedAt: bt.CreatedAt,
		UpdatedAt: bt.UpdatedAt,
	}

	return dto
}

func BonusTypeListFromRepository(bts []repository.BonusType) []*ReturnBonusTypeDto {
	dtos := make([]*ReturnBonusTypeDto, 0, len(bts))

	for _, bt := range bts {
		if dto := BonusTypeFromRepository(&bt); dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
}
