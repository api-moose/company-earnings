// internal/services/company/service_test.go
package company

import (
	"context"
	"testing"

	"github.com/api-moose/company-earnings/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestCompanyService_Search(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		limit   int
		want    []models.Company
		wantErr bool
	}{
		{
			name:  "Valid search",
			query: "Apple",
			limit: 1,
			want: []models.Company{
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
			wantErr: false,
		},
		{
			name:    "Empty query",
			query:   "",
			limit:   10,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService()
			got, err := s.Search(context.Background(), tt.query, tt.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompanyService.Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
