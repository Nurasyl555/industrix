package payment

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

// dealStatusPayload matches deal's dealStatusEvent JSON (id is the deal id).
type dealStatusPayload struct {
	ID string `json:"id"`
	To string `json:"to"`
}

// Consumer subscribes to deal.status.changed and coordinates escrow with the
// deal lifecycle: releasing held funds when a deal completes and refunding them
// when it is cancelled. This closes the escrow loop without deal depending on
// payment.
type Consumer struct {
	svc      Service
	consumer *kafka.Consumer
	log      *logger.Logger
}

// NewConsumer builds the Kafka consumer. Returns nil (with a logged warning) if
// the group can't be created, so the manual release/refund endpoints remain the
// fallback when the event bus is down.
func NewConsumer(ctx context.Context, svc Service, brokers []string, groupID string) *Consumer {
	log := logger.New("payment-consumer")

	c := &Consumer{svc: svc, log: log}
	handlers := map[string]kafka.MessageHandler{
		contracts.TopicDealStatusChanged: c.onDealStatusChanged,
	}
	cfg := &kafka.ConsumerConfig{
		Brokers:       brokers,
		ClientID:      "industrix-payment",
		GroupID:       groupID,
		Topics:        []string{contracts.TopicDealStatusChanged},
		OffsetInitial: sarama.OffsetNewest,
	}
	kc, err := kafka.NewConsumer(ctx, cfg, handlers)
	if err != nil {
		log.Warn().Err(err).Msg("payment Kafka consumer unavailable — escrow will not auto-settle")
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
		c.log.Error().Err(err).Msg("payment consumer stopped")
	}
}

// Close releases the consumer group.
func (c *Consumer) Close() error {
	if c == nil || c.consumer == nil {
		return nil
	}
	return c.consumer.Close()
}

func (c *Consumer) onDealStatusChanged(ctx context.Context, msg *sarama.ConsumerMessage) error {
	var env envelope
	if err := json.Unmarshal(msg.Value, &env); err != nil {
		return err
	}
	var p dealStatusPayload
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		return err
	}
	switch p.To {
	case "completed":
		c.svc.OnDealCompleted(ctx, p.ID)
	case "cancelled":
		c.svc.OnDealCancelled(ctx, p.ID)
	}
	return nil
}
