package tests

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/api-moose/company-earnings/internal/middleware"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

// MockApp implements a minimal version of core.App for testing
type MockApp struct {
	UserData *models.User
}

func (m *MockApp) Dao() *daos.Dao {
	return &daos.Dao{} // This is a stub. Implement methods as needed for testing.
}

func (m *MockApp) FindUserById(id string) (*models.User, error) {
	if m.UserData != nil && m.UserData.Id == id {
		return m.UserData, nil
	}
	return nil, errors.New("user not found")
}

func TestAuthAndRBACMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		role           string
		path           string
		tenantID       string
		expectedStatus int
	}{
		{"Valid admin access", "valid_token", "admin", "/admin", "tenant1", http.StatusOK},
		{"Valid user access", "valid_token", "user", "/user", "tenant1", http.StatusOK},
		{"User accessing admin route", "valid_token", "user", "/admin", "tenant1", http.StatusForbidden},
		{"Missing token", "", "user", "/user", "tenant1", http.StatusUnauthorized},
		{"Invalid token", "invalid_token", "user", "/user", "tenant1", http.StatusUnauthorized},
		{"Cross-tenant access attempt", "valid_token", "admin", "/admin", "tenant2", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockApp := &MockApp{
				UserData: &models.User{
					Id:    "user1",
					Email: "user1@example.com",
				},
			}
			mockApp.UserData.SetDataValue("role", tt.role)

			req, err := http.NewRequest("GET", tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}

			// Set tenant ID in context
			ctx := context.WithValue(req.Context(), middleware.TenantContextKey, tt.tenantID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			// Chain the middlewares
			handler := middleware.AuthMiddleware(mockApp)(middleware.RBACMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})))

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}
