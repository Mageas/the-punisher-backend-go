package sqlcmapper

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func TestRoundResponseFloat(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input float64
		want  float64
	}{
		{
			name:  "repeating decimal",
			input: 85.19999999999999,
			want:  85.2,
		},
		{
			name:  "round down",
			input: 12.344,
			want:  12.34,
		},
		{
			name:  "round up",
			input: 12.345,
			want:  12.35,
		},
		{
			name:  "binary floating edge case",
			input: 1.005,
			want:  1.01,
		},
		{
			name:  "zero stays zero",
			input: 0,
			want:  0,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := roundResponseFloat(tc.input); got != tc.want {
				t.Fatalf("roundResponseFloat(%v) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestRoundOptionalResponseFloat(t *testing.T) {
	t.Parallel()

	if got := roundOptionalResponseFloat(nil); got != nil {
		t.Fatalf("expected nil, got %v", *got)
	}

	value := 42.19999999999999
	got := roundOptionalResponseFloat(&value)
	if got == nil {
		t.Fatal("expected non-nil pointer")
	}
	if *got != 42.2 {
		t.Fatalf("expected 42.2, got %v", *got)
	}
}

func TestBonusFromCreateRowRoundsPointsInResponse(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)
	row := repository.CreateBonusRow{
		ID:               uuid.New(),
		StudentID:        uuid.New(),
		BonusTypeID:      uuid.New(),
		Points:           85.19999999999999,
		CreatedAt:        now,
		OccurredAt:       now,
		StudentFirstName: "Frank",
		StudentLastName:  "Castle",
		BonusTypeName:    "Participation",
	}

	response := BonusFromCreateRow(&row)
	if response == nil {
		t.Fatal("expected non-nil response")
	}
	if response.Points != 85.2 {
		t.Fatalf("expected rounded points 85.2, got %v", response.Points)
	}

	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}
	if strings.Contains(string(payload), "85.19999999999999") {
		t.Fatalf("expected marshaled payload to omit repeating decimals, got %s", payload)
	}
}

func TestStudentFromGetRowRoundsAvailableBonusPoints(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)
	row := repository.GetStudentByUserRow{
		ID:                   uuid.New(),
		FirstName:            "Frank",
		LastName:             "Castle",
		CreatedAt:            now,
		UpdatedAt:            now,
		AvailableBonusPoints: 21.555,
	}

	response := StudentFromGetRow(&row)
	if response == nil {
		t.Fatal("expected non-nil response")
	}
	if response.AvailableBonusPoints != 21.56 {
		t.Fatalf("expected rounded available bonus points 21.56, got %v", response.AvailableBonusPoints)
	}
}

func TestDashboardKpisFromRowRoundsBonusTotals(t *testing.T) {
	t.Parallel()

	row := repository.GetDashboardKpisRow{
		StudentCount:         10,
		AvailableBonusPoints: 64.49999999999999,
		TotalBonusPoints:     91.234,
	}

	response := DashboardKpisFromRow(&row)
	if response == nil {
		t.Fatal("expected non-nil response")
	}
	if response.AvailableBonusPoints != 64.5 {
		t.Fatalf("expected rounded available bonus points 64.5, got %v", response.AvailableBonusPoints)
	}
	if response.TotalBonusPoints != 91.23 {
		t.Fatalf("expected rounded total bonus points 91.23, got %v", response.TotalBonusPoints)
	}
}

func TestStudentKpisFromRowRoundsBonusTotals(t *testing.T) {
	t.Parallel()

	row := repository.GetStudentKpisRow{
		AvailableBonusPoints: 17.105,
		TotalBonusPoints:     19.994,
	}

	response := StudentKpisFromRow(&row)
	if response == nil {
		t.Fatal("expected non-nil response")
	}
	if response.AvailableBonusPoints != 17.11 {
		t.Fatalf("expected rounded available bonus points 17.11, got %v", response.AvailableBonusPoints)
	}
	if response.TotalBonusPoints != 19.99 {
		t.Fatalf("expected rounded total bonus points 19.99, got %v", response.TotalBonusPoints)
	}
}

func TestStudentHistoryFromRowsRoundsBonusPoints(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)
	bonusTypeID := uuid.New()
	bonusTypeName := "Participation"
	points := 5.555
	rows := []repository.ListStudentHistoryRow{
		{
			Type:            "bonus",
			ID:              uuid.New(),
			CreatedAt:       now,
			OccurredAt:      now,
			EvaluationLabel: "Math",
			BonusTypeID:     &bonusTypeID,
			BonusTypeName:   &bonusTypeName,
			Points:          &points,
		},
	}

	history := StudentHistoryFromRows(rows)
	if len(history) != 1 {
		t.Fatalf("expected 1 history item, got %d", len(history))
	}
	if history[0].Points == nil {
		t.Fatal("expected rounded bonus points")
	}
	if *history[0].Points != 5.56 {
		t.Fatalf("expected rounded history points 5.56, got %v", *history[0].Points)
	}
}
