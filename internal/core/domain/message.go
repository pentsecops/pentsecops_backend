package domain

import (
	"time"
)

// MessageStatus represents the status of a message
type MessageStatus string

const (
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
)

// Message represents a 1-to-1 chat message
type Message struct {
	ID          string
	SenderID    string
	ReceiverID  string
	Content     string
	Status      MessageStatus
	SentAt      time.Time
	DeliveredAt *time.Time
	ReadAt      *time.Time
}

// MessageAttachment represents a file attachment in a message
type MessageAttachment struct {
	ID        string
	MessageID string
	FileName  string
	FilePath  string
	FileSize  int64
	MimeType  string
	CreatedAt time.Time
}

// MessageRequest is the request payload for sending a message
type MessageRequest struct {
	ReceiverID string `json:"receiver_id"`
	Content    string `json:"content"`
}

// MessageResponse is the response payload for a message
type MessageResponse struct {
	ID          string                      `json:"id"`
	SenderID    string                      `json:"sender_id"`
	ReceiverID  string                      `json:"receiver_id"`
	Content     string                      `json:"content"`
	Status      MessageStatus               `json:"status"`
	Attachments []MessageAttachmentResponse `json:"attachments,omitempty"`
	SentAt      time.Time                   `json:"sent_at"`
	DeliveredAt *time.Time                  `json:"delivered_at,omitempty"`
	ReadAt      *time.Time                  `json:"read_at,omitempty"`
}

// MessageAttachmentResponse is the response for a message attachment
type MessageAttachmentResponse struct {
	ID        string    `json:"id"`
	FileName  string    `json:"file_name"`
	FilePath  string    `json:"file_path"`
	FileSize  int64     `json:"file_size"`
	MimeType  string    `json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
}

// MessageRepository defines message persistence operations
type MessageRepository interface {
	CreateMessage(message *Message) (*Message, error)
	GetMessageByID(messageID string) (*Message, error)
	GetConversationMessages(userID, otherUserID string, limit, offset int) ([]*Message, error)
	GetUndeliveredMessages(userID string) ([]*Message, error)
	UpdateMessageStatus(messageID string, status MessageStatus) error
	MarkAsDelivered(messageID string) error
	MarkAsRead(messageID string) error
	DeleteMessage(messageID string) error
}

// MessageAttachmentRepository defines message attachment persistence operations
type MessageAttachmentRepository interface {
	CreateAttachment(attachment *MessageAttachment) (*MessageAttachment, error)
	GetAttachmentsByMessageID(messageID string) ([]*MessageAttachment, error)
	DeleteAttachment(attachmentID string) error
	DeleteAttachmentsByMessageID(messageID string) error
}
