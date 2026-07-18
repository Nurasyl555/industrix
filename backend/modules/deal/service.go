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
	// Transition advances a deal to target status per the state machine.
	Transition(ctx context.Context, id, userID, target string) (*Deal, error)
	// Close is a convenience for cancelling a deal (kept for API compatibility).
	Close(ctx context.Context, id, userID string) error

	ListMessages(ctx context.Context, dealID, userID string) ([]*DealMessage, error)
	PostMessage(ctx context.Context, dealID, userID, body string) (*DealMessage, error)

	// Contracts
	contracts.DealProvider
}

type service struct {
	repo     *Repository
	listings contracts.ListingProvider
	notifier contracts.Notifier
	events   contracts.EventPublisher
}

// NewService creates a new deal service
func NewService(repo *Repository, listings contracts.ListingProvider, notifier contracts.Notifier, events contracts.EventPublisher) Service {
	return &service{repo: repo, listings: listings, notifier: notifier, events: events}
}

// dealStatusEvent is the payload published on deal.status.changed.
type dealStatusEvent struct {
	ID        string `json:"id"`
	ListingID string `json:"listing_id"`
	BuyerID   string `json:"buyer_id"`
	SellerID  string `json:"seller_id"`
	From      string `json:"from"`
	To        string `json:"to"`
}

// emit publishes a domain event if a publisher is wired.
func (s *service) emit(ctx context.Context, topic, key string, payload any) {
	if s.events != nil {
		s.events.Publish(ctx, topic, key, payload)
	}
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

// Transition advances a deal through the state machine, validating that the
// requested move is legal from the current status. It emits deal.status.changed
// and notifies the counterparty on success.
func (s *service) Transition(ctx context.Context, id, userID, target string) (*Deal, error) {
	d, err := s.requireParticipant(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if _, ok := allowedTransitions[target]; !ok && target != StatusCompleted && target != StatusCancelled {
		return nil, errors.New(errors.CodeValidation, "Unknown deal status: "+target)
	}
	if !canTransition(d.Status, target) {
		return nil, errors.New(errors.CodeValidation, "Cannot move deal from "+d.Status+" to "+target)
	}
	if err := s.repo.UpdateStatus(ctx, id, target); err != nil {
		return nil, err
	}

	from := d.Status
	d.Status = target
	s.emit(ctx, contracts.TopicDealStatusChanged, id, dealStatusEvent{
		ID: id, ListingID: d.ListingID, BuyerID: d.BuyerID, SellerID: d.SellerID,
		From: from, To: target,
	})
	if s.notifier != nil {
		recipient := d.SellerID
		if userID == d.SellerID {
			recipient = d.BuyerID
		}
		s.notifier.Notify(ctx, recipient, "deal_status", "Deal status changed to "+target, "/shop/deals/"+id)
	}
	return d, nil
}

// Close cancels a deal. Retained for backward compatibility with the existing
// PUT /deals/:id/close route; it delegates to Transition(→cancelled).
func (s *service) Close(ctx context.Context, id, userID string) error {
	_, err := s.Transition(ctx, id, userID, StatusCancelled)
	return err
}

// === Contracts (DealProvider) ===

func (s *service) GetDealBasic(ctx context.Context, dealID string) (*contracts.DealBasic, error) {
	d, err := s.repo.GetDealByID(ctx, dealID)
	if err != nil {
		return nil, err
	}
	return &contracts.DealBasic{
		ID:        d.ID,
		ListingID: d.ListingID,
		BuyerID:   d.BuyerID,
		SellerID:  d.SellerID,
		Status:    d.Status,
	}, nil
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
	if isTerminal(d.Status) {
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
