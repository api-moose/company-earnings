package dberrors

import (
	"errors"
	"testing"
)

func TestNewDBError(t *testing.T) {
	err := NewDBError("connection failed")
	if err.Error() != "DB Error: connection failed" {
		t.Errorf("Expected 'DB Error: connection failed', got '%s'", err.Error())
	}
}

func TestIsDBError(t *testing.T) {
	err := NewDBError("query failed")
	if !IsDBError(err) {
		t.Error("Expected IsDBError to return true")
	}

	regularErr := errors.New("regular error")
	if IsDBError(regularErr) {
		t.Error("Expected IsDBError to return false for regular error")
	}
}
