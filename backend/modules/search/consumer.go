package search

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"

	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/kafka"
	"github.com/industrix/backend/pkg/logger"
)

// envelope mirrors platform/events.Envelope but keeps the payload raw so each
// topic handler can decode it into its own shape.
type envelope struct {
	Payload json.RawMessage `json:"payload"`
}

// equipmentPayload matches catalog's equipmentEvent JSON.
type equipmentPayload struct {
	ID          string `json:"id"`
	OwnerID     string `json:"owner_id"`
	CategoryID  string `json:"category_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Condition   string `json:"condition"`
	Region      string `json:"region"`
	ImageURL    string `json:"image_url"`
}

// listingPayload matches listing's listingEvent JSON.
type listingPayload struct {
	ID          string  `json:"id"`
	EquipmentID string  `json:"equipment_id"`
	SellerID    string  `json:"seller_id"`
	ListingType string  `json:"listing_type"`
	Price       float64 `json:"price"`
	PricePeriod string  `json:"price_period"`
	Status      string  `json:"status"`
}

// Consumer subscribes to equipment.* and listing.* topics and keeps the search
// index in sync with the source-of-truth modules.
type Consumer struct {
	svc      Service
	consumer *kafka.Consumer
	log      *logger.Logger
}

// NewConsumer builds the Kafka consumer. Returns nil (with a logged warning) if
// the consumer group can't be created, so search still serves reads even when
// the event bus is down.
func NewConsumer(ctx context.Context, svc Service, brokers []string, groupID string) *Consumer {
	log := logger.New("search-consumer")

	c := &Consumer{svc: svc, log: log}
	handlers := map[string]kafka.MessageHandler{
		contracts.TopicEquipmentCreated:   c.onEquipmentUpsert,
		contracts.TopicEquipmentUpdated:   c.onEquipmentUpsert,
		contracts.TopicEquipmentDeleted:   c.onEquipmentDeleted,
		contracts.TopicListingPublished:   c.onListingPublished,
		contracts.TopicListingDeactivated: c.onListingDeactivated,
	}
	topics := make([]string, 0, len(handlers))
	for t := range handlers {
		topics = append(topics, t)
	}

	cfg := &kafka.ConsumerConfig{
		Brokers:       brokers,
		ClientID:      "industrix-search",
		GroupID:       groupID,
		Topics:        topics,
		OffsetInitial: sarama.OffsetOldest,
	}
	kc, err := kafka.NewConsumer(ctx, cfg, handlers)
	if err != nil {
		log.Warn().Err(err).Msg("search Kafka consumer unavailable — index will not auto-sync")
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
		c.log.Error().Err(err).Msg("search consumer stopped")
	}
}

// Close releases the consumer group.
func (c *Consumer) Close() error {
	if c == nil || c.consumer == nil {
		return nil
	}
	return c.consumer.Close()
}

func decode[T any](msg *sarama.ConsumerMessage) (T, error) {
	var env envelope
	var payload T
	if err := json.Unmarshal(msg.Value, &env); err != nil {
		return payload, err
	}
	err := json.Unmarshal(env.Payload, &payload)
	return payload, err
}

func (c *Consumer) onEquipmentUpsert(ctx context.Context, msg *sarama.ConsumerMessage) error {
	p, err := decode[equipmentPayload](msg)
	if err != nil {
		return err
	}
	return c.svc.UpsertEquipment(ctx, Doc{
		EquipmentID: p.ID,
		Title:       p.Title,
		Description: p.Description,
		CategoryID:  p.CategoryID,
		Region:      p.Region,
		Condition:   p.Condition,
		ImageURL:    p.ImageURL,
		SellerID:    p.OwnerID,
	})
}

func (c *Consumer) onEquipmentDeleted(ctx context.Context, msg *sarama.ConsumerMessage) error {
	p, err := decode[equipmentPayload](msg)
	if err != nil {
		return err
	}
	return c.svc.DeleteEquipment(ctx, p.ID)
}

func (c *Consumer) onListingPublished(ctx context.Context, msg *sarama.ConsumerMessage) error {
	p, err := decode[listingPayload](msg)
	if err != nil {
		return err
	}
	return c.svc.SetListingActive(ctx, p.EquipmentID, p.ID, p.ListingType, p.Price, p.PricePeriod)
}

func (c *Consumer) onListingDeactivated(ctx context.Context, msg *sarama.ConsumerMessage) error {
	p, err := decode[listingPayload](msg)
	if err != nil {
		return err
	}
	return c.svc.SetListingInactive(ctx, p.EquipmentID)
}
