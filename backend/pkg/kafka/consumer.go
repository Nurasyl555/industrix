package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/industrix/backend/pkg/logger"
)

type ConsumerConfig struct {
	Brokers        []string
	ClientID       string
	GroupID        string
	Topics         []string
	RetryBackoff   time.Duration
	MinBytes       int
	MaxBytes       int
	MaxWaitTime    time.Duration
	OffsetInitial  int64
	CommitInterval time.Duration
}

type Consumer struct {
	consumer sarama.ConsumerGroup
	config   *ConsumerConfig
	log      *logger.Logger
	handlers map[string]MessageHandler
}

type MessageHandler func(ctx context.Context, msg *sarama.ConsumerMessage) error

func NewConsumer(ctx context.Context, cfg *ConsumerConfig, handlers map[string]MessageHandler) (*Consumer, error) {
	if cfg == nil {
		cfg = &ConsumerConfig{
			Brokers:        []string{"localhost:9092"},
			ClientID:       "industrix-consumer",
			GroupID:        "industrix-group",
			Topics:         []string{},
			RetryBackoff:   time.Second,
			MinBytes:       1,
			MaxBytes:       10 * 1024 * 1024,
			MaxWaitTime:    time.Second,
			OffsetInitial:  sarama.OffsetOldest,
			CommitInterval: time.Second,
		}
	}

	log := logger.New("kafka-consumer")

	// Sanitize zero-valued fields so a caller-supplied config that only sets
	// brokers/group/topics still yields a valid sarama config (Fetch.Min must
	// be > 0, etc.).
	if cfg.MinBytes <= 0 {
		cfg.MinBytes = 1
	}
	if cfg.MaxBytes <= 0 {
		cfg.MaxBytes = 10 * 1024 * 1024
	}
	if cfg.MaxWaitTime <= 0 {
		cfg.MaxWaitTime = time.Second
	}
	if cfg.RetryBackoff <= 0 {
		cfg.RetryBackoff = time.Second
	}
	if cfg.CommitInterval <= 0 {
		cfg.CommitInterval = time.Second
	}

	config := sarama.NewConfig()
	config.ClientID = cfg.ClientID
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = cfg.OffsetInitial
	config.Consumer.Retry.Backoff = cfg.RetryBackoff
	config.Consumer.Fetch.Min = int32(cfg.MinBytes)
	config.Consumer.Fetch.Max = int32(cfg.MaxBytes)
	config.Consumer.MaxWaitTime = cfg.MaxWaitTime
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = cfg.CommitInterval

	consumer, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	log.Info().
		Strs("brokers", cfg.Brokers).
		Str("group", cfg.GroupID).
		Msg("Kafka consumer created")

	return &Consumer{
		consumer: consumer,
		config:   cfg,
		log:      log,
		handlers: handlers,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	if len(c.config.Topics) == 0 {
		return fmt.Errorf("no topics specified")
	}

	consumerHandler := &consumerGroupHandler{
		handlers: c.handlers,
		log:      c.log,
	}

	for {
		if err := c.consumer.Consume(ctx, c.config.Topics, consumerHandler); err != nil {
			c.log.Error().Err(err).Msg("Error consuming")
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func (c *Consumer) Close() error {
	if err := c.consumer.Close(); err != nil {
		c.log.Error().Err(err).Msg("Failed to close consumer")
		return err
	}
	c.log.Info().Msg("Kafka consumer closed")
	return nil
}

type consumerGroupHandler struct {
	handlers map[string]MessageHandler
	log      *logger.Logger
}

func (h *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		topic := msg.Topic
		if handler, ok := h.handlers[topic]; ok {
			ctx := session.Context()
			if err := handler(ctx, msg); err != nil {
				h.log.Error().
					Err(err).
					Str("topic", topic).
					Int32("partition", msg.Partition).
					Int64("offset", msg.Offset).
					Msg("Error handling message")
			}
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
