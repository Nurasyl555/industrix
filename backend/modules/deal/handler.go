package deal

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/industrix/backend/pkg/errors"
	"github.com/industrix/backend/pkg/jwt"
)

// Handler handles deal HTTP requests. All routes require authentication —
// there's no anonymous browsing of deals.
type Handler struct {
	service   Service
	hub       *Hub
	jwtClient jwt.Client
}

// NewHandler creates a new deal handler
func NewHandler(service Service, hub *Hub, jwtClient jwt.Client) *Handler {
	return &Handler{service: service, hub: hub, jwtClient: jwtClient}
}

// RegisterRoutes registers all deal routes (all protected).
// "/my" is namespaced separately from "/:id" for the same reason as the
// listing module — see modules/listing/handler.go for why.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	deals := router.Group("/deals")
	deals.Post("/", h.CreateDeal)
	deals.Get("/:id", h.GetDeal)
	deals.Put("/:id/close", h.Close)
	deals.Get("/:id/messages", h.ListMessages)
	deals.Post("/:id/messages", h.PostMessage)

	my := router.Group("/my-deals")
	my.Get("/", h.ListMy)
}

// RegisterWebSocket registers the realtime thread socket. Mounted OUTSIDE the
// /api/v1 JWT middleware group because the WS handshake authenticates itself
// via the access_token cookie (see UpgradeWS), not the Authorization header.
func (h *Handler) RegisterWebSocket(app *fiber.App) {
	app.Use("/ws/deals/:id", h.UpgradeWS)
	app.Get("/ws/deals/:id", websocket.New(h.DealSocket))
}

func respondErr(c *fiber.Ctx, err error) error {
	if domainErr, ok := err.(*errors.Error); ok {
		return c.Status(errors.HTTPStatus(domainErr.Code)).JSON(domainErr)
	}
	return c.Status(http.StatusInternalServerError).JSON(errors.New(errors.CodeInternal, "Something went wrong"))
}

// CreateDeal godoc
// @Summary Inquire about a listing
// @Tags deals
// @Security BearerAuth
// @Param request body CreateDealRequest true "Inquiry details"
// @Success 201 {object} Deal
// @Router /deals [post]
func (h *Handler) CreateDeal(c *fiber.Ctx) error {
	var req CreateDealRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}

	userID := c.Locals("user_id").(string)
	d, err := h.service.CreateDeal(c.Context(), userID, req)
	if err != nil {
		return respondErr(c, err)
	}
	return c.Status(http.StatusCreated).JSON(d)
}

// GetDeal godoc
// @Summary Get a deal (buyer or seller only)
// @Tags deals
// @Security BearerAuth
// @Param id path string true "Deal ID"
// @Success 200 {object} DealView
// @Router /deals/{id} [get]
func (h *Handler) GetDeal(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	d, err := h.service.GetDeal(c.Context(), c.Params("id"), userID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(d)
}

// ListMy godoc
// @Summary List the current user's deals, as buyer or seller
// @Tags deals
// @Security BearerAuth
// @Success 200 {array} DealView
// @Router /my-deals [get]
func (h *Handler) ListMy(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	deals, err := h.service.ListMy(c.Context(), userID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(deals)
}

// Close godoc
// @Summary Close a deal
// @Tags deals
// @Security BearerAuth
// @Param id path string true "Deal ID"
// @Success 200
// @Router /deals/{id}/close [put]
func (h *Handler) Close(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if err := h.service.Close(c.Context(), c.Params("id"), userID); err != nil {
		return respondErr(c, err)
	}
	return c.SendStatus(http.StatusOK)
}

// ListMessages godoc
// @Summary Get the message thread of a deal (buyer or seller only)
// @Tags deals
// @Security BearerAuth
// @Param id path string true "Deal ID"
// @Success 200 {array} DealMessage
// @Router /deals/{id}/messages [get]
func (h *Handler) ListMessages(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	msgs, err := h.service.ListMessages(c.Context(), c.Params("id"), userID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(msgs)
}

// PostMessage godoc
// @Summary Reply within a deal thread
// @Tags deals
// @Security BearerAuth
// @Param id path string true "Deal ID"
// @Param request body PostMessageRequest true "Message body"
// @Success 201 {object} DealMessage
// @Router /deals/{id}/messages [post]
func (h *Handler) PostMessage(c *fiber.Ctx) error {
	var req PostMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errors.New(errors.CodeValidation, "Invalid request body"))
	}
	userID := c.Locals("user_id").(string)
	dealID := c.Params("id")
	msg, err := h.service.PostMessage(c.Context(), dealID, userID, req.Body)
	if err != nil {
		return respondErr(c, err)
	}
	// Fan out to anyone watching this thread over WebSocket.
	if payload, err := json.Marshal(msg); err == nil {
		h.hub.Broadcast(dealID, payload)
	}
	return c.Status(http.StatusCreated).JSON(msg)
}

// UpgradeWS is the pre-upgrade gate: it authenticates the WebSocket handshake
// from the access_token cookie (browsers can't set Authorization headers on
// WS, but cookies are host-scoped and ride along to :8080 automatically) and
// stashes the user id for the socket handler.
func (h *Handler) UpgradeWS(c *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}
	token := c.Cookies("access_token")
	if token == "" {
		return fiber.ErrUnauthorized
	}
	claims, err := h.jwtClient.ParseClaims(token)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	c.Locals("user_id", claims.UserID)
	return c.Next()
}

// DealSocket is the WebSocket handler for a single deal thread. It verifies
// the user is a participant, joins the room, and relays inbound frames
// ({"body": "..."}) through the same PostMessage service path as REST, then
// broadcasts the stored message to the room.
func (h *Handler) DealSocket(c *websocket.Conn) {
	ctx := context.Background()
	dealID := c.Params("id")
	userID, _ := c.Locals("user_id").(string)

	// Authorize: must be a participant. Reuse the service check.
	if _, err := h.service.GetDeal(ctx, dealID, userID); err != nil {
		_ = c.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "not a participant"))
		_ = c.Close()
		return
	}

	h.hub.Join(dealID, c)
	defer h.hub.Leave(dealID, c)

	for {
		_, data, err := c.ReadMessage()
		if err != nil {
			break // client disconnected
		}
		var in PostMessageRequest
		if json.Unmarshal(data, &in) != nil || in.Body == "" {
			continue
		}
		msg, err := h.service.PostMessage(ctx, dealID, userID, in.Body)
		if err != nil {
			continue // e.g. deal closed — ignore silently
		}
		if payload, err := json.Marshal(msg); err == nil {
			h.hub.Broadcast(dealID, payload)
		}
	}
}
