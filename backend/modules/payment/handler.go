package payment

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/industrix/backend/platform/httperr"

	"github.com/industrix/backend/pkg/errors"
)

// Handler handles payment HTTP requests. All routes require authentication.
type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers all payment routes (all protected).
func (h *Handler) RegisterRoutes(router fiber.Router) {
	p := router.Group("/payments")
	p.Post("/", h.InitEscrow)
	p.Get("/:id", h.Get)
	p.Put("/:id/release", h.Release)
	p.Put("/:id/refund", h.Refund)

	router.Get("/my-payments", h.ListMine)
}

// respondErr maps a service error to its HTTP response. See platform/httperr —
// unexpected errors are logged there before the generic 500 goes out.
func respondErr(c *fiber.Ctx, err error) error { return httperr.Respond(c, err) }

// InitEscrow godoc
// @Summary Fund a deal into escrow (buyer only)
// @Tags payments
// @Security BearerAuth
// @Param request body CreatePaymentRequest true "Escrow details"
// @Success 201 {object} Payment
// @Router /payments [post]
func (h *Handler) InitEscrow(c *fiber.Ctx) error {
	var req CreatePaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}
	userID := c.Locals("user_id").(string)
	p, err := h.service.InitEscrow(c.Context(), userID, req)
	if err != nil {
		return respondErr(c, err)
	}
	return c.Status(http.StatusCreated).JSON(p)
}

// Get godoc
// @Summary Get a payment (payer or payee only)
// @Tags payments
// @Security BearerAuth
// @Param id path string true "Payment ID"
// @Success 200 {object} Payment
// @Router /payments/{id} [get]
func (h *Handler) Get(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	p, err := h.service.Get(c.Context(), c.Params("id"), userID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(p)
}

// Release godoc
// @Summary Release a held escrow to the seller (buyer confirms)
// @Tags payments
// @Security BearerAuth
// @Param id path string true "Payment ID"
// @Success 200 {object} Payment
// @Router /payments/{id}/release [put]
func (h *Handler) Release(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	p, err := h.service.Release(c.Context(), c.Params("id"), userID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(p)
}

// Refund godoc
// @Summary Refund a held escrow to the buyer
// @Tags payments
// @Security BearerAuth
// @Param id path string true "Payment ID"
// @Success 200 {object} Payment
// @Router /payments/{id}/refund [put]
func (h *Handler) Refund(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	p, err := h.service.Refund(c.Context(), c.Params("id"), userID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(p)
}

// ListMine godoc
// @Summary List the current user's payments (as payer or payee)
// @Tags payments
// @Security BearerAuth
// @Success 200 {array} Payment
// @Router /my-payments [get]
func (h *Handler) ListMine(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	items, err := h.service.ListMine(c.Context(), userID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(items)
}
