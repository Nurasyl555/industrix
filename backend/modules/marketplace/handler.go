package marketplace

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/industrix/backend/pkg/errors"
)

// Handler handles marketplace HTTP requests
type Handler struct {
	service Service
}

// NewHandler creates a new marketplace handler
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers all marketplace routes (all protected)
func (h *Handler) RegisterRoutes(router fiber.Router) {
	reviews := router.Group("/reviews")
	reviews.Post("/", h.CreateReview)
	reviews.Get("/:entityID", h.GetReviews)
	reviews.Get("/:entityID/reputation", h.GetReputation)
}

// CreateReview godoc
// @Summary Create a review
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateReviewRequest true "Review details"
// @Success 201 {object} Review
// @Router /reviews [post]
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
		if domainErr, ok := err.(*errors.Error); ok {
			return c.Status(errors.HTTPStatus(domainErr.Code)).JSON(domainErr)
		}
		return c.Status(http.StatusInternalServerError).JSON(errors.New(errors.CodeInternal, "Failed to create review"))
	}

	return c.Status(http.StatusCreated).JSON(review)
}

// GetReviews godoc
// @Summary List reviews
// @Tags reviews
// @Param entityID path string true "Entity ID"
// @Success 200 {object} map[string]interface{}
// @Router /reviews/{entityID} [get]
func (h *Handler) GetReviews(c *fiber.Ctx) error {
	entityID := c.Params("entityID")
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

// GetReputation godoc
// @Summary Get reputation score
// @Tags reviews
// @Param entityID path string true "Entity ID"
// @Success 200 {object} ReputationScore
// @Router /reviews/{entityID}/reputation [get]
func (h *Handler) GetReputation(c *fiber.Ctx) error {
	entityID := c.Params("entityID")
	score, err := h.service.GetReputation(c.Context(), entityID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(err)
	}
	return c.JSON(score)
}
