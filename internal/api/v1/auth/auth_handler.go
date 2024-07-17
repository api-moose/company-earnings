package auth

import (
	"context"

	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/api-moose/company-earnings/internal/db/mongo"
)

type FirebaseAuthClient interface {
	VerifyIDToken(ctx context.Context, idToken string) (*firebaseAuth.Token, error)
	GetUser(ctx context.Context, uid string) (*firebaseAuth.UserRecord, error)
}

type AuthenticatorHandler interface {
	AuthenticateUser(ctx context.Context, token string) (*mongo.User, error)
}

type Handler struct {
	client FirebaseAuthClient
}

func NewHandler(client FirebaseAuthClient) *Handler {
	return &Handler{client: client}
}

func (h *Handler) AuthenticateUser(ctx context.Context, token string) (*mongo.User, error) {
	decodedToken, err := h.client.VerifyIDToken(ctx, token)
	if err != nil {
		return nil, err
	}

	firebaseUser, err := h.client.GetUser(ctx, decodedToken.UID)
	if err != nil {
		return nil, err
	}

	user := &mongo.User{
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
