package listing

import (
	"context"
	"fmt"
	"strings"

	"github.com/industrix/backend/pkg/errors"
	"github.com/industrix/backend/pkg/postgres"
)

// Repository handles all listing-related database operations
type Repository struct {
	pg *postgres.Client
}

// NewRepository creates a new listing repository
func NewRepository(pg *postgres.Client) *Repository {
	return &Repository{pg: pg}
}

func (r *Repository) CreateListing(ctx context.Context, l *Listing) error {
	var pricePeriod interface{}
	if l.PricePeriod != "" {
		pricePeriod = l.PricePeriod
	}
	err := r.pg.QueryRow(ctx,
		`INSERT INTO listings (equipment_id, seller_id, listing_type, price, price_period, status)
		 VALUES ($1, $2, $3, $4, $5, 'draft') RETURNING id, status, created_at, updated_at`,
		l.EquipmentID, l.SellerID, l.ListingType, l.Price, pricePeriod,
	).Scan(&l.ID, &l.Status, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create listing: %w", err)
	}
	return nil
}

func (r *Repository) GetListingByID(ctx context.Context, id string) (*Listing, error) {
	var l Listing
	err := r.pg.QueryRow(ctx,
		`SELECT id, equipment_id, seller_id, listing_type, price, COALESCE(price_period, ''), status, created_at, updated_at
		 FROM listings WHERE id = $1`, id,
	).Scan(&l.ID, &l.EquipmentID, &l.SellerID, &l.ListingType, &l.Price, &l.PricePeriod, &l.Status, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "Listing not found")
	}
	return &l, nil
}

func (r *Repository) GetListingViewByID(ctx context.Context, id string) (*ListingView, error) {
	var v ListingView
	err := r.pg.QueryRow(ctx,
		`SELECT l.id, l.equipment_id, e.title, COALESCE(e.description, ''), e.category_id, COALESCE(e.region, ''), e.condition, COALESCE(e.image_url, ''),
		        l.seller_id, l.listing_type, l.price, COALESCE(l.price_period, ''), l.status, l.created_at
		 FROM listings l JOIN equipment e ON e.id = l.equipment_id
		 WHERE l.id = $1`, id,
	).Scan(&v.ID, &v.EquipmentID, &v.Title, &v.Description, &v.CategoryID, &v.Region, &v.Condition, &v.ImageURL,
		&v.SellerID, &v.ListingType, &v.Price, &v.PricePeriod, &v.Status, &v.CreatedAt)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "Listing not found")
	}
	return &v, nil
}

// ListByStatusView returns listings in a given status joined with equipment —
// used by the admin moderation queue.
func (r *Repository) ListByStatusView(ctx context.Context, status string) ([]*ListingView, error) {
	rows, err := r.pg.Query(ctx,
		`SELECT l.id, l.equipment_id, e.title, COALESCE(e.description, ''), e.category_id, COALESCE(e.region, ''), e.condition, COALESCE(e.image_url, ''),
		        l.seller_id, l.listing_type, l.price, COALESCE(l.price_period, ''), l.status, l.created_at
		 FROM listings l JOIN equipment e ON e.id = l.equipment_id
		 WHERE l.status = $1 ORDER BY l.created_at ASC`, status,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*ListingView
	for rows.Next() {
		var v ListingView
		if err := rows.Scan(&v.ID, &v.EquipmentID, &v.Title, &v.Description, &v.CategoryID, &v.Region, &v.Condition, &v.ImageURL,
			&v.SellerID, &v.ListingType, &v.Price, &v.PricePeriod, &v.Status, &v.CreatedAt); err != nil {
			continue
		}
		items = append(items, &v)
	}
	return items, nil
}

// ListActive returns active listings joined with equipment, for public browsing.
func (r *Repository) ListActive(ctx context.Context, f ListListingsFilter) ([]*ListingView, int64, error) {
	where := []string{"l.status = 'active'"}
	args := []interface{}{}
	argN := 1

	if f.CategoryID != "" {
		where = append(where, fmt.Sprintf("e.category_id = $%d", argN))
		args = append(args, f.CategoryID)
		argN++
	}
	if f.Region != "" {
		where = append(where, fmt.Sprintf("e.region = $%d", argN))
		args = append(args, f.Region)
		argN++
	}
	if f.ListingType != "" {
		where = append(where, fmt.Sprintf("l.listing_type = $%d", argN))
		args = append(args, f.ListingType)
		argN++
	}
	if f.Condition != "" {
		where = append(where, fmt.Sprintf("e.condition = $%d", argN))
		args = append(args, f.Condition)
		argN++
	}
	if f.Search != "" {
		where = append(where, fmt.Sprintf("e.title ILIKE $%d", argN))
		args = append(args, "%"+f.Search+"%")
		argN++
	}
	if f.PriceMin > 0 {
		where = append(where, fmt.Sprintf("l.price >= $%d", argN))
		args = append(args, f.PriceMin)
		argN++
	}
	if f.PriceMax > 0 {
		where = append(where, fmt.Sprintf("l.price <= $%d", argN))
		args = append(args, f.PriceMax)
		argN++
	}
	whereClause := strings.Join(where, " AND ")

	orderBy := "l.created_at DESC"
	switch f.Sort {
	case "price_asc":
		orderBy = "l.price ASC"
	case "price_desc":
		orderBy = "l.price DESC"
	}

	offset := (f.Page - 1) * f.Limit
	query := fmt.Sprintf(
		`SELECT l.id, l.equipment_id, e.title, e.category_id, COALESCE(e.region, ''), e.condition, COALESCE(e.image_url, ''),
		        l.seller_id, l.listing_type, l.price, COALESCE(l.price_period, ''), l.status, l.created_at
		 FROM listings l JOIN equipment e ON e.id = l.equipment_id
		 WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, argN, argN+1,
	)
	args = append(args, f.Limit, offset)

	rows, err := r.pg.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*ListingView
	for rows.Next() {
		var v ListingView
		if err := rows.Scan(&v.ID, &v.EquipmentID, &v.Title, &v.CategoryID, &v.Region, &v.Condition, &v.ImageURL,
			&v.SellerID, &v.ListingType, &v.Price, &v.PricePeriod, &v.Status, &v.CreatedAt); err != nil {
			continue
		}
		items = append(items, &v)
	}

	countQuery := fmt.Sprintf(
		"SELECT COUNT(*) FROM listings l JOIN equipment e ON e.id = l.equipment_id WHERE %s", whereClause)
	var total int64
	_ = r.pg.QueryRow(ctx, countQuery, args[:argN-1]...).Scan(&total)

	return items, total, nil
}

func (r *Repository) ListBySeller(ctx context.Context, sellerID string) ([]*Listing, error) {
	rows, err := r.pg.Query(ctx,
		`SELECT id, equipment_id, seller_id, listing_type, price, COALESCE(price_period, ''), status, created_at, updated_at
		 FROM listings WHERE seller_id = $1 ORDER BY created_at DESC`, sellerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*Listing
	for rows.Next() {
		var l Listing
		if err := rows.Scan(&l.ID, &l.EquipmentID, &l.SellerID, &l.ListingType, &l.Price, &l.PricePeriod, &l.Status, &l.CreatedAt, &l.UpdatedAt); err != nil {
			continue
		}
		items = append(items, &l)
	}
	return items, nil
}

func (r *Repository) UpdatePrice(ctx context.Context, id string, price float64, pricePeriod string) error {
	var period interface{}
	if pricePeriod != "" {
		period = pricePeriod
	}
	_, err := r.pg.Exec(ctx,
		"UPDATE listings SET price = $1, price_period = $2, updated_at = NOW() WHERE id = $3",
		price, period, id,
	)
	return err
}

func (r *Repository) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.pg.Exec(ctx, "UPDATE listings SET status = $1, updated_at = NOW() WHERE id = $2", status, id)
	return err
}

func (r *Repository) DeleteListing(ctx context.Context, id string) error {
	_, err := r.pg.Exec(ctx, "DELETE FROM listings WHERE id = $1", id)
	return err
}
