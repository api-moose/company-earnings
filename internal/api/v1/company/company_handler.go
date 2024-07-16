// internal/api/v1/company/handler.go
package company

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/api-moose/company-earnings/internal/services/company"
)

type Handler struct {
	service company.Service
}

func NewHandler(service company.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 10 // Default limit
	}

	companies, err := h.service.Search(r.Context(), query, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := struct {
		Count   int              `json:"count"`
		Results []models.Company `json:"results"`
		NextURL *string          `json:"next_url"`
	}{
		Count:   len(companies),
		Results: companies,
		NextURL: nil, // Implement pagination logic if needed
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
