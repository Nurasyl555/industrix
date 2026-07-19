package dispute

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/industrix/backend/pkg/errors"
	"github.com/industrix/backend/platform/httperr"
)

// Handler serves the dispute endpoints.
type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterProtectedRoutes mounts the participant-facing routes.
func (h *Handler) RegisterProtectedRoutes(router fiber.Router) {
	d := router.Group("/disputes")
	d.Post("/", h.File)
	d.Get("/:id", h.Get)

	router.Get("/my-disputes", h.ListMine)
}

// RegisterAdminRoutes mounts arbitration. The caller applies the admin-role
// middleware to the group it passes in.
func (h *Handler) RegisterAdminRoutes(router fiber.Router) {
	router.Get("/admin/disputes", h.ListOpen)
	router.Put("/admin/disputes/:id/resolve", h.Resolve)
}

// respondErr maps a service error to its HTTP response. See platform/httperr —
// unexpected errors are logged there before the generic 500 goes out.
func respondErr(c *fiber.Ctx, err error) error { return httperr.Respond(c, err) }

// File godoc
// @Summary Open a dispute on a deal (buyer or seller)
// @Tags disputes
// @Security BearerAuth
// @Param request body FileDisputeRequest true "Reason and evidence"
// @Success 201 {object} Dispute
// @Router /disputes [post]
func (h *Handler) File(c *fiber.Ctx) error {
	var req FileDisputeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}
	userID := c.Locals("user_id").(string)
	d, err := h.service.File(c.Context(), userID, req)
	if err != nil {
		return respondErr(c, err)
	}
	return c.Status(http.StatusCreated).JSON(d)
}

// Get godoc
// @Summary Get a dispute (either party to the deal, or an admin)
// @Tags disputes
// @Security BearerAuth
// @Param id path string true "Dispute ID"
// @Success 200 {object} Dispute
// @Router /disputes/{id} [get]
func (h *Handler) Get(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	isAdmin, _ := c.Locals("role").(string)
	d, err := h.service.Get(c.Context(), c.Params("id"), userID, isAdmin == "admin")
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(d)
}

// ListMine godoc
// @Summary List disputes the current user filed
// @Tags disputes
// @Security BearerAuth
// @Success 200 {array} Dispute
// @Router /my-disputes [get]
func (h *Handler) ListMine(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	items, err := h.service.ListMine(c.Context(), userID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(items)
}

// ListOpen godoc
// @Summary [Admin] Disputes awaiting arbitration
// @Tags admin
// @Security BearerAuth
// @Success 200 {array} Dispute
// @Router /admin/disputes [get]
func (h *Handler) ListOpen(c *fiber.Ctx) error {
	items, err := h.service.ListOpen(c.Context())
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(items)
}

// Resolve godoc
// @Summary [Admin] Decide a dispute — refund, release or reject
// @Tags admin
// @Security BearerAuth
// @Param id path string true "Dispute ID"
// @Param request body ResolveRequest true "Decision and note"
// @Success 200 {object} Dispute
// @Router /admin/disputes/{id}/resolve [put]
func (h *Handler) Resolve(c *fiber.Ctx) error {
	var req ResolveRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}
	adminID := c.Locals("user_id").(string)
	d, err := h.service.Resolve(c.Context(), c.Params("id"), adminID, req)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(d)
}
