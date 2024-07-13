package tenancy

import (
	"context"
	"log"
	"net/http"
)

type contextKey string

const TenantContextKey contextKey = "tenantID"

func TenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			log.Println("TenantMiddleware: Tenant ID is required")
			http.Error(w, "Tenant ID is required", http.StatusBadRequest)
			return
		}

		log.Println("TenantMiddleware: Adding tenant ID to context")
		ctx := context.WithValue(r.Context(), TenantContextKey, tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetTenantID retrieves the tenant ID from the request context
func GetTenantID(r *http.Request) (string, bool) {
	tenantID, ok := r.Context().Value(TenantContextKey).(string)
	return tenantID, ok
}