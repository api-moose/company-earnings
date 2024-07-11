package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/api-moose/company-earnings/internal/config"
)

func TestMainHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(mainHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "Welcome to the Financial Data Platform API"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestHealthCheckHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthCheckHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"status":"healthy"}`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestVersionHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/version", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(versionHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}

	if _, exists := response["version"]; !exists {
		t.Errorf("response does not contain version key")
	}
}
func TestMainHandlerWithError(t *testing.T) {
    req, err := http.NewRequest("GET", "/nonexistent", nil)
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(mainHandler)

    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusNotFound {
        t.Errorf("handler returned wrong status code: got %v want %v",
            status, http.StatusNotFound)
    }

    expected := "HTTP 404: Not Found\n"
    if rr.Body.String() != expected {
        t.Errorf("handler returned unexpected body: got %q want %q",
            rr.Body.String(), expected)
    }
}
func TestSetupRouter(t *testing.T) {
    router := setupRouter()

    testCases := []struct {
        name           string
        path           string
        expectedStatus int
        expectedBody   string
    }{
        {"Main route", "/", http.StatusOK, "Welcome to the Financial Data Platform API"},
        {"Health check route", "/health", http.StatusOK, `{"status":"healthy"}`},
        {"Version route", "/version", http.StatusOK, `{"version":"` + version + `"}`},
        {"Nonexistent route", "/nonexistent", http.StatusNotFound, "HTTP 404: Not Found\n"},
    }

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tc.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tc.expectedStatus)
			}

			if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(tc.expectedBody) {
				t.Errorf("handler returned unexpected body: got %v want %v",
					rr.Body.String(), tc.expectedBody)
			}
		})
	}
}

func TestLoggingMiddleware(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(log.Writer())

	router := setupRouter()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	logOutput := buf.String()
	expectedEntries := []string{
		"GET",
		"/",
		"200",
	}

	for _, entry := range expectedEntries {
		if !strings.Contains(logOutput, entry) {
			t.Errorf("Log output missing expected entry: %s", entry)
		}
	}
}

func TestPocketBaseIntegration(t *testing.T) {
	// Create a temporary directory for PocketBase data
	tempDir, err := os.MkdirTemp("", "pocketbase-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set the data directory for PocketBase
	os.Setenv("POCKETBASE_DATA_DIR", tempDir)

	// Initialize PocketBase configuration
	pbConfig := config.NewPocketBaseConfig()

	// Test PocketBase connection
	if pbConfig.App == nil {
		t.Error("PocketBase app is nil")
	}

	// Start and immediately stop PocketBase
	go func() {
		if err := pbConfig.Start(); err != nil {
			t.Errorf("Failed to start PocketBase: %v", err)
		}
	}()

	// Wait a short time for PocketBase to start
	time.Sleep(100 * time.Millisecond)

	// Reset PocketBase state
	pbConfig.App.ResetBootstrapState()
}

func TestMain(m *testing.M) {
	// Setup code here (if needed)

	// Run the tests
	exitCode := m.Run()

	// Teardown code here (if needed)

	os.Exit(exitCode)
}
