package notification

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"

	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/kafka"
	"github.com/industrix/backend/pkg/logger"
)

// envelope mirrors platform/events.Envelope with a raw payload.
type envelope struct {
	Payload json.RawMessage `json:"payload"`
}

// Consumer subscribes to notification.dispatch and fans each event out across
// the requested channels. This is the event-driven path that lets any module
// request multi-channel delivery without a direct dependency on notification.
type Consumer struct {
	svc      Service
	consumer *kafka.Consumer
	log      *logger.Logger
}

// NewConsumer builds the Kafka consumer. Returns nil (with a logged warning) if
// the consumer group can't be created, so the in-app feed still works when the
// event bus is down.
func NewConsumer(ctx context.Context, svc Service, brokers []string, groupID string) *Consumer {
	log := logger.New("notification-consumer")

	c := &Consumer{svc: svc, log: log}
	handlers := map[string]kafka.MessageHandler{
		contracts.TopicNotificationDispatch: c.onDispatch,
	}
	cfg := &kafka.ConsumerConfig{
		Brokers:       brokers,
		ClientID:      "industrix-notification",
		GroupID:       groupID,
		Topics:        []string{contracts.TopicNotificationDispatch},
		OffsetInitial: sarama.OffsetNewest,
	}
	kc, err := kafka.NewConsumer(ctx, cfg, handlers)
	if err != nil {
		log.Warn().Err(err).Msg("notification Kafka consumer unavailable — multi-channel dispatch disabled")
		return nil
	}
	c.consumer = kc
	return c
}

// Start blocks consuming until ctx is cancelled; run it in a goroutine.
func (c *Consumer) Start(ctx context.Context) {
	if c == nil || c.consumer == nil {
		return
	}
	if err := c.consumer.Start(ctx); err != nil && ctx.Err() == nil {
		c.log.Error().Err(err).Msg("notification consumer stopped")
	}
}

// Close releases the consumer group.
func (c *Consumer) Close() error {
	if c == nil || c.consumer == nil {
		return nil
	}
	return c.consumer.Close()
}

func (c *Consumer) onDispatch(ctx context.Context, msg *sarama.ConsumerMessage) error {
	var env envelope
	if err := json.Unmarshal(msg.Value, &env); err != nil {
		return err
	}
	var d Dispatch
	if err := json.Unmarshal(env.Payload, &d); err != nil {
		return err
	}
	c.svc.Dispatch(ctx, d)
	return nil
}
