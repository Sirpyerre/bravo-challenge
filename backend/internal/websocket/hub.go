package websocket

import "sync"

// Hub mantiene el conjunto de clientes activos agrupados por user_id.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*Client]bool // userID → set of clients
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*Client]bool),
	}
}

func (h *Hub) Register(userID string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[userID] == nil {
		h.clients[userID] = make(map[*Client]bool)
	}
	h.clients[userID][c] = true
}

func (h *Hub) Unregister(userID string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.clients[userID]; ok {
		delete(conns, c)
		if len(conns) == 0 {
			delete(h.clients, userID)
		}
	}
}

// BroadcastToUser envía msg a todos los clientes del usuario.
func (h *Hub) BroadcastToUser(userID string, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients[userID] {
		c.send <- msg
	}
}
