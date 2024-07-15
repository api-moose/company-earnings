package auth

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

func TestAuthMiddleware(t *testing.T) {
	mockClient := new(MockFirebaseClient)
	validToken := &auth.Token{UID: "valid_user", Claims: map[string]interface{}{"email": "test@example.com"}}
	mockClient.On("VerifyIDToken", mock.Anything, "valid_token").Return(validToken, nil)
	mockClient.On("VerifyIDToken", mock.Anything, "invalid_token").Return(nil, assert.AnError)

	am := NewAuthMiddleware(mockClient)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r)
		if !ok {
			t.Error("Expected user in context, got none")
		}
		if user.Email != "test@example.com" {
			t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
		}
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := am.Middleware(handler)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{"Valid token", "Bearer valid_token", http.StatusOK},
		{"Missing token", "", http.StatusUnauthorized},
		{"Invalid token", "Bearer invalid_token", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.token != "" {
				req.Header.Add("Authorization", tt.token)
			}
			w := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}

	mockClient.AssertExpectations(t)
}
