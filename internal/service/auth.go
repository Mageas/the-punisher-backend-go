package service

import (
	"context"
	"log/slog"

	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/hash"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type AuthService interface {
	Login(ctx context.Context, req dto.LoginRequestDto) (*dto.LoginResponseDto, error)
}

type authService struct {
	repo repository.Querier
}

func NewAuthService(repo repository.Querier) AuthService {
	return &authService{repo: repo}
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequestDto) (*dto.LoginResponseDto, error) {
	userCredentials, err := s.repo.GetUserCredentialsByEmailForAuth(ctx, req.Email)
	if err != nil {
		slog.Error("failed to get user password", "error", err)
		return nil, api.ErrInternalError
	}

	if err := hash.VerifyPassword(req.Password, userCredentials.PasswordHash); err != nil {
		return nil, api.ErrInvalidCredentialsOrUserDoesntExist
	}

	return &dto.LoginResponseDto{
		AccessToken: "test",
	}, nil
}
