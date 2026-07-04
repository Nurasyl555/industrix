package booking

import (
	"context"
	"strings"

	"github.com/industrix/backend/pkg/errors"
	"github.com/industrix/backend/pkg/postgres"
)

// Repository handles all booking database operations.
type Repository struct {
	pg *postgres.Client
}

func NewRepository(pg *postgres.Client) *Repository {
	return &Repository{pg: pg}
}

// ErrOverlap is returned when the requested dates collide with an existing
// confirmed booking (the DB exclusion constraint rejects the insert).
var ErrOverlap = errors.New(errors.CodeConflict, "Those dates are already booked")

func (r *Repository) Create(ctx context.Context, b *Booking) error {
	err := r.pg.QueryRow(ctx,
		`INSERT INTO bookings (listing_id, renter_id, owner_id, start_date, end_date)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id, status, created_at`,
		b.ListingID, b.RenterID, b.OwnerID, b.StartDate, b.EndDate,
	).Scan(&b.ID, &b.Status, &b.CreatedAt)
	if err != nil {
		// The exclusion constraint surfaces as a 23P01 exclusion_violation.
		if strings.Contains(err.Error(), "exclusion") || strings.Contains(err.Error(), "conflicting key") || strings.Contains(err.Error(), "23P01") {
			return ErrOverlap
		}
		return err
	}
	return nil
}

// ConfirmedRanges returns the booked intervals of a listing (for availability).
func (r *Repository) ConfirmedRanges(ctx context.Context, listingID string) ([]*DateRange, error) {
	rows, err := r.pg.Query(ctx,
		`SELECT to_char(start_date, 'YYYY-MM-DD'), to_char(end_date, 'YYYY-MM-DD')
		 FROM bookings WHERE listing_id = $1 AND status = 'confirmed' AND end_date >= CURRENT_DATE
		 ORDER BY start_date`, listingID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*DateRange
	for rows.Next() {
		var d DateRange
		if err := rows.Scan(&d.StartDate, &d.EndDate); err != nil {
			continue
		}
		out = append(out, &d)
	}
	return out, nil
}

func (r *Repository) ListByRenter(ctx context.Context, renterID string) ([]*Booking, error) {
	rows, err := r.pg.Query(ctx,
		`SELECT id, listing_id, renter_id, owner_id, to_char(start_date, 'YYYY-MM-DD'),
		        to_char(end_date, 'YYYY-MM-DD'), status, created_at
		 FROM bookings WHERE renter_id = $1 ORDER BY start_date DESC`, renterID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Booking
	for rows.Next() {
		var b Booking
		if err := rows.Scan(&b.ID, &b.ListingID, &b.RenterID, &b.OwnerID, &b.StartDate, &b.EndDate, &b.Status, &b.CreatedAt); err != nil {
			continue
		}
		out = append(out, &b)
	}
	return out, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Booking, error) {
	var b Booking
	err := r.pg.QueryRow(ctx,
		`SELECT id, listing_id, renter_id, owner_id, to_char(start_date, 'YYYY-MM-DD'),
		        to_char(end_date, 'YYYY-MM-DD'), status, created_at
		 FROM bookings WHERE id = $1`, id,
	).Scan(&b.ID, &b.ListingID, &b.RenterID, &b.OwnerID, &b.StartDate, &b.EndDate, &b.Status, &b.CreatedAt)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "Booking not found")
	}
	return &b, nil
}

func (r *Repository) Cancel(ctx context.Context, id string) error {
	_, err := r.pg.Exec(ctx, "UPDATE bookings SET status = 'cancelled' WHERE id = $1", id)
	return err
}
