// access_control_test.go
package access_control

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/api-moose/company-earnings/internal/middleware/tenancy"
	"github.com/pocketbase/pocketbase/models"
)

func TestRBACMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		role           string
		path           string
		tenantID       string
		userTenantID   string
		expectedStatus int
	}{
		{"Admin access to admin route", "admin", "/admin", "tenant1", "tenant1", http.StatusOK},
		{"Admin access to user route", "admin", "/user", "tenant1", "tenant1", http.StatusOK},
		{"User access to user route", "user", "/user", "tenant1", "tenant1", http.StatusOK},
		{"User access to admin route", "user", "/admin", "tenant1", "tenant1", http.StatusForbidden},
		{"No role", "", "/user", "tenant1", "tenant1", http.StatusUnauthorized},
		{"Invalid role", "invalid", "/user", "tenant1", "tenant1", http.StatusForbidden},
		{"Cross-tenant access attempt", "admin", "/admin", "tenant1", "tenant2", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			// Set up the context with tenant and user information
			ctx := req.Context()
			ctx = context.WithValue(ctx, tenancy.TenantContextKey, tt.tenantID)
			user := models.NewRecord(&models.Collection{})
			user.Set("role", tt.role)
			user.Set("tenantId", tt.userTenantID)
			ctx = context.WithValue(ctx, "user", user)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler := RBACMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}