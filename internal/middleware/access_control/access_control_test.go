package access_control

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/api-moose/company-earnings/internal/db/models"
	"github.com/api-moose/company-earnings/internal/middleware/auth"
	"github.com/api-moose/company-earnings/internal/middleware/tenancy"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestRBACMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		user           *models.User
		path           string
		tenantID       string
		expectedStatus int
	}{
		{
			name:           "Admin access to admin route",
			user:           models.NewUser("1", "admin", "admin@example.com", "admin", "tenant1"),
			path:           "/admin",
			tenantID:       "tenant1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Admin access to user route",
			user:           models.NewUser("1", "admin", "admin@example.com", "admin", "tenant1"),
			path:           "/user",
			tenantID:       "tenant1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "User access to user route",
			user:           models.NewUser("2", "user", "user@example.com", "user", "tenant1"),
			path:           "/user",
			tenantID:       "tenant1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "User access to admin route",
			user:           models.NewUser("2", "user", "user@example.com", "user", "tenant1"),
			path:           "/admin",
			tenantID:       "tenant1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "No role",
			user:           models.NewUser("3", "norole", "norole@example.com", "", "tenant1"),
			path:           "/user",
			tenantID:       "tenant1",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid role",
			user:           models.NewUser("4", "invalid", "invalid@example.com", "invalid", "tenant1"),
			path:           "/user",
			tenantID:       "tenant1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Cross-tenant access attempt",
			user:           models.NewUser("1", "admin", "admin@example.com", "admin", "tenant1"),
			path:           "/admin",
			tenantID:       "tenant2",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Non-existent route",
			user:           models.NewUser("2", "user", "user@example.com", "user", "tenant1"),
			path:           "/nonexistent",
			tenantID:       "tenant1",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Use(RBACMiddleware)
			r.Get("/admin", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
			r.Get("/user", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

			req, err := http.NewRequest("GET", tt.path, nil)
			assert.NoError(t, err)

			ctx := context.WithValue(req.Context(), tenancy.TenantContextKey, tt.tenantID)
			ctx = context.WithValue(ctx, auth.UserContextKey, tt.user)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
