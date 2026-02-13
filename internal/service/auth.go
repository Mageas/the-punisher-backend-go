package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrInvalidCredentialsOrUserDoesntExist
		}

		return nil, fmt.Errorf("failed to get user credentials: %w", err)
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
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := jwt.Generate(jwt.Config{
		Secret:     s.cfg.RefreshSecret,
		Expiration: s.cfg.RefreshExpiration,
		Issuer:     s.cfg.Issuer,
		Audience:   s.cfg.Audience,
	}, userCredentials.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	_, err = s.repo.CreateRefreshToken(ctx, repository.CreateRefreshTokenParams{
		UserID:    userCredentials.ID,
		Token:     refreshToken,
		UserAgent: "",
		ClientIp:  req.RemoteAddr,
		ExpiresAt: time.Now().Add(s.cfg.RefreshExpiration),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	slog.Info("user logged in", "user_id", userCredentials.ID, "remote_addr", req.RemoteAddr)

	return &dto.LoginResponseDto{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (*dto.RefreshResponseDto, error) {
	claims, err := jwt.Verify(refreshToken, s.cfg.RefreshSecret)
	if err != nil {
		return nil, err
	}

	sub, err := claims.GetSubject()
	if err != nil {
		return nil, api.ErrUnauthorized
	}

	uuid, err := uuid.Parse(sub)
	if err != nil {
		return nil, api.ErrUnauthorized
	}

	_, err = s.repo.GetRefreshToken(ctx, repository.GetRefreshTokenParams{
		UserID: uuid,
		Token:  refreshToken,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.ErrUnauthorized
		}

		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	accessToken, err := jwt.Generate(jwt.Config{
		Secret:     s.cfg.AccessSecret,
		Expiration: s.cfg.AccessExpiration,
		Issuer:     s.cfg.Issuer,
		Audience:   s.cfg.Audience,
	}, sub)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &dto.RefreshResponseDto{
		AccessToken: accessToken,
	}, nil
}
