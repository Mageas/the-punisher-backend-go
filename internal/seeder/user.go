package seeder

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-faker/faker/v4"
	"github.com/mageas/the-punisher-backend/internal/platform/hash"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

func UserSeed(ctx context.Context, repo repository.Querier) error {
	// Create Admin User
	adminEmail := "admin@test.fr"
	adminPassword := "admin@test.fr"
	adminHashedPassword, err := hash.HashPassword(adminPassword)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	exists, err := repo.UserEmailExists(ctx, adminEmail)
	if err != nil {
		return fmt.Errorf("failed to check if admin exists: %w", err)
	}

	if !exists {
		_, err = repo.CreateUser(ctx, repository.CreateUserParams{
			Email:        adminEmail,
			FirstName:    "Admin",
			LastName:     "User",
			PasswordHash: adminHashedPassword,
		})
		if err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}
		slog.Info("Admin user created", "email", adminEmail)
	} else {
		slog.Info("Admin user already exists", "email", adminEmail)
	}

	// Create Random User
	randomEmail := faker.Email()
	randomPassword := randomEmail // Password same as email
	randomHashedPassword, err := hash.HashPassword(randomPassword)
	if err != nil {
		return fmt.Errorf("failed to hash random user password: %w", err)
	}

	_, err = repo.CreateUser(ctx, repository.CreateUserParams{
		Email:        randomEmail,
		FirstName:    faker.FirstName(),
		LastName:     faker.LastName(),
		PasswordHash: randomHashedPassword,
	})
	if err != nil {
		return fmt.Errorf("failed to create random user: %w", err)
	}
	slog.Info("Random user created", "email", randomEmail)

	return nil
}
