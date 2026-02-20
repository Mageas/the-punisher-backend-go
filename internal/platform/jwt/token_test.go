package jwt_test

import (
	"testing"
	"time"

	"github.com/mageas/the-punisher-backend/internal/api"
	platformjwt "github.com/mageas/the-punisher-backend/internal/platform/jwt"
)

func TestVerify(t *testing.T) {
	cfg := platformjwt.Config{
		Secret:     "test-secret",
		Expiration: time.Minute,
		Issuer:     "the-punisher-tests",
		Audience:   "the-punisher-tests",
	}

	t.Run("valid_token", func(t *testing.T) {
		token, err := platformjwt.Generate(cfg, "user-123")
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		claims, err := platformjwt.Verify(token, platformjwt.VerifyConfig{
			Secret:   cfg.Secret,
			Issuer:   cfg.Issuer,
			Audience: cfg.Audience,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		sub, err := claims.GetSubject()
		if err != nil {
			t.Fatalf("expected subject claim, got error: %v", err)
		}
		if sub != "user-123" {
			t.Fatalf("expected subject user-123, got %q", sub)
		}
	})

	t.Run("invalid_issuer", func(t *testing.T) {
		token, err := platformjwt.Generate(cfg, "user-123")
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		_, err = platformjwt.Verify(token, platformjwt.VerifyConfig{
			Secret:   cfg.Secret,
			Issuer:   "another-issuer",
			Audience: cfg.Audience,
		})
		if err == nil {
			t.Fatal("expected verify error")
		}
		if err.Error() != api.ErrJWTInvalidToken.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrJWTInvalidToken.Error(), err.Error())
		}
	})

	t.Run("invalid_audience", func(t *testing.T) {
		token, err := platformjwt.Generate(cfg, "user-123")
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		_, err = platformjwt.Verify(token, platformjwt.VerifyConfig{
			Secret:   cfg.Secret,
			Issuer:   cfg.Issuer,
			Audience: "another-audience",
		})
		if err == nil {
			t.Fatal("expected verify error")
		}
		if err.Error() != api.ErrJWTInvalidToken.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrJWTInvalidToken.Error(), err.Error())
		}
	})

	t.Run("expired_token", func(t *testing.T) {
		token, err := platformjwt.Generate(platformjwt.Config{
			Secret:     cfg.Secret,
			Expiration: -1 * time.Minute,
			Issuer:     cfg.Issuer,
			Audience:   cfg.Audience,
		}, "user-123")
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		_, err = platformjwt.Verify(token, platformjwt.VerifyConfig{
			Secret:   cfg.Secret,
			Issuer:   cfg.Issuer,
			Audience: cfg.Audience,
		})
		if err == nil {
			t.Fatal("expected verify error")
		}
		if err.Error() != api.ErrJWTExpired.Error() {
			t.Fatalf("expected error %q, got %q", api.ErrJWTExpired.Error(), err.Error())
		}
	})
}
