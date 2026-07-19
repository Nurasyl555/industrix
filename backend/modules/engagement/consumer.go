package engagement

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

// priceChangedPayload matches listing's priceChangedEvent JSON.
type priceChangedPayload struct {
	ID       string  `json:"id"`
	OldPrice float64 `json:"old_price"`
	NewPrice float64 `json:"new_price"`
}

// Consumer subscribes to listing.price_changed and drives price-history
// recording plus price-drop alerts to watchers.
type Consumer struct {
	svc      Service
	consumer *kafka.Consumer
	log      *logger.Logger
}

// NewConsumer builds the Kafka consumer. Returns nil (with a logged warning) if
// the group can't be created, so favorites still work when the bus is down.
func NewConsumer(ctx context.Context, svc Service, brokers []string, groupID string) *Consumer {
	log := logger.New("engagement-consumer")

	c := &Consumer{svc: svc, log: log}
	handlers := map[string]kafka.MessageHandler{
		contracts.TopicListingPriceChanged: c.onPriceChanged,
	}
	cfg := &kafka.ConsumerConfig{
		Brokers:       brokers,
		ClientID:      "industrix-engagement",
		GroupID:       groupID,
		Topics:        []string{contracts.TopicListingPriceChanged},
		OffsetInitial: sarama.OffsetNewest,
	}
	kc, err := kafka.NewConsumer(ctx, cfg, handlers)
	if err != nil {
		log.Warn().Err(err).Msg("engagement Kafka consumer unavailable — price alerts disabled")
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
		c.log.Error().Err(err).Msg("engagement consumer stopped")
	}
}

// Close releases the consumer group.
func (c *Consumer) Close() error {
	if c == nil || c.consumer == nil {
		return nil
	}
	return c.consumer.Close()
}

func (c *Consumer) onPriceChanged(ctx context.Context, msg *sarama.ConsumerMessage) error {
	var env envelope
	if err := json.Unmarshal(msg.Value, &env); err != nil {
		return err
	}
	var p priceChangedPayload
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		return err
	}
	c.svc.OnPriceChanged(ctx, p.ID, p.OldPrice, p.NewPrice)
	return nil
}
