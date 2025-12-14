package service

import (
	"context"

	"github.com/mageas/the-punisher-backend/internal/apierr"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
	"github.com/mageas/the-punisher-backend/internal/utils"
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
		return nil, err
	}
	if exists {
		return nil, apierr.ErrEmailAlreadyExists
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.CreateUser(ctx, repository.CreateUserParams{
		Email:        req.Email,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		PasswordHash: hashedPassword,
	})
	if err != nil {
		return nil, err
	}

	return dto.FromRepository(&user), nil
}
