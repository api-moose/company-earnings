// internal/models/company.go
package models

type Company struct {
	Symbol       string `json:"symbol"`
	CIK          string `json:"cik"`
	SecurityName string `json:"securityName"`
	SecurityType string `json:"securityType"`
	Region       string `json:"region"`
	Exchange     string `json:"exchange"`
	Sector       string `json:"sector"`
}
