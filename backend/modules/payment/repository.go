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
	// deal_id / payee_id are absent for subscription charges — store NULL so
	// the UUID columns and the escrow-needs-deal check stay valid.
	var dealID, payeeID interface{}
	if p.DealID != "" {
		dealID = p.DealID
	}
	if p.PayeeID != "" {
		payeeID = p.PayeeID
	}
	kind := p.Kind
	if kind == "" {
		kind = KindEscrow
	}
	return r.pg.QueryRow(ctx,
		`INSERT INTO payments (deal_id, payer_id, payee_id, amount, currency, provider, status, kind, description)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, created_at, updated_at`,
		dealID, p.PayerID, payeeID, p.Amount, p.Currency, p.Provider, p.Status, kind, p.Description,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Payment, error) {
	var p Payment
	err := r.pg.QueryRow(ctx,
		`SELECT id, COALESCE(deal_id::text, ''), payer_id, COALESCE(payee_id::text, ''),
		        amount, currency, provider, status, COALESCE(provider_ref, ''),
		        kind, COALESCE(description, ''), created_at, updated_at
		 FROM payments WHERE id = $1`, id,
	).Scan(&p.ID, &p.DealID, &p.PayerID, &p.PayeeID, &p.Amount, &p.Currency,
		&p.Provider, &p.Status, &p.ProviderRef, &p.Kind, &p.Description, &p.CreatedAt, &p.UpdatedAt)
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

// ListByDealAndStatus returns a deal's payments in a given status (used by the
// deal-status consumer to find held escrow to release/refund).
func (r *Repository) ListByDealAndStatus(ctx context.Context, dealID, status string) ([]*Payment, error) {
	rows, err := r.pg.Query(ctx,
		`SELECT id, COALESCE(deal_id::text, ''), payer_id, COALESCE(payee_id::text, ''),
		        amount, currency, provider, status, COALESCE(provider_ref, ''),
		        kind, COALESCE(description, ''), created_at, updated_at
		 FROM payments WHERE deal_id = $1 AND status = $2`, dealID, status,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Payment
	for rows.Next() {
		var p Payment
		if err := rows.Scan(&p.ID, &p.DealID, &p.PayerID, &p.PayeeID, &p.Amount, &p.Currency,
			&p.Provider, &p.Status, &p.ProviderRef, &p.Kind, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			continue
		}
		out = append(out, &p)
	}
	return out, nil
}

// ListForUser returns payments where the user is payer or payee (their history).
func (r *Repository) ListForUser(ctx context.Context, userID string) ([]*Payment, error) {
	rows, err := r.pg.Query(ctx,
		`SELECT id, COALESCE(deal_id::text, ''), payer_id, COALESCE(payee_id::text, ''),
		        amount, currency, provider, status, COALESCE(provider_ref, ''),
		        kind, COALESCE(description, ''), created_at, updated_at
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
			&p.Provider, &p.Status, &p.ProviderRef, &p.Kind, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			continue
		}
		out = append(out, &p)
	}
	return out, nil
}
