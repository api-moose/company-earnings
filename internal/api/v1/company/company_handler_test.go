package company

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/api-moose/company-earnings/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Search(ctx context.Context, query string, limit int) ([]models.Company, error) {
	args := m.Called(ctx, query, limit)
	return args.Get(0).([]models.Company), args.Error(1)
}

func TestSearchHandler(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		mockResult     []models.Company
		mockError      error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:  "Valid search",
			query: "Apple",
			mockResult: []models.Company{
				{
					Symbol:       "AAPL",
					CIK:          "0000320193",
					SecurityName: "Apple Inc.",
					SecurityType: "Common Stock",
					Region:       "US",
					Exchange:     "NASDAQ",
					Sector:       "Technology",
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"count": float64(1),
				"results": []interface{}{
					map[string]interface{}{
						"symbol":       "AAPL",
						"cik":          "0000320193",
						"securityName": "Apple Inc.",
						"securityType": "Common Stock",
						"region":       "US",
						"exchange":     "NASDAQ",
						"sector":       "Technology",
					},
				},
				"next_url": nil,
			},
		},
		{
			name:           "Empty query",
			query:          "",
			mockResult:     nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"status": float64(http.StatusBadRequest),
				"error":  "query cannot be empty",
			},
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			if tt.query != "" {
				mockRepo.On("Search", mock.Anything, tt.query, 10).Return(tt.mockResult, tt.mockError)
			}

			handler := NewHandler(mockRepo)

			req, err := http.NewRequest("GET", "/companies?query="+tt.query, nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.SearchHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var response map[string]interface{}
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedBody, response)

			mockRepo.AssertExpectations(t)
		})
	}
}
