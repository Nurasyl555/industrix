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
}

// NewService creates a new listing service
func NewService(repo *Repository, equipment contracts.EquipmentProvider) Service {
	return &service{repo: repo, equipment: equipment}
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
	if _, err := s.requireOwner(ctx, id, sellerID); err != nil {
		return err
	}
	return s.repo.UpdateStatus(ctx, id, "moderation")
}

func (s *service) Archive(ctx context.Context, id, sellerID string) error {
	if _, err := s.requireOwner(ctx, id, sellerID); err != nil {
		return err
	}
	return s.repo.UpdateStatus(ctx, id, "archived")
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
	if _, err := s.repo.GetListingByID(ctx, id); err != nil {
		return err
	}
	return s.repo.UpdateStatus(ctx, id, "active")
}

func (s *service) Reject(ctx context.Context, id string) error {
	if _, err := s.repo.GetListingByID(ctx, id); err != nil {
		return err
	}
	return s.repo.UpdateStatus(ctx, id, "rejected")
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
	}, nil
}
