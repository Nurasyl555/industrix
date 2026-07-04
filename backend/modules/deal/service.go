package deal

import (
	"context"

	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/errors"
)

// Service defines the deal service interface
type Service interface {
	CreateDeal(ctx context.Context, buyerID string, req CreateDealRequest) (*Deal, error)
	GetDeal(ctx context.Context, id, userID string) (*DealView, error)
	ListMy(ctx context.Context, userID string) ([]*DealView, error)
	Close(ctx context.Context, id, userID string) error

	ListMessages(ctx context.Context, dealID, userID string) ([]*DealMessage, error)
	PostMessage(ctx context.Context, dealID, userID, body string) (*DealMessage, error)
}

type service struct {
	repo     *Repository
	listings contracts.ListingProvider
	notifier contracts.Notifier
}

// NewService creates a new deal service
func NewService(repo *Repository, listings contracts.ListingProvider, notifier contracts.Notifier) Service {
	return &service{repo: repo, listings: listings, notifier: notifier}
}

func (s *service) CreateDeal(ctx context.Context, buyerID string, req CreateDealRequest) (*Deal, error) {
	if req.ListingID == "" {
		return nil, errors.New(errors.CodeValidation, "Listing is required")
	}

	l, err := s.listings.GetListingBasic(ctx, req.ListingID)
	if err != nil {
		return nil, errors.New(errors.CodeValidation, "Listing does not exist")
	}
	if l.Status != "active" {
		return nil, errors.New(errors.CodeValidation, "Listing is not active")
	}
	if l.SellerID == buyerID {
		return nil, errors.New(errors.CodeValidation, "You cannot inquire about your own listing")
	}

	d := &Deal{
		ListingID: req.ListingID,
		BuyerID:   buyerID,
		SellerID:  l.SellerID,
		Message:   req.Message,
	}
	if err := s.repo.CreateDeal(ctx, d); err != nil {
		return nil, err
	}
	if s.notifier != nil {
		s.notifier.Notify(ctx, d.SellerID, "inquiry", "You have a new inquiry on your listing", "/shop/deals/"+d.ID)
	}
	return d, nil
}

func viewFor(d *Deal, userID string) *DealView {
	role := "seller"
	if d.BuyerID == userID {
		role = "buyer"
	}
	return &DealView{Deal: *d, Role: role}
}

func (s *service) GetDeal(ctx context.Context, id, userID string) (*DealView, error) {
	d, err := s.repo.GetDealByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if d.BuyerID != userID && d.SellerID != userID {
		return nil, errors.New(errors.CodeUnauthorized, "You are not part of this deal")
	}
	return viewFor(d, userID), nil
}

func (s *service) ListMy(ctx context.Context, userID string) ([]*DealView, error) {
	deals, err := s.repo.ListForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	views := make([]*DealView, 0, len(deals))
	for _, d := range deals {
		views = append(views, viewFor(d, userID))
	}
	return views, nil
}

func (s *service) Close(ctx context.Context, id, userID string) error {
	d, err := s.repo.GetDealByID(ctx, id)
	if err != nil {
		return err
	}
	if d.BuyerID != userID && d.SellerID != userID {
		return errors.New(errors.CodeUnauthorized, "You are not part of this deal")
	}
	return s.repo.UpdateStatus(ctx, id, "closed")
}

// requireParticipant loads a deal and asserts the user is buyer or seller.
func (s *service) requireParticipant(ctx context.Context, dealID, userID string) (*Deal, error) {
	d, err := s.repo.GetDealByID(ctx, dealID)
	if err != nil {
		return nil, err
	}
	if d.BuyerID != userID && d.SellerID != userID {
		return nil, errors.New(errors.CodeUnauthorized, "You are not part of this deal")
	}
	return d, nil
}

func (s *service) ListMessages(ctx context.Context, dealID, userID string) ([]*DealMessage, error) {
	if _, err := s.requireParticipant(ctx, dealID, userID); err != nil {
		return nil, err
	}
	return s.repo.ListMessages(ctx, dealID)
}

func (s *service) PostMessage(ctx context.Context, dealID, userID, body string) (*DealMessage, error) {
	if body == "" {
		return nil, errors.New(errors.CodeValidation, "Message cannot be empty")
	}
	d, err := s.requireParticipant(ctx, dealID, userID)
	if err != nil {
		return nil, err
	}
	if d.Status == "closed" {
		return nil, errors.New(errors.CodeValidation, "This deal is closed")
	}
	msg, err := s.repo.AddMessage(ctx, dealID, userID, body)
	if err != nil {
		return nil, err
	}
	// Notify the other participant.
	if s.notifier != nil {
		recipient := d.SellerID
		if userID == d.SellerID {
			recipient = d.BuyerID
		}
		s.notifier.Notify(ctx, recipient, "message", "New message in a deal", "/shop/deals/"+dealID)
	}
	return msg, nil
}
