package auth

import (
	"context"
	"errors"
	"testing"

	"firebase.google.com/go/v4/auth"
	"github.com/api-moose/company-earnings/internal/db/mongo"
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

func (m *MockFirebaseClient) GetUser(ctx context.Context, uid string) (*auth.UserRecord, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) != nil {
		return args.Get(0).(*auth.UserRecord), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestAuthHandler_AuthenticateUser(t *testing.T) {
	mockClient := new(MockFirebaseClient)
	handler := NewHandler(mockClient)

	validToken := &auth.Token{
		UID: "valid_user",
		Claims: map[string]interface{}{
			"role":     "admin",
			"tenantID": "tenant1",
		},
	}
	userRecord := &auth.UserRecord{
		UserInfo: &auth.UserInfo{
			UID:         "valid_user",
			Email:       "test@example.com",
			DisplayName: "Test User",
		},
	}

	mockClient.On("VerifyIDToken", mock.Anything, "valid_token").Return(validToken, nil)
	mockClient.On("VerifyIDToken", mock.Anything, "invalid_token").Return(nil, errors.New("invalid token"))
	mockClient.On("GetUser", mock.Anything, "valid_user").Return(userRecord, nil)

	tests := []struct {
		name         string
		token        string
		expectedUser *mongo.User
		expectedErr  error
	}{
		{
			name:  "Valid token",
			token: "valid_token",
			expectedUser: &mongo.User{
				ID:       "valid_user",
				Username: "Test User",
				Email:    "test@example.com",
				Role:     "admin",
				TenantID: "tenant1",
			},
			expectedErr: nil,
		},
		{
			name:         "Invalid token",
			token:        "invalid_token",
			expectedUser: nil,
			expectedErr:  errors.New("invalid token"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := handler.AuthenticateUser(context.Background(), tt.token)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedUser, user)
		})
	}

	mockClient.AssertExpectations(t)
}
