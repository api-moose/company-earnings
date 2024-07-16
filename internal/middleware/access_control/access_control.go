package access_control

import (
	"log"
	"net/http"

	"github.com/api-moose/company-earnings/internal/middleware/auth"
	"github.com/api-moose/company-earnings/internal/middleware/tenancy"
)

func RBACMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the route exists
		if !routeExists(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

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

		if !isAuthorized(user.Role, r.URL.Path) {
			log.Printf("RBACMiddleware: User not authorized. Role: %s, Path: %s", user.Role, r.URL.Path)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		log.Printf("RBACMiddleware: Access granted. User ID: %s, Role: %s, Path: %s", user.ID, user.Role, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func isAuthorized(role, path string) bool {
	switch role {
	case "admin":
		return true // Admins have access to all routes
	case "user":
		return path == "/" || path == "/user" || path == "/health" || path == "/version"
	default:
		return false // Unknown roles have no access
	}
}

func routeExists(path string) bool {
	knownRoutes := []string{"/", "/admin", "/user", "/health", "/version"}
	for _, route := range knownRoutes {
		if path == route {
			return true
		}
	}
	return false
}
