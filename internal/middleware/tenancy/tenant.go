package tenancy

import (
	"context"
	"log"
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
)

type contextKey string

const TenantContextKey contextKey = "tenantID"

// Define an interface for the Firebase Auth client
type FirebaseAuthClient interface {
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
}

type TenantMiddleware struct {
	client FirebaseAuthClient
}

func NewTenantMiddleware(client FirebaseAuthClient) *TenantMiddleware {
	return &TenantMiddleware{client: client}
}

func (tm *TenantMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Entering TenantMiddleware")

		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			log.Println("TenantMiddleware: Tenant ID is required")
			http.Error(w, "Tenant ID is required", http.StatusBadRequest)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Println("TenantMiddleware: Missing authorization header")
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			log.Println("TenantMiddleware: Invalid authorization header format")
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		decodedToken, err := tm.client.VerifyIDToken(context.Background(), token)
		if err != nil {
			log.Printf("TenantMiddleware: Error verifying ID token: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if decodedToken.Claims["tenantID"] != tenantID {
			log.Println("TenantMiddleware: Tenant ID mismatch")
			http.Error(w, "Tenant ID mismatch", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), TenantContextKey, tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))

		log.Println("Exiting TenantMiddleware")
	})
}

// GetTenantID retrieves the tenant ID from the request context
func GetTenantID(r *http.Request) (string, bool) {
	tenantID, ok := r.Context().Value(TenantContextKey).(string)
	return tenantID, ok
}
