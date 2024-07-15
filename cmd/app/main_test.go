package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	firebase "firebase.google.com/go/v4"
	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/api-moose/company-earnings/internal/db/models"
	authMiddleware "github.com/api-moose/company-earnings/internal/middleware/auth"
	"github.com/api-moose/company-earnings/internal/middleware/tenancy"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/option"
)

var authClient *firebaseAuth.Client

// LoadEnv loads environment variables from a .env file
func LoadEnv(filepath string) error {
	content, err := ioutil.ReadFile(filepath)
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

func TestMain(m *testing.M) {
	// Get the absolute path to the .env file
	absPath, err := filepath.Abs(".env")
	if err != nil {
		log.Fatalf("Error getting absolute path: %v", err)
	}

	// Load the .env file
	err = LoadEnv(absPath)
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize Firebase
	opt := option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}
	authClient, err = app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v", err)
	}

	// Redirect logs to a buffer for testing
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)

	// Run the tests
	exitCode := m.Run()

	// Print the logs if any test failed
	if exitCode != 0 {
		fmt.Println("Test logs:")
		fmt.Println(logBuffer.String())
	}

	// Reset the log output
	log.SetOutput(os.Stderr)

	os.Exit(exitCode)
}

func setupTestRouter() http.Handler {
	return setupRouter(authClient)
}

func TestMainHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := setupTestRouter()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "Welcome to the Financial Data Platform API", rr.Body.String())
}

func TestHealthCheckHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := setupTestRouter()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `{"status":"healthy"}`, strings.TrimSpace(rr.Body.String()))
}

func TestVersionHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/version", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := setupTestRouter()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "version")
}

func TestSetupRouter(t *testing.T) {
	router := setupTestRouter()

	testCases := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{"Main route", "/", http.StatusOK, "Welcome to the Financial Data Platform API"},
		{"Health check route", "/health", http.StatusOK, `{"status":"healthy"}`},
		{"Version route", "/version", http.StatusOK, `{"version":"0.1.0"}`},
		{"Nonexistent route", "/nonexistent", http.StatusNotFound, "404 page not found"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tc.path, nil)
			assert.NoError(t, err)

			// Assuming you have valid tokens and tenant IDs for testing
			req.Header.Set("X-Tenant-ID", "tenant1")
			req.Header.Set("Authorization", "Bearer valid-token")

			log.Printf("TestSetupRouter: Starting test case: %s", tc.name)

			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			// Use the dumpRequestResponse helper function
			dumpRequestResponse(t, req, rr)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.Equal(t, tc.expectedBody, strings.TrimSpace(rr.Body.String()))

			// Check if user is in context after request
			ctx := req.Context()
			userFromCtx, ok := ctx.Value(authMiddleware.UserContextKey).(*models.User)
			if ok {
				log.Printf("TestSetupRouter: User found in context after request: %+v", userFromCtx)
			} else {
				log.Println("TestSetupRouter: User not found in context after request")
			}
		})
	}
}

func TestLoggingMiddleware(t *testing.T) {
	router := setupTestRouter()

	req, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	// Assuming you have valid tokens and tenant IDs for testing
	req.Header.Set("X-Tenant-ID", "tenant1")
	req.Header.Set("Authorization", "Bearer valid-token")

	rr := httptest.NewRecorder()

	log.Println("TestLoggingMiddleware: Starting test")
	log.Printf("TestLoggingMiddleware: Request headers: %v", req.Header)

	router.ServeHTTP(rr, req)

	log.Printf("TestLoggingMiddleware: Response status: %d", rr.Code)
	log.Printf("TestLoggingMiddleware: Response body: %s", rr.Body.String())

	assert.Contains(t, rr.Body.String(), "Welcome to the Financial Data Platform API")
	assert.Equal(t, http.StatusOK, rr.Code)
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

			log.Printf("TestAuthMiddleware: Starting test case: %s", tc.name)
			log.Printf("TestAuthMiddleware: Request headers: %v", req.Header)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			log.Printf("TestAuthMiddleware: Response status: %d", rr.Code)
			log.Printf("TestAuthMiddleware: Response body: %s", rr.Body.String())

			assert.Equal(t, tc.expectedStatus, rr.Code)

			// Check if user is in context after request
			ctx := req.Context()
			userFromCtx, ok := ctx.Value(authMiddleware.UserContextKey).(*models.User)
			if ok {
				log.Printf("TestAuthMiddleware: User found in context after request: %+v", userFromCtx)
			} else {
				log.Println("TestAuthMiddleware: User not found in context after request")
			}
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

			log.Printf("TestTenancyMiddleware: Starting test case: %s", tc.name)
			log.Printf("TestTenancyMiddleware: Request headers: %v", req.Header)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			log.Printf("TestTenancyMiddleware: Response status: %d", rr.Code)
			log.Printf("TestTenancyMiddleware: Response body: %s", rr.Body.String())

			assert.Equal(t, tc.expectedStatus, rr.Code)

			if tc.expectedStatus == http.StatusOK {
				tenantID, ok := tenancy.GetTenantID(req)
				assert.True(t, ok)
				assert.Equal(t, "tenant1", tenantID)
			}
		})
	}
}

// Helper function to dump request and response details
func dumpRequestResponse(t *testing.T, req *http.Request, resp *httptest.ResponseRecorder) {
	t.Logf("Request Method: %s", req.Method)
	t.Logf("Request URL: %s", req.URL.String())
	t.Logf("Request Headers: %v", req.Header)

	t.Logf("Response Status: %d", resp.Code)
	t.Logf("Response Headers: %v", resp.Header())
	t.Logf("Response Body: %s", resp.Body.String())
}
