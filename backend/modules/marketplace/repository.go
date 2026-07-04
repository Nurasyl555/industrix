package marketplace

import (
	"context"
	"fmt"

	"github.com/industrix/backend/pkg/errors"
	"github.com/industrix/backend/pkg/postgres"
)

// Repository handles all review/reputation database operations
type Repository struct {
	pg *postgres.Client
}

// NewRepository creates a new marketplace repository
func NewRepository(pg *postgres.Client) *Repository {
	return &Repository{pg: pg}
}

func (r *Repository) CreateReview(ctx context.Context, rev *Review) error {
	// transaction_id is a nullable UUID column — an empty string is not a
	// valid UUID, so send NULL when the review isn't tied to a transaction.
	var txID interface{}
	if rev.TransactionID != "" {
		txID = rev.TransactionID
	}
	err := r.pg.QueryRow(ctx,
		`INSERT INTO reviews (author_id, target_entity_id, rating, comment, transaction_id)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`,
		rev.AuthorID, rev.TargetEntityID, rev.Rating, rev.Comment, txID,
	).Scan(&rev.ID, &rev.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create review: %w", err)
	}
	return nil
}

func (r *Repository) GetReviewsByEntity(ctx context.Context, entityID string, page, limit int) ([]*Review, int64, error) {
	offset := (page - 1) * limit
	rows, err := r.pg.Query(ctx,
		`SELECT id, author_id, target_entity_id, rating, comment, created_at
		 FROM reviews WHERE target_entity_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		entityID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reviews []*Review
	for rows.Next() {
		var rev Review
		if err := rows.Scan(&rev.ID, &rev.AuthorID, &rev.TargetEntityID, &rev.Rating, &rev.Comment, &rev.CreatedAt); err != nil {
			continue
		}
		reviews = append(reviews, &rev)
	}

	var total int64
	_ = r.pg.QueryRow(ctx, "SELECT COUNT(*) FROM reviews WHERE target_entity_id = $1", entityID).Scan(&total)

	return reviews, total, nil
}

func (r *Repository) GetReputationScore(ctx context.Context, entityID string) (*ReputationScore, error) {
	var s ReputationScore
	err := r.pg.QueryRow(ctx,
		"SELECT entity_id, average_rating, review_count, tier FROM reputation_scores WHERE entity_id = $1",
		entityID,
	).Scan(&s.EntityID, &s.AverageRating, &s.ReviewCount, &s.Tier)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "Reputation not found")
	}
	return &s, nil
}

func (r *Repository) UpdateReputationScore(ctx context.Context, score *ReputationScore) error {
	_, err := r.pg.Exec(ctx,
		`INSERT INTO reputation_scores (entity_id, average_rating, review_count, tier, last_updated)
		 VALUES ($1, $2, $3, $4, NOW())
		 ON CONFLICT (entity_id) DO UPDATE SET
		 average_rating = EXCLUDED.average_rating,
		 review_count = EXCLUDED.review_count,
		 tier = EXCLUDED.tier,
		 last_updated = NOW()`,
		score.EntityID, score.AverageRating, score.ReviewCount, score.Tier,
	)
	return err
}
