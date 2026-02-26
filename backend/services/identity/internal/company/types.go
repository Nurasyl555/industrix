package company

import "time"

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
	BIN     string `json:"bin"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Website string `json:"website"`
}

// UploadDocumentRequest represents upload document request
type UploadDocumentRequest struct {
	DocumentType string `json:"document_type"` // bin_cert, charter, director_id
	FileName     string `json:"file_name"`
	ContentType  string `json:"content_type"`
}

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
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

// VerificationDocument represents a verification document
type VerificationDocument struct {
	ID           string    `json:"id"`
	CompanyID    string    `json:"company_id"`
	DocumentType string    `json:"document_type"`
	FileName     string    `json:"file_name"`
	FileURL      string    `json:"file_url"`
	UploadURL    string    `json:"upload_url,omitempty"`
	UploadedAt   time.Time `json:"uploaded_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// VerificationStatus represents verification status response
type VerificationStatus struct {
	Status       string                  `json:"status"`
	ReviewerNote string                  `json:"reviewer_notes"`
	Documents    []*VerificationDocument `json:"documents"`
	UpdatedAt    time.Time               `json:"updated_at"`
}
