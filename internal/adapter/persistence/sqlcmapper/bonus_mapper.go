package sqlcmapper

import (
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func buildReturnBonusDto(
	id uuid.UUID,
	studentID uuid.UUID,
	studentFirstName string,
	studentLastName string,
	bonusTypeID uuid.UUID,
	bonusTypeName string,
	points float64,
	createdAt time.Time,
	occurredAt time.Time,
	evaluationLabel string,
	usedAt *time.Time,
) *dto.ReturnBonusDto {
	response := &dto.ReturnBonusDto{
		ID:               id,
		StudentID:        studentID,
		StudentFirstName: studentFirstName,
		StudentLastName:  studentLastName,
		BonusTypeID:      bonusTypeID,
		BonusTypeName:    bonusTypeName,
		Points:           points,
		CreatedAt:        normalizeAPITime(createdAt),
		OccurredAt:       normalizeAPITime(occurredAt),
		EvaluationLabel:  evaluationLabel,
	}

	if convertedUsedAt := bonusUsedAt(usedAt); convertedUsedAt != nil {
		response.UsedAt = convertedUsedAt
	}

	return response
}

func BonusFromCreateRow(b *repository.CreateBonusRow) *dto.ReturnBonusDto {
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
		b.OccurredAt,
		b.EvaluationLabel,
		b.UsedAt,
	)
}

func BonusFromGetRow(b *repository.GetBonusByUserRow) *dto.ReturnBonusDto {
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
		b.OccurredAt,
		b.EvaluationLabel,
		b.UsedAt,
	)
}

func BonusListFromListByUserRows(bonuses []repository.ListBonusesByUserRow) []*dto.ReturnBonusDto {
	responses := make([]*dto.ReturnBonusDto, 0, len(bonuses))

	for _, bonus := range bonuses {
		response := buildReturnBonusDto(
			bonus.ID,
			bonus.StudentID,
			bonus.StudentFirstName,
			bonus.StudentLastName,
			bonus.BonusTypeID,
			bonus.BonusTypeName,
			bonus.Points,
			bonus.CreatedAt,
			bonus.OccurredAt,
			bonus.EvaluationLabel,
			bonus.UsedAt,
		)
		if response != nil {
			responses = append(responses, response)
		}
	}

	return responses
}

func BonusListFromListByStudentRows(bonuses []repository.ListBonusesByStudentRow) []*dto.ReturnBonusDto {
	responses := make([]*dto.ReturnBonusDto, 0, len(bonuses))

	for _, bonus := range bonuses {
		response := buildReturnBonusDto(
			bonus.ID,
			bonus.StudentID,
			bonus.StudentFirstName,
			bonus.StudentLastName,
			bonus.BonusTypeID,
			bonus.BonusTypeName,
			bonus.Points,
			bonus.CreatedAt,
			bonus.OccurredAt,
			bonus.EvaluationLabel,
			bonus.UsedAt,
		)
		if response != nil {
			responses = append(responses, response)
		}
	}

	return responses
}

func BonusFromUseRow(b *repository.UseBonusRow) *dto.ReturnBonusDto {
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
		b.OccurredAt,
		b.EvaluationLabel,
		b.UsedAt,
	)
}

func BonusFromUpdateRow(b *repository.UpdateBonusByUserRow) *dto.ReturnBonusDto {
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
		b.OccurredAt,
		b.EvaluationLabel,
		b.UsedAt,
	)
}

func bonusUsedAt(value *time.Time) *time.Time {
	return normalizeOptionalAPITime(value)
}
