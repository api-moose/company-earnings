package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/api-moose/company-earnings/internal/db/mongo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockFirebaseAuthClient struct {
	mock.Mock
}

func (m *MockFirebaseAuthClient) VerifyIDToken(ctx context.Context, idToken string) (*firebaseAuth.Token, error) {
	args := m.Called(ctx, idToken)
	if args.Get(0) != nil {
		return args.Get(0).(*firebaseAuth.Token), args.Error(1)
	}
	return nil, args.Error(1)
}

type MockAuthHandler struct {
	mock.Mock
}

func (m *MockAuthHandler) AuthenticateUser(ctx context.Context, token string) (*mongo.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) != nil {
		return args.Get(0).(*mongo.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestAuthMiddleware(t *testing.T) {
	mockAuthHandler := new(MockAuthHandler)

	validUser := &mongo.User{ID: "valid_user", Email: "test@example.com", Role: "user", TenantID: "tenant1"}

	mockAuthHandler.On("AuthenticateUser", mock.Anything, "valid_token").Return(validUser, nil)
	mockAuthHandler.On("AuthenticateUser", mock.Anything, "invalid_token").Return(nil, assert.AnError)

	am := NewAuthMiddleware(nil, mockAuthHandler)

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
		{"Invalid token", "Bearer invalid_token", http.StatusUnauthorized},
		{"Missing token", "", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}
			w := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}

	mockAuthHandler.AssertExpectations(t)
}
