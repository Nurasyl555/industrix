package main

// @title Industrix API Gateway
// @version 1.0
// @description API Gateway for the Industrial Equipment Marketplace.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/redis/go-redis/v9"

	_ "github.com/industrix/services/gateway/docs"
	"github.com/industrix/services/gateway/internal/config"
	"github.com/industrix/services/gateway/internal/middleware"
	"github.com/industrix/services/gateway/internal/proxy"
)

func main() {
	cfg := config.Load()

	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	})

	app.Use(recover.New())
	app.Use(cors.New())

	app.Get("/swagger/*", swagger.HandlerDefault)
	app.Use(middleware.InjectTraceID())
	app.Use(middleware.RequestLogger())

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	authMiddleware := middleware.NewAuthMiddleware(cfg)
	rateLimiter := middleware.NewRateLimiter(redisClient, cfg.RateLimit.RequestsPerMinute, cfg.RateLimit.Burst)

	router := proxy.NewRouter(app, cfg, authMiddleware, rateLimiter)
	router.RegisterRoutes()

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Starting gateway on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
