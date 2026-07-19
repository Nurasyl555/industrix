package integrity

import (
	"context"
	"time"
)

// GetSubscriptionRow loads a user's subscription row. Returns ok=false when the
// user has no row yet (treated as the free plan by the service).
func (r *Repository) GetSubscriptionRow(ctx context.Context, userID string) (plan, status string, expiresAt *time.Time, updatedAt time.Time, ok bool, err error) {
	err = r.pg.QueryRow(ctx,
		`SELECT plan, status, expires_at, updated_at FROM subscriptions WHERE user_id = $1`, userID,
	).Scan(&plan, &status, &expiresAt, &updatedAt)
	if err != nil {
		// No row (or any read error): fall back to free; not fatal.
		return "", "", nil, time.Time{}, false, nil
	}
	return plan, status, expiresAt, updatedAt, true, nil
}

// UpsertPlan sets a user's plan, creating the subscription row if absent.
// expiresAt is nil for the free plan, which never lapses; paymentID links the
// charge that paid for this period (empty when nothing was charged).
func (r *Repository) UpsertPlan(ctx context.Context, userID, plan string, expiresAt *time.Time, paymentID string) error {
	var pid interface{}
	if paymentID != "" {
		pid = paymentID
	}
	var exp interface{}
	if expiresAt != nil {
		exp = *expiresAt
	}
	_, err := r.pg.Exec(ctx,
		`INSERT INTO subscriptions (user_id, plan, status, expires_at, last_payment_id)
		 VALUES ($1, $2, 'active', $3, $4)
		 ON CONFLICT (user_id)
		 DO UPDATE SET plan = EXCLUDED.plan, status = 'active',
		               expires_at = EXCLUDED.expires_at,
		               last_payment_id = EXCLUDED.last_payment_id,
		               updated_at = NOW()`,
		userID, plan, exp, pid)
	return err
}
