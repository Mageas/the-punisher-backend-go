package repository

import (
	"context"

	"github.com/mageas/the-punisher-backend/internal/db"
	"github.com/mageas/the-punisher-backend/internal/domain"
)

type AuthRepository interface {
	GetUserCredentialsByEmailForAuth(ctx context.Context, email string) (domain.UserCredentials, error)
}

type postgresAuthRepository struct {
	q db.Querier
}

func NewAuthRepository(q db.Querier) AuthRepository {
	return &postgresAuthRepository{q: q}
}

func (r *postgresAuthRepository) GetUserCredentialsByEmailForAuth(ctx context.Context, email string) (domain.UserCredentials, error) {
	user, err := r.q.GetUserCredentialsByEmailForAuth(ctx, email)
	if err != nil {
		return domain.UserCredentials{}, err
	}

	return domain.UserCredentials{
		ID:           user.ID,
		PasswordHash: user.PasswordHash,
	}, nil
}
