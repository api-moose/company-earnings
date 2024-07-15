package auth

import (
	"context"

	"firebase.google.com/go/v4/auth"
)

type FirebaseAuthClient interface {
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
	GetUser(ctx context.Context, uid string) (*auth.UserRecord, error)
}

type AuthService struct {
	client FirebaseAuthClient
}

func NewAuthService(client FirebaseAuthClient) *AuthService {
	return &AuthService{client: client}
}

func (s *AuthService) AuthenticateUser(ctx context.Context, token string) (*auth.UserRecord, error) {
	decodedToken, err := s.client.VerifyIDToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return s.client.GetUser(ctx, decodedToken.UID)
}
