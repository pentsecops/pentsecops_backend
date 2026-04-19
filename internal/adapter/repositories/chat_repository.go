package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/pentsecops/backend/internal/core/domain"
)

// ChatRepository handles chat message persistence
type ChatRepository struct {
	db *sql.DB
}

// NewChatRepository creates a new chat repository
func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// SaveMessage saves a chat message to the database
func (cr *ChatRepository) SaveMessage(message *domain.ChatMessage) error {
	if message.ID == "" {
		return fmt.Errorf("message ID is required")
	}
	if message.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	if message.UserType == "" {
		return fmt.Errorf("user type is required")
	}

	query := `
		INSERT INTO chat_messages (id, user_id, user_type, role, content, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := cr.db.Exec(query, message.ID, message.UserID, message.UserType, message.Role, message.Content, message.CreatedAt)
	if err != nil {
		log.Printf("[ERROR] Failed to save chat message: %v", err)
		return fmt.Errorf("failed to save message: %w", err)
	}

	log.Printf("[SUCCESS] Chat message saved: user_id=%s, user_type=%s, role=%s", message.UserID, message.UserType, message.Role)
	return nil
}

// GetChatHistory retrieves chat history for a user (ordered by creation time)
func (cr *ChatRepository) GetChatHistory(userID string, limit int) ([]domain.ChatMessage, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	if limit <= 0 {
		limit = 50 // Default limit
	}

	query := `
		SELECT id, user_id, user_type, role, content, created_at
		FROM chat_messages
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := cr.db.Query(query, userID, limit)
	if err != nil {
		log.Printf("[ERROR] Failed to query chat history for user %s: %v", userID, err)
		return nil, fmt.Errorf("failed to query chat history: %w", err)
	}
	defer rows.Close()

	var messages []domain.ChatMessage
	for rows.Next() {
		var msg domain.ChatMessage
		err := rows.Scan(&msg.ID, &msg.UserID, &msg.UserType, &msg.Role, &msg.Content, &msg.CreatedAt)
		if err != nil {
			log.Printf("[ERROR] Failed to scan chat message: %v", err)
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		log.Printf("[ERROR] Error iterating chat history rows: %v", err)
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	log.Printf("[SUCCESS] Retrieved %d chat messages for user %s", len(messages), userID)
	return messages, nil
}

// GetChatHistorySince retrieves chat messages since a specific time
func (cr *ChatRepository) GetChatHistorySince(userID string, since time.Time) ([]domain.ChatMessage, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	query := `
		SELECT id, user_id, user_type, role, content, created_at
		FROM chat_messages
		WHERE user_id = $1 AND created_at >= $2
		ORDER BY created_at ASC
	`

	rows, err := cr.db.Query(query, userID, since)
	if err != nil {
		log.Printf("[ERROR] Failed to query chat history since %v for user %s: %v", since, userID, err)
		return nil, fmt.Errorf("failed to query chat history: %w", err)
	}
	defer rows.Close()

	var messages []domain.ChatMessage
	for rows.Next() {
		var msg domain.ChatMessage
		err := rows.Scan(&msg.ID, &msg.UserID, &msg.UserType, &msg.Role, &msg.Content, &msg.CreatedAt)
		if err != nil {
			log.Printf("[ERROR] Failed to scan chat message: %v", err)
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		log.Printf("[ERROR] Error iterating chat history rows: %v", err)
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	log.Printf("[SUCCESS] Retrieved %d chat messages for user %s since %v", len(messages), userID, since)
	return messages, nil
}

// DeleteUserMessages deletes all messages from a specific user
func (cr *ChatRepository) DeleteUserMessages(userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	query := `DELETE FROM chat_messages WHERE user_id = $1`

	result, err := cr.db.Exec(query, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to delete chat messages for user %s: %v", userID, err)
		return fmt.Errorf("failed to delete messages: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("[ERROR] Failed to get rows affected: %v", err)
		return err
	}

	log.Printf("[SUCCESS] Deleted %d chat messages for user %s", rowsAffected, userID)
	return nil
}

// GetMessageCount returns the total number of messages for a user
func (cr *ChatRepository) GetMessageCount(userID string) (int64, error) {
	if userID == "" {
		return 0, fmt.Errorf("user ID is required")
	}

	query := `SELECT COUNT(*) FROM chat_messages WHERE user_id = $1`

	var count int64
	err := cr.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		log.Printf("[ERROR] Failed to count messages for user %s: %v", userID, err)
		return 0, fmt.Errorf("failed to count messages: %w", err)
	}

	return count, nil
}
