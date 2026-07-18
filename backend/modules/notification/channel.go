package notification

import (
	"context"

	"github.com/industrix/backend/pkg/logger"
)

// Dispatch is a single notification to deliver to a user across one or more
// channels.
type Dispatch struct {
	UserID   string   `json:"user_id"`
	Type     string   `json:"type"`
	Message  string   `json:"message"`
	Link     string   `json:"link"`
	Channels []string `json:"channels"` // e.g. ["in_app","email"]; empty => in_app
}

// Channel delivers a notification over one transport (in-app feed, email, SMS,
// push). Delivery failures are returned so the dispatcher can log them, but one
// channel failing must not stop the others.
type Channel interface {
	Name() string
	Send(ctx context.Context, d Dispatch) error
}

// ChannelInApp is the default channel: it persists the notification to the
// user's in-app feed (the notifications table).
const (
	ChannelInApp = "in_app"
	ChannelEmail = "email"
	ChannelSMS   = "sms"
	ChannelPush  = "push"
)

// inAppChannel writes to the notifications table — the existing feed behaviour.
type inAppChannel struct {
	repo *Repository
}

func (c *inAppChannel) Name() string { return ChannelInApp }
func (c *inAppChannel) Send(ctx context.Context, d Dispatch) error {
	return c.repo.Create(ctx, d.UserID, d.Type, d.Message, d.Link)
}

// The email/SMS/push channels are scaffolding: they log the intended delivery
// so the multi-channel fan-out is wired end-to-end, ready for real provider
// integrations (Postal SMTP, Beeline/Kcell SMPP, FCM/APNs) to drop in.

type emailChannel struct{ log *logger.Logger }

func (c *emailChannel) Name() string { return ChannelEmail }
func (c *emailChannel) Send(_ context.Context, d Dispatch) error {
	c.log.Info().Str("user_id", d.UserID).Str("type", d.Type).Msg("email notification (stub)")
	return nil
}

type smsChannel struct{ log *logger.Logger }

func (c *smsChannel) Name() string { return ChannelSMS }
func (c *smsChannel) Send(_ context.Context, d Dispatch) error {
	c.log.Info().Str("user_id", d.UserID).Str("type", d.Type).Msg("SMS notification (stub)")
	return nil
}

type pushChannel struct{ log *logger.Logger }

func (c *pushChannel) Name() string { return ChannelPush }
func (c *pushChannel) Send(_ context.Context, d Dispatch) error {
	c.log.Info().Str("user_id", d.UserID).Str("type", d.Type).Msg("push notification (stub)")
	return nil
}

// defaultChannels builds the standard channel registry keyed by name.
func defaultChannels(repo *Repository) map[string]Channel {
	l := logger.New("notification-channels")
	return map[string]Channel{
		ChannelInApp: &inAppChannel{repo: repo},
		ChannelEmail: &emailChannel{log: l},
		ChannelSMS:   &smsChannel{log: l},
		ChannelPush:  &pushChannel{log: l},
	}
}
