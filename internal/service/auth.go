package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
	"github.com/mageas/the-punisher-backend/internal/utils"
)

var ErrInvalidCredentials = errors.New("invalid credentials or user doesn't exist")

type AuthService interface {
	Login(ctx context.Context, req dto.LoginRequestDto) (*dto.LoginResponseDto, error)
}

type authService struct {
	repo repository.AuthRepository
}

func NewAuthService(repo repository.AuthRepository) AuthService {
	return &authService{repo: repo}
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequestDto) (*dto.LoginResponseDto, error) {
	userCredentials, err := s.repo.GetUserCredentialsByEmailForAuth(ctx, req.Email)
	if err != nil {
		slog.Error("failed to get user password", "error", err)
		return nil, ErrInvalidCredentials
	}

	if err := utils.VerifyPassword(req.Password, userCredentials.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	return &dto.LoginResponseDto{
		AccessToken: "test",
	}, nil
}
