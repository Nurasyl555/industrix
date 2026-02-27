package kafka

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/IBM/sarama"
	"github.com/industrix/pkg/logger"
)

type Config struct {
	Brokers []string
	GroupID string
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func DefaultConfig() *Config {
	brokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	return &Config{
		Brokers: strings.Split(brokers, ","),
		GroupID: "default-group",
	}
}

type Producer interface {
	SendMessage(ctx context.Context, topic string, key []byte, value []byte) error
	Close() error
}

type producer struct {
	syncProducer sarama.SyncProducer
	log          *logger.Logger
}

func NewProducer(cfg *Config) (Producer, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	log := logger.New("kafka-producer")
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = 5

	p, err := sarama.NewSyncProducer(cfg.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	log.Info().Strs("brokers", cfg.Brokers).Msg("Kafka producer connected")
	return &producer{syncProducer: p, log: log}, nil
}

func (p *producer) SendMessage(ctx context.Context, topic string, key []byte, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	partition, offset, err := p.syncProducer.SendMessage(msg)
	if err != nil {
		p.log.Error().Err(err).Str("topic", topic).Msg("Failed to send message")
		return err
	}

	p.log.Debug().
		Str("topic", topic).
		Int32("partition", partition).
		Int64("offset", offset).
		Msg("Message sent")

	return nil
}

func (p *producer) Close() error {
	if p.syncProducer != nil {
		return p.syncProducer.Close()
	}
	return nil
}

type MessageHandler func(ctx context.Context, key, value []byte) error

type Consumer interface {
	Subscribe(topics []string, handler MessageHandler) error
	Close() error
}

type consumer struct {
	group sarama.ConsumerGroup
	log   *logger.Logger
}

func NewConsumer(cfg *Config) (Consumer, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	log := logger.New("kafka-consumer")
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	group, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer group: %w", err)
	}

	log.Info().
		Strs("brokers", cfg.Brokers).
		Str("group_id", cfg.GroupID).
		Msg("Kafka consumer group connected")

	return &consumer{group: group, log: log}, nil
}

type consumerGroupHandler struct {
	handler MessageHandler
	log     *logger.Logger
}

func (h *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.log.Debug().
			Str("topic", msg.Topic).
			Int32("partition", msg.Partition).
			Int64("offset", msg.Offset).
			Msg("Processing message")

		if err := h.handler(sess.Context(), msg.Key, msg.Value); err != nil {
			h.log.Error().Err(err).Msg("Handler failed")
			// Depending on requirements, we might want to not mark message as consumed
			// But for now, we just log and continue
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}

func (c *consumer) Subscribe(topics []string, handler MessageHandler) error {
	ctx := context.Background()
	go func() {
		for {
			if err := c.group.Consume(ctx, topics, &consumerGroupHandler{handler: handler, log: c.log}); err != nil {
				c.log.Error().Err(err).Msg("Error from consumer")
			}
			// Should exit on context cancellation
			if ctx.Err() != nil {
				return
			}
		}
	}()
	return nil
}

func (c *consumer) Close() error {
	if c.group != nil {
		return c.group.Close()
	}
	return nil
}
