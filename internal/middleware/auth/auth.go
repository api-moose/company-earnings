package auth

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase/models"
)

type ContextKey string

const UserContextKey ContextKey = "user"

type AuthApp interface {
	FindAuthRecordByToken(token, secret string) (*models.Record, error)
	GetAuthTokenSecret() string
}

func AuthMiddleware(app AuthApp) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Println("AuthMiddleware: Missing authorization header")
				http.Error(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				log.Println("AuthMiddleware: Invalid authorization header format")
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			token := parts[1]
			secret := app.GetAuthTokenSecret()
			log.Printf("AuthMiddleware: Token: %s Secret: %s", token, secret)
			record, err := app.FindAuthRecordByToken(token, secret)
			if err != nil || record == nil {
				log.Printf("AuthMiddleware: Invalid token: %s", token)
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			log.Println("AuthMiddleware: Token is valid, adding user to context")
			ctx := context.WithValue(r.Context(), UserContextKey, record)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext retrieves the user from the request context
func GetUserFromContext(r *http.Request) (*models.Record, bool) {
	user, ok := r.Context().Value(UserContextKey).(*models.Record)
	return user, ok
}
