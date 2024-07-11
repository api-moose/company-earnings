package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/api-moose/company-earnings/internal/errors/httperrors"
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
	err := httperrors.NewHTTPError(http.StatusBadRequest, "invalid input")

	ErrorResponse(w, err)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]string
	jsonErr := json.Unmarshal(w.Body.Bytes(), &response)
	if jsonErr != nil {
		t.Fatalf("Error unmarshaling response: %v", jsonErr)
	}

	if response["error"] != "invalid input" {
		t.Errorf("Expected error 'invalid input', got '%s'", response["error"])
	}
}
