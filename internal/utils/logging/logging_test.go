package logging

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggingMiddleware(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap the test handler with the logging middleware
	handler := LoggingMiddleware(testHandler)

	// Create a test request
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request using the wrapped handler
	handler.ServeHTTP(rr, req)

	// Check if the status code is still correct
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check if the log output contains expected information
	logOutput := buf.String()
	expectedEntries := []string{
		"127.0.0.1:12345",
		"GET",
		"/test",
		"200",
	}

	for _, entry := range expectedEntries {
		if !strings.Contains(logOutput, entry) {
			t.Errorf("Log output missing expected entry: %s", entry)
		}
	}
}
