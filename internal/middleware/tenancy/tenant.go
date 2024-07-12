package tenancy

import (
	"context"
	"net/http"
)

type contextKey string

const tenantContextKey contextKey = "tenantID"

func TenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			http.Error(w, "Tenant ID is required", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), tenantContextKey, tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetTenantID retrieves the tenant ID from the request context
func GetTenantID(r *http.Request) (string, bool) {
	tenantID, ok := r.Context().Value(tenantContextKey).(string)
	return tenantID, ok
}
