package web

import (
	"net/http/httptest"
	"strings"
	"testing"
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
