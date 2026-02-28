package handler

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestParseBodyRFC3339(t *testing.T) {
	rr := httptest.NewRecorder()
	raw := "2025-01-15T10:30:00Z"

	parsed, ok := parseBodyRFC3339(rr, raw, "due_at")
	if !ok {
		t.Fatalf("expected parse success")
	}
	if parsed.Format(time.RFC3339) != raw {
		t.Fatalf("unexpected parsed value: %s", parsed.Format(time.RFC3339))
	}
}

func TestParseBodyRFC3339Invalid(t *testing.T) {
	rr := httptest.NewRecorder()
	_, ok := parseBodyRFC3339(rr, "not-a-date", "due_at")
	if ok {
		t.Fatalf("expected parse failure")
	}
	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
