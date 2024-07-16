package auth

import (
	"context"

	"firebase.google.com/go/v4/auth"
	"github.com/api-moose/company-earnings/internal/db/models"
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

func (s *AuthService) AuthenticateUser(ctx context.Context, token string) (*models.User, error) {
	decodedToken, err := s.client.VerifyIDToken(ctx, token)
	if err != nil {
		return nil, err
	}

	firebaseUser, err := s.client.GetUser(ctx, decodedToken.UID)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:       firebaseUser.UID,
		Username: firebaseUser.DisplayName,
		Email:    firebaseUser.Email,
		Role:     getRoleFromClaims(decodedToken.Claims),
		TenantID: getTenantIDFromClaims(decodedToken.Claims),
	}

	return user, nil
}

func getRoleFromClaims(claims map[string]interface{}) string {
	if role, ok := claims["role"].(string); ok {
		return role
	}
	return "user" // Default role if not specified
}

func getTenantIDFromClaims(claims map[string]interface{}) string {
	if tenantID, ok := claims["tenantID"].(string); ok {
		return tenantID
	}
	return "" // Empty string if not specified
}
