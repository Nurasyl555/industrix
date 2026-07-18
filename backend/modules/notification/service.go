package notification

import (
	"context"

	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/logger"
)

// Service is the notification feed service. It implements contracts.Notifier
// so other modules can emit events without importing this package.
type Service interface {
	contracts.Notifier

	// Dispatch delivers a notification across the requested channels (defaults
	// to in-app). Used by the Kafka consumer for multi-channel fan-out.
	Dispatch(ctx context.Context, d Dispatch)

	List(ctx context.Context, userID string) ([]*Notification, error)
	UnreadCount(ctx context.Context, userID string) (int, error)
	MarkRead(ctx context.Context, id, userID string) error
	MarkAllRead(ctx context.Context, userID string) error
}

type service struct {
	repo     *Repository
	channels map[string]Channel
	log      *logger.Logger
}

func NewService(repo *Repository) Service {
	return &service{
		repo:     repo,
		channels: defaultChannels(repo),
		log:      logger.New("notification-service"),
	}
}

// Notify is fire-and-forget: it delivers to the in-app feed only. A failure is
// logged but never propagated, so it can't break the operation that triggered
// it. For multi-channel delivery, publish to notification.dispatch instead.
func (s *service) Notify(ctx context.Context, userID, ntype, message, link string) {
	if userID == "" {
		return
	}
	s.Dispatch(ctx, Dispatch{UserID: userID, Type: ntype, Message: message, Link: link})
}

// Dispatch fans a notification out to each requested channel. Unknown channels
// are skipped; per-channel failures are logged but never stop the others.
func (s *service) Dispatch(ctx context.Context, d Dispatch) {
	if d.UserID == "" {
		return
	}
	channels := d.Channels
	if len(channels) == 0 {
		channels = []string{ChannelInApp}
	}
	for _, name := range channels {
		ch, ok := s.channels[name]
		if !ok {
			s.log.Warn().Str("channel", name).Msg("unknown notification channel — skipped")
			continue
		}
		if err := ch.Send(ctx, d); err != nil {
			s.log.Error().Err(err).Str("channel", name).Str("user_id", d.UserID).Msg("channel delivery failed")
		}
	}
}

func (s *service) List(ctx context.Context, userID string) ([]*Notification, error) {
	return s.repo.List(ctx, userID, 50)
}

func (s *service) UnreadCount(ctx context.Context, userID string) (int, error) {
	return s.repo.UnreadCount(ctx, userID)
}

func (s *service) MarkRead(ctx context.Context, id, userID string) error {
	return s.repo.MarkRead(ctx, id, userID)
}

func (s *service) MarkAllRead(ctx context.Context, userID string) error {
	return s.repo.MarkAllRead(ctx, userID)
}
