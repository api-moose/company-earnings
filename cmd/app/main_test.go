package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/api-moose/company-earnings/internal/middleware/access_control"
	auth "github.com/api-moose/company-earnings/internal/middleware/auth"
	"github.com/api-moose/company-earnings/internal/middleware/tenancy"
	"github.com/api-moose/company-earnings/internal/utils/logging"
)

type MockFirebaseAuthClient struct {
	mock.Mock
}

func (m *MockFirebaseAuthClient) VerifyIDToken(ctx context.Context, idToken string) (*firebaseAuth.Token, error) {
	args := m.Called(ctx, idToken)
	if args.Get(0) != nil {
		return args.Get(0).(*firebaseAuth.Token), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestMain(m *testing.M) {
	err := os.Chdir("../..")
	if err != nil {
		log.Fatalf("Error changing directory: %v", err)
	}

	absPath, err := filepath.Abs(".env")
	if err != nil {
		log.Fatalf("Error getting absolute path: %v", err)
	}

	err = LoadEnv(absPath)
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	code := m.Run()
	os.Exit(code)
}

func setupTestRouter() *chi.Mux {
	mockAuth := new(MockFirebaseAuthClient)
	validToken := &firebaseAuth.Token{
		UID: "valid_user",
		Claims: map[string]interface{}{
			"email":    "test@example.com",
			"tenantID": "tenant1",
			"role":     "user",
		},
	}
	mockAuth.On("VerifyIDToken", mock.Anything, "valid-token").Return(validToken, nil)
	mockAuth.On("VerifyIDToken", mock.Anything, "invalid-token").Return(nil, fmt.Errorf("invalid token"))

	r := chi.NewRouter()
	r.Use(logging.LoggingMiddleware)
	r.Use(tenancy.NewTenantMiddleware(mockAuth).Middleware)
	r.Use(auth.NewAuthMiddleware(mockAuth).Middleware)
	r.Use(access_control.RBACMiddleware)

	r.Get("/", mainHandler)
	r.Get("/health", healthCheckHandler)
	r.Get("/version", versionHandler)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, notFoundMessage, http.StatusNotFound)
	})

	return r
}

func TestSetupRouter(t *testing.T) {
	router := setupTestRouter()

	testCases := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   interface{}
	}{
		{"Main route", "/", http.StatusOK, "Welcome to the Financial Data Platform API"},
		{"Health check route", "/health", http.StatusOK, map[string]string{"status": "healthy"}},
		{"Version route", "/version", http.StatusOK, map[string]string{"version": "0.1.0"}},
		{"Nonexistent route", "/nonexistent", http.StatusNotFound, notFoundMessage + "\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tc.path, nil)
			assert.NoError(t, err)

			req.Header.Set("Authorization", "Bearer valid-token")
			req.Header.Set("X-Tenant-ID", "tenant1")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)

			if _, ok := tc.expectedBody.(string); ok {
				assert.Equal(t, tc.expectedBody, rr.Body.String())
			} else {
				var actualBody map[string]string
				err = json.Unmarshal(rr.Body.Bytes(), &actualBody)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedBody, actualBody)
			}
		})
	}
}

func TestAuthMiddleware(t *testing.T) {
	router := setupTestRouter()

	testCases := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{"Valid token", "Bearer valid-token", http.StatusOK},
		{"Invalid token", "Bearer invalid-token", http.StatusUnauthorized},
		{"Missing token", "", http.StatusUnauthorized},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tc.token != "" {
				req.Header.Set("Authorization", tc.token)
			}
			req.Header.Set("X-Tenant-ID", "tenant1")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}

func TestTenancyMiddleware(t *testing.T) {
	router := setupTestRouter()

	testCases := []struct {
		name           string
		tenantID       string
		expectedStatus int
	}{
		{"Valid tenant ID", "tenant1", http.StatusOK},
		{"Missing tenant ID", "", http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			if tc.tenantID != "" {
				req.Header.Set("X-Tenant-ID", tc.tenantID)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}

func LoadEnv(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		os.Setenv(key, value)
	}

	return nil
}
