package handler

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestParsePathUUID(t *testing.T) {
	id := uuid.New()
	req := httptest.NewRequest("GET", "/students/"+id.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("studentID", id.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	parsed, ok := parsePathUUID(rr, req, "studentID")
	if !ok {
		t.Fatalf("expected parse success")
	}
	if parsed != id {
		t.Fatalf("unexpected uuid: %s", parsed)
	}
}

func TestParsePathUUIDInvalid(t *testing.T) {
	req := httptest.NewRequest("GET", "/students/invalid", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("studentID", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	_, ok := parsePathUUID(rr, req, "studentID")
	if ok {
		t.Fatalf("expected parse failure")
	}
	if rr.Code != 404 {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestParseBodyUUIDInvalid(t *testing.T) {
	rr := httptest.NewRecorder()
	_, ok := parseBodyUUID(rr, "invalid", "studentID")
	if ok {
		t.Fatalf("expected parse failure")
	}
	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
