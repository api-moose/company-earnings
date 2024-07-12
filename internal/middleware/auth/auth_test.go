package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pocketbase/pocketbase/models"
)

// MockAuthApp implements the AuthApp interface for testing
type MockAuthApp struct {
	record *models.Record
}

func (m *MockAuthApp) FindAuthRecordByToken(token, secret string) (*models.Record, error) {
	if token == "valid_token" {
		return m.record, nil
	}
	return nil, nil
}

func (m *MockAuthApp) GetAuthTokenSecret() string {
	return "secret_key"
}

func TestAuthMiddleware(t *testing.T) {
	record := &models.Record{}
	app := &MockAuthApp{
		record: record,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(*models.Record)
		if user != record {
			t.Errorf("Expected user record not found in context")
		}
	})

	mw := AuthMiddleware(app)(handler)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{"Valid token", "Bearer valid_token", http.StatusOK},
		{"Missing token", "", http.StatusUnauthorized},
		{"Invalid token", "Bearer invalid_token", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.token != "" {
				req.Header.Add("Authorization", tt.token)
			}
			w := httptest.NewRecorder()

			mw.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
