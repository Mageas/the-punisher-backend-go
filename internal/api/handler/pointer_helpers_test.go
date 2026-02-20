package handler_test

import (
	"time"

	"github.com/google/uuid"
)

func uuidPtr(value uuid.UUID) *uuid.UUID {
	converted := value
	return &converted
}

func doubleTimePtr(value time.Time) *time.Time {
	timeValue := value
	return &timeValue
}
