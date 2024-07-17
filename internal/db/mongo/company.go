// internal/db/mongo/company.go
package mongo

import (
	"context"
	"errors"

	"github.com/api-moose/company-earnings/internal/models"
)

type Repository interface {
	Search(ctx context.Context, query string, limit int) ([]models.Company, error)
}

type companyRepository struct{}

func NewRepository() Repository {
	return &companyRepository{}
}

func (r *companyRepository) Search(ctx context.Context, query string, limit int) ([]models.Company, error) {
	if query == "" {
		return nil, errors.New("query cannot be empty")
	}

	// TODO: Implement actual database search logic
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
