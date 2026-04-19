package services

import (
	"fmt"
	"log"
	"time"

	"github.com/pentsecops/backend/internal/core/domain"
	"github.com/pentsecops/backend/pkg/websocket"
)

// WebSocketService handles WebSocket message processing and integration with use cases
type WebSocketService struct {
	hub                 *websocket.Hub
	messageUseCase      interface{}
	notificationUseCase interface{}
}

// NewWebSocketService creates a new WebSocket service
func NewWebSocketService(
	hub *websocket.Hub,
	messageUseCase interface{},
	notificationUseCase interface{},
) *WebSocketService {
	return &WebSocketService{
		hub:                 hub,
		messageUseCase:      messageUseCase,
		notificationUseCase: notificationUseCase,
	}
}

// ProcessMessage handles incoming WebSocket messages
func (s *WebSocketService) ProcessMessage(userID string, messageType string, data map[string]interface{}) error {
	switch messageType {
	case "message:send":
		return s.handleSendMessage(userID, data)

	case "message:delivered":
		return s.handleMessageDelivered(userID, data)

	case "message:read":
		return s.handleMessageRead(userID, data)

	case "typing":
		return s.handleTypingIndicator(userID, data)

	case "notification:read":
		return s.handleNotificationRead(userID, data)

	default:
		return fmt.Errorf("unknown message type: %s", messageType)
	}
}

// handleSendMessage processes a message send request
func (s *WebSocketService) handleSendMessage(senderID string, data map[string]interface{}) error {
	receiverID, ok := data["receiver_id"].(string)
	if !ok || receiverID == "" {
		return fmt.Errorf("missing or invalid receiver_id")
	}

	content, ok := data["content"].(string)
	if !ok || content == "" {
		return fmt.Errorf("missing or invalid content")
	}

	// Build event for receiver
	event := &websocket.Event{
		Type: "message:received",
		Data: map[string]interface{}{
			"sender_id": senderID,
			"content":   content,
			"timestamp": time.Now(),
		},
		Timestamp: time.Now(),
	}

	// Send to receiver if online
	if s.hub.IsUserOnline(receiverID) {
		if err := s.hub.SendEventToUser(receiverID, event); err != nil {
			log.Printf("[WebSocket] Failed to send message to receiver: %v", err)
		}
	}

	// Send confirmation to sender
	confirmEvent := &websocket.Event{
		Type: "message:sent",
		Data: map[string]interface{}{
			"receiver_id": receiverID,
			"content":     content,
			"timestamp":   time.Now(),
		},
		Timestamp: time.Now(),
	}
	if err := s.hub.SendEventToUser(senderID, confirmEvent); err != nil {
		log.Printf("[WebSocket] Failed to send confirmation to sender: %v", err)
	}

	return nil
}

// handleMessageDelivered marks a message as delivered
func (s *WebSocketService) handleMessageDelivered(userID string, data map[string]interface{}) error {
	messageID, ok := data["message_id"].(string)
	if !ok || messageID == "" {
		return fmt.Errorf("missing or invalid message_id")
	}

	// Broadcast delivery status
	event := &websocket.Event{
		Type: "message:delivered",
		Data: map[string]interface{}{
			"message_id": messageID,
		},
		Timestamp: time.Now(),
	}
	s.hub.BroadcastEvent(event)

	return nil
}

// handleMessageRead marks a message as read
func (s *WebSocketService) handleMessageRead(userID string, data map[string]interface{}) error {
	messageID, ok := data["message_id"].(string)
	if !ok || messageID == "" {
		return fmt.Errorf("missing or invalid message_id")
	}

	// Broadcast read status
	event := &websocket.Event{
		Type: "message:read",
		Data: map[string]interface{}{
			"message_id": messageID,
		},
		Timestamp: time.Now(),
	}
	s.hub.BroadcastEvent(event)

	return nil
}

// handleTypingIndicator handles typing indicator notifications
func (s *WebSocketService) handleTypingIndicator(userID string, data map[string]interface{}) error {
	receiverID, ok := data["receiver_id"].(string)
	if !ok || receiverID == "" {
		return fmt.Errorf("missing or invalid receiver_id")
	}

	isTyping, ok := data["is_typing"].(bool)
	if !ok {
		isTyping = true
	}

	// Send typing indicator to receiver
	event := &websocket.Event{
		Type: "typing",
		Data: map[string]interface{}{
			"user_id":   userID,
			"is_typing": isTyping,
		},
		Timestamp: time.Now(),
	}

	if err := s.hub.SendEventToUser(receiverID, event); err != nil {
		log.Printf("[WebSocket] Failed to send typing indicator: %v", err)
	}

	return nil
}

// handleNotificationRead processes notification read events
func (s *WebSocketService) handleNotificationRead(userID string, data map[string]interface{}) error {
	notificationID, ok := data["notification_id"].(string)
	if !ok || notificationID == "" {
		return fmt.Errorf("missing or invalid notification_id")
	}

	// Send confirmation
	event := &websocket.Event{
		Type: "notification:read_confirmed",
		Data: map[string]interface{}{
			"notification_id": notificationID,
		},
		Timestamp: time.Now(),
	}

	if err := s.hub.SendEventToUser(userID, event); err != nil {
		log.Printf("[WebSocket] Failed to send notification read confirmation: %v", err)
	}

	return nil
}

// BroadcastNotification sends a notification to specific user(s)
func (s *WebSocketService) BroadcastNotification(userID string, notification *domain.Notification) error {
	event := &websocket.Event{
		Type: "notification",
		Data: map[string]interface{}{
			"id":            notification.ID,
			"subject":       notification.Subject,
			"message":       notification.Message,
			"priority":      notification.Priority,
			"sender_id":     notification.SenderID,
			"receiver_type": notification.ReceiverType,
			"created_at":    notification.CreatedAt,
		},
		Timestamp: time.Now(),
	}

	return s.hub.SendEventToUser(userID, event)
}

// BroadcastToAllUsers broadcasts a message to all connected users
func (s *WebSocketService) BroadcastToAllUsers(event *websocket.Event) {
	s.hub.BroadcastEvent(event)
}

// IsUserOnline checks if a user is connected via WebSocket
func (s *WebSocketService) IsUserOnline(userID string) bool {
	return s.hub.IsUserOnline(userID)
}
