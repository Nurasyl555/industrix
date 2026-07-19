package engagement

import (
	"context"

	"github.com/industrix/backend/pkg/postgres"
)

// Repository handles favorites and price-history persistence.
type Repository struct {
	pg *postgres.Client
}

func NewRepository(pg *postgres.Client) *Repository {
	return &Repository{pg: pg}
}

// AddFavorite is idempotent thanks to the (user_id, listing_id) unique index.
func (r *Repository) AddFavorite(ctx context.Context, userID, listingID string) error {
	_, err := r.pg.Exec(ctx,
		`INSERT INTO favorites (user_id, listing_id) VALUES ($1, $2)
		 ON CONFLICT (user_id, listing_id) DO NOTHING`, userID, listingID)
	return err
}

func (r *Repository) RemoveFavorite(ctx context.Context, userID, listingID string) error {
	_, err := r.pg.Exec(ctx,
		`DELETE FROM favorites WHERE user_id = $1 AND listing_id = $2`, userID, listingID)
	return err
}

// ListFavorites returns the user's watchlist joined with listing/equipment.
func (r *Repository) ListFavorites(ctx context.Context, userID string) ([]*FavoriteListing, error) {
	rows, err := r.pg.Query(ctx,
		`SELECT l.id, l.equipment_id, e.title, COALESCE(e.image_url, ''),
		        l.price, COALESCE(l.price_period, ''), l.listing_type, l.status, f.created_at
		 FROM favorites f
		 JOIN listings l ON l.id = f.listing_id
		 JOIN equipment e ON e.id = l.equipment_id
		 WHERE f.user_id = $1 ORDER BY f.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*FavoriteListing
	for rows.Next() {
		var f FavoriteListing
		if err := rows.Scan(&f.ListingID, &f.EquipmentID, &f.Title, &f.ImageURL,
			&f.Price, &f.PricePeriod, &f.ListingType, &f.Status, &f.FavoritedAt); err != nil {
			continue
		}
		out = append(out, &f)
	}
	return out, nil
}

// FavoritersOf returns the user ids watching a listing (for price-drop alerts).
func (r *Repository) FavoritersOf(ctx context.Context, listingID string) ([]string, error) {
	rows, err := r.pg.Query(ctx,
		`SELECT user_id FROM favorites WHERE listing_id = $1`, listingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *Repository) AddPriceHistory(ctx context.Context, listingID string, oldPrice, newPrice float64) error {
	_, err := r.pg.Exec(ctx,
		`INSERT INTO price_history (listing_id, old_price, new_price) VALUES ($1, $2, $3)`,
		listingID, oldPrice, newPrice)
	return err
}

func (r *Repository) ListPriceHistory(ctx context.Context, listingID string) ([]*PriceHistoryEntry, error) {
	rows, err := r.pg.Query(ctx,
		`SELECT listing_id, old_price, new_price, changed_at
		 FROM price_history WHERE listing_id = $1 ORDER BY changed_at DESC`, listingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*PriceHistoryEntry
	for rows.Next() {
		var e PriceHistoryEntry
		if err := rows.Scan(&e.ListingID, &e.OldPrice, &e.NewPrice, &e.ChangedAt); err != nil {
			continue
		}
		out = append(out, &e)
	}
	return out, nil
}
