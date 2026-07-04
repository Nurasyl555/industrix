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

// companySelect is the shared column list. Note the real schema uses
// owner_id and verification_status (see 003_companies.up.sql), not user_id/
// status — the Go struct fields OwnerID/Status map onto those.
const companySelect = `SELECT id, COALESCE(owner_id::text, ''), name, bin, COALESCE(address, ''),
	COALESCE(phone, ''), COALESCE(email, ''), COALESCE(website, ''),
	verification_status, verified, COALESCE(reviewer_note, ''), COALESCE(reputation_score, 0),
	created_at, updated_at FROM companies`

func scanCompany(row interface {
	Scan(dest ...interface{}) error
}) (*Company, error) {
	var c Company
	err := row.Scan(
		&c.ID, &c.OwnerID, &c.Name, &c.BIN, &c.Address, &c.Phone, &c.Email, &c.Website,
		&c.Status, &c.Verified, &c.ReviewerNote, &c.ReputationScore, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repository) CreateCompany(ctx context.Context, company *Company) error {
	if company.ID == "" {
		company.ID = uuid.New().String()
	}
	err := r.pg.QueryRow(ctx,
		`INSERT INTO companies (id, owner_id, name, bin, address, phone, email, website, verification_status, verified, reputation_score, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		RETURNING created_at, updated_at`,
		company.ID, company.OwnerID, company.Name, company.BIN, company.Address, company.Phone,
		company.Email, company.Website, StatusPending, false, 0.0,
	).Scan(&company.CreatedAt, &company.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create company: %w", err)
	}
	company.Status = StatusPending

	// Link the company to its owner's user record.
	if _, err = r.pg.Exec(ctx, "UPDATE users SET company_id = $1 WHERE id = $2", company.ID, company.OwnerID); err != nil {
		return fmt.Errorf("failed to update user company_id: %w", err)
	}
	return nil
}

func (r *Repository) GetCompanyByID(ctx context.Context, id string) (*Company, error) {
	c, err := scanCompany(r.pg.QueryRow(ctx, companySelect+" WHERE id = $1", id))
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "Company not found")
	}
	return c, nil
}

// GetCompanyByOwner returns the company owned by a user, or a NotFound error.
func (r *Repository) GetCompanyByOwner(ctx context.Context, ownerID string) (*Company, error) {
	c, err := scanCompany(r.pg.QueryRow(ctx, companySelect+" WHERE owner_id = $1", ownerID))
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "Company not found")
	}
	return c, nil
}

func (r *Repository) GetCompanyByBIN(ctx context.Context, bin string) (*Company, error) {
	c, err := scanCompany(r.pg.QueryRow(ctx, companySelect+" WHERE bin = $1", bin))
	if err != nil {
		return nil, nil // Not found is not an error for duplicate checking
	}
	return c, nil
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

// === Admin operations ===

// ListByStatus returns companies filtered by verification status (for the
// admin moderation queue). Empty status returns all.
func (r *Repository) ListByStatus(ctx context.Context, status string) ([]*Company, error) {
	q := companySelect + " ORDER BY created_at DESC"
	args := []interface{}{}
	if status != "" {
		q = companySelect + " WHERE verification_status = $1 ORDER BY created_at DESC"
		args = append(args, status)
	}
	rows, err := r.pg.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*Company
	for rows.Next() {
		c, err := scanCompany(rows)
		if err != nil {
			continue
		}
		items = append(items, c)
	}
	return items, nil
}

// SetStatus updates a company's verification status + verified flag and
// records the reviewer's note.
func (r *Repository) SetStatus(ctx context.Context, id string, status CompanyStatus, note string) error {
	_, err := r.pg.Exec(ctx,
		`UPDATE companies SET verification_status = $2, verified = $3, reviewer_note = $4, updated_at = NOW() WHERE id = $1`,
		id, status, status == StatusVerified, note,
	)
	if err != nil {
		return fmt.Errorf("failed to set company status: %w", err)
	}
	return nil
}
