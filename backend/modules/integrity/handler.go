package integrity

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/industrix/backend/pkg/errors"
	"github.com/industrix/backend/platform/httperr"
)

// Handler handles integrity HTTP requests
type Handler struct {
	service Service
}

// NewHandler creates a new integrity handler
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers all integrity routes (all protected).
// "/me" is a distinct group from "/:id" to avoid the static-vs-param
// shadowing issue (see modules/listing/handler.go).
func (h *Handler) RegisterRoutes(router fiber.Router) {
	companies := router.Group("/companies")
	companies.Post("/", h.CreateCompany)
	companies.Get("/:id", h.GetCompany)
	companies.Put("/:id", h.UpdateCompany)

	router.Get("/my-company", h.GetMyCompany)

	sub := router.Group("/subscription")
	sub.Get("/", h.GetSubscription)
	sub.Get("/plans", h.Plans)
	sub.Put("/", h.ChangePlan)
}

// respondErr maps a service error to its HTTP response. See platform/httperr —
// unexpected errors are logged there before the generic 500 goes out.
func respondErr(c *fiber.Ctx, err error) error { return httperr.Respond(c, err) }

// CreateCompany godoc
// @Summary Create a company
// @Tags companies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateCompanyRequest true "Company details"
// @Success 201 {object} Company
// @Failure 400 {object} errors.Error
// @Router /companies [post]
func (h *Handler) CreateCompany(c *fiber.Ctx) error {
	var req CreateCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}

	userID := c.Locals("user_id").(string)
	company := &Company{
		Name:    req.Name,
		BIN:     req.BIN,
		Address: req.Address,
		Phone:   req.Phone,
		Email:   req.Email,
		Website: req.Website,
		OwnerID: userID,
	}

	if err := h.service.CreateCompany(c.Context(), company); err != nil {
		return respondErr(c, err)
	}

	return c.Status(http.StatusCreated).JSON(company)
}

// GetMyCompany godoc
// @Summary Get the current user's company (404 if none yet)
// @Tags companies
// @Security BearerAuth
// @Success 200 {object} Company
// @Router /my-company [get]
func (h *Handler) GetMyCompany(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	company, err := h.service.GetMyCompany(c.Context(), userID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(company)
}

// GetSubscription godoc
// @Summary Get the current user's subscription (defaults to free)
// @Tags subscription
// @Security BearerAuth
// @Success 200 {object} Subscription
// @Router /subscription [get]
func (h *Handler) GetSubscription(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	sub, err := h.service.GetSubscription(c.Context(), userID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(sub)
}

// Plans godoc
// @Summary List subscription plans with prices and listing limits
// @Tags subscription
// @Security BearerAuth
// @Success 200 {array} PlanOption
// @Router /subscription/plans [get]
func (h *Handler) Plans(c *fiber.Ctx) error {
	return c.JSON(Plans())
}

// ChangePlan godoc
// @Summary Change the current user's subscription plan
// @Tags subscription
// @Security BearerAuth
// @Param request body ChangePlanRequest true "New plan"
// @Success 200 {object} Subscription
// @Router /subscription [put]
func (h *Handler) ChangePlan(c *fiber.Ctx) error {
	var req ChangePlanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}
	userID := c.Locals("user_id").(string)
	sub, err := h.service.ChangePlan(c.Context(), userID, req.Plan)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(sub)
}

// GetCompany godoc
// @Summary Get company details
// @Tags companies
// @Param id path string true "Company ID"
// @Success 200 {object} Company
// @Router /companies/{id} [get]
func (h *Handler) GetCompany(c *fiber.Ctx) error {
	id := c.Params("id")
	company, err := h.service.GetCompany(c.Context(), id)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(err)
	}
	return c.JSON(company)
}

// UpdateCompany godoc
// @Summary Update company details
// @Tags companies
// @Security BearerAuth
// @Param id path string true "Company ID"
// @Param request body UpdateCompanyRequest true "Updated details"
// @Success 200 {object} Company
// @Router /companies/{id} [put]
func (h *Handler) UpdateCompany(c *fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}

	company := &Company{
		ID:      id,
		Name:    req.Name,
		Address: req.Address,
		Phone:   req.Phone,
		Email:   req.Email,
		Website: req.Website,
	}

	if err := h.service.UpdateCompany(c.Context(), company); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(err)
	}

	return c.JSON(company)
}
