package booking

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/industrix/backend/pkg/errors"
	"github.com/industrix/backend/platform/httperr"
)

// Handler handles booking HTTP requests.
type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterPublicRoutes exposes read-only availability so the rental calendar
// can show which dates are taken without requiring a login.
func (h *Handler) RegisterPublicRoutes(router fiber.Router) {
	router.Get("/listings/:id/booked-dates", h.BookedDates)
	router.Get("/listings/:id/quote", h.Quote)
}

// RegisterProtectedRoutes registers the auth-gated booking actions.
func (h *Handler) RegisterProtectedRoutes(router fiber.Router) {
	bookings := router.Group("/bookings")
	bookings.Post("/", h.Create)
	bookings.Put("/:id/cancel", h.Cancel)

	router.Get("/my-bookings", h.ListMine)
}

// respondErr maps a service error to its HTTP response. See platform/httperr —
// unexpected errors are logged there before the generic 500 goes out.
func respondErr(c *fiber.Ctx, err error) error { return httperr.Respond(c, err) }

// BookedDates godoc
// @Summary Confirmed booked date ranges for a listing
// @Tags bookings
// @Param id path string true "Listing ID"
// @Success 200 {array} DateRange
// @Router /listings/{id}/booked-dates [get]
func (h *Handler) BookedDates(c *fiber.Ctx) error {
	ranges, err := h.service.BookedRanges(c.Context(), c.Params("id"))
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(ranges)
}

// Quote godoc
// @Summary Estimate rental cost for a listing over a date range
// @Tags bookings
// @Param id path string true "Listing ID"
// @Param start query string true "Start date (YYYY-MM-DD)"
// @Param end query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} Quote
// @Router /listings/{id}/quote [get]
func (h *Handler) Quote(c *fiber.Ctx) error {
	q, err := h.service.Quote(c.Context(), c.Params("id"), c.Query("start"), c.Query("end"))
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(q)
}

// Create godoc
// @Summary Book a rental listing for a date range
// @Tags bookings
// @Security BearerAuth
// @Param request body CreateBookingRequest true "Booking dates"
// @Success 201 {object} Booking
// @Router /bookings [post]
func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateBookingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}
	userID := c.Locals("user_id").(string)
	b, err := h.service.CreateBooking(c.Context(), userID, req)
	if err != nil {
		return respondErr(c, err)
	}
	return c.Status(http.StatusCreated).JSON(b)
}

// ListMine godoc
// @Summary List the current user's bookings
// @Tags bookings
// @Security BearerAuth
// @Success 200 {array} Booking
// @Router /my-bookings [get]
func (h *Handler) ListMine(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	items, err := h.service.ListMine(c.Context(), userID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(items)
}

// Cancel godoc
// @Summary Cancel a booking (renter or owner)
// @Tags bookings
// @Security BearerAuth
// @Param id path string true "Booking ID"
// @Success 200
// @Router /bookings/{id}/cancel [put]
func (h *Handler) Cancel(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if err := h.service.Cancel(c.Context(), c.Params("id"), userID); err != nil {
		return respondErr(c, err)
	}
	return c.SendStatus(http.StatusOK)
}
