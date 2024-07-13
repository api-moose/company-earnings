package access_control

import (
	"log"
	"net/http"

	"github.com/api-moose/company-earnings/internal/middleware/auth"
	"github.com/api-moose/company-earnings/internal/middleware/tenancy"
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

		role, _ := user.Get("role").(string)
		userTenantID, _ := user.Get("tenantId").(string)

		if role == "" {
			log.Println("RBACMiddleware: Role not found in user record")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if userTenantID != tenantID {
			log.Println("RBACMiddleware: Tenant ID mismatch")
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if !isAuthorized(role, r.URL.Path) {
			log.Println("RBACMiddleware: User not authorized")
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isAuthorized(role, path string) bool {
	switch role {
	case "admin":
		return true // Admins have access to all routes
	case "user":
		return path == "/user" // Users only have access to the /user route
	default:
		return false // Unknown roles have no access
	}
}
