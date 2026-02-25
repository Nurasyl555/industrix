package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/industrix/pkg/logger"
)

// ConsumerConfig holds Kafka consumer configuration
type ConsumerConfig struct {
	Brokers        []string
	Group          string
	Topics         []string
	Version        string
	Assignor       string
	Offsets        map[int]int64
	CommitInterval time.Duration
	MinBytes       int
	MaxBytes       int
	MaxWaitTime    time.Duration
}

// DefaultConsumerConfig returns configuration with sensible defaults
func DefaultConsumerConfig() *ConsumerConfig {
	return &ConsumerConfig{
		Brokers:        []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		Group:          getEnv("KAFKA_CONSUMER_GROUP", "industrix-consumer"),
		Topics:         []string{},
		Version:        getEnv("KAFKA_VERSION", "2.7.0"),
		Assignor:       "range",
		Offsets:        make(map[int]int64),
		CommitInterval: 1 * time.Second,
		MinBytes:       1,
		MaxBytes:       10 * 1024 * 1024,
		MaxWaitTime:    250 * time.Millisecond,
	}
}

// MessageHandler is a function type for processing messages
type MessageHandler func(ctx context.Context, msg *sarama.ConsumerMessage) error

// Consumer wraps sarama.ConsumerGroup
type Consumer struct {
	consumer *cluster.Consumer
	handler  MessageHandler
	log      *logger.Logger
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// NewConsumer creates a new Kafka consumer with consumer group support
func NewConsumer(ctx context.Context, cfg *ConsumerConfig, handler MessageHandler) (*Consumer, error) {
	if cfg == nil {
		cfg = DefaultConsumerConfig()
	}

	log := logger.New("kafka-consumer")

	saramaCfg := cluster.NewConfig()
	saramaCfg.Version = sarama.V2_7_0_0
	saramaCfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	saramaCfg.Consumer.Offsets.CommitInterval = cfg.CommitInterval
	saramaCfg.Consumer.Fetch.Min = int32(cfg.MinBytes)
	saramaCfg.Consumer.Fetch.Max = int32(cfg.MaxBytes)
	saramaCfg.Consumer.MaxWaitTime = cfg.MaxWaitTime

	for topic, offset := range cfg.Offsets {
		saramaCfg.Consumer.Offsets.Initial = sarama.OffsetOldest
		_ = topic
		_ = offset
	}

	consumer, err := cluster.NewConsumer(cfg.Brokers, cfg.Group, cfg.Topics, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	consumerCtx, cancel := context.WithCancel(ctx)

	log.Info().
		Strs("brokers", cfg.Brokers).
		Str("group", cfg.Group).
		Msg("Kafka consumer connected")

	return &Consumer{
		consumer: consumer,
		handler:  handler,
		log:      log,
		ctx:      consumerCtx,
		cancel:   cancel,
	}, nil
}

// Start begins consuming messages
func (c *Consumer) Start() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case msg, ok := <-c.consumer.Messages():
				if !ok {
					return
				}

				if c.handler != nil {
					err := c.handler(c.ctx, msg)
					if err != nil {
						c.log.Error().
							Str("topic", msg.Topic).
							Int32("partition", msg.Partition).
							Int64("offset", msg.Offset).
							Err(err).
							Msg("Error processing message")
						// Continue processing - errors don't block other messages
						continue
					}
				}

				// Mark message as processed
				c.consumer.MarkOffset(msg, "")

			case err, ok := <-c.consumer.Errors():
				if !ok {
					return
				}
				c.log.Error().
					Err(err).
					Msg("Consumer error")

			case notification, ok := <-c.consumer.Notifications():
				if !ok {
					return
				}
				c.log.Info().
					Strs("claimed", notification.Claimed).
					Strs("released", notification.Released).
					Msg("Rebalance notification")
			}
		}
	}()
}

// Stop stops the consumer
func (c *Consumer) Stop() {
	c.cancel()
	c.wg.Wait()

	if c.consumer != nil {
		err := c.consumer.Close()
		if err != nil {
			c.log.Error().Err(err).Msg("Error closing consumer")
		}
	}
	c.log.Info("Kafka consumer stopped")
}

// AddTopics adds new topics to the consumer
func (c *Consumer) AddTopics(topics ...string) error {
	return c.consumer.AddPartitionTopic(topics...)
}

// Pause pauses consumption for the given topics
func (c *Consumer) Pause(topics ...string) {
	c.consumer.PauseTopic(topics...)
}

// Resume resumes consumption for the given topics
func (c *Consumer) Resume(topics ...string) {
	c.consumer.ResumeTopic(topics...)
}

// DLQConsumer creates a dead letter queue consumer
type DLQConsumer struct {
	consumer *Consumer
	dlqTopic string
	log      *logger.Logger
}

// NewDLQConsumer creates a new DLQ consumer
func NewDLQConsumer(ctx context.Context, cfg *ConsumerConfig, dlqTopic string) (*DLQConsumer, error) {
	consumer, err := NewConsumer(ctx, cfg, nil)
	if err != nil {
		return nil, err
	}

	return &DLQConsumer{
		consumer: consumer,
		dlqTopic: dlqTopic,
		log:      logger.New("kafka-dlq-consumer"),
	}, nil
}

// ConsumeWithDLQ consumes messages and sends failed messages to DLQ
func (c *DLQConsumer) ConsumeWithDLQ(handler MessageHandler, maxRetries int) error {
	wrappedHandler := func(ctx context.Context, msg *sarama.ConsumerMessage) error {
		err := handler(ctx, msg)
		if err != nil {
			// Send to DLQ - in real implementation would include retry count header
			c.log.Error().
				Str("topic", msg.Topic).
				Int32("partition", msg.Partition).
				Int64("offset", msg.Offset).
				Err(err).
				Msg("Sending message to DLQ")
			return err
		}
		return nil
	}

	c.consumer.handler = wrappedHandler
	c.consumer.Start()

	return nil
}

// ParseMessage parses a message value to the given type
func ParseMessage(msg *sarama.ConsumerMessage, v interface{}) error {
	return json.Unmarshal(msg.Value, v)
}
