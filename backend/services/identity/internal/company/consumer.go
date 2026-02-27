package company

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/industrix/pkg/kafka"
)

type Consumer struct {
	service  *Service
	consumer *kafka.Consumer
}

func NewConsumer(service *Service, consumer *kafka.Consumer) *Consumer {
	return &Consumer{service: service, consumer: consumer}
}

func (c *Consumer) Start(ctx context.Context) error {
	c.consumer.AddTopics("review.created")
	c.consumer.Start(ctx)
	return nil
}

func (c *Consumer) handleReviewEvent(ctx context.Context, msg *sarama.ConsumerMessage) error {
	var event map[string]interface{}
	if err := json.Unmarshal(msg.Value, &event); err != nil { return err }
	return c.service.HandleReviewCreated(ctx, event)
}
