package company

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/industrix/pkg/errors"
)

// Handler handles HTTP requests for company
type Handler struct {
	service *Service
}

// NewHandler creates a new company handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateCompany creates a new company
// POST /companies
func (h *Handler) CreateCompany(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(http.StatusUnauthorized).JSON(errors.New(errors.CodeUnauthorized, "unauthorized"))
	}

	var req CreateCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeBadRequest, "invalid request body"))
	}

	company, err := h.service.CreateCompany(c.Context(), userID.(string), req)
	if err != nil {
		return c.Status(errors.HTTPStatus(err.Code)).JSON(err)
	}

	return c.Status(http.StatusCreated).JSON(company)
}

// UpdateCompany updates the authenticated user's company
// PUT /companies/me
func (h *Handler) UpdateCompany(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(http.StatusUnauthorized).JSON(errors.New(errors.CodeUnauthorized, "unauthorized"))
	}

	var req UpdateCompanyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeBadRequest, "invalid request body"))
	}

	company, err := h.service.UpdateCompany(c.Context(), userID.(string), req)
	if err != nil {
		return c.Status(errors.HTTPStatus(err.Code)).JSON(err)
	}

	return c.JSON(company)
}

// UploadDocument uploads verification documents
// POST /companies/me/documents
func (h *Handler) UploadDocument(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(http.StatusUnauthorized).JSON(errors.New(errors.CodeUnauthorized, "unauthorized"))
	}

	var req UploadDocumentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeBadRequest, "invalid request body"))
	}

	doc, err := h.service.UploadDocument(c.Context(), userID.(string), req)
	if err != nil {
		return c.Status(errors.HTTPStatus(err.Code)).JSON(err)
	}

	return c.JSON(doc)
}

// GetVerificationStatus returns verification status
// GET /companies/me/verification-status
func (h *Handler) GetVerificationStatus(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(http.StatusUnauthorized).JSON(errors.New(errors.CodeUnauthorized, "unauthorized"))
	}

	status, err := h.service.GetVerificationStatus(c.Context(), userID.(string))
	if err != nil {
		return c.Status(errors.HTTPStatus(err.Code)).JSON(err)
	}

	return c.JSON(status)
}

// RegisterRoutes registers company routes
func (h *Handler) RegisterRoutes(router fiber.Router) {
	companies := router.Group("/companies")
	companies.Post("", h.CreateCompany)
	companies.Put("/me", h.UpdateCompany)
	companies.Post("/me/documents", h.UploadDocument)
	companies.Get("/me/verification-status", h.GetVerificationStatus)
}
