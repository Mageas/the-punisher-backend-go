package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

var studentProfileHistoryTimeSentinel = time.Unix(0, 0).UTC()

type StudentProfileStudentDto struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type StudentProfileKpisDto struct {
	AvailableBonusPoints   float64 `json:"available_bonus_points"`
	ActiveBonusCount       int64   `json:"active_bonus_count"`
	TotalPenaltyCount      int64   `json:"total_penalty_count"`
	PendingPunishmentCount int64   `json:"pending_punishment_count"`
}

type StudentProfilePendingPunishmentDto struct {
	ID                 uuid.UUID  `json:"id"`
	PunishmentTypeID   uuid.UUID  `json:"punishment_type_id"`
	PunishmentTypeName string     `json:"punishment_type_name"`
	TriggeringRuleID   *uuid.UUID `json:"triggering_rule_id"`
	TriggeringRuleName *string    `json:"triggering_rule_name"`
	DueAt              time.Time  `json:"due_at"`
	CreatedAt          time.Time  `json:"created_at"`
}

type StudentProfileAvailableBonusDto struct {
	ID            uuid.UUID `json:"id"`
	BonusTypeID   uuid.UUID `json:"bonus_type_id"`
	BonusTypeName string    `json:"bonus_type_name"`
	Points        float64   `json:"points"`
	CreatedAt     time.Time `json:"created_at"`
}

type StudentProfileHistoryItemDto struct {
	Type               string     `json:"type"`
	ID                 uuid.UUID  `json:"id"`
	PenaltyTypeID      *uuid.UUID `json:"penalty_type_id,omitempty"`
	PenaltyTypeName    *string    `json:"penalty_type_name,omitempty"`
	BonusTypeID        *uuid.UUID `json:"bonus_type_id,omitempty"`
	BonusTypeName      *string    `json:"bonus_type_name,omitempty"`
	Points             *float64   `json:"points,omitempty"`
	UsedAt             *time.Time `json:"used_at,omitempty"`
	PunishmentTypeID   *uuid.UUID `json:"punishment_type_id,omitempty"`
	PunishmentTypeName *string    `json:"punishment_type_name,omitempty"`
	TriggeringRuleID   *uuid.UUID `json:"triggering_rule_id,omitempty"`
	TriggeringRuleName *string    `json:"triggering_rule_name,omitempty"`
	DueAt              *time.Time `json:"due_at,omitempty"`
	ResolvedAt         *time.Time `json:"resolved_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

type ReturnStudentProfileDto struct {
	Student            StudentProfileStudentDto             `json:"student"`
	Classrooms         []StudentClassroomDto                `json:"classrooms"`
	Kpis               StudentProfileKpisDto                `json:"kpis"`
	PendingPunishments []StudentProfilePendingPunishmentDto `json:"pending_punishments"`
	AvailableBonuses   []StudentProfileAvailableBonusDto    `json:"available_bonuses"`
	History            []StudentProfileHistoryItemDto       `json:"history"`
}

func StudentProfileFromRows(
	student *repository.GetStudentByUserRow,
	kpis *repository.GetStudentProfileKpisRow,
	classrooms []repository.ListStudentProfileClassroomsRow,
	pendingPunishments []repository.ListStudentProfilePendingPunishmentsRow,
	availableBonuses []repository.ListStudentProfileAvailableBonusesRow,
	history []repository.ListStudentProfileHistoryRow,
) *ReturnStudentProfileDto {
	response := &ReturnStudentProfileDto{
		Classrooms:         studentProfileClassroomsFromRows(classrooms),
		PendingPunishments: studentProfilePendingPunishmentsFromRows(pendingPunishments),
		AvailableBonuses:   studentProfileAvailableBonusesFromRows(availableBonuses),
		History:            studentProfileHistoryFromRows(history),
	}

	if student != nil {
		response.Student = StudentProfileStudentDto{
			ID:        student.ID,
			FirstName: student.FirstName,
			LastName:  student.LastName,
			CreatedAt: student.CreatedAt,
			UpdatedAt: student.UpdatedAt,
		}
	}

	if kpis != nil {
		response.Kpis = StudentProfileKpisDto{
			AvailableBonusPoints:   kpis.AvailableBonusPoints,
			ActiveBonusCount:       kpis.ActiveBonusCount,
			TotalPenaltyCount:      kpis.TotalPenaltyCount,
			PendingPunishmentCount: kpis.PendingPunishmentCount,
		}
	}

	return response
}

func StudentProfileKpisFromRow(kpis *repository.GetStudentProfileKpisRow) *StudentProfileKpisDto {
	if kpis == nil {
		return nil
	}

	return &StudentProfileKpisDto{
		AvailableBonusPoints:   kpis.AvailableBonusPoints,
		ActiveBonusCount:       kpis.ActiveBonusCount,
		TotalPenaltyCount:      kpis.TotalPenaltyCount,
		PendingPunishmentCount: kpis.PendingPunishmentCount,
	}
}

func StudentProfileHistoryFromRows(rows []repository.ListStudentProfileHistoryRow) []StudentProfileHistoryItemDto {
	return studentProfileHistoryFromRows(rows)
}

func studentProfileClassroomsFromRows(rows []repository.ListStudentProfileClassroomsRow) []StudentClassroomDto {
	classrooms := make([]StudentClassroomDto, 0, len(rows))

	for _, row := range rows {
		classrooms = append(classrooms, StudentClassroomDto{
			ID:   row.ID,
			Name: row.Name,
		})
	}

	return classrooms
}

func studentProfilePendingPunishmentsFromRows(rows []repository.ListStudentProfilePendingPunishmentsRow) []StudentProfilePendingPunishmentDto {
	punishments := make([]StudentProfilePendingPunishmentDto, 0, len(rows))

	for _, row := range rows {
		punishments = append(punishments, StudentProfilePendingPunishmentDto{
			ID:                 row.ID,
			PunishmentTypeID:   row.PunishmentTypeID,
			PunishmentTypeName: row.PunishmentTypeName,
			TriggeringRuleID:   studentProfileUUIDPtr(row.TriggeringRuleID),
			TriggeringRuleName: studentProfileTextPtr(row.TriggeringRuleName),
			DueAt:              row.DueAt,
			CreatedAt:          row.CreatedAt,
		})
	}

	return punishments
}

func studentProfileAvailableBonusesFromRows(rows []repository.ListStudentProfileAvailableBonusesRow) []StudentProfileAvailableBonusDto {
	bonuses := make([]StudentProfileAvailableBonusDto, 0, len(rows))

	for _, row := range rows {
		bonuses = append(bonuses, StudentProfileAvailableBonusDto{
			ID:            row.ID,
			BonusTypeID:   row.BonusTypeID,
			BonusTypeName: row.BonusTypeName,
			Points:        row.Points,
			CreatedAt:     row.CreatedAt,
		})
	}

	return bonuses
}

func studentProfileHistoryFromRows(rows []repository.ListStudentProfileHistoryRow) []StudentProfileHistoryItemDto {
	history := make([]StudentProfileHistoryItemDto, 0, len(rows))

	for _, row := range rows {
		item := StudentProfileHistoryItemDto{
			Type:      row.Type,
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
		}

		switch row.Type {
		case "penalty":
			item.PenaltyTypeID = studentProfileUUIDPtrFromSentinel(row.PenaltyTypeID)
			item.PenaltyTypeName = studentProfileTextPtrFromString(row.PenaltyTypeName)
		case "bonus":
			item.BonusTypeID = studentProfileUUIDPtrFromSentinel(row.BonusTypeID)
			item.BonusTypeName = studentProfileTextPtrFromString(row.BonusTypeName)
			item.Points = studentProfileFloatPtrFromFloat(row.Points)
			item.UsedAt = studentProfileTimePtrFromSentinel(row.UsedAt)
		case "punishment":
			punishmentTypeID := row.PunishmentTypeID
			punishmentTypeName := row.PunishmentTypeName
			dueAt := row.DueAt
			item.PunishmentTypeID = &punishmentTypeID
			item.PunishmentTypeName = &punishmentTypeName
			item.TriggeringRuleID = studentProfileUUIDPtr(row.TriggeringRuleID)
			item.TriggeringRuleName = studentProfileTextPtrFromString(row.TriggeringRuleName)
			item.DueAt = &dueAt
			item.ResolvedAt = studentProfileTimePtrFromSentinelPg(row.ResolvedAt)
		}

		history = append(history, item)
	}

	return history
}

func studentProfileUUIDPtr(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}

	id := uuid.UUID(value.Bytes)
	if id == uuid.Nil {
		return nil
	}

	return &id
}

func studentProfileTextPtr(value pgtype.Text) *string {
	if !value.Valid || value.String == "" {
		return nil
	}

	text := value.String
	return &text
}

func studentProfileUUIDPtrFromSentinel(value uuid.UUID) *uuid.UUID {
	if value == uuid.Nil {
		return nil
	}

	id := value
	return &id
}

func studentProfileTextPtrFromString(value string) *string {
	if value == "" {
		return nil
	}

	text := value
	return &text
}

func studentProfileTimePtrFromSentinel(value time.Time) *time.Time {
	if value.Equal(studentProfileHistoryTimeSentinel) {
		return nil
	}

	timeValue := value
	return &timeValue
}

func studentProfileTimePtrFromSentinelPg(value pgtype.Timestamptz) *time.Time {
	if !value.Valid || value.Time.Equal(studentProfileHistoryTimeSentinel) {
		return nil
	}

	timeValue := value.Time
	return &timeValue
}

func studentProfileFloatPtrFromFloat(value float64) *float64 {
	floatValue := value
	return &floatValue
}
