package deal

import (
	"context"
	"fmt"

	"github.com/industrix/backend/pkg/errors"
	"github.com/industrix/backend/pkg/postgres"
)

// Repository handles all deal-related database operations
type Repository struct {
	pg *postgres.Client
}

// NewRepository creates a new deal repository
func NewRepository(pg *postgres.Client) *Repository {
	return &Repository{pg: pg}
}

func (r *Repository) CreateDeal(ctx context.Context, d *Deal) error {
	err := r.pg.QueryRow(ctx,
		`INSERT INTO deals (listing_id, buyer_id, seller_id, message)
		 VALUES ($1, $2, $3, $4) RETURNING id, status, created_at, updated_at`,
		d.ListingID, d.BuyerID, d.SellerID, d.Message,
	).Scan(&d.ID, &d.Status, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create deal: %w", err)
	}
	return nil
}

func (r *Repository) GetDealByID(ctx context.Context, id string) (*Deal, error) {
	var d Deal
	err := r.pg.QueryRow(ctx,
		`SELECT id, listing_id, buyer_id, seller_id, COALESCE(message, ''), status, created_at, updated_at
		 FROM deals WHERE id = $1`, id,
	).Scan(&d.ID, &d.ListingID, &d.BuyerID, &d.SellerID, &d.Message, &d.Status, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "Deal not found")
	}
	return &d, nil
}

// ListForUser returns every deal where the user is either the buyer or the
// seller, most recent first.
func (r *Repository) ListForUser(ctx context.Context, userID string) ([]*Deal, error) {
	rows, err := r.pg.Query(ctx,
		`SELECT id, listing_id, buyer_id, seller_id, COALESCE(message, ''), status, created_at, updated_at
		 FROM deals WHERE buyer_id = $1 OR seller_id = $1 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*Deal
	for rows.Next() {
		var d Deal
		if err := rows.Scan(&d.ID, &d.ListingID, &d.BuyerID, &d.SellerID, &d.Message, &d.Status, &d.CreatedAt, &d.UpdatedAt); err != nil {
			continue
		}
		items = append(items, &d)
	}
	return items, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.pg.Exec(ctx, "UPDATE deals SET status = $1, updated_at = NOW() WHERE id = $2", status, id)
	return err
}
