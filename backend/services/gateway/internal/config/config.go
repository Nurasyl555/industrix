package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server    ServerConfig
	Redis     RedisConfig
	Identity  IdentityConfig
	Services  ServicesConfig
	RateLimit RateLimitConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type IdentityConfig struct {
	URL string
}

type ServicesConfig struct {
	CatalogURL             string
	ListingURL             string
	SearchURL              string
	BookingURL             string
	DealURL                string
	PaymentURL             string
	DocumentURL            string
	ReviewURL              string
	ChatURL                string
	NotificationURL        string
	ServicesMarketplaceURL string
	EngagementURL          string
	IntegrityURL           string
	MediaURL               string
	AnalyticsURL           string
	AdminURL               string
}

type RateLimitConfig struct {
	RequestsPerMinute      int
	Burst                  int
	AdminRequestsPerMinute int
	AdminBurst             int
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("GATEWAY_PORT", "8080"),
			ReadTimeout:  getDurationEnv("GATEWAY_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationEnv("GATEWAY_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getDurationEnv("GATEWAY_IDLE_TIMEOUT", 120*time.Second),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getIntEnv("REDIS_DB", 0),
		},
		Identity: IdentityConfig{
			URL: getEnv("IDENTITY_SERVICE_URL", "localhost:8081"),
		},
		Services: ServicesConfig{
			CatalogURL:             getEnv("CATALOG_SERVICE_URL", "localhost:8082"),
			ListingURL:             getEnv("LISTING_SERVICE_URL", "localhost:8083"),
			SearchURL:              getEnv("SEARCH_SERVICE_URL", "localhost:8084"),
			BookingURL:             getEnv("BOOKING_SERVICE_URL", "localhost:8085"),
			DealURL:                getEnv("DEAL_SERVICE_URL", "localhost:8086"),
			PaymentURL:             getEnv("PAYMENT_SERVICE_URL", "localhost:8087"),
			DocumentURL:            getEnv("DOCUMENT_SERVICE_URL", "localhost:8088"),
			ReviewURL:              getEnv("REVIEW_SERVICE_URL", "localhost:8089"),
			ChatURL:                getEnv("CHAT_SERVICE_URL", "localhost:8090"),
			NotificationURL:        getEnv("NOTIFICATION_SERVICE_URL", "localhost:8091"),
			ServicesMarketplaceURL: getEnv("SERVICES_MARKETPLACE_SERVICE_URL", "localhost:8092"),
			EngagementURL:          getEnv("ENGAGEMENT_SERVICE_URL", "localhost:8093"),
			IntegrityURL:           getEnv("INTEGRITY_SERVICE_URL", "localhost:8094"),
			MediaURL:               getEnv("MEDIA_SERVICE_URL", "localhost:8095"),
			AnalyticsURL:           getEnv("ANALYTICS_SERVICE_URL", "localhost:8096"),
			AdminURL:               getEnv("ADMIN_SERVICE_URL", "localhost:8097"),
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute:      getIntEnv("RATE_LIMIT_REQUESTS", 60),
			Burst:                  getIntEnv("RATE_LIMIT_BURST", 10),
			AdminRequestsPerMinute: getIntEnv("RATE_LIMIT_ADMIN_REQUESTS", 120),
			AdminBurst:             getIntEnv("RATE_LIMIT_ADMIN_BURST", 20),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
