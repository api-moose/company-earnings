package access_control

import (
	"net/http"

	"github.com/api-moose/company-earnings/internal/middleware/tenancy"
	"github.com/pocketbase/pocketbase/models"
)

func RBACMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value("user").(*models.Record)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tenantID, ok := tenancy.GetTenantID(r)
		if !ok {
			http.Error(w, "Tenant context not found", http.StatusInternalServerError)
			return
		}

		role, _ := user.Get("role").(string)
		if !isAuthorized(role, r.URL.Path, tenantID) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isAuthorized(role, path, tenantID string) bool {
	// This is a simple authorization logic. In a real-world scenario,
	// you might want to use a more sophisticated system, possibly involving a database lookup.
	switch role {
	case "admin":
		return true // Admins have access to all routes
	case "user":
		return path == "/user" // Users only have access to the /user route
	default:
		return false // Unknown roles have no access
	}
}
