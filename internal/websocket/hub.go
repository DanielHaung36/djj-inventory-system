// internal/websocket/hub.go
package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Hub manages client connections and broadcasts messages
// supports multiple topics (e.g. "customers", "quotes", "orders", "inventory")
type Hub struct {
	mu      sync.Mutex
	clients map[string]map[*websocket.Conn]bool // topic -> conns
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]map[*websocket.Conn]bool)}
}

// Register adds a connection for a given topic
func (h *Hub) Register(topic string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[topic] == nil {
		h.clients[topic] = make(map[*websocket.Conn]bool)
	}
	h.clients[topic][conn] = true
}

// Unregister removes a connection
func (h *Hub) Unregister(topic string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns := h.clients[topic]; conns != nil {
		delete(conns, conn)
		conn.Close()
	}
}

// Broadcast sends a message to all clients subscribed to a topic
func (h *Hub) Broadcast(topic string, message []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.clients[topic] {
		conn.WriteMessage(websocket.TextMessage, message)
	}
}
