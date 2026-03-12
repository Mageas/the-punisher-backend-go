package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func TestComputeRuleDueAt_DaysUsesUserCalendarDays(t *testing.T) {
	location, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		t.Fatalf("failed to load timezone: %v", err)
	}

	days := int32(1)
	rule := repository.Rule{
		DueAtMode:      ruleDueAtModeDays,
		DueAtAfterDays: &days,
	}
	now := time.Date(2026, 3, 28, 23, 30, 0, 0, time.UTC)

	got, err := computeRuleDueAt(context.Background(), nil, uuid.Nil, rule, nil, now, location)
	if err != nil {
		t.Fatalf("computeRuleDueAt returned error: %v", err)
	}

	want := time.Date(2026, 3, 30, 0, 30, 0, 0, location).UTC()
	if !got.Equal(want) {
		t.Fatalf("unexpected due_at: got=%s want=%s", got.Format(time.RFC3339), want.Format(time.RFC3339))
	}
}
