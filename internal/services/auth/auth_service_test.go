package auth

import (
	"context"
	"errors"
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

func (m *MockFirebaseClient) GetUser(ctx context.Context, uid string) (*auth.UserRecord, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) != nil {
		return args.Get(0).(*auth.UserRecord), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestAuthService_AuthenticateUser(t *testing.T) {
	mockClient := new(MockFirebaseClient)
	validToken := &auth.Token{UID: "valid_user"}
	userRecord := &auth.UserRecord{UserInfo: &auth.UserInfo{UID: "valid_user", Email: "test@example.com"}}

	mockClient.On("VerifyIDToken", mock.Anything, "valid_token").Return(validToken, nil)
	mockClient.On("VerifyIDToken", mock.Anything, "invalid_token").Return(nil, errors.New("invalid token"))
	mockClient.On("GetUser", mock.Anything, "valid_user").Return(userRecord, nil)

	authService := NewAuthService(mockClient)

	tests := []struct {
		name        string
		token       string
		expectedErr error
		expectedRec *auth.UserRecord
	}{
		{"Valid token", "valid_token", nil, userRecord},
		{"Invalid token", "invalid_token", errors.New("invalid token"), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record, err := authService.AuthenticateUser(context.Background(), tt.token)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedRec, record)
		})
	}

	mockClient.AssertExpectations(t)
}
