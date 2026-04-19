package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/pentsecops/backend/internal/core/domain"
)

// MessageRepository handles message persistence
type MessageRepository struct {
	db *sql.DB
}

// ChatInfo represents a chat/conversation info for the chat list
type ChatInfo struct {
	UserID        string    `json:"user_id"`
	Email         string    `json:"email"`
	Role          string    `json:"role"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	LastMessage   string    `json:"last_message"`
	LastMessageAt time.Time `json:"last_message_at"`
	UnreadCount   int       `json:"unread_count"`
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// CreateMessage creates a new message and returns it with generated ID
func (r *MessageRepository) CreateMessage(message *domain.Message) (*domain.Message, error) {
	if message.SenderID == "" || message.ReceiverID == "" || message.Content == "" {
		return nil, fmt.Errorf("missing required fields")
	}

	query := `
		INSERT INTO messages (sender_id, receiver_id, content, status, sent_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, sent_at
	`

	err := r.db.QueryRow(
		query,
		message.SenderID,
		message.ReceiverID,
		message.Content,
		message.Status,
		time.Now(),
	).Scan(&message.ID, &message.SentAt)

	if err != nil {
		log.Printf("[ERROR] Failed to create message: %v", err)
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	log.Printf("[SUCCESS] Message created: id=%s, from=%s, to=%s", message.ID, message.SenderID, message.ReceiverID)
	return message, nil
}

// GetMessageByID retrieves a message by ID
func (r *MessageRepository) GetMessageByID(messageID string) (*domain.Message, error) {
	query := `
		SELECT id, sender_id, receiver_id, content, status, sent_at, delivered_at, read_at
		FROM messages
		WHERE id = $1
	`

	var message domain.Message
	var deliveredAt, readAt sql.NullTime

	err := r.db.QueryRow(query, messageID).Scan(
		&message.ID,
		&message.SenderID,
		&message.ReceiverID,
		&message.Content,
		&message.Status,
		&message.SentAt,
		&deliveredAt,
		&readAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("message not found")
		}
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	if deliveredAt.Valid {
		message.DeliveredAt = &deliveredAt.Time
	}
	if readAt.Valid {
		message.ReadAt = &readAt.Time
	}

	return &message, nil
}

// GetConversationMessages retrieves messages between two users
func (r *MessageRepository) GetConversationMessages(userID, otherUserID string, limit, offset int) ([]*domain.Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, sender_id, receiver_id, content, status, sent_at, delivered_at, read_at
		FROM messages
		WHERE (sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1)
		ORDER BY sent_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(query, userID, otherUserID, limit, offset)
	if err != nil {
		log.Printf("[ERROR] Failed to get conversation: %v", err)
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		var msg domain.Message
		var deliveredAt, readAt sql.NullTime

		if err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.ReceiverID,
			&msg.Content,
			&msg.Status,
			&msg.SentAt,
			&deliveredAt,
			&readAt,
		); err != nil {
			log.Printf("[ERROR] Failed to scan message: %v", err)
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		if deliveredAt.Valid {
			msg.DeliveredAt = &deliveredAt.Time
		}
		if readAt.Valid {
			msg.ReadAt = &readAt.Time
		}

		messages = append(messages, &msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	return messages, nil
}

// UpdateMessageStatus updates the status of a message
func (r *MessageRepository) UpdateMessageStatus(messageID string, status domain.MessageStatus) error {
	query := `UPDATE messages SET status = $1 WHERE id = $2`
	result, err := r.db.Exec(query, string(status), messageID)
	if err != nil {
		log.Printf("[ERROR] Failed to update message status: %v", err)
		return fmt.Errorf("failed to update message status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("message not found")
	}

	return nil
}

// MarkAsDelivered marks a message as delivered
func (r *MessageRepository) MarkAsDelivered(messageID string) error {
	query := `UPDATE messages SET status = $1, delivered_at = $2 WHERE id = $3`
	result, err := r.db.Exec(query, domain.MessageStatusDelivered, time.Now(), messageID)
	if err != nil {
		log.Printf("[ERROR] Failed to mark message as delivered: %v", err)
		return fmt.Errorf("failed to mark message as delivered: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("message not found")
	}

	log.Printf("[SUCCESS] Message marked as delivered: %s", messageID)
	return nil
}

// MarkAsRead marks a message as read
func (r *MessageRepository) MarkAsRead(messageID string) error {
	query := `UPDATE messages SET status = $1, read_at = $2 WHERE id = $3`
	result, err := r.db.Exec(query, domain.MessageStatusRead, time.Now(), messageID)
	if err != nil {
		log.Printf("[ERROR] Failed to mark message as read: %v", err)
		return fmt.Errorf("failed to mark message as read: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("message not found")
	}

	log.Printf("[SUCCESS] Message marked as read: %s", messageID)
	return nil
}

// GetUndeliveredMessages retrieves all undelivered messages for a user
func (r *MessageRepository) GetUndeliveredMessages(userID string) ([]*domain.Message, error) {
	query := `
		SELECT id, sender_id, receiver_id, content, status, sent_at, delivered_at, read_at
		FROM messages
		WHERE receiver_id = $1 AND status = $2
		ORDER BY sent_at ASC
	`

	rows, err := r.db.Query(query, userID, domain.MessageStatusSent)
	if err != nil {
		log.Printf("[ERROR] Failed to get undelivered messages: %v", err)
		return nil, fmt.Errorf("failed to get undelivered messages: %w", err)
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		var msg domain.Message
		var deliveredAt, readAt sql.NullTime

		if err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.ReceiverID,
			&msg.Content,
			&msg.Status,
			&msg.SentAt,
			&deliveredAt,
			&readAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		if deliveredAt.Valid {
			msg.DeliveredAt = &deliveredAt.Time
		}
		if readAt.Valid {
			msg.ReadAt = &readAt.Time
		}

		messages = append(messages, &msg)
	}

	return messages, rows.Err()
}

// GetUserChats retrieves all chats (conversations) for a user with unread counts
func (r *MessageRepository) GetUserChats(userID string) ([]*ChatInfo, error) {
	query := `
		WITH last_messages AS (
			SELECT 
				CASE 
					WHEN sender_id = $1 THEN receiver_id 
					ELSE sender_id 
				END as other_user_id,
				content,
				sent_at,
				status,
				ROW_NUMBER() OVER (
					PARTITION BY CASE 
						WHEN sender_id = $1 THEN receiver_id 
						ELSE sender_id 
					END 
					ORDER BY sent_at DESC
				) as rn
			FROM messages
			WHERE sender_id = $1 OR receiver_id = $1
		),
		chats AS (
			SELECT DISTINCT
				other_user_id,
				(SELECT content FROM last_messages WHERE rn = 1 AND other_user_id = lm.other_user_id LIMIT 1) as last_message,
				(SELECT sent_at FROM last_messages WHERE rn = 1 AND other_user_id = lm.other_user_id LIMIT 1) as last_message_at
			FROM last_messages lm
		),
		unread_counts AS (
			SELECT 
				sender_id,
				COUNT(*) as unread_count
			FROM messages
			WHERE receiver_id = $1 AND status != 'read'
			GROUP BY sender_id
		)
		SELECT 
			c.other_user_id,
			COALESCE(u.email, a.email, '') as email,
			COALESCE(u.role, 'admin') as role,
			COALESCE(p.first_name, s.first_name, '') as first_name,
			COALESCE(p.last_name, s.last_name, '') as last_name,
			COALESCE(c.last_message, '') as last_message,
			COALESCE(c.last_message_at, CURRENT_TIMESTAMP) as last_message_at,
			COALESCE(uc.unread_count, 0) as unread_count
		FROM chats c
		LEFT JOIN users u ON c.other_user_id = u.id
		LEFT JOIN pentesters p ON c.other_user_id = p.user_id
		LEFT JOIN stakeholders s ON c.other_user_id = s.user_id
		LEFT JOIN admin a ON c.other_user_id = a.id
		LEFT JOIN unread_counts uc ON c.other_user_id = uc.sender_id
		ORDER BY c.last_message_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get user chats: %v", err)
		return nil, fmt.Errorf("failed to get user chats: %w", err)
	}
	defer rows.Close()

	var chats []*ChatInfo
	for rows.Next() {
		var chat ChatInfo
		if err := rows.Scan(
			&chat.UserID,
			&chat.Email,
			&chat.Role,
			&chat.FirstName,
			&chat.LastName,
			&chat.LastMessage,
			&chat.LastMessageAt,
			&chat.UnreadCount,
		); err != nil {
			log.Printf("[ERROR] Failed to scan chat: %v", err)
			return nil, fmt.Errorf("failed to scan chat: %w", err)
		}
		chats = append(chats, &chat)
	}

	if err = rows.Err(); err != nil {
		log.Printf("[ERROR] Error iterating chats: %v", err)
		return nil, fmt.Errorf("error iterating chats: %w", err)
	}

	return chats, nil
}

// DeleteMessage deletes a message (soft delete by marking as deleted status)
func (r *MessageRepository) DeleteMessage(messageID string) error {
	query := `DELETE FROM messages WHERE id = $1`
	result, err := r.db.Exec(query, messageID)
	if err != nil {
		log.Printf("[ERROR] Failed to delete message: %v", err)
		return fmt.Errorf("failed to delete message: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("message not found")
	}

	log.Printf("[SUCCESS] Message deleted: %s", messageID)
	return nil
}
