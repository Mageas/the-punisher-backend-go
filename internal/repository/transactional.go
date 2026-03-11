package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WithinTransaction executes a function within a database transaction.
// If q.db is a *pgxpool.Pool, it starts a new transaction.
// If q.db is already a pgx.Tx, it starts a pseudo nested transaction (savepoint).
func (q *Queries) WithinTransaction(ctx context.Context, fn func(Querier) error) error {
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

	tx, ok := q.db.(pgx.Tx)
	if ok {
		nestedTx, err := tx.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin nested transaction: %w", err)
		}
		defer nestedTx.Rollback(ctx)

		qTx := q.WithTx(nestedTx)
		if err := fn(qTx); err != nil {
			return err
		}

		if err := nestedTx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit nested transaction: %w", err)
		}
		return nil
	}

	return fmt.Errorf("database connection is neither *pgxpool.Pool nor pgx.Tx")
}
