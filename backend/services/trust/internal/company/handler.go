package company

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/industrix/pkg/errors"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

type CreateCompanyRequest struct {
	Name    string `json:"name"`
	BIN     string `json:"bin"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Website string `json:"website"`
}

type UpdateCompanyRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Website string `json:"website"`
}

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
		return c.Status(http.StatusConflict).JSON(err)
	}

	return c.Status(http.StatusCreated).JSON(company)
}

func (h *Handler) GetCompany(c *fiber.Ctx) error {
	id := c.Params("id")
	company, err := h.service.GetCompany(c.Context(), id)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(err)
	}
	return c.JSON(company)
}

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

func (h *Handler) RegisterRoutes(app fiber.Router) {
	companies := app.Group("/companies")
	companies.Post("/", h.CreateCompany)
	companies.Get("/:id", h.GetCompany)
	companies.Put("/:id", h.UpdateCompany)
}
