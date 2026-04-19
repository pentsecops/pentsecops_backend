package handlers

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"github.com/pentsecops/backend/pkg/middleware"
	"github.com/pentsecops/backend/pkg/services"
	ws "github.com/pentsecops/backend/pkg/websocket"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub          *ws.Hub
	tokenManager *middleware.TokenManager
	wsService    *services.WebSocketService
}

// WebSocketMessagePayload represents incoming WebSocket message
type WebSocketMessagePayload struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// WebSocketResponsePayload represents outgoing WebSocket message
type WebSocketResponsePayload struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(
	hub *ws.Hub,
	tokenManager *middleware.TokenManager,
	wsService *services.WebSocketService,
) *WebSocketHandler {
	return &WebSocketHandler{
		hub:          hub,
		tokenManager: tokenManager,
		wsService:    wsService,
	}
}

// HandleWebSocket handles WebSocket connection upgrade and message routing
func (h *WebSocketHandler) HandleWebSocket(c *fiber.Ctx) error {
	// Check if the connection is a WebSocket upgrade request
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return websocket.New(h.handleConnection)(c)
	}

	return c.Status(fiber.StatusUpgradeRequired).JSON(fiber.Map{
		"error": "WebSocket upgrade required",
	})
}

// handleConnection handles individual WebSocket connections
func (h *WebSocketHandler) handleConnection(c *websocket.Conn) {
	// Extract user ID from query parameter or token
	userID := c.Query("user_id")
	token := c.Query("token")

	// If no user_id, try to extract from token
	if userID == "" && token != "" {
		claims, err := h.tokenManager.VerifyToken(token)
		if err != nil {
			log.Printf("[WebSocket] Token verification failed: %v", err)
			c.WriteJSON(fiber.Map{
				"error": "Unauthorized",
				"type":  "error",
			})
			c.Close()
			return
		}
		userID = claims.UserID
	}

	if userID == "" {
		log.Println("[WebSocket] Missing user_id or valid token")
		c.WriteJSON(fiber.Map{
			"error": "Missing user_id or valid token",
			"type":  "error",
		})
		c.Close()
		return
	}

	// Initialize connection tracking
	socketID := uuid.New().String()
	log.Printf("[WebSocket] Client %s connected (socket: %s)", userID, socketID)

	// Create a mutex for this connection to protect concurrent writes
	connMutex := &sync.Mutex{}

	// Send connection confirmation
	connEvent := WebSocketResponsePayload{
		Type: "connection:established",
		Data: map[string]interface{}{
			"user_id":   userID,
			"socket_id": socketID,
			"message":   "Connected to WebSocket server",
		},
		Timestamp: time.Now(),
	}

	connMutex.Lock()
	if err := c.WriteJSON(connEvent); err != nil {
		connMutex.Unlock()
		log.Printf("[WebSocket] Failed to send connection confirmation: %v", err)
		return
	}
	connMutex.Unlock()

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run read pump in separate goroutine
	go h.readPump(c, userID, socketID, ctx, cancel, connMutex)

	// Run write pump (ping ticker) in separate goroutine
	go h.writePump(c, userID, ctx, connMutex)

	// Wait for context to be cancelled (when readPump exits)
	<-ctx.Done()
}

// readPump reads messages from the WebSocket connection
func (h *WebSocketHandler) readPump(c *websocket.Conn, userID, socketID string, ctx context.Context, cancel context.CancelFunc, connMutex *sync.Mutex) {
	defer cancel()

	// Set initial read deadline
	c.SetReadDeadline(time.Now().Add(90 * time.Second))

	// Set pong handler to reset read deadline
	c.SetPongHandler(func(appData string) error {
		log.Printf("[WebSocket] Pong received from %s", userID)
		c.SetReadDeadline(time.Now().Add(90 * time.Second))
		return nil
	})

	for {
		var payload WebSocketMessagePayload
		err := c.ReadJSON(&payload)
		if err != nil {
			if websocket.IsCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				log.Printf("[WebSocket] Client %s disconnected gracefully", userID)
			} else {
				log.Printf("[WebSocket] Read error for user %s: %v", userID, err)
			}
			return
		}

		// Reset read deadline on every message received
		c.SetReadDeadline(time.Now().Add(90 * time.Second))

		// Handle ping messages with JSON pong response
		if payload.Type == "ping" {
			pongEvent := WebSocketResponsePayload{
				Type:      "pong",
				Timestamp: time.Now(),
			}
			connMutex.Lock()
			if err := c.WriteJSON(pongEvent); err != nil {
				connMutex.Unlock()
				log.Printf("[WebSocket] Failed to send pong to %s: %v", userID, err)
				return
			}
			connMutex.Unlock()
			log.Printf("[WebSocket] Pong sent to %s in response to ping", userID)
			continue
		}

		// Process message
		if err := h.processMessage(userID, payload.Type, payload.Data); err != nil {
			log.Printf("[WebSocket] Error processing message: %v", err)
			errorEvent := WebSocketResponsePayload{
				Type:      "error",
				Error:     err.Error(),
				Timestamp: time.Now(),
			}
			connMutex.Lock()
			if err := c.WriteJSON(errorEvent); err != nil {
				connMutex.Unlock()
				log.Printf("[WebSocket] Failed to send error: %v", err)
				return
			}
			connMutex.Unlock()
		}
	}
}

// writePump sends ping frames to keep the connection alive
func (h *WebSocketHandler) writePump(c *websocket.Conn, userID string, ctx context.Context, connMutex *sync.Mutex) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send ping frame with mutex protection
			c.SetWriteDeadline(time.Now().Add(5 * time.Second))
			connMutex.Lock()
			if err := c.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				connMutex.Unlock()
				log.Printf("[WebSocket] Ping error for user %s: %v", userID, err)
				return
			}
			connMutex.Unlock()
			log.Printf("[WebSocket] Ping sent to %s", userID)

		case <-ctx.Done():
			return
		}
	}
}

// processMessage handles different message types
func (h *WebSocketHandler) processMessage(userID string, messageType string, data map[string]interface{}) error {
	switch messageType {
	case "message:send":
		return h.wsService.ProcessMessage(userID, "message:send", data)

	case "message:delivered":
		return h.wsService.ProcessMessage(userID, "message:delivered", data)

	case "message:read":
		return h.wsService.ProcessMessage(userID, "message:read", data)

	case "typing":
		return h.wsService.ProcessMessage(userID, "typing", data)

	default:
		return fmt.Errorf("unknown message type: %s", messageType)
	}
}

// BroadcastMessage broadcasts a message to all connected users
func (h *WebSocketHandler) BroadcastMessage(event *ws.Event) {
	h.hub.BroadcastEvent(event)
}

// SendMessageToUser sends a message to a specific user
func (h *WebSocketHandler) SendMessageToUser(userID string, event *ws.Event) error {
	return h.hub.SendEventToUser(userID, event)
}

// IsUserOnline checks if a user has any active WebSocket connections
func (h *WebSocketHandler) IsUserOnline(userID string) bool {
	return h.hub.IsUserOnline(userID)
}
