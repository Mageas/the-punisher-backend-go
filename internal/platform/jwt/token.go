package jwt

import (
	"errors"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
)

type Config struct {
	Secret     string
	Expiration time.Duration
	Issuer     string
	Audience   string
}

type VerifyConfig struct {
	Secret   string
	Issuer   string
	Audience string
}

func Generate(conf Config, subject string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(conf.Expiration)

	claims := jwt.MapClaims{
		"sub": subject,
		"iss": conf.Issuer,
		"aud": conf.Audience,
		"jti": uuid.NewString(),
		"exp": expiresAt.Unix(),
		"iat": now.Unix(),
		"nbf": now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(conf.Secret))
}

func Verify(tokenString string, conf VerifyConfig) (jwt.MapClaims, error) {
	token, err := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, api.ErrJWTInvalidSigningMethod
			}
			return []byte(conf.Secret), nil
		},
		jwt.WithIssuer(conf.Issuer),
		jwt.WithAudience(conf.Audience),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, api.ErrJWTExpired
		}
		slog.Error("failed to verify token", "error", err)
		return nil, api.ErrJWTInvalidToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, api.ErrJWTInvalidToken
}
