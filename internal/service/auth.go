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
	ForgotPassword(ctx context.Context, req dto.ForgotPasswordRequestDto) error
	ResetPassword(ctx context.Context, req dto.ResetPasswordRequestDto) error
}

type PasswordResetEmailSender interface {
	SendPasswordResetEmail(ctx context.Context, toEmail string, firstName string, resetURL string, expiresIn time.Duration) error
}

type passwordResetEmailPayload struct {
	toEmail   string
	firstName string
	resetURL  string
	expiresIn time.Duration
}

type authService struct {
	repo                repository.Querier
	cfg                 config.JWTConfig
	passwordResetCfg    config.PasswordResetConfig
	passwordResetMailer PasswordResetEmailSender
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

func NewAuthServiceWithPasswordReset(
	repo repository.Querier,
	cfg config.JWTConfig,
	passwordResetCfg config.PasswordResetConfig,
	passwordResetMailer PasswordResetEmailSender,
) AuthService {
	return &authService{
		repo:                repo,
		cfg:                 cfg,
		passwordResetCfg:    passwordResetCfg,
		passwordResetMailer: passwordResetMailer,
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

func (s *authService) ForgotPassword(ctx context.Context, req dto.ForgotPasswordRequestDto) error {
	if !s.isPasswordResetEnabled() {
		return fmt.Errorf("password reset is not configured")
	}

	email := strings.TrimSpace(req.Email)
	if email == "" {
		return nil
	}

	txRepo, ok := s.repo.(transactionalAuthRepo)
	if !ok {
		return fmt.Errorf("auth repository does not support transactions")
	}

	var pendingPasswordResetEmail *passwordResetEmailPayload

	err := txRepo.WithinTransaction(ctx, func(txQuerier repository.Querier) error {
		userVerification, getUserErr := txQuerier.GetUserEmailVerificationStateByEmail(ctx, email)
		if getUserErr != nil {
			if errors.Is(getUserErr, repository.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("failed to get user verification state by email: %w", getUserErr)
		}

		if _, invalidateErr := txQuerier.InvalidatePasswordResetTokensByUserID(ctx, userVerification.ID); invalidateErr != nil {
			return fmt.Errorf("failed to invalidate previous password reset tokens: %w", invalidateErr)
		}

		passwordResetEmail, payloadErr := s.createPasswordResetEmailPayload(
			ctx,
			txQuerier,
			userVerification.ID,
			userVerification.Email,
			userVerification.FirstName,
		)
		if payloadErr != nil {
			return payloadErr
		}

		pendingPasswordResetEmail = passwordResetEmail
		return nil
	})
	if err != nil {
		return err
	}

	if pendingPasswordResetEmail != nil {
		if err := s.sendPasswordResetEmail(ctx, *pendingPasswordResetEmail); err != nil {
			return err
		}
	}

	slog.Info("password reset requested", "email", email, "email_sent", pendingPasswordResetEmail != nil)
	return nil
}

func (s *authService) ResetPassword(ctx context.Context, req dto.ResetPasswordRequestDto) error {
	if !s.isPasswordResetEnabled() {
		return fmt.Errorf("password reset is not configured")
	}

	if err := validateResetPasswordRequest(req); err != nil {
		return err
	}

	token := strings.TrimSpace(req.Token)
	if token == "" {
		return api.ErrPasswordResetTokenMissing
	}

	claims, err := jwt.Verify(token, jwt.VerifyConfig{
		Secret:   s.passwordResetCfg.Secret,
		Issuer:   s.passwordResetCfg.Issuer,
		Audience: s.passwordResetCfg.Audience,
	})
	if err != nil {
		if errors.Is(err, api.ErrJWTExpired) {
			return api.ErrPasswordResetTokenExpired
		}
		return api.ErrPasswordResetTokenInvalid
	}

	subject, err := claims.GetSubject()
	if err != nil {
		return api.ErrPasswordResetTokenInvalid
	}

	claimedUserID, err := uuid.Parse(subject)
	if err != nil {
		return api.ErrPasswordResetTokenInvalid
	}

	txRepo, ok := s.repo.(transactionalAuthRepo)
	if !ok {
		return fmt.Errorf("auth repository does not support transactions")
	}

	tokenHash := hash.HashToken(token, s.passwordResetCfg.Secret)
	var invalidatedCount int64

	err = txRepo.WithinTransaction(ctx, func(txQuerier repository.Querier) error {
		resetToken, getTokenErr := txQuerier.GetPasswordResetTokenByHash(ctx, tokenHash)
		if getTokenErr != nil {
			if errors.Is(getTokenErr, repository.ErrNoRows) {
				return api.ErrPasswordResetTokenInvalid
			}
			return fmt.Errorf("failed to get password reset token: %w", getTokenErr)
		}

		if resetToken.UserID != claimedUserID {
			return api.ErrPasswordResetTokenInvalid
		}

		if resetToken.UsedAt != nil {
			return api.ErrPasswordResetTokenAlreadyUsed
		}

		if time.Now().After(resetToken.ExpiresAt) {
			return api.ErrPasswordResetTokenExpired
		}

		hashedPassword, hashErr := hash.HashPassword(req.NewPassword)
		if hashErr != nil {
			return fmt.Errorf("failed to hash password: %w", hashErr)
		}

		updatedRows, updateErr := txQuerier.UpdateUserPasswordByID(ctx, repository.UpdateUserPasswordByIDParams{
			ID:           resetToken.UserID,
			PasswordHash: hashedPassword,
		})
		if updateErr != nil {
			return fmt.Errorf("failed to update user password: %w", updateErr)
		}
		if updatedRows == 0 {
			return api.ErrPasswordResetUserNotFound
		}

		deletedCount, deleteErr := txQuerier.DeleteRefreshTokensByUserId(ctx, resetToken.UserID)
		if deleteErr != nil {
			return fmt.Errorf("failed to delete refresh tokens by user id: %w", deleteErr)
		}
		invalidatedCount = deletedCount

		usedRows, markUsedErr := txQuerier.MarkPasswordResetTokenUsedByID(ctx, resetToken.ID)
		if markUsedErr != nil {
			return fmt.Errorf("failed to mark password reset token as used: %w", markUsedErr)
		}
		if usedRows == 0 {
			return api.ErrPasswordResetTokenAlreadyUsed
		}

		if _, invalidateErr := txQuerier.InvalidatePasswordResetTokensByUserID(ctx, resetToken.UserID); invalidateErr != nil {
			return fmt.Errorf("failed to invalidate previous password reset tokens: %w", invalidateErr)
		}

		return nil
	})
	if err != nil {
		return err
	}

	slog.Info("password reset completed", "user_id", claimedUserID, "invalidated_refresh_token_count", invalidatedCount)
	return nil
}

func (s *authService) isPasswordResetEnabled() bool {
	return strings.TrimSpace(s.passwordResetCfg.Secret) != "" &&
		strings.TrimSpace(s.passwordResetCfg.BaseURL) != "" &&
		s.passwordResetCfg.Expiration > 0 &&
		s.passwordResetMailer != nil
}

func (s *authService) createPasswordResetEmailPayload(
	ctx context.Context,
	txQuerier repository.Querier,
	userID uuid.UUID,
	email string,
	firstName string,
) (*passwordResetEmailPayload, error) {
	resetToken, expiresAt, tokenErr := s.generatePasswordResetToken(userID)
	if tokenErr != nil {
		return nil, tokenErr
	}

	_, createTokenErr := txQuerier.CreatePasswordResetToken(ctx, repository.CreatePasswordResetTokenParams{
		UserID:    userID,
		TokenHash: hash.HashToken(resetToken, s.passwordResetCfg.Secret),
		ExpiresAt: expiresAt,
	})
	if createTokenErr != nil {
		return nil, fmt.Errorf("failed to create password reset token: %w", createTokenErr)
	}

	resetURL, linkErr := buildTokenURL(s.passwordResetCfg.BaseURL, resetToken, "password reset")
	if linkErr != nil {
		return nil, linkErr
	}

	return &passwordResetEmailPayload{
		toEmail:   email,
		firstName: firstName,
		resetURL:  resetURL,
		expiresIn: s.passwordResetCfg.Expiration,
	}, nil
}

func (s *authService) sendPasswordResetEmail(ctx context.Context, resetEmail passwordResetEmailPayload) error {
	if err := s.passwordResetMailer.SendPasswordResetEmail(
		ctx,
		resetEmail.toEmail,
		resetEmail.firstName,
		resetEmail.resetURL,
		resetEmail.expiresIn,
	); err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	return nil
}

func (s *authService) generatePasswordResetToken(userID uuid.UUID) (string, time.Time, error) {
	token, err := jwt.Generate(jwt.Config{
		Secret:     s.passwordResetCfg.Secret,
		Expiration: s.passwordResetCfg.Expiration,
		Issuer:     s.passwordResetCfg.Issuer,
		Audience:   s.passwordResetCfg.Audience,
	}, userID.String())
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate password reset token: %w", err)
	}

	return token, time.Now().Add(s.passwordResetCfg.Expiration), nil
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

func validateResetPasswordRequest(req dto.ResetPasswordRequestDto) error {
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
