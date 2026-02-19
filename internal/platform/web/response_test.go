package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestWriteFromErrorMapsForeignKeyViolationToConflict(t *testing.T) {
	rr := httptest.NewRecorder()

	WriteFromError(rr, fmt.Errorf("wrapped: %w", &pgconn.PgError{Code: "23503"}))

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rr.Code)
	}

	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if body.Error != "related_record_cannot_delete" {
		t.Fatalf("expected error related_record_cannot_delete, got %s", body.Error)
	}
}

func TestWriteFromErrorMapsRestrictViolationToConflict(t *testing.T) {
	rr := httptest.NewRecorder()

	WriteFromError(rr, fmt.Errorf("wrapped: %w", &pgconn.PgError{Code: "23001"}))

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rr.Code)
	}

	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if body.Error != "related_record_cannot_delete" {
		t.Fatalf("expected error related_record_cannot_delete, got %s", body.Error)
	}
}

func TestWriteFromErrorMapsIntegrityAndConcurrencyConflictsToConflict(t *testing.T) {
	t.Parallel()

	conflictCodes := []string{"23000", "23502", "23505", "23514", "23P01", "40001", "40P01", "55P03"}

	for _, code := range conflictCodes {
		code := code
		t.Run(code, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			WriteFromError(rr, fmt.Errorf("wrapped: %w", &pgconn.PgError{Code: code}))

			if rr.Code != http.StatusConflict {
				t.Fatalf("expected status %d, got %d", http.StatusConflict, rr.Code)
			}

			var body struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			if body.Error != "conflict" {
				t.Fatalf("expected error conflict, got %s", body.Error)
			}
		})
	}
}
