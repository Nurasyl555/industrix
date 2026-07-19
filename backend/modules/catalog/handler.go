package catalog

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/industrix/backend/pkg/errors"
)

// Handler handles catalog HTTP requests
type Handler struct {
	service Service
}

// NewHandler creates a new catalog handler
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterPublicRoutes registers routes that don't require authentication —
// browsing the catalog should not require an account.
func (h *Handler) RegisterPublicRoutes(router fiber.Router) {
	catalog := router.Group("/catalog")
	catalog.Get("/categories", h.ListCategories)
	catalog.Get("/compare", h.Compare)
	catalog.Get("/equipment", h.ListEquipment)
	catalog.Get("/equipment/:id", h.GetEquipment)
}

// RegisterProtectedRoutes registers routes that require authentication
func (h *Handler) RegisterProtectedRoutes(router fiber.Router) {
	catalog := router.Group("/catalog")
	catalog.Post("/equipment", h.CreateEquipment)
	catalog.Put("/equipment/:id", h.UpdateEquipment)
	catalog.Delete("/equipment/:id", h.DeleteEquipment)
}

// respondErr maps a domain error to its HTTP status; falls back to 500 for
// anything that isn't a *errors.Error (e.g. a raw DB error).
func respondErr(c *fiber.Ctx, err error) error {
	if domainErr, ok := err.(*errors.Error); ok {
		return c.Status(errors.HTTPStatus(domainErr.Code)).JSON(domainErr)
	}
	return c.Status(http.StatusInternalServerError).JSON(errors.New(errors.CodeInternal, "Something went wrong"))
}

// ListCategories godoc
// @Summary List equipment categories
// @Tags catalog
// @Success 200 {array} Category
// @Router /catalog/categories [get]
func (h *Handler) ListCategories(c *fiber.Ctx) error {
	categories, err := h.service.ListCategories(c.Context())
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(categories)
}

// CreateEquipment godoc
// @Summary Add equipment to the catalog
// @Tags catalog
// @Security BearerAuth
// @Param request body CreateEquipmentRequest true "Equipment details"
// @Success 201 {object} Equipment
// @Router /catalog/equipment [post]
func (h *Handler) CreateEquipment(c *fiber.Ctx) error {
	var req CreateEquipmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}

	userID := c.Locals("user_id").(string)
	eq, err := h.service.CreateEquipment(c.Context(), userID, req)
	if err != nil {
		return respondErr(c, err)
	}
	return c.Status(http.StatusCreated).JSON(eq)
}

// GetEquipment godoc
// @Summary Get equipment details
// @Tags catalog
// @Param id path string true "Equipment ID"
// @Success 200 {object} Equipment
// @Router /catalog/equipment/{id} [get]
func (h *Handler) GetEquipment(c *fiber.Ctx) error {
	eq, err := h.service.GetEquipment(c.Context(), c.Params("id"))
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(eq)
}

// ListEquipment godoc
// @Summary List/search equipment
// @Tags catalog
// @Param category_id query string false "Category ID"
// @Param region query string false "Region"
// @Param search query string false "Search text"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /catalog/equipment [get]
func (h *Handler) ListEquipment(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	filter := ListEquipmentFilter{
		CategoryID: c.Query("category_id"),
		Region:     c.Query("region"),
		Search:     c.Query("search"),
		Page:       page,
		Limit:      limit,
	}

	items, total, err := h.service.ListEquipment(c.Context(), filter)
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

// Compare godoc
// @Summary Compare several equipment items side by side
// @Tags catalog
// @Param ids query string true "Comma-separated equipment IDs (2–10)"
// @Success 200 {object} map[string]interface{}
// @Router /catalog/compare [get]
func (h *Handler) Compare(c *fiber.Ctx) error {
	var ids []string
	for _, part := range strings.Split(c.Query("ids"), ",") {
		if p := strings.TrimSpace(part); p != "" {
			ids = append(ids, p)
		}
	}
	items, err := h.service.CompareEquipment(c.Context(), ids)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(fiber.Map{"items": items})
}

// UpdateEquipment godoc
// @Summary Update equipment details
// @Tags catalog
// @Security BearerAuth
// @Param id path string true "Equipment ID"
// @Param request body UpdateEquipmentRequest true "Updated fields"
// @Success 200 {object} Equipment
// @Router /catalog/equipment/{id} [put]
func (h *Handler) UpdateEquipment(c *fiber.Ctx) error {
	var req UpdateEquipmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}

	userID := c.Locals("user_id").(string)
	eq, err := h.service.UpdateEquipment(c.Context(), c.Params("id"), userID, req)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(eq)
}

// DeleteEquipment godoc
// @Summary Remove equipment from the catalog
// @Tags catalog
// @Security BearerAuth
// @Param id path string true "Equipment ID"
// @Success 204
// @Router /catalog/equipment/{id} [delete]
func (h *Handler) DeleteEquipment(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if err := h.service.DeleteEquipment(c.Context(), c.Params("id"), userID); err != nil {
		return respondErr(c, err)
	}
	return c.SendStatus(http.StatusNoContent)
}
