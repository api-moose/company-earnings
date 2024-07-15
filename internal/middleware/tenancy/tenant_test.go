package tenancy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"firebase.google.com/go/v4/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockFirebaseClient struct {
	mock.Mock
}

func (m *MockFirebaseClient) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	args := m.Called(ctx, idToken)
	if args.Get(0) != nil {
		return args.Get(0).(*auth.Token), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestTenantMiddleware(t *testing.T) {
	mockClient := new(MockFirebaseClient)

	validToken := &auth.Token{
		UID: "valid_user",
		Claims: map[string]interface{}{
			"tenantID": "tenant1",
		},
	}
	mockClient.On("VerifyIDToken", mock.Anything, "valid_token").Return(validToken, nil)
	mockClient.On("VerifyIDToken", mock.Anything, "invalid_token").Return(nil, assert.AnError)

	tm := NewTenantMiddleware(mockClient)

	tests := []struct {
		name           string
		token          string
		tenantID       string
		expectedStatus int
	}{
		{"Valid token and tenant ID", "Bearer valid_token", "tenant1", http.StatusOK},
		{"Missing token", "", "tenant1", http.StatusUnauthorized},
		{"Invalid token", "Bearer invalid_token", "tenant1", http.StatusUnauthorized},
		{"Tenant ID mismatch", "Bearer valid_token", "tenant2", http.StatusUnauthorized},
		{"Missing tenant ID", "Bearer valid_token", "", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			assert.NoError(t, err)

			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}
			if tt.tenantID != "" {
				req.Header.Set("X-Tenant-ID", tt.tenantID)
			}

			rr := httptest.NewRecorder()
			handler := tm.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tenantID, ok := GetTenantID(r)
				assert.True(t, ok)
				assert.Equal(t, tt.tenantID, tenantID)
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rr, req)
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}

	mockClient.AssertExpectations(t)
}
