package service

import (
	"context"

	"github.com/mageas/the-punisher-backend/internal/domain"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
	"github.com/mageas/the-punisher-backend/internal/utils"
)

type UserService interface {
	CreateUser(ctx context.Context, req dto.RequestUserDto) (*dto.ReturnUserDto, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, req dto.RequestUserDto) (*dto.ReturnUserDto, error) {
	exists, err := s.repo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailAlreadyExists
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.Create(ctx, &domain.User{
		Email:        req.Email,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		PasswordHash: hashedPassword,
	})
	if err != nil {
		return nil, err
	}

	return dto.FromDomain(user), nil
}
