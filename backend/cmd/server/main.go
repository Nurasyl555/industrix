package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"

	_ "github.com/industrix/backend/docs"

	"github.com/industrix/backend/modules/admin"
	"github.com/industrix/backend/modules/booking"
	"github.com/industrix/backend/modules/catalog"
	"github.com/industrix/backend/modules/deal"
	"github.com/industrix/backend/modules/identity"
	"github.com/industrix/backend/modules/integrity"
	"github.com/industrix/backend/modules/listing"
	"github.com/industrix/backend/modules/marketplace"
	"github.com/industrix/backend/modules/media"
	"github.com/industrix/backend/modules/notification"
	"github.com/industrix/backend/modules/payment"
	"github.com/industrix/backend/modules/search"
	"github.com/industrix/backend/pkg/jwt"
	"github.com/industrix/backend/pkg/logger"
	"github.com/industrix/backend/pkg/postgres"
	"github.com/industrix/backend/pkg/redis"
	"github.com/industrix/backend/platform/events"
	mw "github.com/industrix/backend/platform/middleware"
)

// @title Industrix API
// @version 1.0
// @description Industrix Industrial Equipment Marketplace API
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	l := logger.New("industrix")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// === Infrastructure ===

	// PostgreSQL
	pgCfg := postgres.DefaultConfig()
	pgCfg.DSN = os.Getenv("DB_DSN")
	pgClient, err := postgres.NewClient(ctx, pgCfg)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}
	defer pgClient.Close()

	// Run migrations
	if err := pgClient.RunMigrations("migrations"); err != nil {
		l.Fatal().Err(err).Msg("Failed to run migrations")
	}

	// Redis
	redisCfg := redis.DefaultConfig()
	redisClient, err := redis.NewClient(ctx, redisCfg)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisClient.Close()

	// JWT
	privateKey, publicKey := loadRSAKeys(l)
	jwtClient := jwt.NewClient(privateKey, publicKey)

	// Event bus — Kafka producer with graceful no-op fallback when unavailable.
	publisher, closePublisher := events.Setup(ctx, l)
	defer func() {
		if err := closePublisher(); err != nil {
			l.Error().Err(err).Msg("Failed to close event publisher")
		}
	}()

	// === Modules ===

	// Notification service is created first — other modules receive it as a
	// contracts.Notifier to emit user-facing events. Its Kafka consumer powers
	// multi-channel dispatch (notification.dispatch topic).
	notificationMod := notification.NewModule(ctx, pgClient,
		splitAndTrim(getEnv("KAFKA_BROKERS", "localhost:9092")),
		"industrix-notification")
	notifier := notificationMod.Service
	if notificationMod.Consumer != nil {
		go notificationMod.Consumer.Start(ctx)
		defer notificationMod.Consumer.Close()
	}

	identityMod := identity.NewModule(pgClient, redisClient, jwtClient)
	integrityMod := integrity.NewModule(pgClient, notifier)
	marketplaceMod := marketplace.NewModule(pgClient)
	catalogMod := catalog.NewModule(pgClient, publisher)
	listingMod := listing.NewModule(pgClient, catalogMod.Service, notifier, publisher, integrityMod.Service)
	dealMod := deal.NewModule(pgClient, listingMod.Service, jwtClient, notifier, publisher)
	bookingMod := booking.NewModule(pgClient, listingMod.Service, notifier)
	paymentMod := payment.NewModule(ctx, pgClient, dealMod.Service, publisher, notifier,
		splitAndTrim(getEnv("KAFKA_BROKERS", "localhost:9092")))
	if paymentMod.Consumer != nil {
		go paymentMod.Consumer.Start(ctx)
		defer paymentMod.Consumer.Close()
	}
	adminMod := admin.NewModule(integrityMod.Service, listingMod.Service)

	// Search — OpenSearch-backed, kept in sync via Kafka consumer.
	searchMod := search.NewModule(ctx, search.Config{
		OpenSearchHosts:    getEnv("OPENSEARCH_HOSTS", "http://localhost:9200"),
		OpenSearchUser:     os.Getenv("OPENSEARCH_USERNAME"),
		OpenSearchPassword: os.Getenv("OPENSEARCH_PASSWORD"),
		Brokers:            splitAndTrim(getEnv("KAFKA_BROKERS", "localhost:9092")),
		ConsumerGroup:      "industrix-search",
	}, redisClient)
	if searchMod.Consumer != nil {
		go searchMod.Consumer.Start(ctx)
		defer searchMod.Consumer.Close()
	}

	mediaMod, err := media.NewModule(media.Config{
		InternalEndpoint: getEnv("MINIO_ENDPOINT", "minio:9000"),
		PublicEndpoint:   getEnv("MINIO_PUBLIC_ENDPOINT", "localhost:9000"),
		AccessKey:        getEnv("MINIO_ROOT_USER", "minio"),
		SecretKey:        getEnv("MINIO_ROOT_PASSWORD", "minio123"),
		UseSSL:           getEnv("MINIO_USE_SSL", "false") == "true",
	})
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to init media module")
	}

	// === HTTP Server ===

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		AppName:               "Industrix",
	})
	app.Use(recover.New())
	app.Use(cors.New())

	// Global middleware
	loggingMw := mw.NewLogging(l)
	rateLimitMw := mw.NewRateLimit(redisClient, 100, time.Minute)
	authMw := mw.NewAuth(jwtClient)

	app.Use(loggingMw.RequestLogger())
	app.Use(rateLimitMw.SlidingWindow())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	app.Get("/swagger/*", swagger.New(swagger.Config{
		URL: "/swagger/doc.json",
	}))

	api := app.Group("/api/v1")

	// Public routes (no auth required)
	identityMod.Handler.RegisterPublicRoutes(api)
	catalogMod.Handler.RegisterPublicRoutes(api)
	listingMod.Handler.RegisterPublicRoutes(api)
	bookingMod.Handler.RegisterPublicRoutes(api)
	searchMod.Handler.RegisterPublicRoutes(api)

	// Protected routes (auth required)
	protected := api.Group("/", authMw.ValidateJWT())
	identityMod.Handler.RegisterProtectedRoutes(protected)
	integrityMod.Handler.RegisterRoutes(protected)
	marketplaceMod.Handler.RegisterRoutes(protected)
	catalogMod.Handler.RegisterProtectedRoutes(protected)
	listingMod.Handler.RegisterProtectedRoutes(protected)
	dealMod.Handler.RegisterRoutes(protected)
	bookingMod.Handler.RegisterProtectedRoutes(protected)
	paymentMod.Handler.RegisterRoutes(protected)
	mediaMod.Handler.RegisterRoutes(protected)
	notificationMod.Handler.RegisterRoutes(protected)

	// Admin routes — protected + admin-role gated
	adminRoutes := api.Group("/", authMw.ValidateJWT(), authMw.RequireAdmin())
	adminMod.Handler.RegisterRoutes(adminRoutes)

	// WebSocket (self-authenticates via cookie, mounted outside /api/v1)
	dealMod.Handler.RegisterWebSocket(app)

	// === Start ===

	httpPort := getEnv("HTTP_PORT", "8080")
	go func() {
		l.Info().Str("port", httpPort).Msg("Starting HTTP server")
		if err := app.Listen(fmt.Sprintf(":%s", httpPort)); err != nil {
			l.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	l.Info().Msg("Shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		l.Error().Err(err).Msg("HTTP server shutdown error")
	}

	l.Info().Msg("Service stopped")
}

func loadRSAKeys(l *logger.Logger) (*rsa.PrivateKey, *rsa.PublicKey) {
	privKeyPEM := os.Getenv("JWT_PRIVATE_KEY")
	pubKeyPEM := os.Getenv("JWT_PUBLIC_KEY")

	if privKeyPEM != "" && pubKeyPEM != "" {
		l.Info().Msg("Loading JWT keys from environment variables")

		block, _ := pem.Decode([]byte(privKeyPEM))
		if block == nil {
			l.Fatal().Msg("Failed to parse JWT private key PEM")
		}
		privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			k, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err2 != nil {
				l.Fatal().Err(err).Msg("Failed to parse JWT private key")
			}
			privKey = k.(*rsa.PrivateKey)
		}

		blockPub, _ := pem.Decode([]byte(pubKeyPEM))
		if blockPub == nil {
			l.Fatal().Msg("Failed to parse JWT public key PEM")
		}
		pubKeyInterface, err := x509.ParsePKIXPublicKey(blockPub.Bytes)
		if err != nil {
			l.Fatal().Err(err).Msg("Failed to parse JWT public key")
		}
		pubKey := pubKeyInterface.(*rsa.PublicKey)

		return privKey, pubKey
	}

	l.Warn().Msg("JWT keys not provided, generating ephemeral keys (NOT FOR PRODUCTION)")
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to generate RSA keys")
	}
	return privateKey, &privateKey.PublicKey
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// splitAndTrim splits a comma-separated env value into a cleaned slice.
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
