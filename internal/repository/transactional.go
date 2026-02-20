package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WithinTransaction executes a function within a database transaction.
// It supports nested transactions if the underlying DB connection supports it (e.g. savepoints),
// but for simplicity here we just check if we are already in a transaction.
// If q.db is a *pgxpool.Pool, it starts a new transaction.
// If q.db is a pgx.Tx, it uses the existing transaction.
func (q *Queries) WithinTransaction(ctx context.Context, fn func(Querier) error) error {
	// Check if the underlying DB connection is a pool
	pool, ok := q.db.(*pgxpool.Pool)
	if ok {
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback(ctx)

		qTx := q.WithTx(tx)
		if err := fn(qTx); err != nil {
			return err
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
		return nil
	}

	// Check if it's already a transaction
	_, ok = q.db.(pgx.Tx)
	if ok {
		// Already in a transaction, just execute the function
		return fn(q)
	}

	return fmt.Errorf("database connection is neither *pgxpool.Pool nor pgx.Tx")
}
