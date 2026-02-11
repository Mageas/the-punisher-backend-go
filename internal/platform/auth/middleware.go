package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/platform/jwt"
	"github.com/mageas/the-punisher-backend/internal/platform/web"
)

type contextKey string

const userIDKey contextKey = "userID"

func AuthMiddleware(accessSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				web.WriteError(w, http.StatusUnauthorized, api.ErrUnauthorized, nil)
				return
			}

			token, found := strings.CutPrefix(authHeader, "Bearer ")
			if !found {
				web.WriteError(w, http.StatusUnauthorized, api.ErrUnauthorized, nil)
				return
			}

			claims, err := jwt.Verify(token, accessSecret)
			if err != nil {
				web.WriteFromError(w, err)
				return
			}

			sub, err := claims.GetSubject()
			if err != nil {
				web.WriteError(w, http.StatusUnauthorized, api.ErrUnauthorized, nil)
				return
			}

			userID, err := uuid.Parse(sub)
			if err != nil {
				web.WriteError(w, http.StatusUnauthorized, api.ErrUnauthorized, nil)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}

func MustUserIDFromContext(ctx context.Context) uuid.UUID {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		panic("auth: userID not found in context — is AuthMiddleware applied?")
	}
	return userID
}
