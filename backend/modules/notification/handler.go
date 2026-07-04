package notification

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/industrix/backend/pkg/errors"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers the protected notification routes.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	n := router.Group("/notifications")
	n.Get("/", h.List)
	n.Get("/unread-count", h.UnreadCount)
	n.Put("/read-all", h.MarkAllRead)
	n.Put("/:id/read", h.MarkRead)
}

func respondErr(c *fiber.Ctx, err error) error {
	return c.Status(http.StatusInternalServerError).JSON(errors.New(errors.CodeInternal, "Something went wrong"))
}

// List godoc
// @Summary List the current user's notifications
// @Tags notifications
// @Security BearerAuth
// @Success 200 {array} Notification
// @Router /notifications [get]
func (h *Handler) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	items, err := h.service.List(c.Context(), userID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(items)
}

// UnreadCount godoc
// @Summary Unread notification count (for the bell badge)
// @Tags notifications
// @Security BearerAuth
// @Success 200 {object} map[string]int
// @Router /notifications/unread-count [get]
func (h *Handler) UnreadCount(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	count, err := h.service.UnreadCount(c.Context(), userID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(fiber.Map{"count": count})
}

// MarkRead godoc
// @Summary Mark one notification read
// @Tags notifications
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 200
// @Router /notifications/{id}/read [put]
func (h *Handler) MarkRead(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if err := h.service.MarkRead(c.Context(), c.Params("id"), userID); err != nil {
		return respondErr(c, err)
	}
	return c.SendStatus(http.StatusOK)
}

// MarkAllRead godoc
// @Summary Mark all notifications read
// @Tags notifications
// @Security BearerAuth
// @Success 200
// @Router /notifications/read-all [put]
func (h *Handler) MarkAllRead(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if err := h.service.MarkAllRead(c.Context(), userID); err != nil {
		return respondErr(c, err)
	}
	return c.SendStatus(http.StatusOK)
}
