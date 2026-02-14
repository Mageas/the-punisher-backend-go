package repository

import (
	"context"
	"fmt"

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
