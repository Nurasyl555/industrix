package company

import (
	"context"
	"encoding/json"
	"log"

	"github.com/industrix/pkg/kafka"
)

// Consumer handles Kafka events for company
type Consumer struct {
	service  *Service
	consumer *kafka.Consumer
}

// NewConsumer creates a new company consumer
func NewConsumer(service *Service, consumer *kafka.Consumer) *Consumer {
	return &Consumer{
		service:  service,
		consumer: consumer,
	}
}

// Start starts consuming events
func (c *Consumer) Start(ctx context.Context) error {
	return c.consumer.Consume(ctx, "review.events", c.handleReviewEvent)
}

// handleReviewEvent handles review events
func (c *Consumer) handleReviewEvent(ctx context.Context, msg []byte) error {
	var event map[string]interface{}
	err := json.Unmarshal(msg, &event)
	if err != nil {
		log.Printf("Failed to unmarshal event: %v", err)
		return err
	}

	eventType, ok := event["event_type"].(string)
	if !ok {
		return nil
	}
	switch eventType {
	case "review.created":
		return c.handleReviewCreated(ctx, event)
	}

	return nil
}

// handleReviewCreated handles review.created event
func (c *Consumer) handleReviewCreated(ctx context.Context, event map[string]interface{}) error {
	targetUserID, ok := event["target_user_id"].(string)
	if !ok {
		return nil
	}

	return c.service.HandleReviewCreated(ctx, event)
}
