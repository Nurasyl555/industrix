package integrity

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/industrix/backend/pkg/errors"
	"github.com/industrix/backend/pkg/postgres"
)

// Repository handles all company database operations
type Repository struct {
	pg *postgres.Client
}

// NewRepository creates a new integrity repository
func NewRepository(pg *postgres.Client) *Repository {
	return &Repository{pg: pg}
}

func (r *Repository) CreateCompany(ctx context.Context, company *Company) error {
	if company.ID == "" {
		company.ID = uuid.New().String()
	}
	err := r.pg.QueryRow(ctx,
		`INSERT INTO companies (id, user_id, name, bin, address, phone, email, website, status, verified, reputation_score, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		RETURNING created_at, updated_at`,
		company.ID, company.OwnerID, company.Name, company.BIN, company.Address, company.Phone,
		company.Email, company.Website, StatusPending, false, 0.0,
	).Scan(&company.CreatedAt, &company.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create company: %w", err)
	}

	// Update user's company_id
	_, err = r.pg.Exec(ctx, "UPDATE users SET company_id = $1 WHERE id = $2", company.ID, company.OwnerID)
	if err != nil {
		return fmt.Errorf("failed to update user company_id: %w", err)
	}

	return nil
}

func (r *Repository) GetCompanyByID(ctx context.Context, id string) (*Company, error) {
	var c Company
	err := r.pg.QueryRow(ctx,
		`SELECT id, COALESCE(user_id::text, ''), name, bin, address, phone, email, website,
		 status, verified, COALESCE(reviewer_note, ''), reputation_score, created_at, updated_at
		 FROM companies WHERE id = $1`, id).Scan(
		&c.ID, &c.OwnerID, &c.Name, &c.BIN, &c.Address, &c.Phone, &c.Email, &c.Website,
		&c.Status, &c.Verified, &c.ReviewerNote, &c.ReputationScore, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "Company not found")
	}
	return &c, nil
}

func (r *Repository) GetCompanyByBIN(ctx context.Context, bin string) (*Company, error) {
	var c Company
	err := r.pg.QueryRow(ctx,
		`SELECT id, COALESCE(user_id::text, ''), name, bin, address, phone, email, website,
		 status, verified, COALESCE(reviewer_note, ''), reputation_score, created_at, updated_at
		 FROM companies WHERE bin = $1`, bin).Scan(
		&c.ID, &c.OwnerID, &c.Name, &c.BIN, &c.Address, &c.Phone, &c.Email, &c.Website,
		&c.Status, &c.Verified, &c.ReviewerNote, &c.ReputationScore, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, nil // Not found is not an error for duplicate checking
	}
	return &c, nil
}

func (r *Repository) UpdateCompany(ctx context.Context, company *Company) error {
	_, err := r.pg.Exec(ctx,
		`UPDATE companies SET name = COALESCE(NULLIF($2, ''), name), address = COALESCE(NULLIF($3, ''), address),
		 phone = COALESCE(NULLIF($4, ''), phone), email = COALESCE(NULLIF($5, ''), email),
		 website = COALESCE(NULLIF($6, ''), website), updated_at = NOW() WHERE id = $1`,
		company.ID, company.Name, company.Address, company.Phone, company.Email, company.Website,
	)
	if err != nil {
		return fmt.Errorf("failed to update company: %w", err)
	}
	return nil
}
