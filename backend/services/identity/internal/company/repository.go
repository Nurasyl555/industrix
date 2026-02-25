package company

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/industrix/pkg/postgres"
)

// Repository handles database operations for company
type Repository struct {
	pg *postgres.Client
}

// NewRepository creates a new company repository
func NewRepository(pg *postgres.Client) *Repository {
	return &Repository{pg: pg}
}

// CreateCompany creates a new company
func (r *Repository) CreateCompany(ctx context.Context, userID string, req CreateCompanyRequest) (*Company, error) {
	companyID := uuid.New().String()

	query := `
		INSERT INTO companies (id, user_id, name, bin, address, phone, email, website, status, verified, reputation_score, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		RETURNING id, user_id, name, bin, address, phone, email, website, status, verified, reviewer_note, reputation_score, created_at, updated_at
	`

	var company Company
	err := r.pg.QueryRow(ctx, query,
		companyID, userID, req.Name, req.BIN, req.Address, req.Phone, req.Email, req.Website,
		StatusPending, false, 0.0,
	).Scan(
		&company.ID, &company.UserID, &company.Name, &company.BIN, &company.Address, &company.Phone,
		&company.Email, &company.Website, &company.Status, &company.Verified, &company.ReviewerNote,
		&company.ReputationScore, &company.CreatedAt, &company.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create company: %w", err)
	}

	// Update user's company_id
	_, err = r.pg.Exec(ctx, "UPDATE users SET company_id = $1 WHERE id = $2", companyID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to update user company_id: %w", err)
	}

	return &company, nil
}

// GetCompanyByUserID retrieves company by user ID
func (r *Repository) GetCompanyByUserID(ctx context.Context, userID string) (*Company, error) {
	query := `
		SELECT id, user_id, name, bin, address, phone, email, website, status, verified, reviewer_note, reputation_score, created_at, updated_at
		FROM companies
		WHERE user_id = $1
	`

	var company Company
	err := r.pg.QueryRow(ctx, query, userID).Scan(
		&company.ID, &company.UserID, &company.Name, &company.BIN, &company.Address, &company.Phone,
		&company.Email, &company.Website, &company.Status, &company.Verified, &company.ReviewerNote,
		&company.ReputationScore, &company.CreatedAt, &company.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	return &company, nil
}

// UpdateCompany updates company info
func (r *Repository) UpdateCompany(ctx context.Context, userID string, req UpdateCompanyRequest) (*Company, error) {
	query := `
		UPDATE companies
		SET name = COALESCE(NULLIF($2, ''), name),
			bin = COALESCE(NULLIF($3, ''), bin),
			address = COALESCE(NULLIF($4, ''), address),
			phone = COALESCE(NULLIF($5, ''), phone),
			email = COALESCE(NULLIF($6, ''), email),
			website = COALESCE(NULLIF($7, ''), website),
			updated_at = NOW()
		WHERE user_id = $1
		RETURNING id, user_id, name, bin, address, phone, email, website, status, verified, reviewer_note, reputation_score, created_at, updated_at
	`

	var company Company
	err := r.pg.QueryRow(ctx, query, userID, req.Name, req.BIN, req.Address, req.Phone, req.Email, req.Website).Scan(
		&company.ID, &company.UserID, &company.Name, &company.BIN, &company.Address, &company.Phone,
		&company.Email, &company.Website, &company.Status, &company.Verified, &company.ReviewerNote,
		&company.ReputationScore, &company.CreatedAt, &company.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update company: %w", err)
	}

	return &company, nil
}

// UpdateStatus updates company status
func (r *Repository) UpdateStatus(ctx context.Context, companyID string, status CompanyStatus) error {
	query := `UPDATE companies SET status = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pg.Exec(ctx, query, companyID, status)
	if err != nil {
		return fmt.Errorf("failed to update company status: %w", err)
	}
	return nil
}

// UpdateVerificationStatus updates verification status and reviewer note
func (r *Repository) UpdateVerificationStatus(ctx context.Context, companyID string, status CompanyStatus, reviewerNote string) error {
	query := `UPDATE companies SET status = $2, reviewer_note = $3, verified = $4, updated_at = NOW() WHERE id = $1`
	verified := status == StatusVerified
	_, err := r.pg.Exec(ctx, query, companyID, status, reviewerNote, verified)
	if err != nil {
		return fmt.Errorf("failed to update verification status: %w", err)
	}
	return nil
}

// CreateDocument creates a verification document
func (r *Repository) CreateDocument(ctx context.Context, companyID string, req UploadDocumentRequest) (*VerificationDocument, error) {
	docID := uuid.New().String()

	query := `
		INSERT INTO company_documents (id, company_id, document_type, file_name, file_url, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, company_id, document_type, file_name, file_url, uploaded_at, created_at
	`

	// File URL will be set after upload via MinIO
	fileURL := fmt.Sprintf("companies/%s/verification/%s", companyID, req.FileName)

	var doc VerificationDocument
	err := r.pg.QueryRow(ctx, query, docID, companyID, req.DocumentType, req.FileName, fileURL).Scan(
		&doc.ID, &doc.CompanyID, &doc.DocumentType, &doc.FileName, &doc.FileURL, &doc.UploadedAt, &doc.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	return &doc, nil
}

// GetDocuments retrieves verification documents for a company
func (r *Repository) GetDocuments(ctx context.Context, companyID string) ([]*VerificationDocument, error) {
	query := `
		SELECT id, company_id, document_type, file_name, file_url, uploaded_at, created_at
		FROM company_documents
		WHERE company_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pg.Query(ctx, query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get documents: %w", err)
	}
	defer rows.Close()

	var docs []*VerificationDocument
	for rows.Next() {
		var doc VerificationDocument
		err := rows.Scan(&doc.ID, &doc.CompanyID, &doc.DocumentType, &doc.FileName, &doc.FileURL, &doc.UploadedAt, &doc.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		docs = append(docs, &doc)
	}

	return docs, nil
}

// UpdateCompanyReputation updates company reputation score based on reviews
func (r *Repository) UpdateCompanyReputation(ctx context.Context, userID string) error {
	// Get user's company
	company, err := r.GetCompanyByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get company: %w", err)
	}

	// Calculate average rating from reviews
	query := `
		UPDATE companies
		SET reputation_score = (
			SELECT COALESCE(AVG(rating), 0)
			FROM reviews r
			JOIN users u ON r.target_user_id = u.id
			WHERE u.company_id = $1
		),
		updated_at = NOW()
		WHERE id = $1
	`

	_, err = r.pg.Exec(ctx, query, company.ID)
	if err != nil {
		return fmt.Errorf("failed to update reputation: %w", err)
	}

	return nil
}

// GetCompanyByID retrieves company by ID
func (r *Repository) GetCompanyByID(ctx context.Context, companyID string) (*Company, error) {
	query := `
		SELECT id, user_id, name, bin, address, phone, email, website, status, verified, reviewer_note, reputation_score, created_at, updated_at
		FROM companies
		WHERE id = $1
	`

	var company Company
	err := r.pg.QueryRow(ctx, query, companyID).Scan(
		&company.ID, &company.UserID, &company.Name, &company.BIN, &company.Address, &company.Phone,
		&company.Email, &company.Website, &company.Status, &company.Verified, &company.ReviewerNote,
		&company.ReputationScore, &company.CreatedAt, &company.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	return &company, nil
}
