package dispute

import (
	"context"

	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/errors"
)

// Service defines the dispute/arbitration service.
type Service interface {
	// File opens a dispute on a deal. Only a participant may file, and only
	// one dispute can be open per deal at a time.
	File(ctx context.Context, userID string, req FileDisputeRequest) (*Dispute, error)
	Get(ctx context.Context, id, userID string, isAdmin bool) (*Dispute, error)
	ListMine(ctx context.Context, userID string) ([]*Dispute, error)

	// Admin arbitration.
	ListOpen(ctx context.Context) ([]*Dispute, error)
	Resolve(ctx context.Context, id, adminID string, req ResolveRequest) (*Dispute, error)
}

type service struct {
	repo     *Repository
	deals    contracts.DealProvider
	escrow   contracts.EscrowSettler
	events   contracts.EventPublisher
	notifier contracts.Notifier
}

func NewService(repo *Repository, deals contracts.DealProvider, escrow contracts.EscrowSettler,
	events contracts.EventPublisher, notifier contracts.Notifier) Service {
	return &service{repo: repo, deals: deals, escrow: escrow, events: events, notifier: notifier}
}

func (s *service) emit(ctx context.Context, topic string, d *Dispute, decidedBy string) {
	if s.events != nil {
		s.events.Publish(ctx, topic, d.ID, disputeEvent{
			ID: d.ID, DealID: d.DealID, FiledBy: d.FiledBy, Status: d.Status, DecidedBy: decidedBy,
		})
	}
}

func (s *service) notify(ctx context.Context, userID, ntype, msg, link string) {
	if s.notifier != nil && userID != "" {
		s.notifier.Notify(ctx, userID, ntype, msg, link)
	}
}

func (s *service) File(ctx context.Context, userID string, req FileDisputeRequest) (*Dispute, error) {
	if req.DealID == "" {
		return nil, errors.New(errors.CodeValidation, "Deal is required")
	}
	if req.Reason == "" {
		return nil, errors.New(errors.CodeValidation, "Please describe what went wrong")
	}

	deal, err := s.deals.GetDealBasic(ctx, req.DealID)
	if err != nil {
		return nil, errors.New(errors.CodeValidation, "Deal does not exist")
	}
	if deal.BuyerID != userID && deal.SellerID != userID {
		return nil, errors.New(errors.CodeUnauthorized, "You are not part of this deal")
	}

	d := &Dispute{
		DealID:       req.DealID,
		FiledBy:      userID,
		Reason:       req.Reason,
		EvidenceURLs: req.EvidenceURLs,
	}
	if d.EvidenceURLs == nil {
		d.EvidenceURLs = []string{}
	}
	if err := s.repo.Create(ctx, d); err != nil {
		return nil, err
	}

	s.emit(ctx, contracts.TopicDisputeFiled, d, "")
	// Tell the other side a complaint was raised against the deal.
	other := deal.SellerID
	if userID == deal.SellerID {
		other = deal.BuyerID
	}
	s.notify(ctx, other, "dispute_filed", "A dispute was opened on your deal", "/shop/deals/"+d.DealID)
	return d, nil
}

func (s *service) Get(ctx context.Context, id, userID string, isAdmin bool) (*Dispute, error) {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if isAdmin {
		return d, nil
	}
	// Either party to the deal may read it, not just whoever filed.
	deal, err := s.deals.GetDealBasic(ctx, d.DealID)
	if err != nil || (deal.BuyerID != userID && deal.SellerID != userID) {
		return nil, errors.New(errors.CodeUnauthorized, "You are not part of this dispute")
	}
	return d, nil
}

func (s *service) ListMine(ctx context.Context, userID string) ([]*Dispute, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *service) ListOpen(ctx context.Context) ([]*Dispute, error) {
	return s.repo.ListByStatus(ctx, StatusOpen)
}

// Resolve applies the admin's decision and settles the escrow to match it.
//
// The status is written first and only from 'open', so two admins deciding at
// once can't both move the money: the loser's update affects no rows and stops
// before touching escrow.
func (s *service) Resolve(ctx context.Context, id, adminID string, req ResolveRequest) (*Dispute, error) {
	status, ok := resolutions[req.Resolution]
	if !ok {
		return nil, errors.New(errors.CodeValidation, "Resolution must be 'refund', 'release' or 'reject'")
	}

	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if d.Status != StatusOpen {
		return nil, errors.New(errors.CodeConflict, "This dispute is already resolved")
	}

	applied, err := s.repo.Resolve(ctx, id, status, req.Note, adminID)
	if err != nil {
		return nil, err
	}
	if !applied {
		return nil, errors.New(errors.CodeConflict, "This dispute is already resolved")
	}

	switch status {
	case StatusResolvedRefund:
		s.escrow.RefundDealEscrow(ctx, d.DealID)
	case StatusResolvedRelease:
		s.escrow.ReleaseDealEscrow(ctx, d.DealID)
	}

	updated, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	s.emit(ctx, contracts.TopicDisputeResolved, updated, adminID)

	// Both sides need to know the outcome, whichever way it went.
	if deal, err := s.deals.GetDealBasic(ctx, d.DealID); err == nil {
		msg := "Your dispute was resolved: " + req.Resolution
		s.notify(ctx, deal.BuyerID, "dispute_resolved", msg, "/shop/deals/"+d.DealID)
		s.notify(ctx, deal.SellerID, "dispute_resolved", msg, "/shop/deals/"+d.DealID)
	}
	return updated, nil
}
