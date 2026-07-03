package deal

import (
	"sync"

	"github.com/gofiber/contrib/websocket"
)

// Hub is an in-memory pub/sub for deal conversation rooms. Each deal id maps
// to the set of live WebSocket connections currently viewing that thread.
// A new message is fanned out to everyone in the room — including the sender,
// which is how the sender's own UI gets the server-assigned id/timestamp.
//
// In-memory means it does not survive a restart and does not fan out across
// multiple backend replicas. For the MVP (single binary) that's fine; the
// architecture.md Chat module is where this graduates to Redis pub/sub.
type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[*websocket.Conn]bool
}

func NewHub() *Hub {
	return &Hub{rooms: make(map[string]map[*websocket.Conn]bool)}
}

func (h *Hub) Join(dealID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[dealID] == nil {
		h.rooms[dealID] = make(map[*websocket.Conn]bool)
	}
	h.rooms[dealID][conn] = true
}

func (h *Hub) Leave(dealID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if room := h.rooms[dealID]; room != nil {
		delete(room, conn)
		if len(room) == 0 {
			delete(h.rooms, dealID)
		}
	}
}

// Broadcast sends payload to every connection in the room. Dead connections
// are dropped so a stale socket can't wedge the room.
func (h *Hub) Broadcast(dealID string, payload []byte) {
	h.mu.RLock()
	conns := make([]*websocket.Conn, 0, len(h.rooms[dealID]))
	for c := range h.rooms[dealID] {
		conns = append(conns, c)
	}
	h.mu.RUnlock()

	for _, c := range conns {
		if err := c.WriteMessage(websocket.TextMessage, payload); err != nil {
			h.Leave(dealID, c)
			_ = c.Close()
		}
	}
}
