package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type RequestBonusDto struct {
	StudentID   string  `json:"student_id" validate:"required,uuid"`
	BonusTypeID string  `json:"bonus_type_id" validate:"required,uuid"`
	Points      float64 `json:"points" validate:"required,gt=0"`
}

type ReturnBonusDto struct {
	ID               uuid.UUID  `json:"id"`
	StudentID        uuid.UUID  `json:"student_id"`
	StudentFirstName string     `json:"student_first_name"`
	StudentLastName  string     `json:"student_last_name"`
	BonusTypeID      uuid.UUID  `json:"bonus_type_id"`
	BonusTypeName    string     `json:"bonus_type_name"`
	Points           float64    `json:"points"`
	CreatedAt        time.Time  `json:"created_at"`
	UsedAt           *time.Time `json:"used_at"`
}

func buildReturnBonusDto(
	id uuid.UUID,
	studentID uuid.UUID,
	studentFirstName string,
	studentLastName string,
	bonusTypeID uuid.UUID,
	bonusTypeName string,
	points float64,
	createdAt time.Time,
	usedAt **time.Time,
) *ReturnBonusDto {
	dto := &ReturnBonusDto{
		ID:               id,
		StudentID:        studentID,
		StudentFirstName: studentFirstName,
		StudentLastName:  studentLastName,
		BonusTypeID:      bonusTypeID,
		BonusTypeName:    bonusTypeName,
		Points:           points,
		CreatedAt:        createdAt,
	}

	if convertedUsedAt := bonusUsedAt(usedAt); convertedUsedAt != nil {
		dto.UsedAt = convertedUsedAt
	}

	return dto
}

func BonusFromCreateRow(b *repository.CreateBonusRow) *ReturnBonusDto {
	if b == nil {
		return nil
	}

	return buildReturnBonusDto(
		b.ID,
		b.StudentID,
		b.StudentFirstName,
		b.StudentLastName,
		b.BonusTypeID,
		b.BonusTypeName,
		b.Points,
		b.CreatedAt,
		b.UsedAt,
	)
}

func BonusFromGetRow(b *repository.GetBonusByUserRow) *ReturnBonusDto {
	if b == nil {
		return nil
	}

	return buildReturnBonusDto(
		b.ID,
		b.StudentID,
		b.StudentFirstName,
		b.StudentLastName,
		b.BonusTypeID,
		b.BonusTypeName,
		b.Points,
		b.CreatedAt,
		b.UsedAt,
	)
}

func BonusListFromListByUserRows(bonuses []repository.ListBonusesByUserRow) []*ReturnBonusDto {
	dtos := make([]*ReturnBonusDto, 0, len(bonuses))

	for _, bonus := range bonuses {
		dto := buildReturnBonusDto(
			bonus.ID,
			bonus.StudentID,
			bonus.StudentFirstName,
			bonus.StudentLastName,
			bonus.BonusTypeID,
			bonus.BonusTypeName,
			bonus.Points,
			bonus.CreatedAt,
			bonus.UsedAt,
		)
		if dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
}

func BonusListFromListByStudentRows(bonuses []repository.ListBonusesByStudentRow) []*ReturnBonusDto {
	dtos := make([]*ReturnBonusDto, 0, len(bonuses))

	for _, bonus := range bonuses {
		dto := buildReturnBonusDto(
			bonus.ID,
			bonus.StudentID,
			bonus.StudentFirstName,
			bonus.StudentLastName,
			bonus.BonusTypeID,
			bonus.BonusTypeName,
			bonus.Points,
			bonus.CreatedAt,
			bonus.UsedAt,
		)
		if dto != nil {
			dtos = append(dtos, dto)
		}
	}

	return dtos
}

func BonusFromUseRow(b *repository.UseBonusRow) *ReturnBonusDto {
	if b == nil {
		return nil
	}

	return buildReturnBonusDto(
		b.ID,
		b.StudentID,
		b.StudentFirstName,
		b.StudentLastName,
		b.BonusTypeID,
		b.BonusTypeName,
		b.Points,
		b.CreatedAt,
		b.UsedAt,
	)
}

func bonusUsedAt(value **time.Time) *time.Time {
	if value == nil || *value == nil {
		return nil
	}

	return *value
}
