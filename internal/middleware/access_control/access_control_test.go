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
		{"Admin access to admin route", models.NewUser("1", "admin", "admin@example.com", "admin", "tenant1"), "/admin", "tenant1", http.StatusOK},
		{"Admin access to user route", models.NewUser("1", "admin", "admin@example.com", "admin", "tenant1"), "/user", "tenant1", http.StatusOK},
		{"User access to user route", models.NewUser("2", "user", "user@example.com", "user", "tenant1"), "/user", "tenant1", http.StatusOK},
		{"User access to admin route", models.NewUser("2", "user", "user@example.com", "user", "tenant1"), "/admin", "tenant1", http.StatusForbidden},
		{"No role", models.NewUser("3", "norole", "norole@example.com", "", "tenant1"), "/user", "tenant1", http.StatusUnauthorized},
		{"Invalid role", models.NewUser("4", "invalid", "invalid@example.com", "invalid", "tenant1"), "/user", "tenant1", http.StatusForbidden},
		{"Cross-tenant access attempt", models.NewUser("1", "admin", "admin@example.com", "admin", "tenant1"), "/admin", "tenant2", http.StatusForbidden},
		{"Non-existent route", models.NewUser("2", "user", "user@example.com", "user", "tenant1"), "/nonexistent", "tenant1", http.StatusNotFound},
		{"User access to root route", models.NewUser("2", "user", "user@example.com", "user", "tenant1"), "/", "tenant1", http.StatusOK},
		{"User access to health route", models.NewUser("2", "user", "user@example.com", "user", "tenant1"), "/health", "tenant1", http.StatusOK},
		{"User access to version route", models.NewUser("2", "user", "user@example.com", "user", "tenant1"), "/version", "tenant1", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Use(RBACMiddleware)
			r.Get("/admin", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
			r.Get("/user", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
			r.Get("/", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
			r.Get("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
			r.Get("/version", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
			r.NotFound(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "404 page not found", http.StatusNotFound)
			})

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

func TestIsAuthorized(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		path     string
		expected bool
	}{
		{"Admin access to admin route", "admin", "/admin", true},
		{"Admin access to user route", "admin", "/user", true},
		{"Admin access to root route", "admin", "/", true},
		{"Admin access to health route", "admin", "/health", true},
		{"Admin access to version route", "admin", "/version", true},
		{"Admin access to non-existent route", "admin", "/nonexistent", true},
		{"User access to user route", "user", "/user", true},
		{"User access to root route", "user", "/", true},
		{"User access to health route", "user", "/health", true},
		{"User access to version route", "user", "/version", true},
		{"User access to admin route", "user", "/admin", false},
		{"User access to non-existent route", "user", "/nonexistent", false},
		{"Invalid role access to user route", "invalid", "/user", false},
		{"Invalid role access to admin route", "invalid", "/admin", false},
		{"Invalid role access to root route", "invalid", "/", false},
		{"Invalid role access to health route", "invalid", "/health", false},
		{"Invalid role access to version route", "invalid", "/version", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAuthorized(tt.role, tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRouteExists(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"Root route", "/", true},
		{"Admin route", "/admin", true},
		{"User route", "/user", true},
		{"Health route", "/health", true},
		{"Version route", "/version", true},
		{"Non-existent route", "/nonexistent", false},
		{"Partial match route", "/admi", false},
		{"Case-sensitive route", "/ADMIN", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := routeExists(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}
