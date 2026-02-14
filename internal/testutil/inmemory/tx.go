package inmemory

import (
	"context"
	"errors"
	"maps"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

var errInMemoryTxUnsupported = errors.New("in-memory transaction does not support low-level pgx operations")

type inMemoryTx struct {
	mu      sync.Mutex
	parent  *Repository
	working *Repository
	closed  bool
}

func (r *Repository) Begin(_ context.Context) (pgx.Tx, error) {
	return &inMemoryTx{
		parent:  r,
		working: r.snapshot(),
	}, nil
}

func (r *Repository) WithTxQuerier(tx pgx.Tx) repository.Querier {
	inMemTx, ok := tx.(*inMemoryTx)
	if !ok {
		return r
	}

	return inMemTx.working
}

func (tx *inMemoryTx) Begin(_ context.Context) (pgx.Tx, error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.closed {
		return nil, pgx.ErrTxClosed
	}

	return &inMemoryTx{
		parent:  tx.working,
		working: tx.working.snapshot(),
	}, nil
}

func (tx *inMemoryTx) Commit(_ context.Context) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.closed {
		return pgx.ErrTxClosed
	}

	tx.parent.mu.Lock()
	tx.parent.applySnapshot(tx.working)
	tx.parent.mu.Unlock()

	tx.closed = true
	return nil
}

func (tx *inMemoryTx) Rollback(_ context.Context) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.closed {
		return pgx.ErrTxClosed
	}

	tx.closed = true
	return nil
}

func (tx *inMemoryTx) CopyFrom(_ context.Context, _ pgx.Identifier, _ []string, _ pgx.CopyFromSource) (int64, error) {
	return 0, errInMemoryTxUnsupported
}

func (tx *inMemoryTx) SendBatch(_ context.Context, _ *pgx.Batch) pgx.BatchResults {
	return inMemoryBatchResults{err: errInMemoryTxUnsupported}
}

func (tx *inMemoryTx) LargeObjects() pgx.LargeObjects {
	return pgx.LargeObjects{}
}

func (tx *inMemoryTx) Prepare(_ context.Context, _, _ string) (*pgconn.StatementDescription, error) {
	return nil, errInMemoryTxUnsupported
}

func (tx *inMemoryTx) Exec(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errInMemoryTxUnsupported
}

func (tx *inMemoryTx) Query(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
	return nil, errInMemoryTxUnsupported
}

func (tx *inMemoryTx) QueryRow(_ context.Context, _ string, _ ...any) pgx.Row {
	return inMemoryRow{err: errInMemoryTxUnsupported}
}

func (tx *inMemoryTx) Conn() *pgx.Conn {
	return nil
}

type inMemoryBatchResults struct {
	err error
}

func (r inMemoryBatchResults) Exec() (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, r.err
}

func (r inMemoryBatchResults) Query() (pgx.Rows, error) {
	return nil, r.err
}

func (r inMemoryBatchResults) QueryRow() pgx.Row {
	return inMemoryRow{err: r.err}
}

func (r inMemoryBatchResults) Close() error {
	return r.err
}

type inMemoryRow struct {
	err error
}

func (r inMemoryRow) Scan(_ ...any) error {
	return r.err
}

func (r *Repository) snapshot() *Repository {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return &Repository{
		users: maps.Clone(r.users),

		usersByEmail:  maps.Clone(r.usersByEmail),
		refreshTokens: maps.Clone(r.refreshTokens),

		students:          maps.Clone(r.students),
		classrooms:        maps.Clone(r.classrooms),
		studentClassrooms: maps.Clone(r.studentClassrooms),

		rules:       maps.Clone(r.rules),
		bonuses:     maps.Clone(r.bonuses),
		penalties:   maps.Clone(r.penalties),
		punishments: maps.Clone(r.punishments),

		bonusTypes:      maps.Clone(r.bonusTypes),
		penaltyTypes:    maps.Clone(r.penaltyTypes),
		punishmentTypes: maps.Clone(r.punishmentTypes),

		errors: maps.Clone(r.errors),
	}
}

func (r *Repository) applySnapshot(snapshot *Repository) {
	r.users = maps.Clone(snapshot.users)

	r.usersByEmail = maps.Clone(snapshot.usersByEmail)
	r.refreshTokens = maps.Clone(snapshot.refreshTokens)

	r.students = maps.Clone(snapshot.students)
	r.classrooms = maps.Clone(snapshot.classrooms)
	r.studentClassrooms = maps.Clone(snapshot.studentClassrooms)

	r.rules = maps.Clone(snapshot.rules)
	r.bonuses = maps.Clone(snapshot.bonuses)
	r.penalties = maps.Clone(snapshot.penalties)
	r.punishments = maps.Clone(snapshot.punishments)

	r.bonusTypes = maps.Clone(snapshot.bonusTypes)
	r.penaltyTypes = maps.Clone(snapshot.penaltyTypes)
	r.punishmentTypes = maps.Clone(snapshot.punishmentTypes)

	r.errors = maps.Clone(snapshot.errors)
}
