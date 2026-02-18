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
	OpUserEmailExists = "UserEmailExists"
	OpCreateUser      = "CreateUser"
	OpGetUserByID     = "GetUserByID"
)

func (r *Repository) SetUserEmailExistsError(err error) {
	r.SetError(OpUserEmailExists, err)
}

func (r *Repository) SetCreateUserError(err error) {
	r.SetError(OpCreateUser, err)
}

func (r *Repository) SetGetUserByIDError(err error) {
	r.SetError(OpGetUserByID, err)
}

func (r *Repository) SeedUser(user repository.User) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = user.CreatedAt
	}

	email := strings.ToLower(user.Email)
	user.Email = email

	r.users[user.ID] = user
	r.usersByEmail[email] = repository.GetUserCredentialsByEmailForAuthRow{
		ID:           user.ID,
		Email:        email,
		PasswordHash: user.PasswordHash,
	}
}

func (r *Repository) UserEmailExists(_ context.Context, email string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpUserEmailExists); err != nil {
		return false, err
	}

	_, exists := r.usersByEmail[strings.ToLower(email)]
	return exists, nil
}

func (r *Repository) CreateUser(_ context.Context, arg repository.CreateUserParams) (repository.CreateUserRow, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.errFor(OpCreateUser); err != nil {
		return repository.CreateUserRow{}, err
	}

	now := time.Now()
	id := uuid.New()
	email := strings.ToLower(arg.Email)

	r.users[id] = repository.User{
		ID:           id,
		Email:        email,
		FirstName:    arg.FirstName,
		LastName:     arg.LastName,
		PasswordHash: arg.PasswordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	r.usersByEmail[email] = repository.GetUserCredentialsByEmailForAuthRow{
		ID:           id,
		Email:        email,
		PasswordHash: arg.PasswordHash,
	}

	return repository.CreateUserRow{
		ID:        id,
		Email:     email,
		FirstName: arg.FirstName,
		LastName:  arg.LastName,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (r *Repository) GetUserByID(_ context.Context, id uuid.UUID) (repository.GetUserByIDRow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.errFor(OpGetUserByID); err != nil {
		return repository.GetUserByIDRow{}, err
	}

	user, ok := r.users[id]
	if !ok {
		return repository.GetUserByIDRow{}, pgx.ErrNoRows
	}

	return repository.GetUserByIDRow{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}
