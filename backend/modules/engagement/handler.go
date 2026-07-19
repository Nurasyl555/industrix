package engagement

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/industrix/backend/pkg/errors"
)

// Handler exposes watchlist and price-history endpoints.
type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterPublicRoutes mounts read-only price history (no auth needed).
func (h *Handler) RegisterPublicRoutes(router fiber.Router) {
	router.Get("/listings/:id/price-history", h.PriceHistory)
}

// RegisterProtectedRoutes mounts the auth-gated watchlist actions.
func (h *Handler) RegisterProtectedRoutes(router fiber.Router) {
	f := router.Group("/favorites")
	f.Post("/", h.AddFavorite)
	f.Delete("/:listingID", h.RemoveFavorite)

	router.Get("/my-favorites", h.ListFavorites)
}

func respondErr(c *fiber.Ctx, err error) error {
	if domainErr, ok := err.(*errors.Error); ok {
		return c.Status(errors.HTTPStatus(domainErr.Code)).JSON(domainErr)
	}
	return c.Status(http.StatusInternalServerError).JSON(errors.New(errors.CodeInternal, "Something went wrong"))
}

// AddFavorite godoc
// @Summary Add a listing to the watchlist
// @Tags engagement
// @Security BearerAuth
// @Param request body AddFavoriteRequest true "Listing to favorite"
// @Success 204
// @Router /favorites [post]
func (h *Handler) AddFavorite(c *fiber.Ctx) error {
	var req AddFavoriteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}
	userID := c.Locals("user_id").(string)
	if err := h.service.AddFavorite(c.Context(), userID, req.ListingID); err != nil {
		return respondErr(c, err)
	}
	return c.SendStatus(http.StatusNoContent)
}

// RemoveFavorite godoc
// @Summary Remove a listing from the watchlist
// @Tags engagement
// @Security BearerAuth
// @Param listingID path string true "Listing ID"
// @Success 204
// @Router /favorites/{listingID} [delete]
func (h *Handler) RemoveFavorite(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if err := h.service.RemoveFavorite(c.Context(), userID, c.Params("listingID")); err != nil {
		return respondErr(c, err)
	}
	return c.SendStatus(http.StatusNoContent)
}

// ListFavorites godoc
// @Summary List the current user's watchlist
// @Tags engagement
// @Security BearerAuth
// @Success 200 {array} FavoriteListing
// @Router /my-favorites [get]
func (h *Handler) ListFavorites(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	items, err := h.service.ListFavorites(c.Context(), userID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(items)
}

// PriceHistory godoc
// @Summary Price change history for a listing
// @Tags engagement
// @Param id path string true "Listing ID"
// @Success 200 {array} PriceHistoryEntry
// @Router /listings/{id}/price-history [get]
func (h *Handler) PriceHistory(c *fiber.Ctx) error {
	items, err := h.service.PriceHistory(c.Context(), c.Params("id"))
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(items)
}
