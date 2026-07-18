package payment

import (
	"context"

	"github.com/industrix/backend/pkg/errors"
	"github.com/industrix/backend/pkg/postgres"
)

// Repository handles all payment database operations.
type Repository struct {
	pg *postgres.Client
}

func NewRepository(pg *postgres.Client) *Repository {
	return &Repository{pg: pg}
}

func (r *Repository) Create(ctx context.Context, p *Payment) error {
	return r.pg.QueryRow(ctx,
		`INSERT INTO payments (deal_id, payer_id, payee_id, amount, currency, provider, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, created_at, updated_at`,
		p.DealID, p.PayerID, p.PayeeID, p.Amount, p.Currency, p.Provider, p.Status,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Payment, error) {
	var p Payment
	err := r.pg.QueryRow(ctx,
		`SELECT id, deal_id, payer_id, payee_id, amount, currency, provider, status,
		        COALESCE(provider_ref, ''), created_at, updated_at
		 FROM payments WHERE id = $1`, id,
	).Scan(&p.ID, &p.DealID, &p.PayerID, &p.PayeeID, &p.Amount, &p.Currency,
		&p.Provider, &p.Status, &p.ProviderRef, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "Payment not found")
	}
	return &p, nil
}

// UpdateStatus updates status and the provider reference in one shot.
func (r *Repository) UpdateStatus(ctx context.Context, id, status, providerRef string) error {
	_, err := r.pg.Exec(ctx,
		`UPDATE payments SET status = $2, provider_ref = $3, updated_at = NOW() WHERE id = $1`,
		id, status, providerRef)
	return err
}

// ListForUser returns payments where the user is payer or payee (their history).
func (r *Repository) ListForUser(ctx context.Context, userID string) ([]*Payment, error) {
	rows, err := r.pg.Query(ctx,
		`SELECT id, deal_id, payer_id, payee_id, amount, currency, provider, status,
		        COALESCE(provider_ref, ''), created_at, updated_at
		 FROM payments WHERE payer_id = $1 OR payee_id = $1 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Payment
	for rows.Next() {
		var p Payment
		if err := rows.Scan(&p.ID, &p.DealID, &p.PayerID, &p.PayeeID, &p.Amount, &p.Currency,
			&p.Provider, &p.Status, &p.ProviderRef, &p.CreatedAt, &p.UpdatedAt); err != nil {
			continue
		}
		out = append(out, &p)
	}
	return out, nil
}
