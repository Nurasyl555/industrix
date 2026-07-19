package dispute

import (
	"context"
	"strings"

	"github.com/industrix/backend/pkg/errors"
	"github.com/industrix/backend/pkg/postgres"
)

// Repository handles dispute persistence.
type Repository struct {
	pg *postgres.Client
}

func NewRepository(pg *postgres.Client) *Repository {
	return &Repository{pg: pg}
}

// ErrAlreadyOpen is returned when a deal already has an unresolved dispute —
// the partial unique index rejects the insert.
var ErrAlreadyOpen = errors.New(errors.CodeConflict, "This deal already has an open dispute")

const disputeCols = `id, deal_id, filed_by, reason, evidence_urls, status,
	COALESCE(resolution_note, ''), COALESCE(resolved_by::text, ''), created_at, updated_at`

func (r *Repository) Create(ctx context.Context, d *Dispute) error {
	err := r.pg.QueryRow(ctx,
		`INSERT INTO disputes (deal_id, filed_by, reason, evidence_urls)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, status, created_at, updated_at`,
		d.DealID, d.FiledBy, d.Reason, d.EvidenceURLs,
	).Scan(&d.ID, &d.Status, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		// The partial unique index surfaces as a 23505 unique_violation.
		if strings.Contains(err.Error(), "idx_disputes_one_open_per_deal") || strings.Contains(err.Error(), "23505") {
			return ErrAlreadyOpen
		}
		return err
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Dispute, error) {
	var d Dispute
	err := r.pg.QueryRow(ctx,
		`SELECT `+disputeCols+` FROM disputes WHERE id = $1`, id,
	).Scan(&d.ID, &d.DealID, &d.FiledBy, &d.Reason, &d.EvidenceURLs, &d.Status,
		&d.ResolutionNote, &d.ResolvedBy, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "Dispute not found")
	}
	return &d, nil
}

// ListByUser returns disputes the user filed.
func (r *Repository) ListByUser(ctx context.Context, userID string) ([]*Dispute, error) {
	return r.query(ctx, `SELECT `+disputeCols+` FROM disputes WHERE filed_by = $1 ORDER BY created_at DESC`, userID)
}

// ListByStatus backs the admin arbitration queue.
func (r *Repository) ListByStatus(ctx context.Context, status string) ([]*Dispute, error) {
	return r.query(ctx, `SELECT `+disputeCols+` FROM disputes WHERE status = $1 ORDER BY created_at ASC`, status)
}

func (r *Repository) query(ctx context.Context, sql string, args ...interface{}) ([]*Dispute, error) {
	rows, err := r.pg.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Dispute
	for rows.Next() {
		var d Dispute
		if err := rows.Scan(&d.ID, &d.DealID, &d.FiledBy, &d.Reason, &d.EvidenceURLs, &d.Status,
			&d.ResolutionNote, &d.ResolvedBy, &d.CreatedAt, &d.UpdatedAt); err != nil {
			continue
		}
		out = append(out, &d)
	}
	return out, nil
}

// Resolve records the arbitration decision. It only touches an open dispute, so
// a second decision on the same dispute is a no-op rather than an overwrite.
func (r *Repository) Resolve(ctx context.Context, id, status, note, resolvedBy string) (bool, error) {
	tag, err := r.pg.Exec(ctx,
		`UPDATE disputes SET status = $2, resolution_note = $3, resolved_by = $4, updated_at = NOW()
		 WHERE id = $1 AND status = 'open'`,
		id, status, note, resolvedBy)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}
