package sqlcmapper

import (
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func buildReturnPenaltyDto(
	id uuid.UUID,
	studentID uuid.UUID,
	studentFirstName string,
	studentLastName string,
	penaltyTypeID uuid.UUID,
	penaltyTypeName string,
	createdAt time.Time,
	occurredAt time.Time,
	evaluationLabel *string,
) *dto.ReturnPenaltyDto {
	return &dto.ReturnPenaltyDto{
		ID:               id,
		StudentID:        studentID,
		StudentFirstName: studentFirstName,
		StudentLastName:  studentLastName,
		PenaltyTypeID:    penaltyTypeID,
		PenaltyTypeName:  penaltyTypeName,
		CreatedAt:        createdAt,
		OccurredAt:       occurredAt,
		EvaluationLabel:  bonusEvaluationLabel(evaluationLabel),
	}
}

func PenaltyFromCreateRow(p *repository.CreatePenaltyRow) *dto.ReturnPenaltyDto {
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
		p.OccurredAt,
		p.EvaluationLabel,
	)
}

func PenaltyFromGetRow(p *repository.GetPenaltyByUserRow) *dto.ReturnPenaltyDto {
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
		p.OccurredAt,
		p.EvaluationLabel,
	)
}

func PenaltyListFromListByUserRows(penalties []repository.ListPenaltiesByUserRow) []*dto.ReturnPenaltyDto {
	responses := make([]*dto.ReturnPenaltyDto, 0, len(penalties))

	for _, penalty := range penalties {
		response := buildReturnPenaltyDto(
			penalty.ID,
			penalty.StudentID,
			penalty.StudentFirstName,
			penalty.StudentLastName,
			penalty.PenaltyTypeID,
			penalty.PenaltyTypeName,
			penalty.CreatedAt,
			penalty.OccurredAt,
			penalty.EvaluationLabel,
		)
		if response != nil {
			responses = append(responses, response)
		}
	}

	return responses
}

func PenaltyListFromListByStudentRows(penalties []repository.ListPenaltiesByStudentRow) []*dto.ReturnPenaltyDto {
	responses := make([]*dto.ReturnPenaltyDto, 0, len(penalties))

	for _, penalty := range penalties {
		response := buildReturnPenaltyDto(
			penalty.ID,
			penalty.StudentID,
			penalty.StudentFirstName,
			penalty.StudentLastName,
			penalty.PenaltyTypeID,
			penalty.PenaltyTypeName,
			penalty.CreatedAt,
			penalty.OccurredAt,
			penalty.EvaluationLabel,
		)
		if response != nil {
			responses = append(responses, response)
		}
	}

	return responses
}

func PenaltyFromUpdateRow(p *repository.UpdatePenaltyByUserRow) *dto.ReturnPenaltyDto {
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
		p.OccurredAt,
		p.EvaluationLabel,
	)
}
