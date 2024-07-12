package tenancy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTenantMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		expectedStatus int
	}{
		{"Valid tenant", "tenant1", http.StatusOK},
		{"Missing tenant", "", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			if tt.tenantID != "" {
				req.Header.Set("X-Tenant-ID", tt.tenantID)
			}

			rr := httptest.NewRecorder()
			handler := TenantMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tenantID, ok := r.Context().Value(TenantContextKey).(string)
				if !ok {
					http.Error(w, "Tenant context not found", http.StatusInternalServerError)
					return
				}
				if tenantID == "" {
					http.Error(w, "Tenant ID is required", http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}
