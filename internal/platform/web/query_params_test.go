package web

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestParseEnumQueryParam(t *testing.T) {
	r := httptest.NewRequest("GET", "/?status=ReSoLvEd", nil)
	value, hasValue, err := ParseEnumQueryParam(r, "status", []string{"resolved", "pending"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasValue || value != "resolved" {
		t.Fatalf("unexpected result: value=%q hasValue=%v", value, hasValue)
	}
}

func TestParseEnumQueryParamMissing(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	value, hasValue, err := ParseEnumQueryParam(r, "status", []string{"resolved", "pending"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasValue || value != "" {
		t.Fatalf("expected empty,false got %q,%v", value, hasValue)
	}
}

func TestParseEnumQueryParamInvalid(t *testing.T) {
	r := httptest.NewRequest("GET", "/?status=invalid", nil)
	_, hasValue, err := ParseEnumQueryParam(r, "status", []string{"resolved", "pending"})
	if !hasValue {
		t.Fatalf("expected hasValue true")
	}
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseEnumQueryParamToBool(t *testing.T) {
	r := httptest.NewRequest("GET", "/?used=yes", nil)
	v, details, err := ParseEnumQueryParamToBool(r, "used", "yes", "no")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(details) != 0 {
		t.Fatalf("unexpected details: %+v", details)
	}
	if v == nil || !*v {
		t.Fatalf("expected true pointer")
	}
}

func TestParseEnumQueryParamToBoolInvalid(t *testing.T) {
	r := httptest.NewRequest("GET", "/?used=maybe", nil)
	v, details, err := ParseEnumQueryParamToBool(r, "used", "yes", "no")
	if err == nil {
		t.Fatalf("expected error")
	}
	if v != nil {
		t.Fatalf("expected nil bool pointer")
	}
	if len(details) != 1 || details[0].Field != "used" {
		t.Fatalf("unexpected details: %+v", details)
	}
	if !strings.Contains(details[0].Error, "yes_or_no") {
		t.Fatalf("expected yes_or_no hint, got %s", details[0].Error)
	}
}

func TestParseEnumQueryParamToBoolMissing(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	v, details, err := ParseEnumQueryParamToBool(r, "used", "yes", "no")
	if err != nil || details != nil || v != nil {
		t.Fatalf("expected nil,nil,nil result, got v=%v details=%v err=%v", v, details, err)
	}
}

func TestParseSearchQueryParam(t *testing.T) {
	r := httptest.NewRequest("GET", "/?search=%20%20John%20%20%20Doe%20%20", nil)
	search := ParseSearchQueryParam(r, "search")
	if search == nil || *search != "John Doe" {
		t.Fatalf("unexpected search value: %v", search)
	}
}

func TestParseSearchQueryParamEmpty(t *testing.T) {
	r := httptest.NewRequest("GET", "/?search=%20%20", nil)
	search := ParseSearchQueryParam(r, "search")
	if search != nil {
		t.Fatalf("expected nil search, got %q", *search)
	}
}

func TestEnumExpected(t *testing.T) {
	got := EnumExpected([]string{"YES", "No"})
	if got != "yes_or_no" {
		t.Fatalf("unexpected EnumExpected result: %s", got)
	}
}

func TestEnumExpectedEmpty(t *testing.T) {
	if got := EnumExpected(nil); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestParseOptionalUUIDQueryParam(t *testing.T) {
	id := uuid.New()
	r := httptest.NewRequest("GET", "/?student_id="+id.String(), nil)

	got, details, err := ParseOptionalUUIDQueryParam(r, "student_id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(details) != 0 {
		t.Fatalf("unexpected details: %+v", details)
	}
	if got == nil || *got != id {
		t.Fatalf("unexpected uuid: %v", got)
	}
}

func TestParseOptionalUUIDQueryParamInvalid(t *testing.T) {
	r := httptest.NewRequest("GET", "/?student_id=invalid", nil)

	got, details, err := ParseOptionalUUIDQueryParam(r, "student_id")
	if err == nil {
		t.Fatalf("expected error")
	}
	if got != nil {
		t.Fatalf("expected nil uuid")
	}
	if len(details) != 1 || details[0].Field != "student_id" {
		t.Fatalf("unexpected details: %+v", details)
	}
}

func TestParseOptionalUUIDQueryParamMissing(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)

	got, details, err := ParseOptionalUUIDQueryParam(r, "student_id")
	if err != nil || details != nil || got != nil {
		t.Fatalf("expected nil,nil,nil, got=%v details=%v err=%v", got, details, err)
	}
}

func TestParseOptionalBoolQueryParam(t *testing.T) {
	r := httptest.NewRequest("GET", "/?overdue=true", nil)

	got, details, err := ParseOptionalBoolQueryParam(r, "overdue")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(details) != 0 {
		t.Fatalf("unexpected details: %+v", details)
	}
	if got == nil || !*got {
		t.Fatalf("expected true pointer")
	}
}

func TestParseOptionalBoolQueryParamInvalid(t *testing.T) {
	r := httptest.NewRequest("GET", "/?overdue=maybe", nil)

	got, details, err := ParseOptionalBoolQueryParam(r, "overdue")
	if err == nil {
		t.Fatalf("expected error")
	}
	if got != nil {
		t.Fatalf("expected nil bool")
	}
	if len(details) != 1 || details[0].Field != "overdue" {
		t.Fatalf("unexpected details: %+v", details)
	}
	if !strings.Contains(details[0].Error, "true_or_false") {
		t.Fatalf("expected true_or_false hint, got %s", details[0].Error)
	}
}

func TestParseOptionalBoolQueryParamMissing(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)

	got, details, err := ParseOptionalBoolQueryParam(r, "overdue")
	if err != nil || details != nil || got != nil {
		t.Fatalf("expected nil,nil,nil, got=%v details=%v err=%v", got, details, err)
	}
}

func TestParseOptionalDateQueryParam(t *testing.T) {
	r := httptest.NewRequest("GET", "/?created_from=2025-01-15", nil)

	got, details, err := ParseOptionalDateQueryParam(r, "created_from")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(details) != 0 {
		t.Fatalf("unexpected details: %+v", details)
	}
	if got == nil || got.Format("2006-01-02") != "2025-01-15" {
		t.Fatalf("unexpected date: %v", got)
	}
}

func TestParseOptionalDateQueryParamInvalid(t *testing.T) {
	r := httptest.NewRequest("GET", "/?created_from=15-01-2025", nil)

	got, details, err := ParseOptionalDateQueryParam(r, "created_from")
	if err == nil {
		t.Fatalf("expected error")
	}
	if got != nil {
		t.Fatalf("expected nil date")
	}
	if len(details) != 1 || details[0].Field != "created_from" {
		t.Fatalf("unexpected details: %+v", details)
	}
	if !strings.Contains(details[0].Error, "yyyy-mm-dd") {
		t.Fatalf("expected yyyy-mm-dd hint, got %s", details[0].Error)
	}
}

func TestParseOptionalDateQueryParamMissing(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)

	got, details, err := ParseOptionalDateQueryParam(r, "created_from")
	if err != nil || details != nil || got != nil {
		t.Fatalf("expected nil,nil,nil, got=%v details=%v err=%v", got, details, err)
	}
}

func TestValidateDateRange(t *testing.T) {
	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	details, err := ValidateDateRange(&from, &to, "created_from", "created_to")
	if err != nil || details != nil {
		t.Fatalf("expected valid range, got details=%v err=%v", details, err)
	}
}

func TestValidateDateRangeInvalid(t *testing.T) {
	from := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	details, err := ValidateDateRange(&from, &to, "created_from", "created_to")
	if err == nil {
		t.Fatalf("expected error")
	}
	if len(details) != 1 || details[0].Field != "created_from" {
		t.Fatalf("unexpected details: %+v", details)
	}
	if !strings.Contains(details[0].Error, "created_from_lte_created_to") {
		t.Fatalf("unexpected detail error: %s", details[0].Error)
	}
}
