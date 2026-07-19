package analytics

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/industrix/backend/pkg/errors"
)

// Handler exposes the seller and admin dashboards.
type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterProtectedRoutes mounts the seller's own dashboard.
func (h *Handler) RegisterProtectedRoutes(router fiber.Router) {
	router.Get("/analytics/seller", h.SellerStats)
}

// RegisterAdminRoutes mounts the platform-wide dashboard (admin-gated).
func (h *Handler) RegisterAdminRoutes(router fiber.Router) {
	router.Get("/admin/analytics", h.AdminStats)
}

func respondErr(c *fiber.Ctx, err error) error {
	if domainErr, ok := err.(*errors.Error); ok {
		return c.Status(errors.HTTPStatus(domainErr.Code)).JSON(domainErr)
	}
	return c.Status(http.StatusInternalServerError).JSON(errors.New(errors.CodeInternal, "Something went wrong"))
}

// SellerStats godoc
// @Summary The current seller's funnel (listings, inquiries, deals, revenue)
// @Tags analytics
// @Security BearerAuth
// @Param days query int false "Window in days (default 30)"
// @Success 200 {object} SellerStats
// @Router /analytics/seller [get]
func (h *Handler) SellerStats(c *fiber.Ctx) error {
	days, _ := strconv.Atoi(c.Query("days", "30"))
	userID := c.Locals("user_id").(string)
	stats, err := h.service.SellerStats(c.Context(), userID, days)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(stats)
}

// AdminStats godoc
// @Summary Platform-wide dashboard (GMV, listings, deals, active sellers)
// @Tags analytics
// @Security BearerAuth
// @Param days query int false "Window in days (default 30)"
// @Success 200 {object} AdminStats
// @Router /admin/analytics [get]
func (h *Handler) AdminStats(c *fiber.Ctx) error {
	days, _ := strconv.Atoi(c.Query("days", "30"))
	stats, err := h.service.AdminStats(c.Context(), days)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(stats)
}
