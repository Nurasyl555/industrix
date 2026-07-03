package listing

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/industrix/backend/pkg/errors"
)

// Handler handles listing HTTP requests
type Handler struct {
	service Service
}

// NewHandler creates a new listing handler
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterPublicRoutes registers routes that don't require authentication —
// browsing listings should not require an account.
func (h *Handler) RegisterPublicRoutes(router fiber.Router) {
	listings := router.Group("/listings")
	listings.Get("/", h.ListActive)
	listings.Get("/:id", h.GetListing)
}

// RegisterProtectedRoutes registers routes that require authentication.
// Deliberately namespaced under /my-listings rather than /listings/:id-style
// paths — mixing a static "/listings/my" with public "/listings/:id" risks
// the static segment being shadowed depending on router internals/order.
// A distinct top-level prefix side-steps that ambiguity entirely.
func (h *Handler) RegisterProtectedRoutes(router fiber.Router) {
	listings := router.Group("/listings")
	listings.Post("/", h.CreateListing)

	my := router.Group("/my-listings")
	my.Get("/", h.ListMy)
	my.Put("/:id", h.UpdateListing)
	my.Put("/:id/publish", h.Publish)
	my.Put("/:id/archive", h.Archive)
	my.Delete("/:id", h.DeleteListing)
}

func respondErr(c *fiber.Ctx, err error) error {
	if domainErr, ok := err.(*errors.Error); ok {
		return c.Status(errors.HTTPStatus(domainErr.Code)).JSON(domainErr)
	}
	return c.Status(http.StatusInternalServerError).JSON(errors.New(errors.CodeInternal, "Something went wrong"))
}

// CreateListing godoc
// @Summary Create a sale/rental listing for owned equipment
// @Tags listings
// @Security BearerAuth
// @Param request body CreateListingRequest true "Listing details"
// @Success 201 {object} Listing
// @Router /listings [post]
func (h *Handler) CreateListing(c *fiber.Ctx) error {
	var req CreateListingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}

	userID := c.Locals("user_id").(string)
	l, err := h.service.CreateListing(c.Context(), userID, req)
	if err != nil {
		return respondErr(c, err)
	}
	return c.Status(http.StatusCreated).JSON(l)
}

// GetListing godoc
// @Summary Get listing details (with equipment info)
// @Tags listings
// @Param id path string true "Listing ID"
// @Success 200 {object} ListingView
// @Router /listings/{id} [get]
func (h *Handler) GetListing(c *fiber.Ctx) error {
	l, err := h.service.GetListing(c.Context(), c.Params("id"))
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(l)
}

// ListActive godoc
// @Summary Browse active listings
// @Tags listings
// @Param category_id query string false "Category ID"
// @Param region query string false "Region"
// @Param listing_type query string false "sale or rental"
// @Param condition query string false "new or used"
// @Param sort query string false "price_asc, price_desc, or newest (default)"
// @Param search query string false "Search text"
// @Param price_min query number false "Minimum price"
// @Param price_max query number false "Maximum price"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /listings [get]
func (h *Handler) ListActive(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	priceMin, _ := strconv.ParseFloat(c.Query("price_min", "0"), 64)
	priceMax, _ := strconv.ParseFloat(c.Query("price_max", "0"), 64)

	filter := ListListingsFilter{
		CategoryID:  c.Query("category_id"),
		Region:      c.Query("region"),
		ListingType: c.Query("listing_type"),
		Condition:   c.Query("condition"),
		Search:      c.Query("search"),
		Sort:        c.Query("sort"),
		PriceMin:    priceMin,
		PriceMax:    priceMax,
		Page:        page,
		Limit:       limit,
	}

	items, total, err := h.service.ListActive(c.Context(), filter)
	if err != nil {
		return respondErr(c, err)
	}

	return c.JSON(fiber.Map{
		"items": items,
		"total": total,
		"page":  filter.Page,
		"limit": filter.Limit,
	})
}

// ListMy godoc
// @Summary List the current user's listings (any status)
// @Tags listings
// @Security BearerAuth
// @Success 200 {array} Listing
// @Router /my-listings [get]
func (h *Handler) ListMy(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	items, err := h.service.ListMy(c.Context(), userID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(items)
}

// UpdateListing godoc
// @Summary Update a listing's price
// @Tags listings
// @Security BearerAuth
// @Param id path string true "Listing ID"
// @Param request body UpdateListingRequest true "Updated fields"
// @Success 200 {object} Listing
// @Router /my-listings/{id} [put]
func (h *Handler) UpdateListing(c *fiber.Ctx) error {
	var req UpdateListingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}

	userID := c.Locals("user_id").(string)
	l, err := h.service.UpdateListing(c.Context(), c.Params("id"), userID, req)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(l)
}

// Publish godoc
// @Summary Publish a draft listing (draft -> active)
// @Tags listings
// @Security BearerAuth
// @Param id path string true "Listing ID"
// @Success 200
// @Router /my-listings/{id}/publish [put]
func (h *Handler) Publish(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if err := h.service.Publish(c.Context(), c.Params("id"), userID); err != nil {
		return respondErr(c, err)
	}
	return c.SendStatus(http.StatusOK)
}

// Archive godoc
// @Summary Archive a listing (-> archived)
// @Tags listings
// @Security BearerAuth
// @Param id path string true "Listing ID"
// @Success 200
// @Router /my-listings/{id}/archive [put]
func (h *Handler) Archive(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if err := h.service.Archive(c.Context(), c.Params("id"), userID); err != nil {
		return respondErr(c, err)
	}
	return c.SendStatus(http.StatusOK)
}

// DeleteListing godoc
// @Summary Delete a listing
// @Tags listings
// @Security BearerAuth
// @Param id path string true "Listing ID"
// @Success 204
// @Router /my-listings/{id} [delete]
func (h *Handler) DeleteListing(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if err := h.service.DeleteListing(c.Context(), c.Params("id"), userID); err != nil {
		return respondErr(c, err)
	}
	return c.SendStatus(http.StatusNoContent)
}
