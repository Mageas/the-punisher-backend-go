//go:build integration

package integration

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mageas/the-punisher-backend/internal/repository"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type integrationDB struct {
	container *postgres.PostgresContainer
	pool      *pgxpool.Pool
}

var (
	integrationDBOnce  sync.Once
	integrationDBState *integrationDB
	integrationDBErr   error
)

func TestMain(m *testing.M) {
	code := m.Run()

	if integrationDBState != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		integrationDBState.pool.Close()
		_ = integrationDBState.container.Terminate(ctx)
	}

	os.Exit(code)
}

func newTestQuerierTx(t *testing.T) (*repository.Queries, context.Context, func()) {
	t.Helper()

	db := getIntegrationDB(t)
	ctx := context.Background()

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		t.Fatalf("failed to begin test tx: %v", err)
	}

	queries := repository.New(db.pool).WithTx(tx)
	cleanup := func() {
		_ = tx.Rollback(ctx)
	}

	return queries, ctx, cleanup
}

func getIntegrationDB(t *testing.T) *integrationDB {
	t.Helper()

	integrationDBOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		integrationDBState, integrationDBErr = startIntegrationDB(ctx)
	})

	if integrationDBErr != nil {
		if isDockerUnavailable(integrationDBErr) {
			t.Skipf("docker unavailable, skipping integration-db tests: %v", integrationDBErr)
		}
		t.Fatalf("failed to init integration db: %v", integrationDBErr)
	}

	return integrationDBState
}

func startIntegrationDB(ctx context.Context) (*integrationDB, error) {
	container, err := postgres.Run(
		ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("punisher_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(90*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("failed to build postgres dsn: %w", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	if err := applyAllMigrations(ctx, pool); err != nil {
		pool.Close()
		_ = container.Terminate(ctx)
		return nil, err
	}

	return &integrationDB{container: container, pool: pool}, nil
}

func applyAllMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrationFiles, err := filepath.Glob(filepath.Join(projectRootDir(), "db", "migrations", "*.up.sql"))
	if err != nil {
		return fmt.Errorf("failed to glob migrations: %w", err)
	}
	if len(migrationFiles) == 0 {
		return errors.New("no migration files found")
	}

	sort.Strings(migrationFiles)

	for _, migrationFile := range migrationFiles {
		content, err := os.ReadFile(migrationFile)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", filepath.Base(migrationFile), err)
		}
		if strings.TrimSpace(string(content)) == "" {
			continue
		}

		if _, err := pool.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", filepath.Base(migrationFile), err)
		}
	}

	return nil
}

func projectRootDir() string {
	_, filePath, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(filePath), "..", ".."))
}

func isDockerUnavailable(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "cannot connect to the docker daemon") ||
		strings.Contains(message, "docker daemon") ||
		strings.Contains(message, "is the docker daemon running") ||
		errors.Is(err, context.DeadlineExceeded)
}
