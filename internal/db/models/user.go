package models

import "errors"

// User represents a user in the system
type User struct {
	ID       string
	Username string
	Email    string
	Role     string
	Token    string
	TenantID string
}

// NewUser creates a new User instance
func NewUser(id, username, email, role, token, tenantID string) *User {
	return &User{
		ID:       id,
		Username: username,
		Email:    email,
		Role:     role,
		Token:    token,
		TenantID: tenantID,
	}
}

// Validate checks if the user data is valid
func (u *User) Validate() error {
	if u.ID == "" {
		return errors.New("id is required")
	}
	if u.Username == "" {
		return errors.New("username is required")
	}
	if u.Email == "" {
		return errors.New("email is required")
	}
	if u.Role == "" {
		return errors.New("role is required")
	}
	if u.Token == "" {
		return errors.New("token is required")
	}
	if u.TenantID == "" {
		return errors.New("tenantID is required")
	}
	return nil
}
