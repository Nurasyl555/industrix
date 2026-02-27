package company

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/industrix/pkg/errors"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateCompany creates a new company
// @Summary Create a new company
// @Description Create a company profile for the authenticated user
// @Tags company
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateCompanyRequest true "Create Company Request"
// @Success 201 {object} Company
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /companies [post]
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
// @Summary Update company profile
// @Description Update the company profile of the authenticated user
// @Tags company
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateCompanyRequest true "Update Company Request"
// @Success 200 {object} Company
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /companies/me [put]
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
// @Summary Upload verification document
// @Description Upload a verification document for the authenticated user's company
// @Tags company
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UploadDocumentRequest true "Upload Document Request"
// @Success 200 {object} VerificationDocument
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /companies/me/documents [post]
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
// @Summary Get verification status
// @Description Get the verification status of the authenticated user's company
// @Tags company
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} VerificationStatus
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /companies/me/verification-status [get]
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

func (h *Handler) RegisterRoutes(router fiber.Router) {
	companies := router.Group("/companies")
	companies.Post("", h.CreateCompany)
	companies.Put("/me", h.UpdateCompany)
	companies.Post("/me/documents", h.UploadDocument)
	companies.Get("/me/verification-status", h.GetVerificationStatus)
}
