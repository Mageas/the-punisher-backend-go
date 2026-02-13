package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mageas/the-punisher-backend/internal/platform/config"
	"github.com/mageas/the-punisher-backend/internal/repository"
	"github.com/mageas/the-punisher-backend/internal/seeder"
)

func main() {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	conn, err := pgxpool.New(ctx, cfg.DB.DSN)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	if err := conn.Ping(ctx); err != nil {
		slog.Error("Failed to ping database", "error", err)
		os.Exit(1)
	}

	slog.Info("Connected to database")

	repo := repository.New(conn)

	if err := seeder.UserSeed(ctx, repo); err != nil {
		slog.Error("Failed to seed users", "error", err)
		os.Exit(1)
	}

	if err := seeder.EducationSeed(ctx, repo); err != nil {
		slog.Error("Failed to seed education data", "error", err)
		os.Exit(1)
	}

	slog.Info("Seeding completed successfully")
}
