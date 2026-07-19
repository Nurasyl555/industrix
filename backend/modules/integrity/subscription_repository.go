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
func (r *Repository) UpsertPlan(ctx context.Context, userID, plan string) error {
	_, err := r.pg.Exec(ctx,
		`INSERT INTO subscriptions (user_id, plan, status)
		 VALUES ($1, $2, 'active')
		 ON CONFLICT (user_id)
		 DO UPDATE SET plan = EXCLUDED.plan, status = 'active', updated_at = NOW()`,
		userID, plan)
	return err
}
