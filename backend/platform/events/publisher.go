// Package events provides the platform-level implementation of
// contracts.EventPublisher. It wraps the Kafka producer and adds a standard
// event envelope, so modules only ever deal with the contract interface and
// never import Kafka directly.
package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/industrix/backend/pkg/kafka"
	"github.com/industrix/backend/pkg/logger"
)

// Envelope is the standard shape every domain event carries on the bus.
// Consumers can rely on these fields regardless of the payload type.
type Envelope struct {
	ID         string `json:"id"`          // unique event id (dedup / tracing)
	Type       string `json:"type"`        // equals the topic name
	Key        string `json:"key"`         // partition key (entity id)
	OccurredAt string `json:"occurred_at"` // RFC3339 UTC
	Payload    any    `json:"payload"`     // event-specific body
}

// KafkaPublisher implements contracts.EventPublisher over a Kafka producer.
type KafkaPublisher struct {
	producer *kafka.Producer
	log      *logger.Logger
}

// NewKafkaPublisher wraps a live Kafka producer.
func NewKafkaPublisher(p *kafka.Producer) *KafkaPublisher {
	return &KafkaPublisher{producer: p, log: logger.New("event-publisher")}
}

// Publish wraps payload in an Envelope and sends it to the topic. It is
// fire-and-forget: marshalling or transport failures are logged, never
// returned, so an event-bus outage can't break the core operation that
// produced the event.
func (k *KafkaPublisher) Publish(ctx context.Context, topic, key string, payload any) {
	env := Envelope{
		ID:         uuid.NewString(),
		Type:       topic,
		Key:        key,
		OccurredAt: time.Now().UTC().Format(time.RFC3339),
		Payload:    payload,
	}
	b, err := json.Marshal(env)
	if err != nil {
		k.log.Error().Err(err).Str("topic", topic).Msg("failed to marshal event")
		return
	}
	if err := k.producer.SendMessage(ctx, topic, []byte(key), b); err != nil {
		k.log.Error().Err(err).Str("topic", topic).Str("key", key).Msg("failed to publish event")
	}
}

// NoopPublisher satisfies contracts.EventPublisher but drops every event. It is
// used as a graceful fallback when Kafka is unavailable (e.g. local dev without
// the infra stack), so the service still boots and serves traffic.
type NoopPublisher struct{}

// Publish does nothing.
func (NoopPublisher) Publish(_ context.Context, _, _ string, _ any) {}
