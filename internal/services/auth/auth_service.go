package auth

import (
	"github.com/api-moose/company-earnings/internal/config"
	"github.com/pocketbase/pocketbase/models"
)

type AuthService struct {
	adapter config.AuthProvider
}

func NewAuthService(adapter config.AuthProvider) *AuthService {
	return &AuthService{adapter: adapter}
}

func (s *AuthService) AuthenticateUser(token string) (*models.Record, error) {
	secret := s.adapter.GetAuthTokenSecret()
	return s.adapter.FindAuthRecordByToken(token, secret)
}
