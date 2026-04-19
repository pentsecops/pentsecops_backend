package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/pentsecops/backend/internal/core/domain"
)

// NotificationRepository handles notification persistence
type NotificationRepository struct {
	db *sql.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// CreateNotification creates a new notification
func (r *NotificationRepository) CreateNotification(notification *domain.Notification) (*domain.Notification, error) {
	if notification.Subject == "" || notification.Message == "" {
		return nil, fmt.Errorf("missing required fields")
	}

	// Ensure default values
	if notification.Priority == "" {
		notification.Priority = domain.NotificationPriorityMedium
	}
	if notification.ReceiverType == "" {
		notification.ReceiverType = domain.NotificationReceiverTypeSingle
	}

	// Convert receiver IDs to JSON
	receiverIDsJSON, err := json.Marshal(notification.ReceiverIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal receiver IDs: %w", err)
	}

	query := `
		INSERT INTO notifications (sender_id, receiver_type, receiver_ids, subject, message, priority, in_app, email, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`

	err = r.db.QueryRow(
		query,
		notification.SenderID,
		notification.ReceiverType,
		string(receiverIDsJSON),
		notification.Subject,
		notification.Message,
		notification.Priority,
		notification.InApp,
		notification.Email,
		time.Now(),
	).Scan(&notification.ID, &notification.CreatedAt)

	if err != nil {
		log.Printf("[ERROR] Failed to create notification: %v", err)
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	log.Printf("[SUCCESS] Notification created: id=%s, type=%s, priority=%s", notification.ID, notification.ReceiverType, notification.Priority)
	return notification, nil
}

// GetNotificationByID retrieves a notification by ID
func (r *NotificationRepository) GetNotificationByID(notificationID string) (*domain.Notification, error) {
	query := `
		SELECT id, sender_id, receiver_type, receiver_ids, subject, message, priority, in_app, email, created_at
		FROM notifications
		WHERE id = $1
	`

	var notification domain.Notification
	var receiverIDsJSON string

	err := r.db.QueryRow(query, notificationID).Scan(
		&notification.ID,
		&notification.SenderID,
		&notification.ReceiverType,
		&receiverIDsJSON,
		&notification.Subject,
		&notification.Message,
		&notification.Priority,
		&notification.InApp,
		&notification.Email,
		&notification.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("notification not found")
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	// Parse receiver IDs
	err = json.Unmarshal([]byte(receiverIDsJSON), &notification.ReceiverIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal receiver IDs: %w", err)
	}

	return &notification, nil
}

// GetNotificationsByReceiver retrieves notifications for a specific receiver
func (r *NotificationRepository) GetNotificationsByReceiver(receiverID string, limit, offset int) ([]*domain.Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT n.id, n.sender_id, n.receiver_type, n.receiver_ids, n.subject, n.message, n.priority, n.in_app, n.email, n.created_at
		FROM notifications n
		WHERE (n.receiver_type = $1 AND n.receiver_ids::jsonb @> $2::jsonb)
		   OR (n.receiver_type = $3 AND n.receiver_ids::jsonb @> $2::jsonb)
		   OR n.receiver_type = $4
		ORDER BY n.created_at DESC
		LIMIT $5 OFFSET $6
	`

	rows, err := r.db.Query(
		query,
		domain.NotificationReceiverTypeSingle,
		fmt.Sprintf(`["%s"]`, receiverID),
		domain.NotificationReceiverTypeMultiple,
		domain.NotificationReceiverTypeBroadcast,
		limit,
		offset,
	)
	if err != nil {
		log.Printf("[ERROR] Failed to get notifications: %v", err)
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*domain.Notification
	for rows.Next() {
		var n domain.Notification
		var receiverIDsJSON string

		if err := rows.Scan(
			&n.ID,
			&n.SenderID,
			&n.ReceiverType,
			&receiverIDsJSON,
			&n.Subject,
			&n.Message,
			&n.Priority,
			&n.InApp,
			&n.Email,
			&n.CreatedAt,
		); err != nil {
			log.Printf("[ERROR] Failed to scan notification: %v", err)
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		// Parse receiver IDs
		err := json.Unmarshal([]byte(receiverIDsJSON), &n.ReceiverIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal receiver IDs: %w", err)
		}

		notifications = append(notifications, &n)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notifications: %w", err)
	}

	return notifications, nil
}

// DeleteNotification deletes a notification
func (r *NotificationRepository) DeleteNotification(notificationID string) error {
	query := `DELETE FROM notifications WHERE id = $1`
	result, err := r.db.Exec(query, notificationID)
	if err != nil {
		log.Printf("[ERROR] Failed to delete notification: %v", err)
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("notification not found")
	}

	log.Printf("[SUCCESS] Notification deleted: %s", notificationID)
	return nil
}

// UpdateNotificationDeliveredAt updates the delivered_at timestamp
func (r *NotificationRepository) UpdateNotificationDeliveredAt(notificationID string) error {
	query := `UPDATE notifications SET delivered_at = $1 WHERE id = $2`
	result, err := r.db.Exec(query, time.Now(), notificationID)
	if err != nil {
		log.Printf("[ERROR] Failed to update notification: %v", err)
		return fmt.Errorf("failed to update notification: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}
