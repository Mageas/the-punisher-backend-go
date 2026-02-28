package service

import (
	"time"

	"github.com/google/uuid"
)

type BonusState string

const (
	BonusStateUsed   BonusState = "used"
	BonusStateUnused BonusState = "unused"
)

func (s BonusState) Used() bool {
	return s == BonusStateUsed
}

type PunishmentState string

const (
	PunishmentStatePending  PunishmentState = "pending"
	PunishmentStateResolved PunishmentState = "resolved"
)

func (s PunishmentState) Resolved() bool {
	return s == PunishmentStateResolved
}

type ListBonusesFilters struct {
	StudentID   *uuid.UUID
	ClassroomID *uuid.UUID
	BonusTypeID *uuid.UUID
	State       *BonusState
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	Limit       int32
	Offset      int32
}

type ListPenaltiesFilters struct {
	StudentID     *uuid.UUID
	ClassroomID   *uuid.UUID
	PenaltyTypeID *uuid.UUID
	CreatedFrom   *time.Time
	CreatedTo     *time.Time
	Limit         int32
	Offset        int32
}

type ListPunishmentsFilters struct {
	StudentID        *uuid.UUID
	ClassroomID      *uuid.UUID
	PunishmentTypeID *uuid.UUID
	State            *PunishmentState
	Automated        *bool
	Overdue          *bool
	CreatedFrom      *time.Time
	CreatedTo        *time.Time
	DueFrom          *time.Time
	DueTo            *time.Time
	Limit            int32
	Offset           int32
}
