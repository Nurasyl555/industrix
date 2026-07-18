package payment

import (
	"context"

	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/errors"
)

// Service defines the payment/escrow service interface.
type Service interface {
	// InitEscrow funds a deal into escrow (buyer only). Holds the money with
	// the provider and records a held payment.
	InitEscrow(ctx context.Context, payerID string, req CreatePaymentRequest) (*Payment, error)
	// Release pays a held escrow out to the seller (buyer confirms delivery).
	Release(ctx context.Context, id, userID string) (*Payment, error)
	// Refund returns a held escrow to the buyer (deal fell through).
	Refund(ctx context.Context, id, userID string) (*Payment, error)
	Get(ctx context.Context, id, userID string) (*Payment, error)
	ListMine(ctx context.Context, userID string) ([]*Payment, error)
}

type service struct {
	repo     *Repository
	deals    contracts.DealProvider
	provider Provider
	events   contracts.EventPublisher
	notifier contracts.Notifier
}

func NewService(repo *Repository, deals contracts.DealProvider, provider Provider, events contracts.EventPublisher, notifier contracts.Notifier) Service {
	return &service{repo: repo, deals: deals, provider: provider, events: events, notifier: notifier}
}

func (s *service) emit(ctx context.Context, topic string, p *Payment) {
	if s.events != nil {
		s.events.Publish(ctx, topic, p.ID, toEvent(p))
	}
}

func (s *service) notify(ctx context.Context, userID, ntype, msg string) {
	if s.notifier != nil && userID != "" {
		s.notifier.Notify(ctx, userID, ntype, msg, "")
	}
}

func (s *service) InitEscrow(ctx context.Context, payerID string, req CreatePaymentRequest) (*Payment, error) {
	if req.Amount <= 0 {
		return nil, errors.New(errors.CodeValidation, "Amount must be greater than 0")
	}
	d, err := s.deals.GetDealBasic(ctx, req.DealID)
	if err != nil {
		return nil, errors.New(errors.CodeValidation, "Deal does not exist")
	}
	if d.BuyerID != payerID {
		return nil, errors.New(errors.CodeUnauthorized, "Only the buyer can fund this deal")
	}
	if d.Status == "cancelled" {
		return nil, errors.New(errors.CodeValidation, "Cannot pay for a cancelled deal")
	}

	currency := req.Currency
	if currency == "" {
		currency = "KZT"
	}

	p := &Payment{
		DealID:   req.DealID,
		PayerID:  payerID,
		PayeeID:  d.SellerID,
		Amount:   req.Amount,
		Currency: currency,
		Provider: s.provider.Name(),
		Status:   StatusPending,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	ref, err := s.provider.Hold(ctx, p.Amount, p.Currency)
	if err != nil {
		p.Status = StatusFailed
		_ = s.repo.UpdateStatus(ctx, p.ID, StatusFailed, "")
		s.emit(ctx, contracts.TopicPaymentFailed, p)
		return nil, errors.New(errors.CodeInternal, "Payment hold failed")
	}

	p.Status = StatusHeld
	p.ProviderRef = ref
	if err := s.repo.UpdateStatus(ctx, p.ID, StatusHeld, ref); err != nil {
		return nil, err
	}
	s.notify(ctx, p.PayeeID, "payment_held", "A buyer has funded a deal — payment is held in escrow")
	return p, nil
}

func (s *service) Release(ctx context.Context, id, userID string) (*Payment, error) {
	p, err := s.requireParticipant(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	// Only the payer (buyer) confirms release of funds to the seller.
	if p.PayerID != userID {
		return nil, errors.New(errors.CodeUnauthorized, "Only the payer can release escrow")
	}
	if p.Status != StatusHeld {
		return nil, errors.New(errors.CodeConflict, "Only a held payment can be released")
	}
	if err := s.provider.Release(ctx, p.ProviderRef); err != nil {
		return nil, errors.New(errors.CodeInternal, "Payment release failed")
	}
	p.Status = StatusReleased
	if err := s.repo.UpdateStatus(ctx, p.ID, StatusReleased, p.ProviderRef); err != nil {
		return nil, err
	}
	s.emit(ctx, contracts.TopicPaymentCompleted, p)
	s.notify(ctx, p.PayeeID, "payment_released", "Escrow released — funds are on their way to you")
	return p, nil
}

func (s *service) Refund(ctx context.Context, id, userID string) (*Payment, error) {
	p, err := s.requireParticipant(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if p.Status != StatusHeld {
		return nil, errors.New(errors.CodeConflict, "Only a held payment can be refunded")
	}
	if err := s.provider.Refund(ctx, p.ProviderRef); err != nil {
		return nil, errors.New(errors.CodeInternal, "Payment refund failed")
	}
	p.Status = StatusRefunded
	if err := s.repo.UpdateStatus(ctx, p.ID, StatusRefunded, p.ProviderRef); err != nil {
		return nil, err
	}
	s.emit(ctx, contracts.TopicPaymentRefunded, p)
	s.notify(ctx, p.PayerID, "payment_refunded", "Your escrow payment has been refunded")
	return p, nil
}

func (s *service) Get(ctx context.Context, id, userID string) (*Payment, error) {
	return s.requireParticipant(ctx, id, userID)
}

func (s *service) ListMine(ctx context.Context, userID string) ([]*Payment, error) {
	return s.repo.ListForUser(ctx, userID)
}

// requireParticipant loads a payment and asserts the user is payer or payee.
func (s *service) requireParticipant(ctx context.Context, id, userID string) (*Payment, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.PayerID != userID && p.PayeeID != userID {
		return nil, errors.New(errors.CodeUnauthorized, "You are not part of this payment")
	}
	return p, nil
}
