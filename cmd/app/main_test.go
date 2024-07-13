package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/api-moose/company-earnings/internal/config"
	"github.com/api-moose/company-earnings/internal/middleware/auth"
	"github.com/api-moose/company-earnings/internal/middleware/tenancy"
	"github.com/pocketbase/pocketbase/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthProvider struct {
	mock.Mock
}

func (m *MockAuthProvider) FindAuthRecordByToken(token, secret string) (*models.Record, error) {
	args := m.Called(token, secret)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Record), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthProvider) GetAuthTokenSecret() string {
	args := m.Called()
	return args.String(0)
}

func createMockRecord() *models.Record {
	collection := &models.Collection{
		BaseModel: models.BaseModel{
			Id: "mockCollectionId",
		},
		Name: "mockCollection",
	}

	record := models.NewRecord(collection)
	record.SetId("mockId")
	record.Set("role", "admin")
	record.Set("tenantId", "test-tenant")
	return record
}

func mockAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockRecord := createMockRecord()
		ctx := context.WithValue(r.Context(), auth.UserContextKey, mockRecord)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func setupTestRouter(pbConfig *config.PocketBaseConfig) http.Handler {
	return setupRouter(pbConfig, mockAuthMiddleware)
}

func TestMainHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(mainHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "Welcome to the Financial Data Platform API", rr.Body.String())
}

func TestHealthCheckHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthCheckHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `{"status":"healthy"}`, rr.Body.String())
}

func TestVersionHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/version", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(versionHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "version")
}

func TestSetupRouter(t *testing.T) {
	mockAuth := new(MockAuthProvider)
	mockRecord := createMockRecord()

	mockAuth.On("FindAuthRecordByToken", mock.Anything, mock.Anything).Return(mockRecord, nil)
	mockAuth.On("GetAuthTokenSecret").Return("test_secret")

	appConfig := &config.PocketBaseConfig{Adapter: mockAuth}

	router := setupTestRouter(appConfig)

	testCases := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{"Main route", "/", http.StatusOK, "Welcome to the Financial Data Platform API"},
		{"Health check route", "/health", http.StatusOK, `{"status":"healthy"}`},
		{"Version route", "/version", http.StatusOK, `{"version":"0.1.0"}`},
		{"Nonexistent route", "/nonexistent", http.StatusNotFound, "404 page not found\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tc.path, nil)
			assert.NoError(t, err)

			req.Header.Set("X-Tenant-ID", "test-tenant")
			req.Header.Set("Authorization", "Bearer test-token")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.Equal(t, tc.expectedBody, rr.Body.String())
		})
	}
}

func TestLoggingMiddleware(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	mockAuth := new(MockAuthProvider)
	mockRecord := createMockRecord()

	mockAuth.On("FindAuthRecordByToken", mock.Anything, mock.Anything).Return(mockRecord, nil)
	mockAuth.On("GetAuthTokenSecret").Return("test_secret")

	appConfig := &config.PocketBaseConfig{Adapter: mockAuth}

	router := setupTestRouter(appConfig)

	req, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	req.Header.Set("X-Tenant-ID", "test-tenant")
	req.Header.Set("Authorization", "Bearer test-token")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	logOutput := buf.String()
	assert.Contains(t, logOutput, "GET")
	assert.Contains(t, logOutput, "/")
	assert.Contains(t, logOutput, "200")
}

func TestAuthMiddleware(t *testing.T) {
	mockAuth := new(MockAuthProvider)
	mockRecord := createMockRecord()

	mockAuth.On("FindAuthRecordByToken", "valid-token", "test_secret").Return(mockRecord, nil)
	mockAuth.On("FindAuthRecordByToken", "invalid-token", "test_secret").Return(nil, nil)
	mockAuth.On("GetAuthTokenSecret").Return("test_secret")

	appConfig := &config.PocketBaseConfig{Adapter: mockAuth}

	router := setupTestRouter(appConfig)

	t.Run("Valid token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		req.Header.Set("X-Tenant-ID", "test-tenant")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Invalid token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		req.Header.Set("X-Tenant-ID", "test-tenant")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code) // Updated to match expected behavior
	})
}

func TestTenancyMiddleware(t *testing.T) {
	mockAuth := new(MockAuthProvider)
	mockRecord := createMockRecord()

	mockAuth.On("FindAuthRecordByToken", mock.Anything, mock.Anything).Return(mockRecord, nil)
	mockAuth.On("GetAuthTokenSecret").Return("test_secret")

	appConfig := &config.PocketBaseConfig{Adapter: mockAuth}

	router := setupTestRouter(appConfig)

	t.Run("Valid tenant ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Tenant-ID", "test-tenant")
		req.Header.Set("Authorization", "Bearer test-token")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)

		tenantID, ok := tenancy.GetTenantID(req)
		assert.True(t, ok)
		assert.Equal(t, "test-tenant", tenantID)
	})

	t.Run("Missing tenant ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestMain(m *testing.M) {
	// Setup code here (if needed)
	exitCode := m.Run()
	// Teardown code here (if needed)
	os.Exit(exitCode)
}
