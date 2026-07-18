package events

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/IBM/sarama"

	"github.com/industrix/backend/contracts"
	"github.com/industrix/backend/pkg/kafka"
	"github.com/industrix/backend/pkg/logger"
)

// Setup builds the process-wide event publisher. It tries to connect a Kafka
// producer; if Kafka is disabled (KAFKA_ENABLED=false) or unreachable, it logs
// a warning and returns a NoopPublisher so the service still starts. The
// returned close func is a no-op for the noop path.
func Setup(ctx context.Context, l *logger.Logger) (contracts.EventPublisher, func() error) {
	if strings.EqualFold(os.Getenv("KAFKA_ENABLED"), "false") {
		l.Warn().Msg("Kafka disabled (KAFKA_ENABLED=false) — domain events will be dropped")
		return NoopPublisher{}, func() error { return nil }
	}

	cfg := kafka.DefaultConfig()
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		cfg.Brokers = strings.Split(brokers, ",")
	}

	pcfg := kafka.ProducerConfig{
		Brokers:       cfg.Brokers,
		ClientID:      "industrix-producer",
		RetryMax:      3,
		RetryBackoff:  time.Second,
		Acks:          sarama.WaitForAll,
		Timeout:       10 * time.Second,
		FlushInterval: time.Millisecond,
	}
	producer, err := kafka.NewProducer(ctx, &pcfg)
	if err != nil {
		l.Warn().Err(err).Msg("Kafka producer unavailable — falling back to no-op event publisher")
		return NoopPublisher{}, func() error { return nil }
	}

	l.Info().Strs("brokers", cfg.Brokers).Msg("Kafka event publisher ready")
	return NewKafkaPublisher(producer), producer.Close
}
