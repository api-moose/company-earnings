package models

import (
	"testing"

	"github.com/pocketbase/pocketbase/models"
)

func TestNewUser(t *testing.T) {
	user := NewUser()
	if user == nil {
		t.Fatal("Expected non-nil User")
	}
	if user.Collection != "users" {
		t.Errorf("Expected Collection to be 'users', got %s", user.Collection)
	}
}

func TestUserValidation(t *testing.T) {
	tests := []struct {
		name      string
		user      *User
		wantError bool
	}{
		{
			name: "Valid user",
			user: &User{
				Model: models.Model{
					Id: "1234567890",
				},
				Username: "testuser",
				Email:    "test@example.com",
				Name:     "Test User",
			},
			wantError: false,
		},
		{
			name: "Missing username",
			user: &User{
				Model: models.Model{
					Id: "1234567890",
				},
				Email: "test@example.com",
				Name:  "Test User",
			},
			wantError: true,
		},
		{
			name: "Missing email",
			user: &User{
				Model: models.Model{
					Id: "1234567890",
				},
				Username: "testuser",
				Name:     "Test User",
			},
			wantError: true,
		},
		{
			name: "Missing name",
			user: &User{
				Model: models.Model{
					Id: "1234567890",
				},
				Username: "testuser",
				Email:    "test@example.com",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("User.Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
