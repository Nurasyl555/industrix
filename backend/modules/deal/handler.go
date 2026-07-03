package deal

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/industrix/backend/pkg/errors"
)

// Handler handles deal HTTP requests. All routes require authentication —
// there's no anonymous browsing of deals.
type Handler struct {
	service Service
}

// NewHandler creates a new deal handler
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers all deal routes (all protected).
// "/my" is namespaced separately from "/:id" for the same reason as the
// listing module — see modules/listing/handler.go for why.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	deals := router.Group("/deals")
	deals.Post("/", h.CreateDeal)
	deals.Get("/:id", h.GetDeal)
	deals.Put("/:id/close", h.Close)

	my := router.Group("/my-deals")
	my.Get("/", h.ListMy)
}

func respondErr(c *fiber.Ctx, err error) error {
	if domainErr, ok := err.(*errors.Error); ok {
		return c.Status(errors.HTTPStatus(domainErr.Code)).JSON(domainErr)
	}
	return c.Status(http.StatusInternalServerError).JSON(errors.New(errors.CodeInternal, "Something went wrong"))
}

// CreateDeal godoc
// @Summary Inquire about a listing
// @Tags deals
// @Security BearerAuth
// @Param request body CreateDealRequest true "Inquiry details"
// @Success 201 {object} Deal
// @Router /deals [post]
func (h *Handler) CreateDeal(c *fiber.Ctx) error {
	var req CreateDealRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}

	userID := c.Locals("user_id").(string)
	d, err := h.service.CreateDeal(c.Context(), userID, req)
	if err != nil {
		return respondErr(c, err)
	}
	return c.Status(http.StatusCreated).JSON(d)
}

// GetDeal godoc
// @Summary Get a deal (buyer or seller only)
// @Tags deals
// @Security BearerAuth
// @Param id path string true "Deal ID"
// @Success 200 {object} DealView
// @Router /deals/{id} [get]
func (h *Handler) GetDeal(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	d, err := h.service.GetDeal(c.Context(), c.Params("id"), userID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(d)
}

// ListMy godoc
// @Summary List the current user's deals, as buyer or seller
// @Tags deals
// @Security BearerAuth
// @Success 200 {array} DealView
// @Router /my-deals [get]
func (h *Handler) ListMy(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	deals, err := h.service.ListMy(c.Context(), userID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(deals)
}

// Close godoc
// @Summary Close a deal
// @Tags deals
// @Security BearerAuth
// @Param id path string true "Deal ID"
// @Success 200
// @Router /deals/{id}/close [put]
func (h *Handler) Close(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if err := h.service.Close(c.Context(), c.Params("id"), userID); err != nil {
		return respondErr(c, err)
	}
	return c.SendStatus(http.StatusOK)
}
