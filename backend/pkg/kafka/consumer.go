package kafka

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"github.com/industrix/pkg/logger"
)

type ConsumerConfig struct {
	Brokers []string
	Group   string
	Topics  []string
}

func DefaultConsumerConfig() *ConsumerConfig {
	return &ConsumerConfig{
		Brokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		Group:   "industrix-group",
	}
}

type MessageHandler func(ctx context.Context, msg *sarama.ConsumerMessage) error

type Consumer struct {
	group   sarama.ConsumerGroup
	handler MessageHandler
	topics  []string
	log     *logger.Logger
}

func NewConsumer(ctx context.Context, cfg *ConsumerConfig, handler MessageHandler) (*Consumer, error) {
	if cfg == nil { cfg = DefaultConsumerConfig() }
	log := logger.New("kafka-consumer")
	saramaCfg := sarama.NewConfig()
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	group, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.Group, saramaCfg)
	if err != nil { return nil, err }
	return &Consumer{group: group, handler: handler, topics: cfg.Topics, log: log}, nil
}

type groupHandler struct {
	handler MessageHandler
	log     *logger.Logger
}
func (h *groupHandler) Setup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *groupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *groupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		if h.handler != nil {
			if err := h.handler(sess.Context(), msg); err != nil {
				h.log.Error().Err(err).Msg("Error processing message")
			}
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}

func (c *Consumer) Start(ctx context.Context) {
	go func() {
		h := &groupHandler{handler: c.handler, log: c.log}
		for {
			if err := c.group.Consume(ctx, c.topics, h); err != nil {
				if ctx.Err() != nil { return }
				time.Sleep(time.Second)
			}
		}
	}()
}

func (c *Consumer) AddTopics(topics ...string) { c.topics = append(c.topics, topics...) }
func (c *Consumer) Close() error { return c.group.Close() }
