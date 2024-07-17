package auth

import (
	"context"
	"log"
	"net/http"
	"strings"

	firebaseAuth "firebase.google.com/go/v4/auth"
	authHandler "github.com/api-moose/company-earnings/internal/api/v1/auth"
	"github.com/api-moose/company-earnings/internal/db/mongo"
)

type ContextKey string

const UserContextKey ContextKey = "user"

type FirebaseAuthClient interface {
	VerifyIDToken(ctx context.Context, idToken string) (*firebaseAuth.Token, error)
}

type AuthMiddleware struct {
	client      FirebaseAuthClient
	authHandler authHandler.AuthenticatorHandler
}

func NewAuthMiddleware(client FirebaseAuthClient, authHandler authHandler.AuthenticatorHandler) *AuthMiddleware {
	return &AuthMiddleware{client: client, authHandler: authHandler}
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

		// Use the authHandler to authenticate the user
		user, err := am.authHandler.AuthenticateUser(r.Context(), token)
		if err != nil {
			log.Printf("AuthMiddleware: Error authenticating user: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if user == nil {
			log.Println("AuthMiddleware: User is nil after authentication")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		log.Printf("AuthMiddleware: User authenticated: ID=%s, Email=%s, Role=%s, TenantID=%s", user.ID, user.Email, user.Role, user.TenantID)
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(r *http.Request) (*mongo.User, bool) {
	user, ok := r.Context().Value(UserContextKey).(*mongo.User)
	if !ok {
		log.Println("GetUserFromContext: User not found in context")
	} else {
		log.Printf("GetUserFromContext: User found in context: ID=%s, Email=%s, Role=%s, TenantID=%s", user.ID, user.Email, user.Role, user.TenantID)
	}
	return user, ok
}
