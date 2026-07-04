package admin

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/industrix/backend/modules/integrity"
	"github.com/industrix/backend/modules/listing"
	"github.com/industrix/backend/pkg/errors"
)

// Handler exposes the admin moderation surface. Admin is a cross-cutting
// aggregator that sits above the domain modules, so it's allowed to depend on
// their services directly (nothing depends on admin in turn).
type Handler struct {
	integrity integrity.Service
	listing   listing.Service
}

// NewHandler wires the admin handler with the services it moderates.
func NewHandler(integritySvc integrity.Service, listingSvc listing.Service) *Handler {
	return &Handler{integrity: integritySvc, listing: listingSvc}
}

// RegisterRoutes mounts /admin/* — the caller must apply an admin-role
// middleware to the router group it passes in.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	admin := router.Group("/admin")

	admin.Get("/companies", h.ListCompanies)
	admin.Put("/companies/:id/verify", h.VerifyCompany)
	admin.Put("/companies/:id/reject", h.RejectCompany)

	admin.Get("/listings/moderation", h.ListModerationQueue)
	admin.Put("/listings/:id/approve", h.ApproveListing)
	admin.Put("/listings/:id/reject", h.RejectListing)
}

func respondErr(c *fiber.Ctx, err error) error {
	if domainErr, ok := err.(*errors.Error); ok {
		return c.Status(errors.HTTPStatus(domainErr.Code)).JSON(domainErr)
	}
	return c.Status(http.StatusInternalServerError).JSON(errors.New(errors.CodeInternal, "Something went wrong"))
}

type noteRequest struct {
	Note string `json:"note"`
}

// ListCompanies godoc
// @Summary [Admin] List companies, optionally filtered by ?status=pending
// @Tags admin
// @Security BearerAuth
// @Success 200 {array} integrity.Company
// @Router /admin/companies [get]
func (h *Handler) ListCompanies(c *fiber.Ctx) error {
	companies, err := h.integrity.ListCompaniesByStatus(c.Context(), c.Query("status"))
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(companies)
}

// VerifyCompany godoc
// @Summary [Admin] Verify a company
// @Tags admin
// @Security BearerAuth
// @Param id path string true "Company ID"
// @Router /admin/companies/{id}/verify [put]
func (h *Handler) VerifyCompany(c *fiber.Ctx) error {
	if err := h.integrity.SetCompanyStatus(c.Context(), c.Params("id"), integrity.StatusVerified, ""); err != nil {
		return respondErr(c, err)
	}
	return c.SendStatus(http.StatusOK)
}

// RejectCompany godoc
// @Summary [Admin] Reject a company (with a note)
// @Tags admin
// @Security BearerAuth
// @Param id path string true "Company ID"
// @Param request body noteRequest false "Reviewer note"
// @Router /admin/companies/{id}/reject [put]
func (h *Handler) RejectCompany(c *fiber.Ctx) error {
	var req noteRequest
	_ = c.BodyParser(&req)
	if err := h.integrity.SetCompanyStatus(c.Context(), c.Params("id"), integrity.StatusRejected, req.Note); err != nil {
		return respondErr(c, err)
	}
	return c.SendStatus(http.StatusOK)
}

// ListModerationQueue godoc
// @Summary [Admin] Listings awaiting moderation
// @Tags admin
// @Security BearerAuth
// @Success 200 {array} listing.ListingView
// @Router /admin/listings/moderation [get]
func (h *Handler) ListModerationQueue(c *fiber.Ctx) error {
	items, err := h.listing.ListForModeration(c.Context())
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(items)
}

// ApproveListing godoc
// @Summary [Admin] Approve a listing → active
// @Tags admin
// @Security BearerAuth
// @Param id path string true "Listing ID"
// @Router /admin/listings/{id}/approve [put]
func (h *Handler) ApproveListing(c *fiber.Ctx) error {
	if err := h.listing.Approve(c.Context(), c.Params("id")); err != nil {
		return respondErr(c, err)
	}
	return c.SendStatus(http.StatusOK)
}

// RejectListing godoc
// @Summary [Admin] Reject a listing → rejected
// @Tags admin
// @Security BearerAuth
// @Param id path string true "Listing ID"
// @Router /admin/listings/{id}/reject [put]
func (h *Handler) RejectListing(c *fiber.Ctx) error {
	if err := h.listing.Reject(c.Context(), c.Params("id")); err != nil {
		return respondErr(c, err)
	}
	return c.SendStatus(http.StatusOK)
}
