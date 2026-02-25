package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/redis/go-redis/v9"

	"gateway/internal/config"
	"gateway/internal/middleware"
	"gateway/internal/proxy"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(middleware.InjectTraceID())
	app.Use(middleware.RequestLogger())

	// Initialize Redis for rate limiting
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg)
	rateLimiter := middleware.NewRateLimiter(redisClient, cfg.RateLimit.RequestsPerMinute, cfg.RateLimit.Burst)

	// Initialize router
	router := proxy.NewRouter(app, cfg, authMiddleware, rateLimiter)
	router.RegisterRoutes()

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Starting gateway on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
