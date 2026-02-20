package integration

import (
	"context"
	"testing"
	"time"

	"github.com/mageas/the-punisher-backend/internal/platform/hash"
	"github.com/mageas/the-punisher-backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthFlow(t *testing.T) {
	if testDB == nil {
		t.Skip("Database not initialized")
	}
	truncateTables(t)

	ctx := context.Background()
	repo := repository.New(testDB)

	// 1. Register (Manual Insert for now as we are testing DB integration, not API handler here directly)
	// Actually, we should test the repository methods.

	email := "test@example.com"
	password := "password123"
	hashedPassword, err := hash.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user, err := repo.CreateUser(ctx, repository.CreateUserParams{
		Email:        email,
		FirstName:    "Test",
		LastName:     "User",
		PasswordHash: hashedPassword,
	})
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.Email != email {
		t.Errorf("Expected email %s, got %s", email, user.Email)
	}

	// 2. Login (Verify credentials)
	creds, err := repo.GetUserCredentialsByEmailForAuth(ctx, email)
	if err != nil {
		t.Fatalf("Failed to get credentials: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte(password)); err != nil {
		t.Errorf("Password verification failed: %v", err)
	}

	// 3. Create Refresh Token
	tokenString := "some_refresh_token"
	expiresAt := time.Now().Add(24 * time.Hour)
	refreshToken, err := repo.CreateRefreshToken(ctx, repository.CreateRefreshTokenParams{
		UserID:    user.ID,
		Token:     tokenString,
		UserAgent: "test-agent",
		ClientIp:  "127.0.0.1",
		ExpiresAt: expiresAt,
	})
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	// 4. Get Refresh Token
	retrievedToken, err := repo.GetRefreshToken(ctx, repository.GetRefreshTokenParams{
		UserID: user.ID,
		Token:  tokenString,
	})
	if err != nil {
		t.Fatalf("Failed to retrieve refresh token: %v", err)
	}

	if retrievedToken.ID != refreshToken.ID {
		t.Errorf("Expected token ID %s, got %s", refreshToken.ID, retrievedToken.ID)
	}

	// 5. Revoke Token
	revoked, err := repo.RevokeRefreshToken(ctx, tokenString)
	if err != nil {
		t.Fatalf("Failed to revoke token: %v", err)
	}
	if revoked.RevokedAt == nil {
		t.Error("Expected RevokedAt to be set")
	}

	// 6. Verify Revoked (should not be returned by GetRefreshToken with revoked check)
	_, err = repo.GetRefreshToken(ctx, repository.GetRefreshTokenParams{
		UserID: user.ID,
		Token:  tokenString,
	})
	if err == nil {
		t.Error("Expected error when getting revoked token, got nil")
	}
}
