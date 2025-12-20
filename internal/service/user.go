package service

import (
	"context"
	"fmt"

	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/hash"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type UserService interface {
	CreateUser(ctx context.Context, req dto.RequestUserDto) (*dto.ReturnUserDto, error)
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

	return dto.FromRepository(&user), nil
}
