package auth

import (
	"errors"
	"testing"

	"github.com/pocketbase/pocketbase/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPocketBaseAdapter mocks the PocketBaseAdapter for testing
type MockPocketBaseAdapter struct {
	mock.Mock
}

func (m *MockPocketBaseAdapter) FindAuthRecordByToken(token, secret string) (*models.Record, error) {
	args := m.Called(token, secret)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Record), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPocketBaseAdapter) GetAuthTokenSecret() string {
	args := m.Called()
	return args.String(0)
}

// Define ErrInvalidToken in config package
var ErrInvalidToken = errors.New("invalid token")

func TestAuthService_AuthenticateUser(t *testing.T) {
	mockAdapter := new(MockPocketBaseAdapter)
	mockRecord := &models.Record{}

	mockAdapter.On("FindAuthRecordByToken", "valid_token", "secret_key").Return(mockRecord, nil)
	mockAdapter.On("FindAuthRecordByToken", "invalid_token", "secret_key").Return(nil, ErrInvalidToken)
	mockAdapter.On("GetAuthTokenSecret").Return("secret_key")

	authService := NewAuthService(mockAdapter)

	tests := []struct {
		name        string
		token       string
		expectedErr error
		expectedRec *models.Record
	}{
		{"Valid token", "valid_token", nil, mockRecord},
		{"Invalid token", "invalid_token", ErrInvalidToken, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record, err := authService.AuthenticateUser(tt.token)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedRec, record)
		})
	}

	mockAdapter.AssertExpectations(t)
}
