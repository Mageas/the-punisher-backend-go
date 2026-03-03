package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
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
	Logout(ctx context.Context, refreshToken string) error
	LogoutAll(ctx context.Context, userID uuid.UUID) error
	ChangePassword(ctx context.Context, userID uuid.UUID, req dto.ChangePasswordRequestDto) error
}

type authService struct {
	repo repository.Querier
	cfg  config.JWTConfig
}

type transactionalAuthRepo interface {
	repository.Querier
	WithinTransaction(ctx context.Context, fn func(repository.Querier) error) error
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
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrInvalidCredentialsOrUserDoesntExist
		}

		return nil, fmt.Errorf("failed to get user credentials: %w", err)
	}

	if err := hash.VerifyPassword(req.Password, userCredentials.PasswordHash); err != nil {
		return nil, api.ErrInvalidCredentialsOrUserDoesntExist
	}

	if userCredentials.EmailVerifiedAt == nil {
		return nil, api.ErrEmailNotVerified
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

	refreshTokenHash := hash.HashToken(refreshToken, s.cfg.RefreshSecret)

	_, err = s.repo.CreateRefreshToken(ctx, repository.CreateRefreshTokenParams{
		UserID:    userCredentials.ID,
		Token:     refreshTokenHash,
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
	claims, err := jwt.Verify(refreshToken, jwt.VerifyConfig{
		Secret:   s.cfg.RefreshSecret,
		Issuer:   s.cfg.Issuer,
		Audience: s.cfg.Audience,
	})
	if err != nil {
		return nil, err
	}

	sub, err := claims.GetSubject()
	if err != nil {
		return nil, api.ErrUnauthorized
	}

	userID, err := uuid.Parse(sub)
	if err != nil {
		return nil, api.ErrUnauthorized
	}

	refreshTokenHash := hash.HashToken(refreshToken, s.cfg.RefreshSecret)

	accessToken, err := jwt.Generate(jwt.Config{
		Secret:     s.cfg.AccessSecret,
		Expiration: s.cfg.AccessExpiration,
		Issuer:     s.cfg.Issuer,
		Audience:   s.cfg.Audience,
	}, sub)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	rotatedRefreshToken, err := jwt.Generate(jwt.Config{
		Secret:     s.cfg.RefreshSecret,
		Expiration: s.cfg.RefreshExpiration,
		Issuer:     s.cfg.Issuer,
		Audience:   s.cfg.Audience,
	}, sub)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	rotatedRefreshTokenHash := hash.HashToken(rotatedRefreshToken, s.cfg.RefreshSecret)

	txRepo, ok := s.repo.(transactionalAuthRepo)
	if !ok {
		return nil, fmt.Errorf("auth repository does not support transactions")
	}

	err = txRepo.WithinTransaction(ctx, func(txQuerier repository.Querier) error {
		storedRefreshToken, getErr := txQuerier.GetRefreshToken(ctx, repository.GetRefreshTokenParams{
			UserID: userID,
			Token:  refreshTokenHash,
		})
		if getErr != nil {
			if errors.Is(getErr, repository.ErrNoRows) {
				return api.ErrUnauthorized
			}
			return fmt.Errorf("failed to get refresh token: %w", getErr)
		}

		_, revokeErr := txQuerier.RevokeRefreshToken(ctx, refreshTokenHash)
		if revokeErr != nil {
			if errors.Is(revokeErr, repository.ErrNoRows) {
				return api.ErrUnauthorized
			}
			return fmt.Errorf("failed to revoke refresh token: %w", revokeErr)
		}

		_, createErr := txQuerier.CreateRefreshToken(ctx, repository.CreateRefreshTokenParams{
			UserID:    userID,
			Token:     rotatedRefreshTokenHash,
			UserAgent: storedRefreshToken.UserAgent,
			ClientIp:  storedRefreshToken.ClientIp,
			ExpiresAt: time.Now().Add(s.cfg.RefreshExpiration),
		})
		if createErr != nil {
			return fmt.Errorf("failed to create refresh token: %w", createErr)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &dto.RefreshResponseDto{
		AccessToken:  accessToken,
		RefreshToken: rotatedRefreshToken,
	}, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil
	}

	refreshTokenHash := hash.HashToken(refreshToken, s.cfg.RefreshSecret)

	_, err := s.repo.RevokeRefreshToken(ctx, refreshTokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil
		}

		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

func (s *authService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	deletedCount, err := s.repo.DeleteRefreshTokensByUserId(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to delete refresh tokens by user id: %w", err)
	}

	slog.Info("all user refresh tokens invalidated", "user_id", userID, "deleted_count", deletedCount)
	return nil
}

func (s *authService) ChangePassword(ctx context.Context, userID uuid.UUID, req dto.ChangePasswordRequestDto) error {
	if err := validatePasswordChangeRequest(req); err != nil {
		return err
	}

	txRepo, ok := s.repo.(transactionalAuthRepo)
	if !ok {
		return fmt.Errorf("auth repository does not support transactions")
	}

	var invalidatedCount int64

	err := txRepo.WithinTransaction(ctx, func(txQuerier repository.Querier) error {
		userCredentials, getErr := txQuerier.GetUserCredentialsByIDForAuth(ctx, userID)
		if getErr != nil {
			if errors.Is(getErr, repository.ErrNoRows) {
				return api.ErrUnauthorized
			}
			return fmt.Errorf("failed to get user credentials by id: %w", getErr)
		}

		if err := hash.VerifyPassword(req.CurrentPassword, userCredentials.PasswordHash); err != nil {
			return api.ErrInvalidCurrentPassword
		}

		hashedPassword, hashErr := hash.HashPassword(req.NewPassword)
		if hashErr != nil {
			return fmt.Errorf("failed to hash password: %w", hashErr)
		}

		updatedRows, updateErr := txQuerier.UpdateUserPasswordByID(ctx, repository.UpdateUserPasswordByIDParams{
			ID:           userID,
			PasswordHash: hashedPassword,
		})
		if updateErr != nil {
			return fmt.Errorf("failed to update user password: %w", updateErr)
		}
		if updatedRows == 0 {
			return api.ErrUnauthorized
		}

		deletedCount, deleteErr := txQuerier.DeleteRefreshTokensByUserId(ctx, userID)
		if deleteErr != nil {
			return fmt.Errorf("failed to delete refresh tokens by user id: %w", deleteErr)
		}
		invalidatedCount = deletedCount

		return nil
	})
	if err != nil {
		return err
	}

	slog.Info("user password changed", "user_id", userID, "invalidated_refresh_token_count", invalidatedCount)
	return nil
}

func validatePasswordChangeRequest(req dto.ChangePasswordRequestDto) error {
	details := make([]api.ErrorDetail, 0, 1)

	if req.NewPassword != req.ConfirmPassword {
		details = append(details, api.ErrorDetail{
			Field: "confirm_password",
			Error: api.KeyValidationPasswordConfirmationMismatch,
		})
	}

	if len(details) == 0 {
		return nil
	}

	return api.NewAPIError(http.StatusBadRequest, api.ErrValidationFailed.Message, details...)
}
