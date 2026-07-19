package integrity

import (
	"context"
	"time"

	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/errors"
)

// Service defines the integrity service interface
type Service interface {
	CreateCompany(ctx context.Context, company *Company) error
	GetCompany(ctx context.Context, id string) (*Company, error)
	GetMyCompany(ctx context.Context, ownerID string) (*Company, error)
	UpdateCompany(ctx context.Context, company *Company) error

	// Admin
	ListCompaniesByStatus(ctx context.Context, status string) ([]*Company, error)
	SetCompanyStatus(ctx context.Context, id string, status CompanyStatus, note string) error

	// Subscriptions
	GetSubscription(ctx context.Context, userID string) (*Subscription, error)
	ChangePlan(ctx context.Context, userID, plan string) (*Subscription, error)
	// SetCharger closes the billing dependency cycle after construction.
	SetCharger(c contracts.Charger)

	// Contracts
	contracts.CompanyProvider
	contracts.SubscriptionProvider
}

type service struct {
	repo     *Repository
	notifier contracts.Notifier
	events   contracts.EventPublisher
	charger  contracts.Charger
}

// NewService creates a new integrity service
func NewService(repo *Repository, notifier contracts.Notifier, events contracts.EventPublisher) Service {
	return &service{repo: repo, notifier: notifier, events: events}
}

// SetCharger wires the billing dependency after construction. It can't be a
// constructor argument: payment depends on deal, which depends on listing,
// which depends on this module — so the cycle is closed here once both halves
// exist. Until it is set, paid plans are refused rather than granted free.
func (s *service) SetCharger(c contracts.Charger) { s.charger = c }

func (s *service) CreateCompany(ctx context.Context, company *Company) error {
	if len(company.BIN) != 12 {
		return errors.New(errors.CodeValidation, "Invalid BIN format")
	}

	existing, err := s.repo.GetCompanyByBIN(ctx, company.BIN)
	if err == nil && existing != nil {
		return errors.New(errors.CodeConflict, "Company with this BIN already exists")
	}

	return s.repo.CreateCompany(ctx, company)
}

func (s *service) GetCompany(ctx context.Context, id string) (*Company, error) {
	return s.repo.GetCompanyByID(ctx, id)
}

func (s *service) GetMyCompany(ctx context.Context, ownerID string) (*Company, error) {
	return s.repo.GetCompanyByOwner(ctx, ownerID)
}

func (s *service) UpdateCompany(ctx context.Context, company *Company) error {
	return s.repo.UpdateCompany(ctx, company)
}

// === Admin ===

func (s *service) ListCompaniesByStatus(ctx context.Context, status string) ([]*Company, error) {
	return s.repo.ListByStatus(ctx, status)
}

func (s *service) SetCompanyStatus(ctx context.Context, id string, status CompanyStatus, note string) error {
	if status != StatusVerified && status != StatusRejected && status != StatusPending {
		return errors.New(errors.CodeValidation, "Invalid status")
	}
	company, err := s.repo.GetCompanyByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.SetStatus(ctx, id, status, note); err != nil {
		return err
	}
	if s.notifier != nil && company.OwnerID != "" {
		switch status {
		case StatusVerified:
			s.notifier.Notify(ctx, company.OwnerID, "company_verified", "Your company was verified", "/account/company")
		case StatusRejected:
			s.notifier.Notify(ctx, company.OwnerID, "company_rejected", "Your company registration was rejected", "/account/company")
		}
	}
	return nil
}

// === Subscriptions ===

// GetSubscription returns the user's subscription, defaulting to the free plan
// when no row exists. A paid plan whose period has lapsed is reported as
// expired and falls back to free limits until it is renewed.
func (s *service) GetSubscription(ctx context.Context, userID string) (*Subscription, error) {
	plan, status, expiresAt, updatedAt, ok, _ := s.repo.GetSubscriptionRow(ctx, userID)
	if !ok {
		return &Subscription{
			UserID: userID, Plan: PlanFree, Status: "active",
			Price: priceFor(PlanFree), ListingLimit: listingLimitFor(PlanFree),
		}, nil
	}
	if expiresAt != nil && time.Now().After(*expiresAt) {
		// Lapsed: keep the record for history but grant only free entitlements.
		return &Subscription{
			UserID: userID, Plan: plan, Status: "expired",
			Price: priceFor(plan), ListingLimit: listingLimitFor(PlanFree),
			ExpiresAt: expiresAt, UpdatedAt: updatedAt,
		}, nil
	}
	return &Subscription{
		UserID: userID, Plan: plan, Status: status,
		Price: priceFor(plan), ListingLimit: listingLimitFor(plan),
		ExpiresAt: expiresAt, UpdatedAt: updatedAt,
	}, nil
}

// ChangePlan bills the plan's fee and, only if the charge succeeds, switches
// the user over for one billing period. Downgrading to free costs nothing and
// never expires.
func (s *service) ChangePlan(ctx context.Context, userID, plan string) (*Subscription, error) {
	if !validPlans[plan] {
		return nil, errors.New(errors.CodeValidation, "Unknown plan: "+plan)
	}

	price := priceFor(plan)
	var paymentID string
	var expiresAt *time.Time

	if price > 0 {
		if s.charger == nil {
			return nil, errors.New(errors.CodeInternal, "Billing is not available")
		}
		// Charge first: a failed payment must leave the plan untouched.
		id, err := s.charger.Charge(ctx, userID, price, "KZT", "Subscription: "+plan)
		if err != nil {
			return nil, err
		}
		paymentID = id
		until := time.Now().Add(planPeriod)
		expiresAt = &until
	}

	if err := s.repo.UpsertPlan(ctx, userID, plan, expiresAt, paymentID); err != nil {
		return nil, err
	}
	if s.events != nil {
		s.events.Publish(ctx, contracts.TopicSubscriptionActivated, userID, subscriptionEvent{
			UserID: userID, Plan: plan, Price: price, PaymentID: paymentID,
		})
	}
	if s.notifier != nil {
		s.notifier.Notify(ctx, userID, "subscription_changed", "Your plan is now "+plan, "/account/subscription")
	}
	return s.GetSubscription(ctx, userID)
}

// subscriptionEvent is the payload published on subscription.activated.
type subscriptionEvent struct {
	UserID    string  `json:"user_id"`
	Plan      string  `json:"plan"`
	Price     float64 `json:"price"`
	PaymentID string  `json:"payment_id,omitempty"`
}

// === Contracts (SubscriptionProvider) ===

// ListingLimit returns the current plan's listing cap (-1 = unlimited).
func (s *service) ListingLimit(ctx context.Context, userID string) int {
	sub, err := s.GetSubscription(ctx, userID)
	if err != nil {
		return listingLimitFor(PlanFree)
	}
	return sub.ListingLimit
}

// === Contracts (CompanyProvider) ===

func (s *service) GetCompanyBasic(ctx context.Context, companyID string) (*contracts.CompanyBasic, error) {
	c, err := s.repo.GetCompanyByID(ctx, companyID)
	if err != nil {
		return nil, err
	}
	return &contracts.CompanyBasic{
		ID:       c.ID,
		Name:     c.Name,
		Verified: c.Verified,
	}, nil
}
