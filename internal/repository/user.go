package repository

import (
	"context"

	"github.com/mageas/the-punisher-backend/internal/db"
	"github.com/mageas/the-punisher-backend/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
}

type postgresUserRepository struct {
	q db.Querier
}

func NewUserRepository(q db.Querier) UserRepository {
	return &postgresUserRepository{q: q}
}

func (r *postgresUserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	row, err := r.q.CreateUser(ctx, db.CreateUserParams{
		Email:        user.Email,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		PasswordHash: user.PasswordHash,
	})
	if err != nil {
		return nil, err
	}

	return &domain.User{
		ID:           row.ID,
		Email:        row.Email,
		FirstName:    row.FirstName,
		LastName:     row.LastName,
		PasswordHash: user.PasswordHash,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}, nil
}

func (r *postgresUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	return r.q.UserEmailExists(ctx, email)
}
