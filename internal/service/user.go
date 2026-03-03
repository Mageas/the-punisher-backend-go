package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/adapter/persistence/sqlcmapper"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/config"
	"github.com/mageas/the-punisher-backend/internal/platform/hash"
	"github.com/mageas/the-punisher-backend/internal/platform/jwt"
	"github.com/mageas/the-punisher-backend/internal/repository"
)

type UserService interface {
	CreateUser(ctx context.Context, req dto.RequestUserDto) (*dto.ReturnUserDto, error)
	ConfirmEmail(ctx context.Context, token string) error
	ResendEmailConfirmation(ctx context.Context, email string) error
	GetCurrentUser(ctx context.Context, userID uuid.UUID) (*dto.ReturnUserDto, error)
}

type ConfirmationEmailSender interface {
	SendConfirmationEmail(ctx context.Context, toEmail string, firstName string, confirmationURL string, expiresIn time.Duration) error
}

type confirmationEmailPayload struct {
	toEmail         string
	firstName       string
	confirmationURL string
	expiresIn       time.Duration
}

type userService struct {
	repo               repository.Querier
	emailConfirmCfg    config.EmailConfirmationConfig
	confirmationMailer ConfirmationEmailSender
}

type transactionalUserRepo interface {
	repository.Querier
	WithinTransaction(ctx context.Context, fn func(repository.Querier) error) error
}

func NewUserService(repo repository.Querier) UserService {
	return &userService{
		repo: repo,
	}
}

func NewUserServiceWithEmailConfirmation(
	repo repository.Querier,
	emailConfirmCfg config.EmailConfirmationConfig,
	confirmationMailer ConfirmationEmailSender,
) UserService {
	return &userService{
		repo:               repo,
		emailConfirmCfg:    emailConfirmCfg,
		confirmationMailer: confirmationMailer,
	}
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

	if !s.isEmailConfirmationEnabled() {
		user, createErr := s.repo.CreateUser(ctx, repository.CreateUserParams{
			Email:        req.Email,
			FirstName:    req.FirstName,
			LastName:     req.LastName,
			PasswordHash: hashedPassword,
		})
		if createErr != nil {
			if repository.IsUniqueViolation(createErr) {
				return nil, api.ErrEmailAlreadyExists
			}
			return nil, fmt.Errorf("failed to create user: %w", createErr)
		}

		slog.Info("user created", "user_id", user.ID, "email", user.Email)
		return sqlcmapper.UserFromRepository(&user), nil
	}

	txRepo, ok := s.repo.(transactionalUserRepo)
	if !ok {
		return nil, fmt.Errorf("user repository does not support transactions")
	}

	var user repository.CreateUserRow
	var pendingConfirmationEmail *confirmationEmailPayload

	err = txRepo.WithinTransaction(ctx, func(txQuerier repository.Querier) error {
		createdUser, createErr := txQuerier.CreateUser(ctx, repository.CreateUserParams{
			Email:        req.Email,
			FirstName:    req.FirstName,
			LastName:     req.LastName,
			PasswordHash: hashedPassword,
		})
		if createErr != nil {
			if repository.IsUniqueViolation(createErr) {
				return api.ErrEmailAlreadyExists
			}
			return fmt.Errorf("failed to create user: %w", createErr)
		}
		user = createdUser

		confirmationEmail, err := s.createConfirmationEmailPayload(ctx, txQuerier, user.ID, user.Email, user.FirstName)
		if err != nil {
			return err
		}
		pendingConfirmationEmail = confirmationEmail

		return nil
	})
	if err != nil {
		return nil, err
	}

	if pendingConfirmationEmail != nil {
		if err := s.sendConfirmationEmail(ctx, *pendingConfirmationEmail); err != nil {
			return nil, err
		}
	}

	slog.Info("user created", "user_id", user.ID, "email", user.Email, "email_confirmation_sent", true)

	return sqlcmapper.UserFromRepository(&user), nil
}

func (s *userService) ResendEmailConfirmation(ctx context.Context, email string) error {
	if !s.isEmailConfirmationEnabled() {
		return fmt.Errorf("email confirmation is not configured")
	}

	email = strings.TrimSpace(email)
	if email == "" {
		return nil
	}

	txRepo, ok := s.repo.(transactionalUserRepo)
	if !ok {
		return fmt.Errorf("user repository does not support transactions")
	}

	var pendingConfirmationEmail *confirmationEmailPayload

	err := txRepo.WithinTransaction(ctx, func(txQuerier repository.Querier) error {
		userVerification, getUserErr := txQuerier.GetUserEmailVerificationStateByEmail(ctx, email)
		if getUserErr != nil {
			if errors.Is(getUserErr, repository.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("failed to get user verification state by email: %w", getUserErr)
		}

		if userVerification.EmailVerifiedAt != nil {
			return nil
		}

		if _, invalidateErr := txQuerier.InvalidateEmailConfirmationTokensByUserID(ctx, userVerification.ID); invalidateErr != nil {
			return fmt.Errorf("failed to invalidate previous email confirmation tokens: %w", invalidateErr)
		}

		confirmationEmail, err := s.createConfirmationEmailPayload(ctx, txQuerier, userVerification.ID, userVerification.Email, userVerification.FirstName)
		if err != nil {
			return err
		}
		pendingConfirmationEmail = confirmationEmail

		return nil
	})
	if err != nil {
		return err
	}

	if pendingConfirmationEmail != nil {
		if err := s.sendConfirmationEmail(ctx, *pendingConfirmationEmail); err != nil {
			return err
		}
	}

	slog.Info("email confirmation resend requested", "email", email)
	return nil
}

func (s *userService) ConfirmEmail(ctx context.Context, token string) error {
	if !s.isEmailConfirmationEnabled() {
		return fmt.Errorf("email confirmation is not configured")
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return api.ErrEmailConfirmationTokenMissing
	}

	claims, err := jwt.Verify(token, jwt.VerifyConfig{
		Secret:   s.emailConfirmCfg.Secret,
		Issuer:   s.emailConfirmCfg.Issuer,
		Audience: s.emailConfirmCfg.Audience,
	})
	if err != nil {
		if errors.Is(err, api.ErrJWTExpired) {
			return api.ErrEmailConfirmationTokenExpired
		}
		return api.ErrEmailConfirmationTokenInvalid
	}

	subject, err := claims.GetSubject()
	if err != nil {
		return api.ErrEmailConfirmationTokenInvalid
	}

	claimedUserID, err := uuid.Parse(subject)
	if err != nil {
		return api.ErrEmailConfirmationTokenInvalid
	}

	txRepo, ok := s.repo.(transactionalUserRepo)
	if !ok {
		return fmt.Errorf("user repository does not support transactions")
	}

	tokenHash := hash.HashToken(token, s.emailConfirmCfg.Secret)

	err = txRepo.WithinTransaction(ctx, func(txQuerier repository.Querier) error {
		confirmationToken, getTokenErr := txQuerier.GetEmailConfirmationTokenByHash(ctx, tokenHash)
		if getTokenErr != nil {
			if errors.Is(getTokenErr, repository.ErrNoRows) {
				return api.ErrEmailConfirmationTokenInvalid
			}
			return fmt.Errorf("failed to get email confirmation token: %w", getTokenErr)
		}

		if confirmationToken.UserID != claimedUserID {
			return api.ErrEmailConfirmationTokenInvalid
		}

		if confirmationToken.UsedAt != nil {
			return api.ErrEmailConfirmationTokenAlreadyUsed
		}

		if time.Now().After(confirmationToken.ExpiresAt) {
			return api.ErrEmailConfirmationTokenExpired
		}

		userVerification, getUserErr := txQuerier.GetUserEmailVerificationStateByID(ctx, confirmationToken.UserID)
		if getUserErr != nil {
			if errors.Is(getUserErr, repository.ErrNoRows) {
				return api.ErrEmailConfirmationUserNotFound
			}
			return fmt.Errorf("failed to get user verification state: %w", getUserErr)
		}

		if userVerification.EmailVerifiedAt != nil {
			return api.ErrEmailAlreadyVerified
		}

		verifiedRows, verifyErr := txQuerier.VerifyUserEmailByID(ctx, confirmationToken.UserID)
		if verifyErr != nil {
			return fmt.Errorf("failed to verify user email: %w", verifyErr)
		}
		if verifiedRows == 0 {
			return api.ErrEmailAlreadyVerified
		}

		usedRows, markUsedErr := txQuerier.MarkEmailConfirmationTokenUsedByID(ctx, confirmationToken.ID)
		if markUsedErr != nil {
			return fmt.Errorf("failed to mark confirmation token as used: %w", markUsedErr)
		}
		if usedRows == 0 {
			return api.ErrEmailConfirmationTokenAlreadyUsed
		}

		return nil
	})
	if err != nil {
		return err
	}

	slog.Info("email confirmed", "user_id", claimedUserID)
	return nil
}

func (s *userService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*dto.ReturnUserDto, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, api.ErrUnauthorized
		}
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	return sqlcmapper.UserFromGetByIDRow(&user), nil
}

func (s *userService) isEmailConfirmationEnabled() bool {
	return strings.TrimSpace(s.emailConfirmCfg.Secret) != "" &&
		strings.TrimSpace(s.emailConfirmCfg.BaseURL) != "" &&
		s.emailConfirmCfg.Expiration > 0 &&
		s.confirmationMailer != nil
}

func (s *userService) createConfirmationEmailPayload(
	ctx context.Context,
	txQuerier repository.Querier,
	userID uuid.UUID,
	email string,
	firstName string,
) (*confirmationEmailPayload, error) {
	confirmationToken, expiresAt, tokenErr := s.generateEmailConfirmationToken(userID)
	if tokenErr != nil {
		return nil, tokenErr
	}

	_, createTokenErr := txQuerier.CreateEmailConfirmationToken(ctx, repository.CreateEmailConfirmationTokenParams{
		UserID:    userID,
		TokenHash: hash.HashToken(confirmationToken, s.emailConfirmCfg.Secret),
		ExpiresAt: expiresAt,
	})
	if createTokenErr != nil {
		return nil, fmt.Errorf("failed to create email confirmation token: %w", createTokenErr)
	}

	confirmationURL, linkErr := buildTokenURL(s.emailConfirmCfg.BaseURL, confirmationToken, "email confirmation")
	if linkErr != nil {
		return nil, linkErr
	}

	return &confirmationEmailPayload{
		toEmail:         email,
		firstName:       firstName,
		confirmationURL: confirmationURL,
		expiresIn:       s.emailConfirmCfg.Expiration,
	}, nil
}

func (s *userService) sendConfirmationEmail(ctx context.Context, confirmationEmail confirmationEmailPayload) error {
	if err := s.confirmationMailer.SendConfirmationEmail(
		ctx,
		confirmationEmail.toEmail,
		confirmationEmail.firstName,
		confirmationEmail.confirmationURL,
		confirmationEmail.expiresIn,
	); err != nil {
		return fmt.Errorf("failed to send confirmation email: %w", err)
	}

	return nil
}

func (s *userService) generateEmailConfirmationToken(userID uuid.UUID) (string, time.Time, error) {
	token, err := jwt.Generate(jwt.Config{
		Secret:     s.emailConfirmCfg.Secret,
		Expiration: s.emailConfirmCfg.Expiration,
		Issuer:     s.emailConfirmCfg.Issuer,
		Audience:   s.emailConfirmCfg.Audience,
	}, userID.String())
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate email confirmation token: %w", err)
	}

	return token, time.Now().Add(s.emailConfirmCfg.Expiration), nil
}
