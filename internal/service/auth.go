package service

import (
	"context"
	"log/slog"

	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/config"
	"github.com/mageas/the-punisher-backend/internal/platform/hash"
	"github.com/mageas/the-punisher-backend/internal/platform/jwt"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type AuthService interface {
	Login(ctx context.Context, req dto.LoginRequestDto) (*dto.LoginResponseDto, error)
	Refresh(ctx context.Context, refreshToken string) (*dto.RefreshResponseDto, error)
}

type authService struct {
	repo repository.Querier
	cfg  config.JWTConfig
}

func NewAuthService(repo repository.Querier, cfg config.JWTConfig) AuthService {
	return &authService{
		repo: repo,
		cfg:  cfg,
	}
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

	accessToken, err := jwt.Generate(jwt.Config{
		Secret:     s.cfg.AccessSecret,
		Expiration: s.cfg.AccessExpiration,
		Issuer:     s.cfg.Issuer,
		Audience:   s.cfg.Audience,
	}, userCredentials.ID.String())
	if err != nil {
		return nil, api.ErrInternalError
	}

	refreshToken, err := jwt.Generate(jwt.Config{
		Secret:     s.cfg.RefreshSecret,
		Expiration: s.cfg.RefreshExpiration,
		Issuer:     s.cfg.Issuer,
		Audience:   s.cfg.Audience,
	}, userCredentials.ID.String())
	if err != nil {
		return nil, api.ErrInternalError
	}

	return &dto.LoginResponseDto{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (*dto.RefreshResponseDto, error) {
	claims, err := jwt.Verify(refreshToken, s.cfg.RefreshSecret)
	if err != nil {
		slog.Error("failed to verify refresh token", "error", err)
		return nil, api.ErrUnauthorized
	}

	sub, err := claims.GetSubject()
	if err != nil {
		slog.Error("failed to get subject from refresh token", "error", err)
		return nil, api.ErrUnauthorized
	}

	accessToken, err := jwt.Generate(jwt.Config{
		Secret:     s.cfg.AccessSecret,
		Expiration: s.cfg.AccessExpiration,
		Issuer:     s.cfg.Issuer,
		Audience:   s.cfg.Audience,
	}, sub)
	if err != nil {
		slog.Error("failed to generate access token", "error", err)
		return nil, api.ErrInternalError
	}

	return &dto.RefreshResponseDto{
		AccessToken: accessToken,
	}, nil
}
