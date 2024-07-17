package response

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ErrorMessage struct {
	Status  int    `json:"status"`
	Message string `json:"error"`
}

// Implement the error interface for ErrorMessage
func (e *ErrorMessage) Error() string {
	return fmt.Sprintf("Status %d: %s", e.Status, e.Message)
}

func JSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func ErrorResponse(w http.ResponseWriter, err error) {
	var errMsg *ErrorMessage
	if e, ok := err.(*ErrorMessage); ok {
		errMsg = e
	} else {
		errMsg = &ErrorMessage{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	JSONResponse(w, errMsg.Status, errMsg)
}
