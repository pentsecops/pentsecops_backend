package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// NotificationPriority represents notification priority level
type NotificationPriority string

const (
	NotificationPriorityLow      NotificationPriority = "low"
	NotificationPriorityMedium   NotificationPriority = "medium"
	NotificationPriorityHigh     NotificationPriority = "high"
	NotificationPriorityCritical NotificationPriority = "critical"
)

// NotificationReceiverType represents the type of notification receiver
type NotificationReceiverType string

const (
	NotificationReceiverTypeSingle    NotificationReceiverType = "single"
	NotificationReceiverTypeMultiple  NotificationReceiverType = "multiple"
	NotificationReceiverTypeBroadcast NotificationReceiverType = "broadcast"
)

// DeliveryChannel represents the delivery method
type DeliveryChannel string

const (
	DeliveryChannelInApp DeliveryChannel = "in_app"
	DeliveryChannelEmail DeliveryChannel = "email"
)

// DeliveryStatus represents the status of notification delivery
type DeliveryStatus string

const (
	DeliveryStatusPending   DeliveryStatus = "pending"
	DeliveryStatusSent      DeliveryStatus = "sent"
	DeliveryStatusDelivered DeliveryStatus = "delivered"
	DeliveryStatusFailed    DeliveryStatus = "failed"
	DeliveryStatusBounced   DeliveryStatus = "bounced"
)

// Notification represents a notification sent to users
type Notification struct {
	ID           string
	SenderID     *string
	ReceiverType NotificationReceiverType
	ReceiverIDs  StringArray // JSON array stored as string
	Subject      string
	Message      string
	Priority     NotificationPriority
	InApp        bool
	Email        bool
	CreatedAt    time.Time
	DeliveredAt  *time.Time
}

// StringArray is a custom type for JSON arrays in database
type StringArray []string

// Scan implements sql.Scanner for JSON array type
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		// Try to convert to string then unmarshal
		data, ok := value.(string)
		if !ok {
			return nil
		}
		return json.Unmarshal([]byte(data), s)
	}
}

// Value implements driver.Valuer for JSON array type
func (s StringArray) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return json.Marshal(s)
}

// NotificationRequest is the request payload for creating a notification
type NotificationRequest struct {
	ReceiverType NotificationReceiverType `json:"receiver_type"`
	ReceiverIDs  []string                 `json:"receiver_ids,omitempty"`
	Subject      string                   `json:"subject"`
	Message      string                   `json:"message"`
	Priority     NotificationPriority     `json:"priority"`
	InApp        bool                     `json:"in_app"`
	Email        bool                     `json:"email"`
}

// NotificationResponse is the response payload for a notification
type NotificationResponse struct {
	ID           string                   `json:"id"`
	SenderID     *string                  `json:"sender_id,omitempty"`
	ReceiverType NotificationReceiverType `json:"receiver_type"`
	ReceiverIDs  StringArray              `json:"receiver_ids,omitempty"`
	Subject      string                   `json:"subject"`
	Message      string                   `json:"message"`
	Priority     NotificationPriority     `json:"priority"`
	InApp        bool                     `json:"in_app"`
	Email        bool                     `json:"email"`
	CreatedAt    time.Time                `json:"created_at"`
	DeliveredAt  *time.Time               `json:"delivered_at,omitempty"`
}

// NotificationDeliveryLog tracks delivery status for each user/channel
type NotificationDeliveryLog struct {
	ID             string
	NotificationID string
	ReceiverID     string
	Channel        DeliveryChannel
	Status         DeliveryStatus
	AttemptedAt    *time.Time
	DeliveredAt    *time.Time
	ErrorMessage   *string
	CreatedAt      time.Time
}

// NotificationRepository defines notification persistence operations
type NotificationRepository interface {
	CreateNotification(notification *Notification) (*Notification, error)
	GetNotificationByID(notificationID string) (*Notification, error)
	GetNotificationsByReceiver(receiverID string, limit, offset int) ([]*Notification, error)
	UpdateNotificationDeliveredAt(notificationID string) error
	DeleteNotification(notificationID string) error
}

// NotificationDeliveryLogRepository defines delivery log persistence operations
type NotificationDeliveryLogRepository interface {
	CreateDeliveryLog(log *NotificationDeliveryLog) (*NotificationDeliveryLog, error)
	GetDeliveryLogByID(logID string) (*NotificationDeliveryLog, error)
	GetDeliveryLogsByNotification(notificationID string) ([]*NotificationDeliveryLog, error)
	GetDeliveryLogsByReceiver(receiverID string, channel DeliveryChannel) ([]*NotificationDeliveryLog, error)
	UpdateDeliveryStatus(logID string, status DeliveryStatus) error
	MarkAsDelivered(logID string) error
	MarkAsFailed(logID string) error
	MarkAsBounced(logID string) error
	GetPendingDeliveryLogs(channel string, limit int) ([]*NotificationDeliveryLog, error)
	DeleteDeliveryLog(logID string) error
}
