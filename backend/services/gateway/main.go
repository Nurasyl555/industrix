package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/industrix/pkg/logger"
	"github.com/industrix/pkg/redis"
	"github.com/industrix/services/gateway/internal/middleware"
	"github.com/industrix/services/gateway/internal/proxy"
	trustpb "github.com/industrix/gen/go/trust/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	serviceName := "gateway-service"
	l := logger.New(serviceName)
	logger.SetGlobal(l)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Redis client for rate limiting
	redisCfg := redis.DefaultConfig()
	redisClient, err := redis.NewClient(ctx, redisCfg)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisClient.Close()

	// gRPC connection to Trust service for Auth
	trustAddr := getEnv("TRUST_SERVICE_ADDR", "trust:50051")
	conn, err := grpc.NewClient(trustAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to connect to Trust Service")
	}
	defer conn.Close()

	trustClient := trustpb.NewTrustServiceClient(conn)

	// Initialize middleware
	authMiddleware := middleware.NewAuth(trustClient)
	rateLimitMiddleware := middleware.NewRateLimit(redisClient, 100, time.Minute) // 100 req/min/IP
	loggingMiddleware := middleware.NewLogging(l)

	// Configure Proxy
	proxyCfg := proxy.Config{
		TrustServiceURL:         getEnv("TRUST_SERVICE_URL", "http://trust:8081"),
		InventoryServiceURL:     getEnv("INVENTORY_SERVICE_URL", "http://inventory:8081"),
		TransactionServiceURL:   getEnv("TRANSACTION_SERVICE_URL", "http://transaction:8081"),
		ContentServiceURL:       getEnv("CONTENT_SERVICE_URL", "http://content:8081"),
		CommunicationServiceURL: getEnv("COMMUNICATION_SERVICE_URL", "http://communication:8081"),
		ServicesMarketplaceURL:  getEnv("SERVICES_MARKETPLACE_URL", "http://services-marketplace:8081"),
		AnalyticsServiceURL:     getEnv("ANALYTICS_SERVICE_URL", "http://analytics:8081"),
		AdminServiceURL:         getEnv("ADMIN_SERVICE_URL", "http://admin:8081"),
	}

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		AppName:               "Industrix API Gateway",
	})
	app.Use(recover.New())
	app.Use(cors.New())

	// Apply global middleware
	app.Use(loggingMiddleware.RequestLogger())
	app.Use(rateLimitMiddleware.SlidingWindow())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	proxy.RegisterRoutes(app, proxyCfg, authMiddleware, rateLimitMiddleware, loggingMiddleware)

	httpPort := getEnv("HTTP_PORT", "8080")

	go func() {
		l.Info().Str("port", httpPort).Msg("Starting API Gateway")
		if err := app.Listen(fmt.Sprintf(":%s", httpPort)); err != nil {
			l.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	l.Info().Msg("Shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		l.Error().Err(err).Msg("HTTP server shutdown error")
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
