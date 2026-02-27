package review

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

type CreateReviewRequest struct {
	TargetEntityID string `json:"target_entity_id"`
	Rating         int    `json:"rating"`
	Comment        string `json:"comment"`
	TransactionID  string `json:"transaction_id"`
}

func (h *Handler) CreateReview(c *fiber.Ctx) error {
	var req CreateReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}

	userID := c.Locals("user_id").(string)

	review := &Review{
		AuthorID:       userID,
		TargetEntityID: req.TargetEntityID,
		Rating:         req.Rating,
		Comment:        req.Comment,
		TransactionID:  req.TransactionID,
	}

	if err := h.service.CreateReview(c.Context(), review); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(err)
	}

	return c.Status(http.StatusCreated).JSON(review)
}

func (h *Handler) GetReviews(c *fiber.Ctx) error {
	entityID := c.Params("entityID")
	// Pagination logic (simplified)
	page := 1
	limit := 10

	reviews, total, err := h.service.GetReviews(c.Context(), entityID, page, limit)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(err)
	}

	return c.JSON(fiber.Map{
		"items": reviews,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *Handler) GetReputation(c *fiber.Ctx) error {
	entityID := c.Params("entityID")
	score, err := h.service.GetReputation(c.Context(), entityID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(err)
	}
	return c.JSON(score)
}

func (h *Handler) RegisterRoutes(app fiber.Router) {
	reviews := app.Group("/reviews")
	reviews.Post("/", h.CreateReview)
	reviews.Get("/:entityID", h.GetReviews)
	reviews.Get("/:entityID/reputation", h.GetReputation)
}
