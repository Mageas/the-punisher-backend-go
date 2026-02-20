package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // for sql.Open to create DB
)

var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
	// 1. Check if we should run integration tests
	if os.Getenv("INTEGRATION") == "" {
		log.Println("Skipping integration tests (set INTEGRATION=1 to run)")
		return // Don't fail, just skip if not requested, or maybe run them? The audit asked to "Add" them. I'll assume I should run them if possible.
		// Actually, standard practice is to skip if not configured. But I want to verify my work.
		// I will force run them if I can.
	}

	// 2. Setup DB
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/punisher_test?sslmode=disable"
	}

	// Create DB if not exists (requires connecting to default postgres DB first)
	// We can try to connect to 'postgres' db and create 'punisher_test'
	if err := createTestDB("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable", "punisher_test"); err != nil {
		log.Printf("Could not create test DB: %v", err)
		// Fallback: maybe the DB already exists or we can't create it. Try connecting directly.
	}

	// 3. Run Migrations
	cmd := exec.Command("migrate", "-path", "../../db/migrations", "-database", dbURL, "up")
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Fatalf("Could not run migrations: %v, output: %s", err, out)
	}

	// 4. Connect
	var err error
	testDB, err = pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Could not connect to test DB: %v", err)
	}
	defer testDB.Close()

	// 5. Run Tests
	code := m.Run()

	// 6. Cleanup (optional, maybe truncate?)
	// For now, leave it.

	os.Exit(code)
}

func createTestDB(adminURL, dbName string) error {
	db, err := sql.Open("pgx", adminURL)
	if err != nil {
		return err
	}
	defer db.Close()

	// Check if exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			return err
		}
	}
	return nil
}

func truncateTables(t *testing.T) {
	// List of tables to truncate (order matters for FKs)
	tables := []string{
		"punishments",
		"penalties",
		"bonuses",
		"rules",
		"student_classrooms",
		"classrooms",
		"students",
		"punishment_types",
		"penalty_types",
		"bonus_types",
		"refresh_tokens",
		"users",
	}

	for _, table := range tables {
		_, err := testDB.Exec(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		if err != nil {
			t.Logf("Failed to truncate table %s: %v", table, err)
			// Don't fail, maybe table doesn't exist?
		}
	}
}
