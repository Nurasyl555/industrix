package analytics

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

// The payloads below only pick the fields the dashboards need.

type listingPayload struct {
	ID       string `json:"id"`
	SellerID string `json:"seller_id"`
}

type equipmentPayload struct {
	ID      string `json:"id"`
	OwnerID string `json:"owner_id"`
}

type dealPayload struct {
	ID       string `json:"id"`
	SellerID string `json:"seller_id"`
	To       string `json:"to"`
}

type paymentPayload struct {
	ID      string  `json:"id"`
	PayeeID string  `json:"payee_id"`
	Amount  float64 `json:"amount"`
}

// Consumer feeds the analytics event store from every topic the dashboards
// aggregate over.
type Consumer struct {
	svc      Service
	consumer *kafka.Consumer
	log      *logger.Logger
}

// NewConsumer builds the Kafka consumer. Returns nil (with a logged warning) if
// the group can't be created, so the dashboards still serve existing data.
func NewConsumer(ctx context.Context, svc Service, brokers []string, groupID string) *Consumer {
	log := logger.New("analytics-consumer")

	c := &Consumer{svc: svc, log: log}
	handlers := map[string]kafka.MessageHandler{
		contracts.TopicEquipmentCreated:  c.onEquipment,
		contracts.TopicListingPublished:  c.onListing,
		contracts.TopicDealStatusChanged: c.onDeal,
		contracts.TopicPaymentCompleted:  c.onPayment,
		contracts.TopicPaymentRefunded:   c.onPayment,
	}
	topics := make([]string, 0, len(handlers))
	for t := range handlers {
		topics = append(topics, t)
	}

	cfg := &kafka.ConsumerConfig{
		Brokers:       brokers,
		ClientID:      "industrix-analytics",
		GroupID:       groupID,
		Topics:        topics,
		OffsetInitial: sarama.OffsetOldest, // replayable: rebuild history
	}
	kc, err := kafka.NewConsumer(ctx, cfg, handlers)
	if err != nil {
		log.Warn().Err(err).Msg("analytics Kafka consumer unavailable — dashboards will not update")
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
		c.log.Error().Err(err).Msg("analytics consumer stopped")
	}
}

// Close releases the consumer group.
func (c *Consumer) Close() error {
	if c == nil || c.consumer == nil {
		return nil
	}
	return c.consumer.Close()
}

func payloadOf[T any](msg *sarama.ConsumerMessage) (T, error) {
	var env envelope
	var out T
	if err := json.Unmarshal(msg.Value, &env); err != nil {
		return out, err
	}
	err := json.Unmarshal(env.Payload, &out)
	return out, err
}

func (c *Consumer) onEquipment(ctx context.Context, msg *sarama.ConsumerMessage) error {
	p, err := payloadOf[equipmentPayload](msg)
	if err != nil {
		return err
	}
	return c.svc.Record(ctx, Event{
		EventType: msg.Topic, EntityID: p.ID, SellerID: p.OwnerID,
	})
}

func (c *Consumer) onListing(ctx context.Context, msg *sarama.ConsumerMessage) error {
	p, err := payloadOf[listingPayload](msg)
	if err != nil {
		return err
	}
	return c.svc.Record(ctx, Event{
		EventType: msg.Topic, EntityID: p.ID, SellerID: p.SellerID,
	})
}

// onDeal records the target status in the event type so each funnel stage can
// be counted separately (deal.status.changed:inquiry, :completed, ...).
func (c *Consumer) onDeal(ctx context.Context, msg *sarama.ConsumerMessage) error {
	p, err := payloadOf[dealPayload](msg)
	if err != nil {
		return err
	}
	return c.svc.Record(ctx, Event{
		EventType: msg.Topic + ":" + p.To, EntityID: p.ID, SellerID: p.SellerID,
	})
}

func (c *Consumer) onPayment(ctx context.Context, msg *sarama.ConsumerMessage) error {
	p, err := payloadOf[paymentPayload](msg)
	if err != nil {
		return err
	}
	return c.svc.Record(ctx, Event{
		EventType: msg.Topic, EntityID: p.ID, SellerID: p.PayeeID, Amount: p.Amount,
	})
}
