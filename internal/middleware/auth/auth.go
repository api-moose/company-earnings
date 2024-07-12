package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase/models"
)

type AuthApp interface {
	FindAuthRecordByToken(token, secret string) (*models.Record, error)
	GetAuthTokenSecret() string
}

func AuthMiddleware(app AuthApp) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			token := parts[1]
			record, err := app.FindAuthRecordByToken(token, app.GetAuthTokenSecret())
			if err != nil || record == nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "user", record)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
