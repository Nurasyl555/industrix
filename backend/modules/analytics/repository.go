package analytics

import (
	"context"
	"fmt"

	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/postgres"
)

// Repository owns the analytics event store and its aggregate queries.
type Repository struct {
	pg *postgres.Client
}

func NewRepository(pg *postgres.Client) *Repository {
	return &Repository{pg: pg}
}

// Record appends one event. Empty ids are stored as NULL so the UUID columns
// stay valid.
func (r *Repository) Record(ctx context.Context, e Event) error {
	var entityID, sellerID interface{}
	if e.EntityID != "" {
		entityID = e.EntityID
	}
	if e.SellerID != "" {
		sellerID = e.SellerID
	}
	var amount interface{}
	if e.Amount != 0 {
		amount = e.Amount
	}
	_, err := r.pg.Exec(ctx,
		`INSERT INTO analytics_events (event_type, entity_id, seller_id, amount)
		 VALUES ($1, $2, $3, $4)`,
		e.EventType, entityID, sellerID, amount)
	return err
}

// countsFor returns per-event-type counts for a seller (or platform-wide when
// sellerID is empty) within the window.
func (r *Repository) countsFor(ctx context.Context, sellerID string, days int) (map[string]int, error) {
	query := fmt.Sprintf(
		`SELECT event_type, COUNT(*) FROM analytics_events
		 WHERE occurred_at >= NOW() - INTERVAL '%d days'`, days)
	args := []interface{}{}
	if sellerID != "" {
		query += " AND seller_id = $1"
		args = append(args, sellerID)
	}
	query += " GROUP BY event_type"

	rows, err := r.pg.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]int{}
	for rows.Next() {
		var t string
		var n int
		if err := rows.Scan(&t, &n); err != nil {
			continue
		}
		out[t] = n
	}
	return out, nil
}

// revenue sums completed-payment amounts for a seller (or platform-wide when
// sellerID is empty) within the window.
func (r *Repository) revenue(ctx context.Context, sellerID string, days int) (float64, error) {
	query := fmt.Sprintf(
		`SELECT COALESCE(SUM(amount), 0) FROM analytics_events
		 WHERE event_type = '%s' AND occurred_at >= NOW() - INTERVAL '%d days'`,
		contracts.TopicPaymentCompleted, days)
	args := []interface{}{}
	if sellerID != "" {
		query += " AND seller_id = $1"
		args = append(args, sellerID)
	}
	var total float64
	err := r.pg.QueryRow(ctx, query, args...).Scan(&total)
	return total, err
}

// activeSellers counts distinct sellers with any event in the window.
func (r *Repository) activeSellers(ctx context.Context, days int) (int, error) {
	var n int
	err := r.pg.QueryRow(ctx, fmt.Sprintf(
		`SELECT COUNT(DISTINCT seller_id) FROM analytics_events
		 WHERE seller_id IS NOT NULL AND occurred_at >= NOW() - INTERVAL '%d days'`, days)).Scan(&n)
	return n, err
}

// dailyGMV buckets completed-payment amounts per day.
func (r *Repository) dailyGMV(ctx context.Context, days int) ([]DailyGMVBucket, error) {
	rows, err := r.pg.Query(ctx, fmt.Sprintf(
		`SELECT to_char(occurred_at::date, 'YYYY-MM-DD'), COALESCE(SUM(amount), 0)
		 FROM analytics_events
		 WHERE event_type = '%s' AND occurred_at >= NOW() - INTERVAL '%d days'
		 GROUP BY occurred_at::date ORDER BY occurred_at::date`,
		contracts.TopicPaymentCompleted, days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []DailyGMVBucket
	for rows.Next() {
		var b DailyGMVBucket
		if err := rows.Scan(&b.Day, &b.GMV); err != nil {
			continue
		}
		out = append(out, b)
	}
	return out, nil
}
