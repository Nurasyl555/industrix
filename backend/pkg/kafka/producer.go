package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/industrix/pkg/logger"
)

// Config holds Kafka producer configuration
type ProducerConfig struct {
	Brokers      []string
	Version      string
	Partitions   int
	Replication  int
	Ack          string
	Retries      int
	RetryMax     int
	BatchSize    int
	LingerMs     int
	Compression  string
	MaxOpenReqs  int
	WriteTimeout time.Duration
}

// getEnv returns environment variable or default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// DefaultProducerConfig returns configuration with sensible defaults
func DefaultProducerConfig() *ProducerConfig {
	return &ProducerConfig{
		Brokers:     []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		Version:     getEnv("KAFKA_VERSION", "2.7.0"),
		Ack:         "all",
		Retries:     3,
		RetryMax:    3,
		BatchSize:   16384,
		LingerMs:    10,
		Compression: "lz4",
		MaxOpenReqs: 5,
	}
}

// Producer wraps sarama.SyncProducer
type Producer struct {
	producer sarama.SyncProducer
	log      *logger.Logger
	mu       sync.RWMutex
}

// NewProducer creates a new Kafka producer
func NewProducer(ctx context.Context, cfg *ProducerConfig) (*Producer, error) {
	if cfg == nil {
		cfg = DefaultProducerConfig()
	}

	log := logger.New("kafka-producer")

	saramaCfg := sarama.NewConfig()
	saramaCfg.Version = sarama.V2_7_0_0
	saramaCfg.Producer.RequiredAcks = sarama.WaitForAll
	saramaCfg.Producer.Retry.Max = cfg.Retries
	saramaCfg.Producer.Retry.Backoff = 100 * time.Millisecond
	saramaCfg.Producer.Batch.Size = cfg.BatchSize
	saramaCfg.Producer.Batch.Linger = time.Duration(cfg.LingerMs) * time.Millisecond
	saramaCfg.Producer.Compression = getCompressionCodec(cfg.Compression)
	saramaCfg.Producer.MaxOpenRequests = cfg.MaxOpenReqs
	saramaCfg.Producer.Return.Successes = true
	saramaCfg.Producer.Return.Errors = true

	producer, err := sarama.NewSyncProducer(cfg.Brokers, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	log.Info().
		Strs("brokers", cfg.Brokers).
		Msg("Kafka producer connected")

	return &Producer{
		producer: producer,
		log:      log,
	}, nil
}

// getCompressionCodec returns the appropriate compression codec
func getCompressionCodec(compression string) sarama.CompressionCodec {
	switch compression {
	case "snappy":
		return sarama.CompressionSnappy
	case "gzip":
		return sarama.CompressionGZIP
	case "lz4":
		return sarama.CompressionLZ4
	case "zstd":
		return sarama.CompressionZSTD
	default:
		return sarama.CompressionNone
	}
}

// SendMessage sends a message to a topic
func (p *Producer) SendMessage(ctx context.Context, topic string, key string, value interface{}) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var keyBytes []byte
	var valueBytes []byte
	var err error

	// Serialize key
	if key != "" {
		keyBytes = []byte(key)
	}

	// Serialize value
	switch v := value.(type) {
	case string:
		valueBytes = []byte(v)
	case []byte:
		valueBytes = v
	default:
		valueBytes, err = json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal message value: %w", err)
		}
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(keyBytes),
		Value: sarama.ByteEncoder(valueBytes),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		p.log.Error().
			Str("topic", topic).
			Err(err).
			Msg("Failed to send message")
		return fmt.Errorf("failed to send message: %w", err)
	}

	p.log.Debug().
		Str("topic", topic).
		Int32("partition", partition).
		Int64("offset", offset).
		Msg("Message sent successfully")

	return nil
}

// SendMessageWithHeaders sends a message with headers
func (p *Producer) SendMessageWithHeaders(ctx context.Context, topic string, key string, value interface{}, headers map[string]string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var keyBytes []byte
	var valueBytes []byte
	var err error

	if key != "" {
		keyBytes = []byte(key)
	}

	switch v := value.(type) {
	case string:
		valueBytes = []byte(v)
	case []byte:
		valueBytes = v
	default:
		valueBytes, err = json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal message value: %w", err)
		}
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(keyBytes),
		Value: sarama.ByteEncoder(valueBytes),
	}

	for k, v := range headers {
		msg.Headers = append(msg.Headers, sarama.RecordHeader{
			Key:   []byte(k),
			Value: []byte(v),
		})
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		p.log.Error().
			Str("topic", topic).
			Err(err).
			Msg("Failed to send message with headers")
		return fmt.Errorf("failed to send message: %w", err)
	}

	p.log.Debug().
		Str("topic", topic).
		Int32("partition", partition).
		Int64("offset", offset).
		Msg("Message with headers sent successfully")

	return nil
}

// Close closes the producer
func (p *Producer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.producer != nil {
		err := p.producer.AsyncClose()
		if err != nil {
			return err
		}
		p.log.Info("Kafka producer closed")
	}
	return nil
}
