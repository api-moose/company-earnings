package auth

import (
	"context"
	"log"
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/api-moose/company-earnings/internal/db/models"
)

type ContextKey string

const UserContextKey ContextKey = "user"

// Define an interface for the Firebase Auth client
type FirebaseAuthClient interface {
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
}

type AuthMiddleware struct {
	client FirebaseAuthClient
}

func NewAuthMiddleware(client FirebaseAuthClient) *AuthMiddleware {
	return &AuthMiddleware{client: client}
}

func (am *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Entering AuthMiddleware")
		defer log.Println("Exiting AuthMiddleware")

		authHeader := r.Header.Get("Authorization")
		log.Printf("AuthMiddleware: Authorization header: %s", authHeader)

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
		log.Printf("AuthMiddleware: Extracted token: %s", token)

		decodedToken, err := am.client.VerifyIDToken(context.Background(), token)
		if err != nil {
			log.Printf("AuthMiddleware: Error verifying ID token: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		user := &models.User{
			ID:    decodedToken.UID,
			Email: decodedToken.Claims["email"].(string),
			// Populate other fields from claims as needed
		}

		log.Printf("AuthMiddleware: User authenticated: ID=%s, Email=%s", user.ID, user.Email)
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(r *http.Request) (*models.User, bool) {
	user, ok := r.Context().Value(UserContextKey).(*models.User)
	if !ok {
		log.Println("GetUserFromContext: User not found in context")
	} else {
		log.Printf("GetUserFromContext: User found in context: ID=%s, Email=%s", user.ID, user.Email)
	}
	return user, ok
}
