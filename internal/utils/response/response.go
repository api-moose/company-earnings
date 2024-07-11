package response

import (
	"encoding/json"
	"net/http"

	"github.com/api-moose/company-earnings/internal/errors/httperrors"
)

func JSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func ErrorResponse(w http.ResponseWriter, err error) {
	httpErr, ok := err.(*httperrors.HTTPError)
	if !ok {
		httpErr = httperrors.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	JSONResponse(w, httpErr.StatusCode, map[string]string{"error": httpErr.Message})
}
