package company

import (
	"net/http"
	"strconv"

	"github.com/api-moose/company-earnings/internal/db/mongo"
	"github.com/api-moose/company-earnings/internal/models"
	"github.com/api-moose/company-earnings/internal/utils/response"
)

type Handler struct {
	repo mongo.Repository
}

func NewHandler(repo mongo.Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		response.ErrorResponse(w, &response.ErrorMessage{
			Status:  http.StatusBadRequest,
			Message: "query cannot be empty",
		})
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 10 // Default limit
	}

	companies, err := h.repo.Search(r.Context(), query, limit)
	if err != nil {
		response.ErrorResponse(w, err)
		return
	}

	resp := struct {
		Count   int              `json:"count"`
		Results []models.Company `json:"results"`
		NextURL *string          `json:"next_url"`
	}{
		Count:   len(companies),
		Results: companies,
		NextURL: nil, // Implement pagination logic if needed
	}

	response.JSONResponse(w, http.StatusOK, resp)
}
