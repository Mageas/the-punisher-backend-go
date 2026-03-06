package sqlcmapper

import (
	"testing"
	"time"
)

func TestNormalizeAPITime(t *testing.T) {
	input := time.Date(2026, 3, 6, 15, 4, 5, 123456789, time.FixedZone("CET", 3600))

	got := normalizeAPITime(input)

	if got.Location() != time.UTC {
		t.Fatalf("expected UTC location, got %s", got.Location())
	}
	if got.Nanosecond()%1000 != 0 {
		t.Fatalf("expected microsecond precision, got nanoseconds=%d", got.Nanosecond())
	}
	if got.Format(time.RFC3339Nano) != "2026-03-06T14:04:05.123456Z" {
		t.Fatalf("unexpected normalized format: %s", got.Format(time.RFC3339Nano))
	}
}

func TestNormalizeOptionalAPITime(t *testing.T) {
	input := time.Date(2026, 3, 6, 15, 4, 5, 123456789, time.FixedZone("CET", 3600))

	got := normalizeOptionalAPITime(&input)
	if got == nil {
		t.Fatalf("expected non-nil pointer")
	}
	if got.Format(time.RFC3339Nano) != "2026-03-06T14:04:05.123456Z" {
		t.Fatalf("unexpected normalized format: %s", got.Format(time.RFC3339Nano))
	}
}

func TestNormalizeOptionalAPITimeNil(t *testing.T) {
	if normalizeOptionalAPITime(nil) != nil {
		t.Fatalf("expected nil pointer")
	}
}
