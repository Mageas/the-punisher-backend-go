package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type RequestBonusDto struct {
	StudentID   string  `json:"student_id" validate:"required,uuid"`
	BonusTypeID string  `json:"bonus_type_id" validate:"required,uuid"`
	Points      float64 `json:"points" validate:"required,gt=0"`
}

type ReturnBonusDto struct {
	ID          uuid.UUID  `json:"id"`
	StudentID   uuid.UUID  `json:"student_id"`
	BonusTypeID uuid.UUID  `json:"bonus_type_id"`
	Points      float64    `json:"points"`
	CreatedAt   time.Time  `json:"created_at"`
	UsedAt      *time.Time `json:"used_at"`
}

func BonusFromRepository(b *repository.Bonuse) *ReturnBonusDto {
	if b == nil {
		return nil
	}

	dto := &ReturnBonusDto{
		ID:          b.ID,
		StudentID:   b.StudentID,
		BonusTypeID: b.BonusTypeID,
		Points:      b.Points,
		CreatedAt:   b.CreatedAt,
	}

	if usedAt := bonusUsedAt(b.UsedAt); usedAt != nil {
		dto.UsedAt = usedAt
	}

	return dto
}

func BonusListFromRepository(bonuses []repository.Bonuse) []*ReturnBonusDto {
	dtos := make([]*ReturnBonusDto, 0, len(bonuses))

	for _, bonus := range bonuses {
		if dto := BonusFromRepository(&bonus); dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
}

func bonusUsedAt(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	return &value.Time
}
