// internal/api/v1/company/handler_test.go
package company

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/api-moose/company-earnings/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCompanyService struct {
	mock.Mock
}

func (m *MockCompanyService) Search(ctx context.Context, query string, limit int) ([]models.Company, error) {
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
		expectedBody   string
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
			expectedBody:   `{"count":1,"results":[{"symbol":"AAPL","cik":"0000320193","securityName":"Apple Inc.","securityType":"Common Stock","region":"US","exchange":"NASDAQ","sector":"Technology"}],"next_url":null}`,
		},
		{
			name:           "Empty query",
			query:          "",
			mockResult:     nil,
			mockError:      errors.New("query cannot be empty"),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"query cannot be empty"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockCompanyService)
			mockService.On("Search", mock.Anything, tt.query, 10).Return(tt.mockResult, tt.mockError)

			handler := NewHandler(mockService)

			req, err := http.NewRequest("GET", "/companies?query="+tt.query, nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.SearchHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.JSONEq(t, tt.expectedBody, rr.Body.String())

			mockService.AssertExpectations(t)
		})
	}
}
