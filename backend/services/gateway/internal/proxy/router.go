package proxy

import (
	"github.com/gofiber/fiber/v2"

	"gateway/internal/config"
	"gateway/internal/middleware"
)

type Router struct {
	app         *fiber.App
	cfg         *config.Config
	auth        *middleware.AuthMiddleware
	rateLimiter *middleware.RateLimiter
}

func NewRouter(app *fiber.App, cfg *config.Config, auth *middleware.AuthMiddleware, rateLimiter *middleware.RateLimiter) *Router {
	return &Router{
		app:         app,
		cfg:         cfg,
		auth:        auth,
		rateLimiter: rateLimiter,
	}
}

// RegisterRoutes maps all API routes to downstream services
func (r *Router) RegisterRoutes() {
	// Health check - no auth required
	r.app.Get("/health", r.handleHealth)

	// API v1 routes
	v1 := r.app.Group("/api/v1")

	// Public routes - rate limited but no auth
	public := v1.Group("", r.rateLimiter.SlidingWindow())

	// Auth routes - no auth required for register/login
	auth := public.Group("/auth")
	auth.Post("/register", r.proxyToIdentity("/auth/register"))
	auth.Post("/login", r.proxyToIdentity("/auth/login"))
	auth.Post("/otp/send", r.proxyToIdentity("/auth/otp/send"))
	auth.Post("/otp/verify", r.proxyToIdentity("/auth/otp/verify"))
	auth.Post("/refresh", r.proxyToIdentity("/auth/refresh"))
	auth.Post("/logout", r.proxyToIdentity("/auth/logout"))
	auth.Post("/password/forgot", r.proxyToIdentity("/auth/password/forgot"))
	auth.Post("/password/reset", r.proxyToIdentity("/auth/password/reset"))

	// Public catalog routes
	public.Get("/catalog/categories", r.proxyToCatalog("/categories"))
	public.Get("/catalog/equipment", r.proxyToCatalog("/equipment"))
	public.Get("/catalog/equipment/:id", r.proxyToCatalog("/equipment/:id"))

	// Public search
	public.Get("/search", r.proxyToSearch("/"))

	// Protected routes - require auth
	protected := v1.Group("", r.auth.ValidateJWT())

	// Profile routes
	profile := protected.Group("/profile")
	profile.Get("/", r.proxyToIdentity("/profile"))
	profile.Put("/", r.proxyToIdentity("/profile"))
	profile.Put("/avatar", r.proxyToIdentity("/profile/avatar"))
	profile.Get("/:userID", r.proxyToIdentity("/profile/:userID"))
	profile.Put("/notifications", r.proxyToIdentity("/profile/notifications"))

	// Company routes
	company := protected.Group("/companies")
	company.Post("/", r.proxyToIdentity("/companies"))
	company.Put("/me", r.proxyToIdentity("/companies/me"))
	company.Get("/me", r.proxyToIdentity("/companies/me"))
	company.Post("/me/documents", r.proxyToIdentity("/companies/me/documents"))
	company.Get("/me/verification-status", r.proxyToIdentity("/companies/me/verification-status"))

	// Listing routes
	listing := protected.Group("/listings")
	listing.Post("/", r.proxyToListing("/"))
	listing.Get("/", r.proxyToListing("/"))
	listing.Get("/:id", r.proxyToListing("/:id"))
	listing.Put("/:id", r.proxyToListing("/:id"))
	listing.Delete("/:id", r.proxyToListing("/:id"))
	listing.Get("/:id/stats", r.proxyToListing("/:id/stats"))

	// Booking routes
	booking := protected.Group("/bookings")
	booking.Post("/", r.proxyToBooking("/"))
	booking.Get("/", r.proxyToBooking("/"))
	booking.Get("/:id", r.proxyToBooking("/:id"))
	booking.Put("/:id", r.proxyToBooking("/:id"))
	booking.Delete("/:id", r.proxyToBooking("/:id"))

	// Deal routes
	deals := protected.Group("/deals")
	deals.Post("/", r.proxyToDeal("/"))
	deals.Get("/", r.proxyToDeal("/"))
	deals.Get("/:id", r.proxyToDeal("/:id"))
	deals.Put("/:id", r.proxyToDeal("/:id"))
	deals.Put("/:id/status", r.proxyToDeal("/:id/status"))

	// Payment routes
	payments := protected.Group("/payments")
	payments.Post("/initiate", r.proxyToPayment("/initiate"))
	payments.Get("/:id", r.proxyToPayment("/:id"))
	payments.Post("/:id/complete", r.proxyToPayment("/:id/complete"))
	payments.Post("/:id/refund", r.proxyToPayment("/:id/refund"))

	// Document routes
	documents := protected.Group("/documents")
	documents.Post("/", r.proxyToDocument("/"))
	documents.Get("/:id", r.proxyToDocument("/:id"))
	documents.Get("/:id/download", r.proxyToDocument("/:id/download"))

	// Review routes
	reviews := protected.Group("/reviews")
	reviews.Post("/", r.proxyToReview("/"))
	reviews.Get("/", r.proxyToReview("/"))
	reviews.Get("/:id", r.proxyToReview("/:id"))
	reviews.Put("/:id", r.proxyToReview("/:id"))

	// Chat routes
	chat := protected.Group("/chat")
	chat.Get("/conversations", r.proxyToChat("/conversations"))
	chat.Get("/conversations/:id", r.proxyToChat("/conversations/:id"))
	chat.Post("/conversations/:id/messages", r.proxyToChat("/conversations/:id/messages"))
	chat.Get("/conversations/:id/messages", r.proxyToChat("/conversations/:id/messages"))

	// Notification routes
	notifications := protected.Group("/notifications")
	notifications.Get("/", r.proxyToNotification("/"))
	notifications.Get("/:id", r.proxyToNotification("/:id"))
	notifications.Put("/:id/read", r.proxyToNotification("/:id/read"))
	notifications.Put("/read-all", r.proxyToNotification("/read-all"))

	// Services marketplace
	services := protected.Group("/services")
	services.Post("/", r.proxyToServicesMarketplace("/"))
	services.Get("/", r.proxyToServicesMarketplace("/"))
	services.Get("/:id", r.proxyToServicesMarketplace("/:id"))

	// Engagement routes
	engagement := protected.Group("/engagement")
	engagement.Get("/favorites", r.proxyToEngagement("/favorites"))
	engagement.Post("/favorites", r.proxyToEngagement("/favorites"))
	engagement.Delete("/favorites/:id", r.proxyToEngagement("/favorites/:id"))
	engagement.Get("/history", r.proxyToEngagement("/history"))

	// Media upload
	media := protected.Group("/media")
	media.Post("/upload-url", r.proxyToMedia("/upload-url"))
	media.Delete("/:id", r.proxyToMedia("/:id"))

	// Admin routes - stricter auth scope check
	admin := protected.Group("/admin", r.auth.RequireRole("ADMIN", "MODERATOR"))
	admin.Get("/users", r.proxyToAdmin("/users"))
	admin.Get("/users/:id", r.proxyToAdmin("/users/:id"))
	admin.Put("/users/:id/status", r.proxyToAdmin("/users/:id/status"))
	admin.Get("/companies", r.proxyToAdmin("/companies"))
	admin.Get("/companies/:id", r.proxyToAdmin("/companies/:id"))
	admin.Put("/companies/:id/verify", r.proxyToAdmin("/companies/:id/verify"))
	admin.Put("/companies/:id/reject", r.proxyToAdmin("/companies/:id/reject"))
	admin.Get("/listings", r.proxyToAdmin("/listings"))
	admin.Put("/listings/:id/moderate", r.proxyToAdmin("/listings/:id/moderate"))
	admin.Get("/stats", r.proxyToAdmin("/stats"))
}

// handleHealth handles gateway liveness check
func (r *Router) handleHealth(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"service": "gateway",
	})
}

// Proxy helper methods - these create fiber handlers that reverse proxy to downstream services
func (r *Router) proxyToIdentity(path string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return r.proxy(c, r.cfg.Services.CatalogURL, path)
	}
}

func (r *Router) proxyToCatalog(path string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return r.proxy(c, r.cfg.Services.CatalogURL, path)
	}
}

func (r *Router) proxyToListing(path string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return r.proxy(c, r.cfg.Services.ListingURL, path)
	}
}

func (r *Router) proxyToSearch(path string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return r.proxy(c, r.cfg.Services.SearchURL, path)
	}
}

func (r *Router) proxyToBooking(path string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return r.proxy(c, r.cfg.Services.BookingURL, path)
	}
}

func (r *Router) proxyToDeal(path string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return r.proxy(c, r.cfg.Services.DealURL, path)
	}
}

func (r *Router) proxyToPayment(path string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return r.proxy(c, r.cfg.Services.PaymentURL, path)
	}
}

func (r *Router) proxyToDocument(path string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return r.proxy(c, r.cfg.Services.DocumentURL, path)
	}
}

func (r *Router) proxyToReview(path string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return r.proxy(c, r.cfg.Services.ReviewURL, path)
	}
}

func (r *Router) proxyToChat(path string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return r.proxy(c, r.cfg.Services.ChatURL, path)
	}
}

func (r *Router) proxyToNotification(path string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return r.proxy(c, r.cfg.Services.NotificationURL, path)
	}
}

func (r *Router) proxyToServicesMarketplace(path string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return r.proxy(c, r.cfg.Services.ServicesMarketplaceURL, path)
	}
}

func (r *Router) proxyToEngagement(path string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return r.proxy(c, r.cfg.Services.EngagementURL, path)
	}
}

func (r *Router) proxyToMedia(path string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return r.proxy(c, r.cfg.Services.MediaURL, path)
	}
}

func (r *Router) proxyToAdmin(path string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return r.proxy(c, r.cfg.Services.AdminURL, path)
	}
}

// proxy performs the actual reverse proxy operation
func (r *Router) proxy(c *fiber.Ctx, backendURL string, path string) error {
	// Forward the request to the backend service
	// This is a simplified version - in production you'd use a proper HTTP client
	// like httputil.ReverseProxy or a library like reverseproxy

	// TODO: Implement actual reverse proxy with:
	// - Header forwarding (X-User-ID, X-Trace-ID, etc.)
	// - Response forwarding
	// - Error handling

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "proxy not implemented",
	})
}
