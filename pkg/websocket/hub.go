package websocket

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Event types for WebSocket messages
const (
	EventTypeMessage          = "message"
	EventTypeMessageDelivered = "message:delivered"
	EventTypeMessageRead      = "message:read"
	EventTypeNotification     = "notification"
	EventTypeTypingIndicator  = "typing"
	EventTypePing             = "ping"
	EventTypePong             = "pong"
)

// Connection represents a single WebSocket connection
type Connection struct {
	UserID   string
	SocketID string
	Conn     *websocket.Conn
	Hub      *Hub
	Send     chan interface{}
	Done     chan bool
	LastPong time.Time
	Mu       sync.Mutex
}

// Event represents a WebSocket event
type Event struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// Hub manages all active WebSocket connections
type Hub struct {
	clients    map[string][]*Connection // userID -> []*Connection
	broadcast  chan *Event
	register   chan *Connection
	unregister chan *Connection
	mu         sync.RWMutex

	// Ping-Pong config
	pingInterval time.Duration
	pongWait     time.Duration
}

// NewHub creates a new WebSocket hub
func NewHub(pingInterval, pongWait time.Duration) *Hub {
	return &Hub{
		clients:      make(map[string][]*Connection),
		broadcast:    make(chan *Event, 256),
		register:     make(chan *Connection),
		unregister:   make(chan *Connection),
		pingInterval: pingInterval,
		pongWait:     pongWait,
	}
}

// Run starts the hub event loop
func (h *Hub) Run() {
	ticker := time.NewTicker(h.pingInterval)
	defer ticker.Stop()

	for {
		select {
		case conn := <-h.register:
			h.registerConnection(conn)
			log.Printf("[WebSocket] User %s connected (socket: %s)", conn.UserID, conn.SocketID)

		case conn := <-h.unregister:
			h.unregisterConnection(conn)
			log.Printf("[WebSocket] User %s disconnected (socket: %s)", conn.UserID, conn.SocketID)

		case event := <-h.broadcast:
			h.broadcastEvent(event)

		case <-ticker.C:
			h.sendPings()
		}
	}
}

// registerConnection registers a new connection
func (h *Hub) registerConnection(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.clients[conn.UserID]; !exists {
		h.clients[conn.UserID] = make([]*Connection, 0)
	}
	h.clients[conn.UserID] = append(h.clients[conn.UserID], conn)
}

// unregisterConnection removes a connection
func (h *Hub) unregisterConnection(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conns, exists := h.clients[conn.UserID]; exists {
		for i, c := range conns {
			if c.SocketID == conn.SocketID {
				// Remove connection from slice
				h.clients[conn.UserID] = append(conns[:i], conns[i+1:]...)
				break
			}
		}

		// Remove user entry if no more connections
		if len(h.clients[conn.UserID]) == 0 {
			delete(h.clients, conn.UserID)
		}
	}

	close(conn.Send)
}

// GetUserConnections returns all connections for a user
func (h *Hub) GetUserConnections(userID string) []*Connection {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns, exists := h.clients[userID]
	if !exists {
		return nil
	}

	result := make([]*Connection, len(conns))
	copy(result, conns)
	return result
}

// IsUserOnline checks if a user has any active connections
func (h *Hub) IsUserOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns, exists := h.clients[userID]
	return exists && len(conns) > 0
}

// SendEventToUser sends an event to a specific user (all their connections)
func (h *Hub) SendEventToUser(userID string, event *Event) error {
	h.mu.RLock()
	conns, exists := h.clients[userID]
	h.mu.RUnlock()

	if !exists || len(conns) == 0 {
		return fmt.Errorf("user %s is offline", userID)
	}

	for _, conn := range conns {
		select {
		case conn.Send <- event:
		case <-time.After(time.Second):
			log.Printf("[WebSocket] Send timeout for user %s", userID)
		}
	}

	return nil
}

// Register registers a new connection with the hub
func (h *Hub) Register(conn *Connection) {
	select {
	case h.register <- conn:
	case <-time.After(time.Second):
		log.Printf("[WebSocket] Register timeout for user %s", conn.UserID)
	}
}

// Unregister removes a connection from the hub
func (h *Hub) Unregister(conn *Connection) {
	select {
	case h.unregister <- conn:
	case <-time.After(time.Second):
		log.Printf("[WebSocket] Unregister timeout for user %s", conn.UserID)
	}
}

// BroadcastEvent broadcasts an event to all connected users
func (h *Hub) BroadcastEvent(event *Event) {
	select {
	case h.broadcast <- event:
	default:
		log.Printf("[WebSocket] Broadcast channel full, dropping event")
	}
}

// broadcastEvent sends an event to all clients
func (h *Hub) broadcastEvent(event *Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, conns := range h.clients {
		for _, conn := range conns {
			select {
			case conn.Send <- event:
			default:
				log.Printf("[WebSocket] Send buffer full for user %s", conn.UserID)
			}
		}
	}
}

// sendPings sends ping frames to all clients
func (h *Hub) sendPings() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for userID, conns := range h.clients {
		for i, conn := range conns {
			pingEvent := &Event{
				Type:      EventTypePing,
				Timestamp: time.Now(),
			}

			// Non-blocking send with timeout
			select {
			case conn.Send <- pingEvent:
			case <-time.After(100 * time.Millisecond):
				// If send buffers are full, mark connection for removal
				log.Printf("[WebSocket] Ping timeout for user %s, removing connection", userID)
				conns[i] = nil
			}
		}

		// Clean up nil connections
		validConns := make([]*Connection, 0)
		for _, c := range conns {
			if c != nil {
				validConns = append(validConns, c)
			}
		}
		h.clients[userID] = validConns
	}
}

// NewConnection creates a new WebSocket connection
func NewConnection(userID, socketID string, ws *websocket.Conn, hub *Hub) *Connection {
	return &Connection{
		UserID:   userID,
		SocketID: socketID,
		Conn:     ws,
		Hub:      hub,
		Send:     make(chan interface{}, 256),
		Done:     make(chan bool),
		LastPong: time.Now(),
	}
}

// ReadPump reads messages from the WebSocket
func (c *Connection) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(c.Hub.pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Mu.Lock()
		c.LastPong = time.Now()
		c.Mu.Unlock()
		c.Conn.SetReadDeadline(time.Now().Add(c.Hub.pongWait))
		return nil
	})

	for {
		var msg map[string]interface{}
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WebSocket] Read error for user %s: %v", c.UserID, err)
			}
			return
		}

		// Handle received events
		if msgType, ok := msg["type"].(string); ok {
			switch msgType {
			case EventTypePong:
				c.Mu.Lock()
				c.LastPong = time.Now()
				c.Mu.Unlock()
			case EventTypeMessage:
				// Message will be handled by application logic
				// Broadcast to target or specific user
			case EventTypeTypingIndicator:
				// Handle typing indicator
			}
		}
	}
}

// WritePump writes messages to the WebSocket
func (c *Connection) WritePump() {
	ticker := time.NewTicker(c.Hub.pingInterval)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case event, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

			if !ok {
				// Channel closed
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteJSON(event); err != nil {
				log.Printf("[WebSocket] Write error for user %s: %v", c.UserID, err)
				return
			}

		case <-c.Done:
			return
		}
	}
}

// Close closes the connection
func (c *Connection) Close() {
	select {
	case c.Done <- true:
	default:
	}
}
