package listing

import (
	"context"

	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/errors"
)

// Service defines the listing service interface
type Service interface {
	CreateListing(ctx context.Context, sellerID string, req CreateListingRequest) (*Listing, error)
	GetListing(ctx context.Context, id string) (*ListingView, error)
	ListActive(ctx context.Context, f ListListingsFilter) ([]*ListingView, int64, error)
	ListMy(ctx context.Context, sellerID string) ([]*Listing, error)
	UpdateListing(ctx context.Context, id, sellerID string, req UpdateListingRequest) (*Listing, error)
	Publish(ctx context.Context, id, sellerID string) error
	Archive(ctx context.Context, id, sellerID string) error
	DeleteListing(ctx context.Context, id, sellerID string) error

	// Admin moderation
	ListForModeration(ctx context.Context) ([]*ListingView, error)
	Approve(ctx context.Context, id string) error
	Reject(ctx context.Context, id string) error

	// Contracts
	contracts.ListingProvider
}

type service struct {
	repo      *Repository
	equipment contracts.EquipmentProvider
	notifier  contracts.Notifier
	events    contracts.EventPublisher
	subs      contracts.SubscriptionProvider
}

// NewService creates a new listing service
func NewService(repo *Repository, equipment contracts.EquipmentProvider, notifier contracts.Notifier, events contracts.EventPublisher, subs contracts.SubscriptionProvider) Service {
	return &service{repo: repo, equipment: equipment, notifier: notifier, events: events, subs: subs}
}

// liveStatuses are the listing states that count against a seller's plan limit:
// pending moderation and live listings both occupy a slot.
var liveStatuses = []string{"moderation", "active"}

// listingEvent is the payload published on listing.* topics — enough for search
// indexing without consumers calling back into listing.
type listingEvent struct {
	ID          string  `json:"id"`
	EquipmentID string  `json:"equipment_id"`
	SellerID    string  `json:"seller_id"`
	ListingType string  `json:"listing_type"`
	Price       float64 `json:"price"`
	PricePeriod string  `json:"price_period"`
	Status      string  `json:"status"`
}

func toListingEvent(l *Listing) listingEvent {
	return listingEvent{
		ID:          l.ID,
		EquipmentID: l.EquipmentID,
		SellerID:    l.SellerID,
		ListingType: l.ListingType,
		Price:       l.Price,
		PricePeriod: l.PricePeriod,
		Status:      l.Status,
	}
}

// emit publishes a domain event if a publisher is wired.
func (s *service) emit(ctx context.Context, topic, key string, payload any) {
	if s.events != nil {
		s.events.Publish(ctx, topic, key, payload)
	}
}

var validListingTypes = map[string]bool{"sale": true, "rental": true}
var validPricePeriods = map[string]bool{"": true, "day": true, "week": true, "month": true}

func (s *service) CreateListing(ctx context.Context, sellerID string, req CreateListingRequest) (*Listing, error) {
	if req.EquipmentID == "" {
		return nil, errors.New(errors.CodeValidation, "Equipment is required")
	}
	if !validListingTypes[req.ListingType] {
		return nil, errors.New(errors.CodeValidation, "Listing type must be 'sale' or 'rental'")
	}
	if req.Price <= 0 {
		return nil, errors.New(errors.CodeValidation, "Price must be greater than 0")
	}
	if !validPricePeriods[req.PricePeriod] {
		return nil, errors.New(errors.CodeValidation, "Price period must be 'day', 'week' or 'month'")
	}

	eq, err := s.equipment.GetEquipmentBasic(ctx, req.EquipmentID)
	if err != nil {
		return nil, errors.New(errors.CodeValidation, "Equipment does not exist")
	}
	if eq.OwnerID != sellerID {
		return nil, errors.New(errors.CodeUnauthorized, "You do not own this equipment")
	}

	l := &Listing{
		EquipmentID: req.EquipmentID,
		SellerID:    sellerID,
		ListingType: req.ListingType,
		Price:       req.Price,
		PricePeriod: req.PricePeriod,
	}
	if err := s.repo.CreateListing(ctx, l); err != nil {
		return nil, err
	}
	return l, nil
}

func (s *service) GetListing(ctx context.Context, id string) (*ListingView, error) {
	// Count this detail view before reading, so the returned count is fresh.
	// Best-effort: a counter failure must not break the read.
	_ = s.repo.IncrementViewCount(ctx, id)
	return s.repo.GetListingViewByID(ctx, id)
}

func (s *service) ListActive(ctx context.Context, f ListListingsFilter) ([]*ListingView, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
	return s.repo.ListActive(ctx, f)
}

func (s *service) ListMy(ctx context.Context, sellerID string) ([]*Listing, error) {
	return s.repo.ListBySeller(ctx, sellerID)
}

func (s *service) requireOwner(ctx context.Context, id, sellerID string) (*Listing, error) {
	l, err := s.repo.GetListingByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if l.SellerID != sellerID {
		return nil, errors.New(errors.CodeUnauthorized, "You do not own this listing")
	}
	return l, nil
}

func (s *service) UpdateListing(ctx context.Context, id, sellerID string, req UpdateListingRequest) (*Listing, error) {
	l, err := s.requireOwner(ctx, id, sellerID)
	if err != nil {
		return nil, err
	}
	if req.Price <= 0 {
		return nil, errors.New(errors.CodeValidation, "Price must be greater than 0")
	}
	if !validPricePeriods[req.PricePeriod] {
		return nil, errors.New(errors.CodeValidation, "Price period must be 'day', 'week' or 'month'")
	}

	if err := s.repo.UpdatePrice(ctx, id, req.Price, req.PricePeriod); err != nil {
		return nil, err
	}
	l.Price = req.Price
	l.PricePeriod = req.PricePeriod
	return l, nil
}

// Publish submits a draft for moderation — it does NOT go live directly. An
// admin approves it (Approve) before buyers can see it.
func (s *service) Publish(ctx context.Context, id, sellerID string) error {
	l, err := s.requireOwner(ctx, id, sellerID)
	if err != nil {
		return err
	}
	// Enforce the seller's subscription plan limit before the listing occupies a
	// live/pending slot. A draft going to moderation counts against the cap.
	if s.subs != nil {
		if limit := s.subs.ListingLimit(ctx, sellerID); limit >= 0 {
			count, err := s.repo.CountBySellerStatuses(ctx, sellerID, liveStatuses)
			if err != nil {
				return err
			}
			if count >= limit {
				return errors.New(errors.CodeConflict, "Listing limit reached for your plan — upgrade to publish more")
			}
		}
	}
	if err := s.repo.UpdateStatus(ctx, id, "moderation"); err != nil {
		return err
	}
	l.Status = "moderation"
	s.emit(ctx, contracts.TopicListingSubmitted, id, toListingEvent(l))
	return nil
}

func (s *service) Archive(ctx context.Context, id, sellerID string) error {
	l, err := s.requireOwner(ctx, id, sellerID)
	if err != nil {
		return err
	}
	if err := s.repo.UpdateStatus(ctx, id, "archived"); err != nil {
		return err
	}
	l.Status = "archived"
	s.emit(ctx, contracts.TopicListingDeactivated, id, toListingEvent(l))
	return nil
}

func (s *service) DeleteListing(ctx context.Context, id, sellerID string) error {
	if _, err := s.requireOwner(ctx, id, sellerID); err != nil {
		return err
	}
	return s.repo.DeleteListing(ctx, id)
}

// === Admin moderation ===

func (s *service) ListForModeration(ctx context.Context) ([]*ListingView, error) {
	return s.repo.ListByStatusView(ctx, "moderation")
}

func (s *service) Approve(ctx context.Context, id string) error {
	l, err := s.repo.GetListingByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.UpdateStatus(ctx, id, "active"); err != nil {
		return err
	}
	l.Status = "active"
	s.emit(ctx, contracts.TopicListingPublished, id, toListingEvent(l))
	if s.notifier != nil {
		s.notifier.Notify(ctx, l.SellerID, "listing_approved", "Your listing was approved and is now live", "/shop/details?id="+id)
	}
	return nil
}

func (s *service) Reject(ctx context.Context, id string) error {
	l, err := s.repo.GetListingByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.UpdateStatus(ctx, id, "rejected"); err != nil {
		return err
	}
	l.Status = "rejected"
	s.emit(ctx, contracts.TopicListingDeactivated, id, toListingEvent(l))
	if s.notifier != nil {
		s.notifier.Notify(ctx, l.SellerID, "listing_rejected", "Your listing was rejected by moderation", "")
	}
	return nil
}

// === Contracts (ListingProvider) ===

func (s *service) GetListingBasic(ctx context.Context, listingID string) (*contracts.ListingBasic, error) {
	l, err := s.repo.GetListingByID(ctx, listingID)
	if err != nil {
		return nil, err
	}
	return &contracts.ListingBasic{
		ID:          l.ID,
		EquipmentID: l.EquipmentID,
		SellerID:    l.SellerID,
		Status:      l.Status,
		ListingType: l.ListingType,
	}, nil
}
