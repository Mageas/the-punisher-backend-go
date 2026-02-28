package jwt

import (
	"errors"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/mageas/the-punisher-backend/internal/api"
)

func TestGenerateAndVerify(t *testing.T) {
	conf := Config{
		Secret:     "access-secret",
		Expiration: time.Minute,
		Issuer:     "test-issuer",
		Audience:   "test-aud",
	}

	token, err := Generate(conf, "subject-123")
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	claims, err := Verify(token, VerifyConfig{
		Secret:   conf.Secret,
		Issuer:   conf.Issuer,
		Audience: conf.Audience,
	})
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}

	sub, err := claims.GetSubject()
	if err != nil {
		t.Fatalf("GetSubject returned error: %v", err)
	}
	if sub != "subject-123" {
		t.Fatalf("unexpected subject: %s", sub)
	}
}

func TestVerifyExpiredToken(t *testing.T) {
	token, err := Generate(Config{
		Secret:     "refresh-secret",
		Expiration: -1 * time.Second,
		Issuer:     "issuer",
		Audience:   "aud",
	}, "sub")
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	_, err = Verify(token, VerifyConfig{Secret: "refresh-secret", Issuer: "issuer", Audience: "aud"})
	if !errors.Is(err, api.ErrJWTExpired) {
		t.Fatalf("expected ErrJWTExpired, got: %v", err)
	}
}

func TestVerifyInvalidToken(t *testing.T) {
	token, err := Generate(Config{
		Secret:     "secret-a",
		Expiration: time.Minute,
		Issuer:     "issuer",
		Audience:   "aud",
	}, "sub")
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	_, err = Verify(token, VerifyConfig{Secret: "secret-b", Issuer: "issuer", Audience: "aud"})
	if !errors.Is(err, api.ErrJWTInvalidToken) {
		t.Fatalf("expected ErrJWTInvalidToken, got: %v", err)
	}
}

func TestVerifyInvalidSigningMethod(t *testing.T) {
	now := time.Now()
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodNone, jwtlib.MapClaims{
		"sub": "sub",
		"iss": "issuer",
		"aud": "aud",
		"exp": now.Add(time.Minute).Unix(),
		"iat": now.Unix(),
		"nbf": now.Unix(),
	})

	tokenString, err := token.SignedString(jwtlib.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("failed to sign none token: %v", err)
	}

	_, err = Verify(tokenString, VerifyConfig{Secret: "secret", Issuer: "issuer", Audience: "aud"})
	if !errors.Is(err, api.ErrJWTInvalidToken) {
		t.Fatalf("expected ErrJWTInvalidToken, got: %v", err)
	}
}
