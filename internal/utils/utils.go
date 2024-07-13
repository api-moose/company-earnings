package main

import "fmt"

// AppError represents an application-level error.
type AppError struct {
	StatusCode int
	Message    string
}

func (e *AppError) Error() string {
	return fmt.Sprintf("Error %d: %s", e.StatusCode, e.Message)
}

// NewAppError creates a new AppError instance.
func NewAppError(statusCode int, message string) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Message:    message,
	}
}