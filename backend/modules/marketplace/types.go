package marketplace

import "time"

// Review represents a review
type Review struct {
	ID             string    `json:"id"`
	AuthorID       string    `json:"author_id"`
	TargetEntityID string    `json:"target_entity_id"`
	Rating         int       `json:"rating"`
	Comment        string    `json:"comment"`
	TransactionID  string    `json:"transaction_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ReputationScore represents an entity's aggregated reputation
type ReputationScore struct {
	EntityID      string  `json:"entity_id"`
	AverageRating float64 `json:"average_rating"`
	ReviewCount   int     `json:"review_count"`
	Tier          string  `json:"tier"` // gold, silver, bronze, none
}

// CreateReviewRequest represents create review request
type CreateReviewRequest struct {
	TargetEntityID string `json:"target_entity_id"`
	Rating         int    `json:"rating"`
	Comment        string `json:"comment"`
	TransactionID  string `json:"transaction_id"`
}
