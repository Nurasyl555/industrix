package kafka

import (
	"os"
	"strings"
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
