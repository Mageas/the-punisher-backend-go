package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mageas/the-punisher-backend/internal/adapter/persistence/sqlcmapper"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/hash"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type UserService interface {
	CreateUser(ctx context.Context, req dto.RequestUserDto) (*dto.ReturnUserDto, error)
	GetCurrentUser(ctx context.Context, userID uuid.UUID) (*dto.ReturnUserDto, error)
}

type userService struct {
	repo repository.Querier
}

func NewUserService(repo repository.Querier) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, req dto.RequestUserDto) (*dto.ReturnUserDto, error) {
	exists, err := s.repo.UserEmailExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, api.ErrEmailAlreadyExists
	}

	hashedPassword, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, repository.CreateUserParams{
		Email:        req.Email,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		PasswordHash: hashedPassword,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	slog.Info("user created", "user_id", user.ID, "email", user.Email)

	return sqlcmapper.UserFromRepository(&user), nil
}

func (s *userService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*dto.ReturnUserDto, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrUnauthorized
		}
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	return sqlcmapper.UserFromGetByIDRow(&user), nil
}
