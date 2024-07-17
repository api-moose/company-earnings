package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONResponse(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"message": "success"}

	JSONResponse(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}

	if response["message"] != "success" {
		t.Errorf("Expected message 'success', got '%s'", response["message"])
	}
}

func TestErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()
	err := &ErrorMessage{Status: http.StatusBadRequest, Message: "invalid input"}

	ErrorResponse(w, err)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response ErrorMessage
	jsonErr := json.Unmarshal(w.Body.Bytes(), &response)
	if jsonErr != nil {
		t.Fatalf("Error unmarshaling response: %v", jsonErr)
	}

	if response.Message != "invalid input" {
		t.Errorf("Expected error 'invalid input', got '%s'", response.Message)
	}

	if response.Status != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, response.Status)
	}
}

func TestErrorResponseWithRegularError(t *testing.T) {
	w := httptest.NewRecorder()
	err := errors.New("regular error")

	ErrorResponse(w, err)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var response ErrorMessage
	jsonErr := json.Unmarshal(w.Body.Bytes(), &response)
	if jsonErr != nil {
		t.Fatalf("Error unmarshaling response: %v", jsonErr)
	}

	if response.Message != "regular error" {
		t.Errorf("Expected error 'regular error', got '%s'", response.Message)
	}

	if response.Status != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, response.Status)
	}
}
