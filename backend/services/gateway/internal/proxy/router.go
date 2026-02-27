package proxy

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/industrix/services/gateway/internal/middleware"
)

type Config struct {
	TrustServiceURL         string
	InventoryServiceURL     string
	TransactionServiceURL   string
	ContentServiceURL       string
	CommunicationServiceURL string
	ServicesMarketplaceURL  string
	AnalyticsServiceURL     string
	AdminServiceURL         string
}

func RegisterRoutes(app *fiber.App, cfg Config, auth *middleware.AuthMiddleware, limiter *middleware.RateLimitMiddleware, logger *middleware.LoggingMiddleware) {
	// Apply global middleware
	app.Use(logger.RequestLogger())
	app.Use(limiter.SlidingWindow())

	api := app.Group("/api/v1")

	// Public Routes (No Auth)
	// Auth endpoints
	api.Post("/auth/register", func(c *fiber.Ctx) error { return proxy.Do(c, cfg.TrustServiceURL+"/api/v1/auth/register") })
	api.Post("/auth/login", func(c *fiber.Ctx) error { return proxy.Do(c, cfg.TrustServiceURL+"/api/v1/auth/login") })
	api.Post("/auth/verify-otp", func(c *fiber.Ctx) error { return proxy.Do(c, cfg.TrustServiceURL+"/api/v1/auth/verify-otp") })
	api.Post("/auth/refresh", func(c *fiber.Ctx) error { return proxy.Do(c, cfg.TrustServiceURL+"/api/v1/auth/refresh") })

	// Public Inventory Routes (Search, Get)
	api.Get("/equipment/*", func(c *fiber.Ctx) error { return proxy.Do(c, cfg.InventoryServiceURL+c.Path()) })
	api.Get("/categories/*", func(c *fiber.Ctx) error { return proxy.Do(c, cfg.InventoryServiceURL+c.Path()) })
	api.Get("/reviews/:entityID", func(c *fiber.Ctx) error { return proxy.Do(c, cfg.TrustServiceURL+c.Path()) })

	// Protected Routes (Apply Auth Middleware)
	protected := api.Group("/")
	protected.Use(auth.ValidateJWT())

	// Trust Service Protected
	setupProxy(protected, "/users", cfg.TrustServiceURL)
	setupProxy(protected, "/companies", cfg.TrustServiceURL)
	setupProxy(protected, "/reviews", cfg.TrustServiceURL) // POST reviews is protected

	// Inventory Service Protected (Create/Update/Delete)
	protected.Post("/equipment", func(c *fiber.Ctx) error { return proxy.Do(c, cfg.InventoryServiceURL+c.Path()) })
	protected.Put("/equipment/*", func(c *fiber.Ctx) error { return proxy.Do(c, cfg.InventoryServiceURL+c.Path()) })
	protected.Delete("/equipment/*", func(c *fiber.Ctx) error { return proxy.Do(c, cfg.InventoryServiceURL+c.Path()) })

	// Transaction Service Protected
	setupProxy(protected, "/deals", cfg.TransactionServiceURL)
	setupProxy(protected, "/bookings", cfg.TransactionServiceURL)
	setupProxy(protected, "/payments", cfg.TransactionServiceURL)

	// Content Service Protected
	setupProxy(protected, "/documents", cfg.ContentServiceURL)
	setupProxy(protected, "/media", cfg.ContentServiceURL)

	// Communication Service Protected
	setupProxy(protected, "/messages", cfg.CommunicationServiceURL)
	setupProxy(protected, "/notifications", cfg.CommunicationServiceURL)

	// Services Marketplace Protected
	setupProxy(protected, "/services", cfg.ServicesMarketplaceURL)

	// Analytics Protected
	setupProxy(protected, "/analytics", cfg.AnalyticsServiceURL)

	// Admin Routes
	admin := app.Group("/admin")
	admin.Use(auth.ValidateJWT())
	// Add admin role check middleware here if needed
	admin.All("/*", func(c *fiber.Ctx) error {
		path := strings.TrimPrefix(c.Path(), "/admin")
		return proxy.Do(c, cfg.AdminServiceURL+"/api/v1/admin"+path)
	})
}

func setupProxy(router fiber.Router, prefix string, target string) {
	router.All(prefix+"/*", func(c *fiber.Ctx) error {
		return proxy.Do(c, target+c.Path())
	})
}
