package handlers

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pentsecops/backend/internal/core/domain"
	"github.com/pentsecops/backend/internal/core/usecases"
	"github.com/pentsecops/backend/pkg/services"
)

// NotificationHandler handles notification-related HTTP requests
type NotificationHandler struct {
	notificationUseCase *usecases.NotificationUseCase
	emailService        *services.EmailService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notificationUseCase *usecases.NotificationUseCase) *NotificationHandler {
	return &NotificationHandler{
		notificationUseCase: notificationUseCase,
	}
}

// NewNotificationHandlerWithEmail creates a new notification handler with email service
func NewNotificationHandlerWithEmail(notificationUseCase *usecases.NotificationUseCase, emailService *services.EmailService) *NotificationHandler {
	return &NotificationHandler{
		notificationUseCase: notificationUseCase,
		emailService:        emailService,
	}
}

// CreateNotificationRequest represents the request payload for creating a notification
type CreateNotificationRequest struct {
	ReceiverType string   `json:"receiver_type"` // "single", "multiple", "broadcast"
	ReceiverIDs  []string `json:"receiver_ids,omitempty"`
	Subject      string   `json:"subject"`
	Message      string   `json:"message"`
	Priority     string   `json:"priority"` // "low", "medium", "high", "critical"
	InApp        bool     `json:"in_app"`
	Email        bool     `json:"email"`
}

// NotificationResponse represents the response payload for a notification
type NotificationResponse struct {
	ID           string   `json:"id"`
	SenderID     *string  `json:"sender_id,omitempty"`
	ReceiverType string   `json:"receiver_type"`
	ReceiverIDs  []string `json:"receiver_ids,omitempty"`
	Subject      string   `json:"subject"`
	Message      string   `json:"message"`
	Priority     string   `json:"priority"`
	InApp        bool     `json:"in_app"`
	Email        bool     `json:"email"`
	CreatedAt    string   `json:"created_at"`
	DeliveredAt  *string  `json:"delivered_at,omitempty"`
}

// CreateNotification handles POST /api/v1/notifications
// Creates a notification for administrators or system
func (h *NotificationHandler) CreateNotification(c *fiber.Ctx) error {
	// Get authenticated user ID from context
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		log.Printf("[ERROR] CreateNotification - User not authenticated")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Parse request
	req := new(CreateNotificationRequest)
	if err := c.BodyParser(req); err != nil {
		log.Printf("[ERROR] CreateNotification - Failed to parse request: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request format",
		})
	}

	// Validate request
	if req.Subject == "" || req.Message == "" || req.ReceiverType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "subject, message, and receiver_type are required",
		})
	}

	// Convert receiver_type string to domain type
	receiverType := domain.NotificationReceiverType(req.ReceiverType)
	priority := domain.NotificationPriority(req.Priority)

	// Create notification
	notification, err := h.notificationUseCase.CreateNotification(
		userID,
		receiverType,
		req.ReceiverIDs,
		req.Subject,
		req.Message,
		priority,
		req.InApp,
		req.Email,
	)
	if err != nil {
		log.Printf("[ERROR] CreateNotification - Failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Create delivery logs
	channels := make([]domain.DeliveryChannel, 0)
	if req.InApp {
		channels = append(channels, domain.DeliveryChannelInApp)
	}
	if req.Email {
		channels = append(channels, domain.DeliveryChannelEmail)
	}

	if len(channels) > 0 {
		if err := h.notificationUseCase.CreateDeliveryLogs(notification.ID, channels); err != nil {
			log.Printf("[WARN] CreateNotification - Failed to create delivery logs: %v", err)
			// Continue even if delivery log creation fails
		}
	}

	// Send emails if requested
	if req.Email {
		go h.notificationUseCase.SendNotificationEmails(notification)
	}

	// Build response
	resp := h.notificationToResponse(notification)

	log.Printf("[SUCCESS] CreateNotification - Notification created: id=%s, type=%s", notification.ID, req.ReceiverType)
	return c.Status(fiber.StatusCreated).JSON(resp)
}

// GetUserNotifications handles GET /api/v1/notifications
// Retrieves notifications for authenticated user
func (h *NotificationHandler) GetUserNotifications(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Parse pagination parameters
	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Get notifications
	notifications, err := h.notificationUseCase.GetUserNotifications(userID, limit, offset)
	if err != nil {
		log.Printf("[ERROR] GetUserNotifications - Failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Convert to responses
	responses := make([]NotificationResponse, len(notifications))
	for i, notif := range notifications {
		responses[i] = h.notificationToResponse(notif)
	}

	return c.JSON(fiber.Map{
		"notifications": responses,
		"count":         len(responses),
		"limit":         limit,
		"offset":        offset,
	})
}

// GetNotificationStats handles GET /api/v1/notifications/:id/stats
// Returns delivery statistics for a notification
func (h *NotificationHandler) GetNotificationStats(c *fiber.Ctx) error {
	notificationID := c.Params("id")
	if notificationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "notification id is required",
		})
	}

	// Get notification to verify it exists
	_, err := h.notificationUseCase.GetNotificationByID(notificationID)
	if err != nil {
		log.Printf("[ERROR] GetNotificationStats - Notification not found: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "notification not found",
		})
	}

	// Get stats
	stats, err := h.notificationUseCase.GetNotificationStats(notificationID)
	if err != nil {
		log.Printf("[ERROR] GetNotificationStats - Failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"notification_id": notificationID,
		"stats":           stats,
	})
}

// DeleteNotification handles DELETE /api/v1/notifications/:id
// Deletes a notification
func (h *NotificationHandler) DeleteNotification(c *fiber.Ctx) error {
	notificationID := c.Params("id")
	if notificationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "notification id is required",
		})
	}

	// Delete notification
	if err := h.notificationUseCase.DeleteNotification(notificationID); err != nil {
		log.Printf("[ERROR] DeleteNotification - Failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	log.Printf("[SUCCESS] DeleteNotification - Notification deleted: %s", notificationID)
	return c.JSON(fiber.Map{
		"success": true,
		"message": "notification deleted",
	})
}

// Helper function to convert domain.Notification to NotificationResponse
func (h *NotificationHandler) notificationToResponse(notif *domain.Notification) NotificationResponse {
	resp := NotificationResponse{
		ID:           notif.ID,
		SenderID:     notif.SenderID,
		ReceiverType: string(notif.ReceiverType),
		ReceiverIDs:  notif.ReceiverIDs,
		Subject:      notif.Subject,
		Message:      notif.Message,
		Priority:     string(notif.Priority),
		InApp:        notif.InApp,
		Email:        notif.Email,
		CreatedAt:    notif.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if notif.DeliveredAt != nil {
		deliveredAt := notif.DeliveredAt.Format("2006-01-02T15:04:05Z07:00")
		resp.DeliveredAt = &deliveredAt
	}

	return resp
}
