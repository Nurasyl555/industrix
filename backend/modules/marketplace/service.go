package marketplace

import (
	"context"

	"github.com/industrix/backend/pkg/errors"
)

// Service defines the marketplace service interface
type Service interface {
	CreateReview(ctx context.Context, review *Review) error
	GetReviews(ctx context.Context, entityID string, page, limit int) ([]*Review, int64, error)
	GetReputation(ctx context.Context, entityID string) (*ReputationScore, error)
}

type service struct {
	repo *Repository
}

// NewService creates a new marketplace service
func NewService(repo *Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateReview(ctx context.Context, review *Review) error {
	if review.Rating < 1 || review.Rating > 5 {
		return errors.New(errors.CodeValidation, "Rating must be between 1 and 5")
	}

	if err := s.repo.CreateReview(ctx, review); err != nil {
		return err
	}

	return s.updateReputation(ctx, review.TargetEntityID)
}

func (s *service) updateReputation(ctx context.Context, entityID string) error {
	reviews, total, err := s.repo.GetReviewsByEntity(ctx, entityID, 1, 1000)
	if err != nil {
		return err
	}

	var sum int
	for _, r := range reviews {
		sum += r.Rating
	}

	avg := 0.0
	if total > 0 {
		avg = float64(sum) / float64(total)
	}

	tier := "none"
	if total >= 10 && avg >= 4.5 {
		tier = "gold"
	} else if total >= 5 && avg >= 4.0 {
		tier = "silver"
	} else if total >= 3 && avg >= 3.5 {
		tier = "bronze"
	}

	score := &ReputationScore{
		EntityID:      entityID,
		AverageRating: avg,
		ReviewCount:   int(total),
		Tier:          tier,
	}

	return s.repo.UpdateReputationScore(ctx, score)
}

func (s *service) GetReviews(ctx context.Context, entityID string, page, limit int) ([]*Review, int64, error) {
	return s.repo.GetReviewsByEntity(ctx, entityID, page, limit)
}

func (s *service) GetReputation(ctx context.Context, entityID string) (*ReputationScore, error) {
	return s.repo.GetReputationScore(ctx, entityID)
}
