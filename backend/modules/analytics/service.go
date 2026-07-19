package analytics

import (
	"context"

	"github.com/industrix/backend/contracts"
)

// Deal-status events are stored as "deal.status.changed:<to>" so the funnel
// stages can be counted separately from one topic.
const (
	evtDealInquiry   = contracts.TopicDealStatusChanged + ":inquiry"
	evtDealCompleted = contracts.TopicDealStatusChanged + ":completed"
	evtDealCancelled = contracts.TopicDealStatusChanged + ":cancelled"
)

// defaultWindowDays is the reporting window when the caller doesn't pick one.
const defaultWindowDays = 30

// Service exposes the dashboards plus the recording entry point used by the
// Kafka consumer.
type Service interface {
	SellerStats(ctx context.Context, sellerID string, days int) (*SellerStats, error)
	AdminStats(ctx context.Context, days int) (*AdminStats, error)

	// Record persists one domain event. Driven by the consumer.
	Record(ctx context.Context, e Event) error
}

type service struct {
	repo *Repository
}

func NewService(repo *Repository) Service {
	return &service{repo: repo}
}

func normalizeDays(days int) int {
	if days < 1 || days > 365 {
		return defaultWindowDays
	}
	return days
}

func (s *service) Record(ctx context.Context, e Event) error {
	return s.repo.Record(ctx, e)
}

func (s *service) SellerStats(ctx context.Context, sellerID string, days int) (*SellerStats, error) {
	days = normalizeDays(days)
	counts, err := s.repo.countsFor(ctx, sellerID, days)
	if err != nil {
		return nil, err
	}
	revenue, err := s.repo.revenue(ctx, sellerID, days)
	if err != nil {
		return nil, err
	}

	st := &SellerStats{
		Days:              days,
		ListingsPublished: counts[contracts.TopicListingPublished],
		Inquiries:         counts[evtDealInquiry],
		DealsCompleted:    counts[evtDealCompleted],
		DealsCancelled:    counts[evtDealCancelled],
		Revenue:           revenue,
	}
	if st.Inquiries > 0 {
		st.ConversionRate = float64(st.DealsCompleted) / float64(st.Inquiries)
	}
	return st, nil
}

func (s *service) AdminStats(ctx context.Context, days int) (*AdminStats, error) {
	days = normalizeDays(days)
	counts, err := s.repo.countsFor(ctx, "", days)
	if err != nil {
		return nil, err
	}
	gmv, err := s.repo.revenue(ctx, "", days)
	if err != nil {
		return nil, err
	}
	sellers, err := s.repo.activeSellers(ctx, days)
	if err != nil {
		return nil, err
	}
	daily, err := s.repo.dailyGMV(ctx, days)
	if err != nil {
		return nil, err
	}

	return &AdminStats{
		Days:              days,
		GMV:               gmv,
		PaymentsCompleted: counts[contracts.TopicPaymentCompleted],
		PaymentsRefunded:  counts[contracts.TopicPaymentRefunded],
		ListingsPublished: counts[contracts.TopicListingPublished],
		Inquiries:         counts[evtDealInquiry],
		DealsCompleted:    counts[evtDealCompleted],
		ActiveSellers:     sellers,
		EventsByType:      counts,
		Daily:             daily,
	}, nil
}
