package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

type txBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Begin starts a transaction from the underlying DB handle used by sqlc queries.
func (q *Queries) Begin(ctx context.Context) (pgx.Tx, error) {
	beginner, ok := q.db.(txBeginner)
	if !ok {
		return nil, fmt.Errorf("underlying db does not support transactions")
	}

	return beginner.Begin(ctx)
}

// WithTxQuerier binds a transaction and returns a querier scoped to it.
func (q *Queries) WithTxQuerier(tx pgx.Tx) Querier {
	return q.WithTx(tx)
}

// WithinTransaction executes fn atomically with a transaction-scoped querier.
func (q *Queries) WithinTransaction(ctx context.Context, fn func(Querier) error) error {
	tx, err := q.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			slog.Error("failed to rollback transaction", "error", rollbackErr)
		}
	}()

	if err := fn(q.WithTxQuerier(tx)); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
