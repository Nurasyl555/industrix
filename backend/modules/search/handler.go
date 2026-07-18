package search

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/industrix/backend/pkg/errors"
)

// Handler exposes the public search endpoint.
type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterPublicRoutes mounts search under the public API group.
func (h *Handler) RegisterPublicRoutes(router fiber.Router) {
	router.Get("/search", h.Search)
}

// Search godoc
// @Summary Full-text + faceted marketplace search
// @Tags search
// @Param q query string false "Search text"
// @Param category_id query string false "Category filter"
// @Param region query string false "Region filter"
// @Param condition query string false "Condition filter (new/used)"
// @Param listing_type query string false "Listing type (sale/rental)"
// @Param min_price query number false "Minimum price"
// @Param max_price query number false "Maximum price"
// @Param sort query string false "price_asc | price_desc | newest"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} Result
// @Router /search [get]
func (h *Handler) Search(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	minPrice, _ := strconv.ParseFloat(c.Query("min_price", "0"), 64)
	maxPrice, _ := strconv.ParseFloat(c.Query("max_price", "0"), 64)

	q := Query{
		Text:        c.Query("q"),
		CategoryID:  c.Query("category_id"),
		Region:      c.Query("region"),
		Condition:   c.Query("condition"),
		ListingType: c.Query("listing_type"),
		PriceMin:    minPrice,
		PriceMax:    maxPrice,
		Sort:        c.Query("sort"),
		Page:        page,
		Limit:       limit,
	}

	res, err := h.service.Search(c.Context(), q)
	if err != nil {
		return c.Status(http.StatusServiceUnavailable).JSON(
			errors.New(errors.CodeInternal, "Search is temporarily unavailable"))
	}
	return c.JSON(res)
}
