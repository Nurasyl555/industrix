package engagement

import (
	"context"
	"fmt"

	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/errors"
)

// Service exposes the watchlist API plus the price-change handler the Kafka
// consumer drives.
type Service interface {
	AddFavorite(ctx context.Context, userID, listingID string) error
	RemoveFavorite(ctx context.Context, userID, listingID string) error
	ListFavorites(ctx context.Context, userID string) ([]*FavoriteListing, error)
	PriceHistory(ctx context.Context, listingID string) ([]*PriceHistoryEntry, error)

	// OnPriceChanged records a price change and, on a drop, alerts every user
	// watching the listing. Driven by the listing.price_changed consumer.
	OnPriceChanged(ctx context.Context, listingID string, oldPrice, newPrice float64)
}

type service struct {
	repo     *Repository
	notifier contracts.Notifier
}

func NewService(repo *Repository, notifier contracts.Notifier) Service {
	return &service{repo: repo, notifier: notifier}
}

func (s *service) AddFavorite(ctx context.Context, userID, listingID string) error {
	if listingID == "" {
		return errors.New(errors.CodeValidation, "Listing is required")
	}
	return s.repo.AddFavorite(ctx, userID, listingID)
}

func (s *service) RemoveFavorite(ctx context.Context, userID, listingID string) error {
	return s.repo.RemoveFavorite(ctx, userID, listingID)
}

func (s *service) ListFavorites(ctx context.Context, userID string) ([]*FavoriteListing, error) {
	return s.repo.ListFavorites(ctx, userID)
}

func (s *service) PriceHistory(ctx context.Context, listingID string) ([]*PriceHistoryEntry, error) {
	return s.repo.ListPriceHistory(ctx, listingID)
}

func (s *service) OnPriceChanged(ctx context.Context, listingID string, oldPrice, newPrice float64) {
	// Record history (best-effort).
	_ = s.repo.AddPriceHistory(ctx, listingID, oldPrice, newPrice)

	// Only a genuine drop triggers alerts.
	if newPrice >= oldPrice {
		return
	}
	if s.notifier == nil {
		return
	}
	watchers, err := s.repo.FavoritersOf(ctx, listingID)
	if err != nil {
		return
	}
	msg := fmt.Sprintf("Price dropped from %.0f to %.0f on a listing you follow", oldPrice, newPrice)
	link := "/shop/details?id=" + listingID
	for _, uid := range watchers {
		s.notifier.Notify(ctx, uid, "price_drop", msg, link)
	}
}
