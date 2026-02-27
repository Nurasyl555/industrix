package integrity

import "time"

// CompanyStatus represents the verification status of a company
type CompanyStatus string

const (
	StatusPending  CompanyStatus = "pending"
	StatusVerified CompanyStatus = "verified"
	StatusRejected CompanyStatus = "rejected"
)

// Company represents a company
type Company struct {
	ID              string        `json:"id"`
	UserID          string        `json:"user_id"`
	Name            string        `json:"name"`
	BIN             string        `json:"bin"`
	Address         string        `json:"address"`
	Phone           string        `json:"phone"`
	Email           string        `json:"email"`
	Website         string        `json:"website"`
	Status          CompanyStatus `json:"status"`
	Verified        bool          `json:"verified"`
	ReviewerNote    string        `json:"reviewer_note"`
	ReputationScore float64       `json:"reputation_score"`
	OwnerID         string        `json:"owner_id"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

// CreateCompanyRequest represents create company request
type CreateCompanyRequest struct {
	Name    string `json:"name"`
	BIN     string `json:"bin"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Website string `json:"website"`
}

// UpdateCompanyRequest represents update company request
type UpdateCompanyRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Website string `json:"website"`
}
