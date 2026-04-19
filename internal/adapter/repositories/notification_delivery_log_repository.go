package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/pentsecops/backend/internal/core/domain"
)

// NotificationDeliveryLogRepository handles notification delivery log persistence
type NotificationDeliveryLogRepository struct {
	db *sql.DB
}

// NewNotificationDeliveryLogRepository creates a new notification delivery log repository
func NewNotificationDeliveryLogRepository(db *sql.DB) *NotificationDeliveryLogRepository {
	return &NotificationDeliveryLogRepository{db: db}
}

// CreateDeliveryLog creates a new delivery log entry
func (r *NotificationDeliveryLogRepository) CreateDeliveryLog(deliveryLog *domain.NotificationDeliveryLog) (*domain.NotificationDeliveryLog, error) {
	if deliveryLog.NotificationID == "" || deliveryLog.ReceiverID == "" || deliveryLog.Channel == "" {
		return nil, fmt.Errorf("missing required fields")
	}

	// Set default values
	if deliveryLog.Status == "" {
		deliveryLog.Status = domain.DeliveryStatusPending
	}

	query := `
		INSERT INTO notification_delivery_logs (notification_id, receiver_id, channel, status, attempted_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := r.db.QueryRow(
		query,
		deliveryLog.NotificationID,
		deliveryLog.ReceiverID,
		deliveryLog.Channel,
		deliveryLog.Status,
		time.Now(),
	).Scan(&deliveryLog.ID)

	if err != nil {
		log.Printf("[ERROR] Failed to create delivery log: %v", err)
		return nil, fmt.Errorf("failed to create delivery log: %w", err)
	}

	log.Printf("[SUCCESS] Delivery log created: id=%s, notification=%s, receiver=%s", deliveryLog.ID, deliveryLog.NotificationID, deliveryLog.ReceiverID)
	return deliveryLog, nil
}

// GetDeliveryLogByID retrieves a delivery log by ID
func (r *NotificationDeliveryLogRepository) GetDeliveryLogByID(logID string) (*domain.NotificationDeliveryLog, error) {
	query := `
		SELECT id, notification_id, receiver_id, channel, status, attempted_at, delivered_at
		FROM notification_delivery_logs
		WHERE id = $1
	`

	var deliveryLog domain.NotificationDeliveryLog
	var deliveredAt sql.NullTime

	err := r.db.QueryRow(query, logID).Scan(
		&deliveryLog.ID,
		&deliveryLog.NotificationID,
		&deliveryLog.ReceiverID,
		&deliveryLog.Channel,
		&deliveryLog.Status,
		&deliveryLog.AttemptedAt,
		&deliveredAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("delivery log not found")
		}
		return nil, fmt.Errorf("failed to get delivery log: %w", err)
	}

	if deliveredAt.Valid {
		deliveryLog.DeliveredAt = &deliveredAt.Time
	}

	return &deliveryLog, nil
}

// GetPendingDeliveryLogs retrieves all pending delivery logs for a specific channel
func (r *NotificationDeliveryLogRepository) GetPendingDeliveryLogs(channel string, limit int) ([]*domain.NotificationDeliveryLog, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	query := `
		SELECT id, notification_id, receiver_id, channel, status, attempted_at, delivered_at
		FROM notification_delivery_logs
		WHERE channel = $1 AND status = $2
		ORDER BY attempted_at ASC
		LIMIT $3
	`

	rows, err := r.db.Query(query, channel, domain.DeliveryStatusPending, limit)
	if err != nil {
		log.Printf("[ERROR] Failed to get pending delivery logs: %v", err)
		return nil, fmt.Errorf("failed to get pending delivery logs: %w", err)
	}
	defer rows.Close()

	var logs []*domain.NotificationDeliveryLog
	for rows.Next() {
		var deliveryLog domain.NotificationDeliveryLog
		var deliveredAt sql.NullTime

		if err := rows.Scan(
			&deliveryLog.ID,
			&deliveryLog.NotificationID,
			&deliveryLog.ReceiverID,
			&deliveryLog.Channel,
			&deliveryLog.Status,
			&deliveryLog.AttemptedAt,
			&deliveredAt,
		); err != nil {
			log.Printf("[ERROR] Failed to scan delivery log: %v", err)
			return nil, fmt.Errorf("failed to scan delivery log: %w", err)
		}

		if deliveredAt.Valid {
			deliveryLog.DeliveredAt = &deliveredAt.Time
		}

		logs = append(logs, &deliveryLog)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating delivery logs: %w", err)
	}

	return logs, nil
}

// GetDeliveryLogsByNotification retrieves all delivery logs for a notification
func (r *NotificationDeliveryLogRepository) GetDeliveryLogsByNotification(notificationID string) ([]*domain.NotificationDeliveryLog, error) {
	query := `
		SELECT id, notification_id, receiver_id, channel, status, attempted_at, delivered_at
		FROM notification_delivery_logs
		WHERE notification_id = $1
		ORDER BY attempted_at ASC
	`

	rows, err := r.db.Query(query, notificationID)
	if err != nil {
		log.Printf("[ERROR] Failed to get delivery logs: %v", err)
		return nil, fmt.Errorf("failed to get delivery logs: %w", err)
	}
	defer rows.Close()

	var logs []*domain.NotificationDeliveryLog
	for rows.Next() {
		var deliveryLog domain.NotificationDeliveryLog
		var deliveredAt sql.NullTime

		if err := rows.Scan(
			&deliveryLog.ID,
			&deliveryLog.NotificationID,
			&deliveryLog.ReceiverID,
			&deliveryLog.Channel,
			&deliveryLog.Status,
			&deliveryLog.AttemptedAt,
			&deliveredAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan delivery log: %w", err)
		}

		if deliveredAt.Valid {
			deliveryLog.DeliveredAt = &deliveredAt.Time
		}

		logs = append(logs, &deliveryLog)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating delivery logs: %w", err)
	}

	return logs, nil
}

// UpdateDeliveryStatus updates the status of a delivery log
func (r *NotificationDeliveryLogRepository) UpdateDeliveryStatus(logID string, status domain.DeliveryStatus) error {
	query := `UPDATE notification_delivery_logs SET status = $1, delivered_at = $2 WHERE id = $3`
	result, err := r.db.Exec(query, string(status), time.Now(), logID)
	if err != nil {
		log.Printf("[ERROR] Failed to update delivery status: %v", err)
		return fmt.Errorf("failed to update delivery status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("delivery log not found")
	}

	log.Printf("[SUCCESS] Delivery status updated: %s -> %s", logID, status)
	return nil
}

// MarkAsDelivered marks a delivery log as delivered
func (r *NotificationDeliveryLogRepository) MarkAsDelivered(logID string) error {
	return r.UpdateDeliveryStatus(logID, domain.DeliveryStatusDelivered)
}

// MarkAsFailed marks a delivery log as failed
func (r *NotificationDeliveryLogRepository) MarkAsFailed(logID string) error {
	return r.UpdateDeliveryStatus(logID, domain.DeliveryStatusFailed)
}

// MarkAsBounced marks a delivery log as bounced (invalid email)
func (r *NotificationDeliveryLogRepository) MarkAsBounced(logID string) error {
	return r.UpdateDeliveryStatus(logID, domain.DeliveryStatusBounced)
}

// DeleteDeliveryLog deletes a delivery log
func (r *NotificationDeliveryLogRepository) DeleteDeliveryLog(logID string) error {
	query := `DELETE FROM notification_delivery_logs WHERE id = $1`
	result, err := r.db.Exec(query, logID)
	if err != nil {
		log.Printf("[ERROR] Failed to delete delivery log: %v", err)
		return fmt.Errorf("failed to delete delivery log: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("delivery log not found")
	}

	log.Printf("[SUCCESS] Delivery log deleted: %s", logID)
	return nil
}

// GetDeliveryLogsByReceiver retrieves delivery logs for a specific receiver and channel
func (r *NotificationDeliveryLogRepository) GetDeliveryLogsByReceiver(receiverID string, channel domain.DeliveryChannel) ([]*domain.NotificationDeliveryLog, error) {
	query := `
		SELECT id, notification_id, receiver_id, channel, status, attempted_at, delivered_at, error_message, created_at
		FROM notification_delivery_logs
		WHERE receiver_id = $1 AND channel = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, receiverID, string(channel))
	if err != nil {
		log.Printf("[ERROR] Failed to get delivery logs: %v", err)
		return nil, fmt.Errorf("failed to get delivery logs: %w", err)
	}
	defer rows.Close()

	var deliveryLogs []*domain.NotificationDeliveryLog
	for rows.Next() {
		var deliveryLog domain.NotificationDeliveryLog
		var attemptedAt, deliveredAt sql.NullTime
		var errorMessage sql.NullString

		if err := rows.Scan(
			&deliveryLog.ID,
			&deliveryLog.NotificationID,
			&deliveryLog.ReceiverID,
			&deliveryLog.Channel,
			&deliveryLog.Status,
			&attemptedAt,
			&deliveredAt,
			&errorMessage,
			&deliveryLog.CreatedAt,
		); err != nil {
			log.Printf("[ERROR] Failed to scan delivery log: %v", err)
			return nil, fmt.Errorf("failed to scan delivery log: %w", err)
		}

		if attemptedAt.Valid {
			deliveryLog.AttemptedAt = &attemptedAt.Time
		}
		if deliveredAt.Valid {
			deliveryLog.DeliveredAt = &deliveredAt.Time
		}
		if errorMessage.Valid {
			deliveryLog.ErrorMessage = &errorMessage.String
		}

		deliveryLogs = append(deliveryLogs, &deliveryLog)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating logs: %w", err)
	}

	return deliveryLogs, nil
}
