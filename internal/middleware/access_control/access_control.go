package access_control

import (
	"log"
	"net/http"

	"github.com/api-moose/company-earnings/internal/middleware/auth"
	"github.com/api-moose/company-earnings/internal/middleware/tenancy"
	"github.com/go-chi/chi/v5"
)

func RBACMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.GetUserFromContext(r)
		if !ok {
			log.Println("RBACMiddleware: User not found in context")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tenantID, ok := tenancy.GetTenantID(r)
		if !ok {
			log.Println("RBACMiddleware: Tenant context not found")
			http.Error(w, "Tenant context not found", http.StatusInternalServerError)
			return
		}

		if user.Role == "" {
			log.Println("RBACMiddleware: Role not found in user record")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if user.TenantID != tenantID {
			log.Printf("RBACMiddleware: Tenant ID mismatch. User TenantID: %s, Request TenantID: %s", user.TenantID, tenantID)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Check if the route exists
		rctx := chi.RouteContext(r.Context())
		if rctx.RoutePattern() != "" {
			// This is an existing route, apply RBAC
			if !isAuthorized(user.Role, r.URL.Path) {
				log.Printf("RBACMiddleware: User not authorized. Role: %s, Path: %s", user.Role, r.URL.Path)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}
		// If the route doesn't exist, let it pass through to be handled by NotFound

		log.Printf("RBACMiddleware: Access granted. User ID: %s, Role: %s, Path: %s", user.ID, user.Role, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func isAuthorized(role, path string) bool {
	switch role {
	case "admin":
		return true // Admins have access to all routes
	case "user":
		return path == "/" || path == "/user" || path == "/health" || path == "/version" // Users have access to these routes
	default:
		return false // Unknown roles have no access
	}
}
