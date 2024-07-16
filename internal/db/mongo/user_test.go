package mongo

import (
	"testing"
)

func TestNewUser(t *testing.T) {
	user := NewUser("1", "testuser", "test@example.com", "user", "tenant1")

	if user.ID != "1" {
		t.Errorf("Expected ID '1', got '%s'", user.ID)
	}
	if user.Username != "testuser" {
		t.Errorf("Expected Username 'testuser', got '%s'", user.Username)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected Email 'test@example.com', got '%s'", user.Email)
	}
	if user.Role != "user" {
		t.Errorf("Expected Role 'user', got '%s'", user.Role)
	}
	if user.TenantID != "tenant1" {
		t.Errorf("Expected TenantID 'tenant1', got '%s'", user.TenantID)
	}
}

func TestUserValidate(t *testing.T) {
	tests := []struct {
		name    string
		user    *User
		wantErr bool
	}{
		{
			name:    "Valid user",
			user:    NewUser("1", "testuser", "test@example.com", "user", "tenant1"),
			wantErr: false,
		},
		{
			name:    "Missing ID",
			user:    NewUser("", "testuser", "test@example.com", "user", "tenant1"),
			wantErr: true,
		},
		{
			name:    "Missing Username",
			user:    NewUser("1", "", "test@example.com", "user", "tenant1"),
			wantErr: false, // Username is not required in the current Validate method
		},
		{
			name:    "Missing Email",
			user:    NewUser("1", "testuser", "", "user", "tenant1"),
			wantErr: true,
		},
		{
			name:    "Missing Role",
			user:    NewUser("1", "testuser", "test@example.com", "", "tenant1"),
			wantErr: true,
		},
		{
			name:    "Missing TenantID",
			user:    NewUser("1", "testuser", "test@example.com", "user", ""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("User.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
