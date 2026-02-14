package inmemory

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

const (
	OpGetUserCredentialsByEmailForAuth = "GetUserCredentialsByEmailForAuth"
	OpCreateRefreshToken               = "CreateRefreshToken"
	OpGetRefreshToken                  = "GetRefreshToken"
)

func (r *Repository) SetGetUserCredentialsError(err error) {
	r.SetError(OpGetUserCredentialsByEmailForAuth, err)
}

func (r *Repository) SetCreateRefreshTokenError(err error) {
	r.SetError(OpCreateRefreshToken, err)
}

func (r *Repository) SetGetRefreshTokenError(err error) {
	r.SetError(OpGetRefreshToken, err)
}

func (r *Repository) SeedAuthUser(id uuid.UUID, email, passwordHash string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.setAuthUser(id, email, passwordHash)
}

func (r *Repository) SeedRefreshToken(token repository.RefreshToken) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	if token.CreatedAt.IsZero() {
		token.CreatedAt = time.Now()
	}

	r.refreshTokens[refreshTokenKey(token.UserID, token.Token)] = token
}

func (r *Repository) StoredRefreshToken(userID uuid.UUID, token string) (repository.RefreshToken, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rt, ok := r.refreshTokens[refreshTokenKey(userID, token)]
	return rt, ok
}

func (r *Repository) GetUserCredentialsByEmailForAuth(_ context.Context, email string) (repository.GetUserCredentialsByEmailForAuthRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpGetUserCredentialsByEmailForAuth); err != nil {
		return repository.GetUserCredentialsByEmailForAuthRow{}, err
	}

	u, ok := r.usersByEmail[strings.ToLower(email)]
	if !ok {
		return repository.GetUserCredentialsByEmailForAuthRow{}, pgx.ErrNoRows
	}

	return u, nil
}

func (r *Repository) CreateRefreshToken(_ context.Context, arg repository.CreateRefreshTokenParams) (repository.RefreshToken, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpCreateRefreshToken); err != nil {
		return repository.RefreshToken{}, err
	}

	token := repository.RefreshToken{
		ID:        uuid.New(),
		UserID:    arg.UserID,
		Token:     arg.Token,
		UserAgent: arg.UserAgent,
		ClientIp:  arg.ClientIp,
		ExpiresAt: arg.ExpiresAt,
		CreatedAt: time.Now(),
	}

	r.refreshTokens[refreshTokenKey(token.UserID, token.Token)] = token
	return token, nil
}

func (r *Repository) GetRefreshToken(_ context.Context, arg repository.GetRefreshTokenParams) (repository.RefreshToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpGetRefreshToken); err != nil {
		return repository.RefreshToken{}, err
	}

	token, ok := r.refreshTokens[refreshTokenKey(arg.UserID, arg.Token)]
	if !ok || token.RevokedAt.Valid {
		return repository.RefreshToken{}, pgx.ErrNoRows
	}

	return token, nil
}

func refreshTokenKey(userID uuid.UUID, token string) string {
	return userID.String() + ":" + token
}
