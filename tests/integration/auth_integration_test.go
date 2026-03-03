//go:build integration

package integration

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/config"
	"github.com/mageas/the-punisher-backend/internal/platform/hash"
	platformjwt "github.com/mageas/the-punisher-backend/internal/platform/jwt"
	"github.com/mageas/the-punisher-backend/internal/repository"
	. "github.com/mageas/the-punisher-backend/internal/service"
)

func integrationJWTConfig() config.JWTConfig {
	return config.JWTConfig{
		AccessSecret:      "access-secret",
		AccessExpiration:  15 * time.Minute,
		RefreshSecret:     "refresh-secret",
		RefreshExpiration: 7 * 24 * time.Hour,
		Issuer:            "test-issuer",
		Audience:          "test-audience",
	}
}

type capturedPasswordResetEmail struct {
	ToEmail   string
	FirstName string
	ResetURL  string
	ExpiresIn time.Duration
}

type fakePasswordResetMailer struct {
	emails []capturedPasswordResetEmail
	err    error
}

func (f *fakePasswordResetMailer) SendPasswordResetEmail(
	_ context.Context,
	toEmail string,
	firstName string,
	resetURL string,
	expiresIn time.Duration,
) error {
	if f.err != nil {
		return f.err
	}

	f.emails = append(f.emails, capturedPasswordResetEmail{
		ToEmail:   toEmail,
		FirstName: firstName,
		ResetURL:  resetURL,
		ExpiresIn: expiresIn,
	})
	return nil
}

func integrationPasswordResetConfig() config.PasswordResetConfig {
	return config.PasswordResetConfig{
		Secret:     "password-reset-secret",
		Expiration: time.Hour,
		BaseURL:    "http://localhost:3000/reset-password",
		Issuer:     "test-issuer",
		Audience:   "test-audience",
	}
}

func tokenFromPasswordResetURL(t *testing.T, resetURL string) string {
	t.Helper()

	parsed, err := url.Parse(resetURL)
	if err != nil {
		t.Fatalf("failed to parse password reset URL: %v", err)
	}

	token := strings.TrimSpace(parsed.Query().Get("token"))
	if token == "" {
		t.Fatalf("missing token in password reset URL: %s", resetURL)
	}

	return token
}

func createVerifiedAuthUser(t *testing.T, repo *repository.Queries, ctx context.Context, req dto.RequestUserDto) *dto.ReturnUserDto {
	t.Helper()

	userSvc := NewUserService(repo)
	user, err := userSvc.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	verifiedRows, err := repo.VerifyUserEmailByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to verify user email in test setup: %v", err)
	}
	if verifiedRows != 1 {
		t.Fatalf("expected one user email to be verified, got %d", verifiedRows)
	}

	return user
}

func TestAuthService_LoginRefreshLogoutFlow_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := createVerifiedAuthUser(t, repo, ctx, dto.RequestUserDto{
		Email:     "john@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Password:  "VeryStrongPassword123",
	})

	cfg := integrationJWTConfig()
	authSvc := NewAuthService(repo, cfg)

	loginResp, err := authSvc.Login(ctx, dto.LoginRequestDto{
		Email:      "john@example.com",
		Password:   "VeryStrongPassword123",
		RemoteAddr: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if loginResp.AccessToken == "" || loginResp.RefreshToken == "" {
		t.Fatalf("expected non-empty tokens")
	}

	if _, err := platformjwt.Verify(loginResp.AccessToken, platformjwt.VerifyConfig{
		Secret:   cfg.AccessSecret,
		Issuer:   cfg.Issuer,
		Audience: cfg.Audience,
	}); err != nil {
		t.Fatalf("access token verification failed: %v", err)
	}

	tokens, err := repo.ListRefreshTokensByUserId(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to list refresh tokens: %v", err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected 1 refresh token row, got %d", len(tokens))
	}
	if tokens[0].Token != hash.HashToken(loginResp.RefreshToken, cfg.RefreshSecret) {
		t.Fatalf("expected stored hashed refresh token")
	}

	refreshResp, err := authSvc.Refresh(ctx, loginResp.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}
	if refreshResp.RefreshToken == "" || refreshResp.AccessToken == "" {
		t.Fatalf("expected non-empty rotated tokens")
	}
	if refreshResp.RefreshToken == loginResp.RefreshToken {
		t.Fatalf("expected rotated refresh token to differ")
	}

	tokensAfterRefresh, err := repo.ListRefreshTokensByUserId(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to list refresh tokens after refresh: %v", err)
	}
	if len(tokensAfterRefresh) != 2 {
		t.Fatalf("expected 2 refresh token rows after rotation, got %d", len(tokensAfterRefresh))
	}

	activeCount := 0
	for _, tk := range tokensAfterRefresh {
		if tk.RevokedAt == nil {
			activeCount++
		}
	}
	if activeCount != 1 {
		t.Fatalf("expected exactly one active refresh token, got %d", activeCount)
	}

	if err := authSvc.Logout(ctx, refreshResp.RefreshToken); err != nil {
		t.Fatalf("Logout returned error: %v", err)
	}

	tokensAfterLogout, err := repo.ListRefreshTokensByUserId(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to list refresh tokens after logout: %v", err)
	}
	for _, tk := range tokensAfterLogout {
		if tk.RevokedAt == nil {
			t.Fatalf("expected all refresh tokens to be revoked")
		}
	}
}

func TestAuthService_LoginInvalidCredentials_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	createVerifiedAuthUser(t, repo, ctx, dto.RequestUserDto{
		Email:     "john@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Password:  "VeryStrongPassword123",
	})

	authSvc := NewAuthService(repo, integrationJWTConfig())

	_, err := authSvc.Login(ctx, dto.LoginRequestDto{
		Email:      "john@example.com",
		Password:   "WrongPassword123",
		RemoteAddr: "127.0.0.1",
	})
	if !errors.Is(err, api.ErrInvalidCredentialsOrUserDoesntExist) {
		t.Fatalf("expected ErrInvalidCredentialsOrUserDoesntExist, got %v", err)
	}
}

func TestAuthService_RefreshWithRevokedTokenReturnsUnauthorized_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	createVerifiedAuthUser(t, repo, ctx, dto.RequestUserDto{
		Email:     "john@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Password:  "VeryStrongPassword123",
	})

	authSvc := NewAuthService(repo, integrationJWTConfig())
	loginResp, err := authSvc.Login(ctx, dto.LoginRequestDto{
		Email:      "john@example.com",
		Password:   "VeryStrongPassword123",
		RemoteAddr: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}

	if err := authSvc.Logout(ctx, loginResp.RefreshToken); err != nil {
		t.Fatalf("Logout returned error: %v", err)
	}

	_, err = authSvc.Refresh(ctx, loginResp.RefreshToken)
	if !errors.Is(err, api.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestAuthService_LogoutAll_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := createVerifiedAuthUser(t, repo, ctx, dto.RequestUserDto{
		Email:     "john@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Password:  "VeryStrongPassword123",
	})

	authSvc := NewAuthService(repo, integrationJWTConfig())

	_, err := authSvc.Login(ctx, dto.LoginRequestDto{Email: user.Email, Password: "VeryStrongPassword123", RemoteAddr: "127.0.0.1"})
	if err != nil {
		t.Fatalf("first login failed: %v", err)
	}
	_, err = authSvc.Login(ctx, dto.LoginRequestDto{Email: user.Email, Password: "VeryStrongPassword123", RemoteAddr: "127.0.0.1"})
	if err != nil {
		t.Fatalf("second login failed: %v", err)
	}

	if err := authSvc.LogoutAll(ctx, user.ID); err != nil {
		t.Fatalf("LogoutAll returned error: %v", err)
	}

	tokens, err := repo.ListRefreshTokensByUserId(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to list refresh tokens: %v", err)
	}
	if len(tokens) != 0 {
		t.Fatalf("expected no refresh tokens after logout all, got %d", len(tokens))
	}
}

func TestAuthService_LoginWithUnverifiedEmailReturnsForbidden_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	userSvc := NewUserService(repo)
	_, err := userSvc.CreateUser(ctx, dto.RequestUserDto{
		Email:     "john@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Password:  "VeryStrongPassword123",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	authSvc := NewAuthService(repo, integrationJWTConfig())

	_, err = authSvc.Login(ctx, dto.LoginRequestDto{
		Email:      "john@example.com",
		Password:   "VeryStrongPassword123",
		RemoteAddr: "127.0.0.1",
	})
	if !errors.Is(err, api.ErrEmailNotVerified) {
		t.Fatalf("expected ErrEmailNotVerified, got %v", err)
	}
}

func TestAuthService_ChangePasswordSuccessInvalidatesRefreshTokens_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := createVerifiedAuthUser(t, repo, ctx, dto.RequestUserDto{
		Email:     "change-password.success@example.com",
		FirstName: "Change",
		LastName:  "Password",
		Password:  "CurrentPass1!",
	})

	authSvc := NewAuthService(repo, integrationJWTConfig())

	_, err := authSvc.Login(ctx, dto.LoginRequestDto{
		Email:      user.Email,
		Password:   "CurrentPass1!",
		RemoteAddr: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("first login failed: %v", err)
	}
	_, err = authSvc.Login(ctx, dto.LoginRequestDto{
		Email:      user.Email,
		Password:   "CurrentPass1!",
		RemoteAddr: "127.0.0.2",
	})
	if err != nil {
		t.Fatalf("second login failed: %v", err)
	}

	beforeChange, err := repo.GetUserCredentialsByIDForAuth(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to load credentials before password change: %v", err)
	}
	if beforeChange.PasswordChangedAt != nil {
		t.Fatalf("expected password_changed_at to be nil before change")
	}

	err = authSvc.ChangePassword(ctx, user.ID, dto.ChangePasswordRequestDto{
		CurrentPassword: "CurrentPass1!",
		NewPassword:     "NewSecurePass2@",
		ConfirmPassword: "NewSecurePass2@",
	})
	if err != nil {
		t.Fatalf("ChangePassword returned error: %v", err)
	}

	afterChange, err := repo.GetUserCredentialsByIDForAuth(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to load credentials after password change: %v", err)
	}
	if afterChange.PasswordChangedAt == nil {
		t.Fatalf("expected password_changed_at to be set")
	}

	tokens, err := repo.ListRefreshTokensByUserId(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to list refresh tokens: %v", err)
	}
	if len(tokens) != 0 {
		t.Fatalf("expected all refresh tokens to be invalidated, got %d", len(tokens))
	}

	_, err = authSvc.Login(ctx, dto.LoginRequestDto{
		Email:      user.Email,
		Password:   "CurrentPass1!",
		RemoteAddr: "127.0.0.1",
	})
	if !errors.Is(err, api.ErrInvalidCredentialsOrUserDoesntExist) {
		t.Fatalf("expected old password to fail, got %v", err)
	}

	_, err = authSvc.Login(ctx, dto.LoginRequestDto{
		Email:      user.Email,
		Password:   "NewSecurePass2@",
		RemoteAddr: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("expected login with new password to succeed, got %v", err)
	}
}

func TestAuthService_ChangePasswordWrongCurrentPassword_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := createVerifiedAuthUser(t, repo, ctx, dto.RequestUserDto{
		Email:     "change-password.invalid-current@example.com",
		FirstName: "Change",
		LastName:  "Password",
		Password:  "CurrentPass1!",
	})

	authSvc := NewAuthService(repo, integrationJWTConfig())

	_, err := authSvc.Login(ctx, dto.LoginRequestDto{
		Email:      user.Email,
		Password:   "CurrentPass1!",
		RemoteAddr: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	err = authSvc.ChangePassword(ctx, user.ID, dto.ChangePasswordRequestDto{
		CurrentPassword: "WrongCurrentPass1!",
		NewPassword:     "NewSecurePass2@",
		ConfirmPassword: "NewSecurePass2@",
	})
	if !errors.Is(err, api.ErrInvalidCurrentPassword) {
		t.Fatalf("expected ErrInvalidCurrentPassword, got %v", err)
	}

	tokens, err := repo.ListRefreshTokensByUserId(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to list refresh tokens: %v", err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected refresh tokens to remain unchanged after failure, got %d", len(tokens))
	}

	_, err = authSvc.Login(ctx, dto.LoginRequestDto{
		Email:      user.Email,
		Password:   "CurrentPass1!",
		RemoteAddr: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("expected login with unchanged password to succeed, got %v", err)
	}
}

func TestAuthService_ChangePasswordMismatchConfirmation_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := createVerifiedAuthUser(t, repo, ctx, dto.RequestUserDto{
		Email:     "change-password.confirmation@example.com",
		FirstName: "Change",
		LastName:  "Password",
		Password:  "CurrentPass1!",
	})

	authSvc := NewAuthService(repo, integrationJWTConfig())

	err := authSvc.ChangePassword(ctx, user.ID, dto.ChangePasswordRequestDto{
		CurrentPassword: "CurrentPass1!",
		NewPassword:     "NewSecurePass2@",
		ConfirmPassword: "DifferentPass2@",
	})

	var apiErr *api.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %v", err)
	}
	if apiErr.Message != api.ErrValidationFailed.Message {
		t.Fatalf("expected validation_failed, got %s", apiErr.Message)
	}
	if !hasErrorDetail(apiErr.Details, "confirm_password", api.KeyValidationPasswordConfirmationMismatch) {
		t.Fatalf("expected confirm_password mismatch detail, got %+v", apiErr.Details)
	}
}

func TestAuthService_ForgotAndResetPasswordFlow_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := createVerifiedAuthUser(t, repo, ctx, dto.RequestUserDto{
		Email:     "forgot-reset.success@example.com",
		FirstName: "Forgot",
		LastName:  "Reset",
		Password:  "CurrentPass1!",
	})

	mailer := &fakePasswordResetMailer{}
	authSvc := NewAuthServiceWithPasswordReset(repo, integrationJWTConfig(), integrationPasswordResetConfig(), mailer)

	_, err := authSvc.Login(ctx, dto.LoginRequestDto{
		Email:      user.Email,
		Password:   "CurrentPass1!",
		RemoteAddr: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("first login failed: %v", err)
	}
	_, err = authSvc.Login(ctx, dto.LoginRequestDto{
		Email:      user.Email,
		Password:   "CurrentPass1!",
		RemoteAddr: "127.0.0.2",
	})
	if err != nil {
		t.Fatalf("second login failed: %v", err)
	}

	if err := authSvc.ForgotPassword(ctx, dto.ForgotPasswordRequestDto{Email: user.Email}); err != nil {
		t.Fatalf("ForgotPassword first call failed: %v", err)
	}
	if len(mailer.emails) != 1 {
		t.Fatalf("expected one password reset email, got %d", len(mailer.emails))
	}
	firstToken := tokenFromPasswordResetURL(t, mailer.emails[0].ResetURL)

	if err := authSvc.ForgotPassword(ctx, dto.ForgotPasswordRequestDto{Email: "FORGOT-RESET.SUCCESS@EXAMPLE.COM"}); err != nil {
		t.Fatalf("ForgotPassword second call failed: %v", err)
	}
	if len(mailer.emails) != 2 {
		t.Fatalf("expected two password reset emails, got %d", len(mailer.emails))
	}
	secondToken := tokenFromPasswordResetURL(t, mailer.emails[1].ResetURL)
	if secondToken == firstToken {
		t.Fatalf("expected newly generated password reset token")
	}

	err = authSvc.ResetPassword(ctx, dto.ResetPasswordRequestDto{
		Token:           firstToken,
		NewPassword:     "NewSecurePass2@",
		ConfirmPassword: "NewSecurePass2@",
	})
	if !errors.Is(err, api.ErrPasswordResetTokenAlreadyUsed) {
		t.Fatalf("expected ErrPasswordResetTokenAlreadyUsed for invalidated token, got %v", err)
	}

	err = authSvc.ResetPassword(ctx, dto.ResetPasswordRequestDto{
		Token:           secondToken,
		NewPassword:     "NewSecurePass2@",
		ConfirmPassword: "NewSecurePass2@",
	})
	if err != nil {
		t.Fatalf("ResetPassword returned error: %v", err)
	}

	afterChange, err := repo.GetUserCredentialsByIDForAuth(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to load credentials after reset: %v", err)
	}
	if afterChange.PasswordChangedAt == nil {
		t.Fatalf("expected password_changed_at to be set after reset")
	}

	tokens, err := repo.ListRefreshTokensByUserId(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to list refresh tokens after reset: %v", err)
	}
	if len(tokens) != 0 {
		t.Fatalf("expected all refresh tokens to be invalidated after reset, got %d", len(tokens))
	}

	_, err = authSvc.Login(ctx, dto.LoginRequestDto{
		Email:      user.Email,
		Password:   "CurrentPass1!",
		RemoteAddr: "127.0.0.1",
	})
	if !errors.Is(err, api.ErrInvalidCredentialsOrUserDoesntExist) {
		t.Fatalf("expected old password to fail, got %v", err)
	}

	_, err = authSvc.Login(ctx, dto.LoginRequestDto{
		Email:      user.Email,
		Password:   "NewSecurePass2@",
		RemoteAddr: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("expected login with new password to succeed, got %v", err)
	}

	err = authSvc.ResetPassword(ctx, dto.ResetPasswordRequestDto{
		Token:           secondToken,
		NewPassword:     "AnotherPass3#",
		ConfirmPassword: "AnotherPass3#",
	})
	if !errors.Is(err, api.ErrPasswordResetTokenAlreadyUsed) {
		t.Fatalf("expected token to be unusable after success, got %v", err)
	}
}

func TestAuthService_ForgotPasswordUnknownUserIsNeutral_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	mailer := &fakePasswordResetMailer{}
	authSvc := NewAuthServiceWithPasswordReset(repo, integrationJWTConfig(), integrationPasswordResetConfig(), mailer)

	if err := authSvc.ForgotPassword(ctx, dto.ForgotPasswordRequestDto{Email: "missing.user@example.com"}); err != nil {
		t.Fatalf("ForgotPassword should be neutral for unknown user, got %v", err)
	}
	if len(mailer.emails) != 0 {
		t.Fatalf("expected no email for unknown user, got %d", len(mailer.emails))
	}
}

func TestAuthService_ResetPasswordInvalidToken_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	authSvc := NewAuthServiceWithPasswordReset(repo, integrationJWTConfig(), integrationPasswordResetConfig(), &fakePasswordResetMailer{})

	err := authSvc.ResetPassword(ctx, dto.ResetPasswordRequestDto{
		Token:           "not-a-jwt",
		NewPassword:     "NewSecurePass2@",
		ConfirmPassword: "NewSecurePass2@",
	})
	if !errors.Is(err, api.ErrPasswordResetTokenInvalid) {
		t.Fatalf("expected ErrPasswordResetTokenInvalid, got %v", err)
	}
}

func TestAuthService_ResetPasswordExpiredToken_WithQuerier(t *testing.T) {
	repo, ctx, cleanup := newTestQuerierTx(t)
	defer cleanup()

	user := createVerifiedAuthUser(t, repo, ctx, dto.RequestUserDto{
		Email:     "forgot-reset.expired@example.com",
		FirstName: "Forgot",
		LastName:  "Reset",
		Password:  "CurrentPass1!",
	})

	expiredConfig := integrationPasswordResetConfig()
	expiredConfig.Expiration = time.Millisecond

	mailer := &fakePasswordResetMailer{}
	authSvc := NewAuthServiceWithPasswordReset(repo, integrationJWTConfig(), expiredConfig, mailer)

	if err := authSvc.ForgotPassword(ctx, dto.ForgotPasswordRequestDto{Email: user.Email}); err != nil {
		t.Fatalf("ForgotPassword returned error: %v", err)
	}
	if len(mailer.emails) != 1 {
		t.Fatalf("expected one password reset email, got %d", len(mailer.emails))
	}

	token := tokenFromPasswordResetURL(t, mailer.emails[0].ResetURL)
	time.Sleep(25 * time.Millisecond)

	err := authSvc.ResetPassword(ctx, dto.ResetPasswordRequestDto{
		Token:           token,
		NewPassword:     "NewSecurePass2@",
		ConfirmPassword: "NewSecurePass2@",
	})
	if !errors.Is(err, api.ErrPasswordResetTokenExpired) {
		t.Fatalf("expected ErrPasswordResetTokenExpired, got %v", err)
	}
}

func hasErrorDetail(details []api.ErrorDetail, field string, key string) bool {
	for _, d := range details {
		if d.Field == field && d.Error == key {
			return true
		}
	}
	return false
}
