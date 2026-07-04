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

	List(ctx context.Context, userID string) ([]*Notification, error)
	UnreadCount(ctx context.Context, userID string) (int, error)
	MarkRead(ctx context.Context, id, userID string) error
	MarkAllRead(ctx context.Context, userID string) error
}

type service struct {
	repo *Repository
	log  *logger.Logger
}

func NewService(repo *Repository) Service {
	return &service{repo: repo, log: logger.New("notification-service")}
}

// Notify is fire-and-forget: a failure to record a notification is logged but
// never propagated, so it can't break the operation that triggered it.
func (s *service) Notify(ctx context.Context, userID, ntype, message, link string) {
	if userID == "" {
		return
	}
	if err := s.repo.Create(ctx, userID, ntype, message, link); err != nil {
		s.log.Error().Err(err).Str("user_id", userID).Str("type", ntype).Msg("failed to record notification")
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
