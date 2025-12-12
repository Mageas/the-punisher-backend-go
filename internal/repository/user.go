package repository

import (
	"context"

	repo "github.com/mageas/the-punisher-backend/internal/adapters/storage/postgres"
)

type UserRepository interface {
	CreateUser(ctx context.Context, params repo.CreateUserParams) (repo.CreateUserRow, error)
	EmailExists(ctx context.Context, email string) (bool, error)
}

type userRepository struct {
	q repo.Querier
}

func NewUserRepository(q repo.Querier) *userRepository {
	return &userRepository{
		q: q,
	}
}

func (r *userRepository) CreateUser(ctx context.Context, params repo.CreateUserParams) (repo.CreateUserRow, error) {
	return r.q.CreateUser(ctx, params)
}

func (r *userRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	return r.q.UserEmailExists(ctx, email)
}
