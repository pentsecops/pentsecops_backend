package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/pentsecops/backend/internal/core/domain"
)

// MessageAttachmentRepository handles message attachment persistence
type MessageAttachmentRepository struct {
	db *sql.DB
}

// NewMessageAttachmentRepository creates a new message attachment repository
func NewMessageAttachmentRepository(db *sql.DB) *MessageAttachmentRepository {
	return &MessageAttachmentRepository{db: db}
}

// CreateAttachment creates a new message attachment
func (r *MessageAttachmentRepository) CreateAttachment(attachment *domain.MessageAttachment) (*domain.MessageAttachment, error) {
	if attachment.MessageID == "" || attachment.FileName == "" || attachment.FilePath == "" {
		return nil, fmt.Errorf("missing required fields")
	}

	query := `
		INSERT INTO message_attachments (message_id, file_name, file_path, file_size, mime_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(
		query,
		attachment.MessageID,
		attachment.FileName,
		attachment.FilePath,
		attachment.FileSize,
		attachment.MimeType,
		time.Now(),
	).Scan(&attachment.ID, &attachment.CreatedAt)

	if err != nil {
		log.Printf("[ERROR] Failed to create attachment: %v", err)
		return nil, fmt.Errorf("failed to create attachment: %w", err)
	}

	log.Printf("[SUCCESS] Attachment created: id=%s, message_id=%s, file=%s", attachment.ID, attachment.MessageID, attachment.FileName)
	return attachment, nil
}

// GetAttachmentByID retrieves an attachment by ID
func (r *MessageAttachmentRepository) GetAttachmentByID(attachmentID string) (*domain.MessageAttachment, error) {
	query := `
		SELECT id, message_id, file_name, file_path, file_size, mime_type, created_at
		FROM message_attachments
		WHERE id = $1
	`

	var attachment domain.MessageAttachment
	err := r.db.QueryRow(query, attachmentID).Scan(
		&attachment.ID,
		&attachment.MessageID,
		&attachment.FileName,
		&attachment.FilePath,
		&attachment.FileSize,
		&attachment.MimeType,
		&attachment.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("attachment not found")
		}
		return nil, fmt.Errorf("failed to get attachment: %w", err)
	}

	return &attachment, nil
}

// GetAttachmentsByMessageID retrieves all attachments for a message
func (r *MessageAttachmentRepository) GetAttachmentsByMessageID(messageID string) ([]*domain.MessageAttachment, error) {
	query := `
		SELECT id, message_id, file_name, file_path, file_size, mime_type, created_at
		FROM message_attachments
		WHERE message_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, messageID)
	if err != nil {
		log.Printf("[ERROR] Failed to get attachments: %v", err)
		return nil, fmt.Errorf("failed to get attachments: %w", err)
	}
	defer rows.Close()

	var attachments []*domain.MessageAttachment
	for rows.Next() {
		var att domain.MessageAttachment
		if err := rows.Scan(
			&att.ID,
			&att.MessageID,
			&att.FileName,
			&att.FilePath,
			&att.FileSize,
			&att.MimeType,
			&att.CreatedAt,
		); err != nil {
			log.Printf("[ERROR] Failed to scan attachment: %v", err)
			return nil, fmt.Errorf("failed to scan attachment: %w", err)
		}
		attachments = append(attachments, &att)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating attachments: %w", err)
	}

	return attachments, nil
}

// DeleteAttachment deletes an attachment
func (r *MessageAttachmentRepository) DeleteAttachment(attachmentID string) error {
	query := `DELETE FROM message_attachments WHERE id = $1`
	result, err := r.db.Exec(query, attachmentID)
	if err != nil {
		log.Printf("[ERROR] Failed to delete attachment: %v", err)
		return fmt.Errorf("failed to delete attachment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("attachment not found")
	}

	log.Printf("[SUCCESS] Attachment deleted: %s", attachmentID)
	return nil
}

// DeleteAttachmentsByMessageID deletes all attachments for a message
func (r *MessageAttachmentRepository) DeleteAttachmentsByMessageID(messageID string) error {
	query := `DELETE FROM message_attachments WHERE message_id = $1`
	result, err := r.db.Exec(query, messageID)
	if err != nil {
		log.Printf("[ERROR] Failed to delete attachments: %v", err)
		return fmt.Errorf("failed to delete attachments: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	log.Printf("[SUCCESS] Message attachments deleted: message_id=%s, count=%d", messageID, rows)
	return nil
}
