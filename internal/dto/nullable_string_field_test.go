package dto

import (
	"encoding/json"
	"testing"
)

type nullableStringFieldPayload struct {
	EvaluationLabel NullableStringField `json:"evaluation_label"`
}

func TestNullableStringFieldAbsent(t *testing.T) {
	var payload nullableStringFieldPayload
	if err := json.Unmarshal([]byte(`{}`), &payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.EvaluationLabel.Set {
		t.Fatalf("expected Set=false when field is absent")
	}
	if payload.EvaluationLabel.Value != nil {
		t.Fatalf("expected Value=nil when field is absent")
	}
}

func TestNullableStringFieldNull(t *testing.T) {
	var payload nullableStringFieldPayload
	if err := json.Unmarshal([]byte(`{"evaluation_label":null}`), &payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !payload.EvaluationLabel.Set {
		t.Fatalf("expected Set=true when field is present")
	}
	if payload.EvaluationLabel.Value != nil {
		t.Fatalf("expected Value=nil when field is explicitly null")
	}
}

func TestNullableStringFieldValue(t *testing.T) {
	var payload nullableStringFieldPayload
	if err := json.Unmarshal([]byte(`{"evaluation_label":"Retard"}`), &payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !payload.EvaluationLabel.Set {
		t.Fatalf("expected Set=true when field is present")
	}
	if payload.EvaluationLabel.Value == nil || *payload.EvaluationLabel.Value != "Retard" {
		t.Fatalf("unexpected Value=%v", payload.EvaluationLabel.Value)
	}
}

func TestNullableStringFieldInvalid(t *testing.T) {
	var payload nullableStringFieldPayload
	if err := json.Unmarshal([]byte(`{"evaluation_label":12}`), &payload); err == nil {
		t.Fatalf("expected unmarshal error")
	}
}
