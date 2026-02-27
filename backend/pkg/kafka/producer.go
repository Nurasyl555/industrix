package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/industrix/backend/pkg/logger"
)

type ProducerConfig struct {
	Brokers       []string
	ClientID      string
	RetryMax      int
	RetryBackoff  time.Duration
	Acks          sarama.RequiredAcks
	Timeout       time.Duration
	FlushInterval time.Duration
}

type Producer struct {
	producer sarama.SyncProducer
	config   *ProducerConfig
	log      *logger.Logger
}

func NewProducer(ctx context.Context, cfg *ProducerConfig) (*Producer, error) {
	if cfg == nil {
		cfg = &ProducerConfig{
			Brokers:       []string{"localhost:9092"},
			ClientID:      "industrix-producer",
			RetryMax:      3,
			RetryBackoff:  time.Second,
			Acks:          sarama.WaitForAll,
			Timeout:       10 * time.Second,
			FlushInterval: time.Millisecond,
		}
	}

	log := logger.New("kafka-producer")

	config := sarama.NewConfig()
	config.ClientID = cfg.ClientID
	config.Producer.RequiredAcks = cfg.Acks
	config.Producer.Timeout = cfg.Timeout
	config.Producer.Flush.Frequency = cfg.FlushInterval
	config.Producer.Retry.Max = cfg.RetryMax
	config.Producer.Retry.Backoff = cfg.RetryBackoff
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	log.Info().
		Strs("brokers", cfg.Brokers).
		Msg("Kafka producer created")

	return &Producer{
		producer: producer,
		config:   cfg,
		log:      log,
	}, nil
}

func (p *Producer) SendMessage(ctx context.Context, topic string, key, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		p.log.Error().
			Err(err).
			Str("topic", topic).
			Msg("Failed to send message")
		return err
	}

	p.log.Debug().
		Str("topic", topic).
		Int32("partition", partition).
		Int64("offset", offset).
		Msg("Message sent")

	return nil
}

func (p *Producer) SendMessages(ctx context.Context, messages []*sarama.ProducerMessage) error {
	err := p.producer.SendMessages(messages)
	if err != nil {
		p.log.Error().
			Err(err).
			Int("count", len(messages)).
			Msg("Failed to send messages")
		return err
	}

	p.log.Debug().
		Int("count", len(messages)).
		Msg("Messages sent")

	return nil
}

func (p *Producer) Close() error {
	if err := p.producer.Close(); err != nil {
		p.log.Error().Err(err).Msg("Failed to close producer")
		return err
	}
	p.log.Info().Msg("Kafka producer closed")
	return nil
}
