package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/IBM/sarama"
	"github.com/industrix/pkg/logger"
)

type ProducerConfig struct {
	Brokers []string
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func DefaultProducerConfig() *ProducerConfig {
	return &ProducerConfig{
		Brokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
	}
}

type Producer struct {
	producer sarama.SyncProducer
	log      *logger.Logger
}

func NewProducer(ctx context.Context, cfg *ProducerConfig) (*Producer, error) {
	if cfg == nil {
		cfg = DefaultProducerConfig()
	}
	log := logger.New("kafka-producer")
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.RequiredAcks = sarama.WaitForAll
	saramaCfg.Producer.Retry.Max = 5
	saramaCfg.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(cfg.Brokers, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}
	log.Info().Msg("Kafka producer connected")
	return &Producer{producer: producer, log: log}, nil
}

func (p *Producer) Publish(ctx context.Context, topic string, key string, value interface{}) error {
	var valueBytes []byte
	var err error
	switch v := value.(type) {
	case string: valueBytes = []byte(v)
	case []byte: valueBytes = v
	default:
		valueBytes, err = json.Marshal(v)
		if err != nil { return err }
	}
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(valueBytes),
	}
	_, _, err = p.producer.SendMessage(msg)
	return err
}

func (p *Producer) Close() error {
	if p.producer != nil {
		err := p.producer.Close()
		p.log.Info().Msg("Kafka producer closed")
		return err
	}
	return nil
}
