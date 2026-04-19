package usecases

import (
	"fmt"
	"log"
	"time"

	"github.com/pentsecops/backend/internal/core/domain"
	"github.com/pentsecops/backend/pkg/services"
)

// NotificationUseCase handles business logic for notifications
type NotificationUseCase struct {
	notificationRepo domain.NotificationRepository
	deliveryLogRepo  domain.NotificationDeliveryLogRepository
	userRepo         domain.UserRepository
	adminRepo        domain.AdminRepository
	emailService     interface{} // *services.EmailService
}

// NewNotificationUseCase creates a new notification use case
func NewNotificationUseCase(
	notificationRepo domain.NotificationRepository,
	deliveryLogRepo domain.NotificationDeliveryLogRepository,
	userRepo domain.UserRepository,
	adminRepo domain.AdminRepository,
) *NotificationUseCase {
	return &NotificationUseCase{
		notificationRepo: notificationRepo,
		deliveryLogRepo:  deliveryLogRepo,
		userRepo:         userRepo,
		adminRepo:        adminRepo,
	}
}

// NewNotificationUseCaseWithEmail creates a new notification use case with email service
func NewNotificationUseCaseWithEmail(
	notificationRepo domain.NotificationRepository,
	deliveryLogRepo domain.NotificationDeliveryLogRepository,
	userRepo domain.UserRepository,
	adminRepo domain.AdminRepository,
	emailService interface{},
) *NotificationUseCase {
	return &NotificationUseCase{
		notificationRepo: notificationRepo,
		deliveryLogRepo:  deliveryLogRepo,
		userRepo:         userRepo,
		adminRepo:        adminRepo,
		emailService:     emailService,
	}
}

// CreateNotification creates a new notification
// receiverIDs can be nil for broadcast (receiverType = broadcast)
func (uc *NotificationUseCase) CreateNotification(
	senderID string,
	receiverType domain.NotificationReceiverType,
	receiverIDs []string,
	subject string,
	message string,
	priority domain.NotificationPriority,
	inApp bool,
	email bool,
) (*domain.Notification, error) {
	if senderID == "" {
		return nil, fmt.Errorf("sender_id is required")
	}

	if subject == "" || message == "" {
		return nil, fmt.Errorf("subject and message are required")
	}

	// Verify sender exists - check both users and admins
	_, userErr := uc.userRepo.GetUserByID(senderID)
	_, adminErr := uc.adminRepo.GetAdminByID(senderID)

	if userErr != nil && adminErr != nil {
		log.Printf("[ERROR] CreateNotification - Sender not found in users or admins: %v", userErr)
		return nil, fmt.Errorf("sender not found")
	}

	// Validate receiver IDs based on receiver type
	if receiverType == domain.NotificationReceiverTypeSingle || receiverType == domain.NotificationReceiverTypeMultiple {
		if len(receiverIDs) == 0 {
			return nil, fmt.Errorf("receiver_ids required for %s receiver type", receiverType)
		}

		// Verify all receivers exist - check both users and admins
		for _, receiverID := range receiverIDs {
			_, userErr := uc.userRepo.GetUserByID(receiverID)
			_, adminErr := uc.adminRepo.GetAdminByID(receiverID)

			if userErr != nil && adminErr != nil {
				log.Printf("[ERROR] CreateNotification - Receiver not found: %s", receiverID)
				return nil, fmt.Errorf("receiver %s not found", receiverID)
			}
		}
	}

	// Create notification
	senderIDPtr := &senderID
	notification := &domain.Notification{
		SenderID:     senderIDPtr,
		ReceiverType: receiverType,
		ReceiverIDs:  domain.StringArray(receiverIDs),
		Subject:      subject,
		Message:      message,
		Priority:     priority,
		InApp:        inApp,
		Email:        email,
		CreatedAt:    time.Now(),
	}

	// Save to database
	savedNotification, err := uc.notificationRepo.CreateNotification(notification)
	if err != nil {
		log.Printf("[ERROR] CreateNotification - Failed to save notification: %v", err)
		return nil, fmt.Errorf("failed to save notification: %w", err)
	}

	log.Printf("[SUCCESS] CreateNotification - Notification created: id=%s, priority=%s, type=%s",
		savedNotification.ID, priority, receiverType)

	return savedNotification, nil
}

// CreateDeliveryLogs creates delivery log entries for a notification
func (uc *NotificationUseCase) CreateDeliveryLogs(notificationID string, channels []domain.DeliveryChannel) error {
	if notificationID == "" {
		return fmt.Errorf("notification_id is required")
	}

	// Get notification to find receivers
	notification, err := uc.notificationRepo.GetNotificationByID(notificationID)
	if err != nil {
		log.Printf("[ERROR] CreateDeliveryLogs - Notification not found: %v", err)
		return fmt.Errorf("notification not found")
	}

	// Determine receiver IDs based on notification type
	var receiverIDs []string
	if notification.ReceiverType == domain.NotificationReceiverTypeBroadcast {
		// For broadcast, would fetch all users - simplified for now
		log.Printf("[INFO] CreateDeliveryLogs - Broadcast notification skipped user fetch")
		receiverIDs = notification.ReceiverIDs
	} else {
		receiverIDs = notification.ReceiverIDs
	}

	// Create delivery log for each receiver and channel
	for _, receiverID := range receiverIDs {
		for _, channel := range channels {
			now := time.Now()
			deliveryLog := &domain.NotificationDeliveryLog{
				NotificationID: notificationID,
				ReceiverID:     receiverID,
				Channel:        channel,
				Status:         domain.DeliveryStatusPending,
				AttemptedAt:    &now,
			}

			_, err := uc.deliveryLogRepo.CreateDeliveryLog(deliveryLog)
			if err != nil {
				log.Printf("[ERROR] CreateDeliveryLogs - Failed to create delivery log: %v", err)
				// Continue with other receivers
			}
		}
	}

	log.Printf("[SUCCESS] CreateDeliveryLogs - Delivery logs created for notification: %s", notificationID)
	return nil
}

// SendNotificationEmails sends emails for a notification to all receivers
func (uc *NotificationUseCase) SendNotificationEmails(notification *domain.Notification) error {
	if notification == nil {
		return fmt.Errorf("notification is required")
	}

	// Check if email service is available
	if uc.emailService == nil {
		log.Printf("[WARN] SendNotificationEmails - Email service not available")
		return nil // Don't fail if email service is not configured
	}

	emailSvc, ok := uc.emailService.(*services.EmailService)
	if !ok || emailSvc == nil {
		log.Printf("[WARN] SendNotificationEmails - Email service not properly initialized")
		return nil
	}

	// Get receiver emails
	receiverIDs := notification.ReceiverIDs
	if len(receiverIDs) == 0 {
		log.Printf("[INFO] SendNotificationEmails - No receivers for notification: %s", notification.ID)
		return nil
	}

	// Send email to each receiver
	for _, receiverID := range receiverIDs {
		// Get receiver email - try users first, then admins
		var receiverEmail string
		user, err := uc.userRepo.GetUserByID(receiverID)
		if err == nil && user != nil {
			receiverEmail = user.Email
		} else {
			// Try to get from admin table
			admin, err := uc.adminRepo.GetAdminByID(receiverID)
			if err == nil && admin != nil {
				receiverEmail = admin.Email
			}
		}

		if receiverEmail == "" {
			log.Printf("[WARN] SendNotificationEmails - Could not find email for receiver: %s", receiverID)
			continue
		}

		// Compose email
		subject := notification.Subject
		message := notification.Message

		// Send email asynchronously
		go func(email, emailSubject, emailMessage string) {
			err := emailSvc.SendEmail(email, emailSubject, emailMessage, "")
			if err != nil {
				log.Printf("[ERROR] SendNotificationEmails - Failed to send email to %s: %v", email, err)
			} else {
				log.Printf("[SUCCESS] SendNotificationEmails - Email sent to %s", email)
			}
		}(receiverEmail, subject, message)
	}

	return nil
}

// GetNotificationByID retrieves a notification by ID
func (uc *NotificationUseCase) GetNotificationByID(notificationID string) (*domain.Notification, error) {
	if notificationID == "" {
		return nil, fmt.Errorf("notification_id is required")
	}

	notification, err := uc.notificationRepo.GetNotificationByID(notificationID)
	if err != nil {
		log.Printf("[ERROR] GetNotificationByID - Failed to retrieve notification: %v", err)
		return nil, fmt.Errorf("failed to retrieve notification: %w", err)
	}

	return notification, nil
}

// GetUserNotifications retrieves notifications for a specific user
func (uc *NotificationUseCase) GetUserNotifications(userID string, limit, offset int) ([]*domain.Notification, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	if offset < 0 {
		offset = 0
	}

	notifications, err := uc.notificationRepo.GetNotificationsByReceiver(userID, limit, offset)
	if err != nil {
		log.Printf("[ERROR] GetUserNotifications - Failed to retrieve notifications: %v", err)
		return nil, fmt.Errorf("failed to retrieve notifications: %w", err)
	}

	return notifications, nil
}

// GetPendingDeliveryLogs retrieves pending delivery logs for a specific channel
func (uc *NotificationUseCase) GetPendingDeliveryLogs(channel domain.DeliveryChannel, limit int) ([]*domain.NotificationDeliveryLog, error) {
	logs, err := uc.deliveryLogRepo.GetPendingDeliveryLogs(string(channel), limit)
	if err != nil {
		log.Printf("[ERROR] GetPendingDeliveryLogs - Failed to retrieve logs: %v", err)
		return nil, fmt.Errorf("failed to retrieve delivery logs: %w", err)
	}

	return logs, nil
}

// MarkDeliveryAsDelivered marks a delivery log as successfully delivered
func (uc *NotificationUseCase) MarkDeliveryAsDelivered(deliveryLogID string) error {
	if deliveryLogID == "" {
		return fmt.Errorf("delivery_log_id is required")
	}

	if err := uc.deliveryLogRepo.MarkAsDelivered(deliveryLogID); err != nil {
		log.Printf("[ERROR] MarkDeliveryAsDelivered - Failed to update: %v", err)
		return fmt.Errorf("failed to mark as delivered: %w", err)
	}

	return nil
}

// MarkDeliveryAsFailed marks a delivery log as failed
func (uc *NotificationUseCase) MarkDeliveryAsFailed(deliveryLogID string) error {
	if deliveryLogID == "" {
		return fmt.Errorf("delivery_log_id is required")
	}

	if err := uc.deliveryLogRepo.MarkAsFailed(deliveryLogID); err != nil {
		log.Printf("[ERROR] MarkDeliveryAsFailed - Failed to update: %v", err)
		return fmt.Errorf("failed to mark as failed: %w", err)
	}

	return nil
}

// MarkDeliveryAsBounced marks a delivery log as bounced (invalid email, etc.)
func (uc *NotificationUseCase) MarkDeliveryAsBounced(deliveryLogID string) error {
	if deliveryLogID == "" {
		return fmt.Errorf("delivery_log_id is required")
	}

	if err := uc.deliveryLogRepo.MarkAsBounced(deliveryLogID); err != nil {
		log.Printf("[ERROR] MarkDeliveryAsBounced - Failed to update: %v", err)
		return fmt.Errorf("failed to mark as bounced: %w", err)
	}

	return nil
}

// GetNotificationStats returns statistics for a notification
func (uc *NotificationUseCase) GetNotificationStats(notificationID string) (map[string]interface{}, error) {
	if notificationID == "" {
		return nil, fmt.Errorf("notification_id is required")
	}

	// Get all delivery logs for this notification
	deliveryLogs, err := uc.deliveryLogRepo.GetDeliveryLogsByNotification(notificationID)
	if err != nil {
		log.Printf("[ERROR] GetNotificationStats - Failed to retrieve logs: %v", err)
		return nil, fmt.Errorf("failed to retrieve notification stats: %w", err)
	}

	// Calculate stats
	stats := map[string]interface{}{
		"total":     len(deliveryLogs),
		"pending":   0,
		"delivered": 0,
		"failed":    0,
		"bounced":   0,
	}

	for _, log := range deliveryLogs {
		switch log.Status {
		case domain.DeliveryStatusPending:
			stats["pending"] = stats["pending"].(int) + 1
		case domain.DeliveryStatusDelivered:
			stats["delivered"] = stats["delivered"].(int) + 1
		case domain.DeliveryStatusFailed:
			stats["failed"] = stats["failed"].(int) + 1
		case domain.DeliveryStatusBounced:
			stats["bounced"] = stats["bounced"].(int) + 1
		}
	}

	return stats, nil
}

// DeleteNotification deletes a notification
func (uc *NotificationUseCase) DeleteNotification(notificationID string) error {
	if notificationID == "" {
		return fmt.Errorf("notification_id is required")
	}

	if err := uc.notificationRepo.DeleteNotification(notificationID); err != nil {
		log.Printf("[ERROR] DeleteNotification - Failed to delete: %v", err)
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	log.Printf("[SUCCESS] DeleteNotification - Notification deleted: %s", notificationID)
	return nil
}
