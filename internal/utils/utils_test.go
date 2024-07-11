package main

import (
	"net/http"
	"testing"
)

func TestAppError(t *testing.T) {
	err := NewAppError(http.StatusNotFound, "Resource not found")
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err.Error() != "Error 404: Resource not found" {
		t.Errorf("Unexpected error message: got %q", err.Error())
	}
}
