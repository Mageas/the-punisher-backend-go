package repository

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsUniqueViolation(t *testing.T) {
	if !IsUniqueViolation(&pgconn.PgError{Code: "23505"}) {
		t.Fatalf("expected unique violation")
	}
	if IsUniqueViolation(&pgconn.PgError{Code: "99999"}) {
		t.Fatalf("did not expect unique violation")
	}
	if IsUniqueViolation(errors.New("boom")) {
		t.Fatalf("did not expect unique violation for generic error")
	}
}
