package dto

import (
	"encoding/json"
	"testing"
)

func assertEvaluationLabelEmptyString(t *testing.T, payload any) {
	t.Helper()

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	value, exists := decoded["evaluation_label"]
	if !exists {
		t.Fatalf("expected evaluation_label field to be present: %s", string(data))
	}
	if value != "" {
		t.Fatalf("expected evaluation_label to be empty string, got %v", value)
	}
}

func TestReturnBonusDto_AlwaysSerializesEvaluationLabel(t *testing.T) {
	assertEvaluationLabelEmptyString(t, ReturnBonusDto{})
}

func TestReturnPenaltyDto_AlwaysSerializesEvaluationLabel(t *testing.T) {
	assertEvaluationLabelEmptyString(t, ReturnPenaltyDto{})
}

func TestReturnPunishmentDto_AlwaysSerializesEvaluationLabel(t *testing.T) {
	assertEvaluationLabelEmptyString(t, ReturnPunishmentDto{})
}

func TestStudentHistoryItemDto_AlwaysSerializesEvaluationLabel(t *testing.T) {
	assertEvaluationLabelEmptyString(t, StudentHistoryItemDto{})
}
