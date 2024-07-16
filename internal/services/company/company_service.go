// internal/services/company/service.go
package company

import (
	"context"
	"errors"

	"github.com/api-moose/company-earnings/internal/models"
)

type Service interface {
	Search(ctx context.Context, query string, limit int) ([]models.Company, error)
}

type companyService struct{}

func NewService() Service {
	return &companyService{}
}

func (s *companyService) Search(ctx context.Context, query string, limit int) ([]models.Company, error) {
	if query == "" {
		return nil, errors.New("query cannot be empty")
	}

	// TODO: Implement actual search logic
	// For now, we'll return a mock result
	if query == "Apple" {
		return []models.Company{
			{
				Symbol:       "AAPL",
				CIK:          "0000320193",
				SecurityName: "Apple Inc.",
				SecurityType: "Common Stock",
				Region:       "US",
				Exchange:     "NASDAQ",
				Sector:       "Technology",
			},
		}, nil
	}

	return []models.Company{}, nil
}
