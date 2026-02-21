package repository

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ErrNoRows is an alias for pgx.ErrNoRows to avoid direct dependency on pgx in the service layer.
var ErrNoRows = pgx.ErrNoRows

// IsUniqueViolation checks if the error is a PostgreSQL unique constraint violation (code 23505).
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
