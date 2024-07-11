package httperrors

import (
	"errors"
	"net/http"
	"testing"
)

func TestNewHTTPError(t *testing.T) {
	err := NewHTTPError(http.StatusNotFound, "resource not found")
	if err.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, err.StatusCode)
	}
	if err.Message != "resource not found" {
		t.Errorf("Expected message 'resource not found', got '%s'", err.Message)
	}
}

func TestHTTPError_Error(t *testing.T) {
	err := NewHTTPError(http.StatusBadRequest, "invalid input")
	if err.Error() != "HTTP 400: invalid input" {
		t.Errorf("Expected 'HTTP 400: invalid input', got '%s'", err.Error())
	}
}

func TestIsHTTPError(t *testing.T) {
	err := NewHTTPError(http.StatusUnauthorized, "unauthorized access")
	if !IsHTTPError(err) {
		t.Error("Expected IsHTTPError to return true")
	}

	regularErr := errors.New("regular error")
	if IsHTTPError(regularErr) {
		t.Error("Expected IsHTTPError to return false for regular error")
	}
}
