package usecases

import (
	"fmt"
	"log"
	"time"

	"github.com/pentsecops/backend/internal/core/domain"
)

// MessageUseCase handles business logic for messages
type MessageUseCase struct {
	messageRepo           domain.MessageRepository
	messageAttachmentRepo domain.MessageAttachmentRepository
	userRepo              domain.UserRepository
}

// NewMessageUseCase creates a new message use case
func NewMessageUseCase(
	messageRepo domain.MessageRepository,
	messageAttachmentRepo domain.MessageAttachmentRepository,
	userRepo domain.UserRepository,
) *MessageUseCase {
	return &MessageUseCase{
		messageRepo:           messageRepo,
		messageAttachmentRepo: messageAttachmentRepo,
		userRepo:              userRepo,
	}
}

// SendMessage sends a message from sender to receiver
// Returns the created message with ID
func (uc *MessageUseCase) SendMessage(senderID, receiverID, content string) (*domain.Message, error) {
	if senderID == "" || receiverID == "" {
		return nil, fmt.Errorf("sender_id and receiver_id are required")
	}

	if content == "" {
		return nil, fmt.Errorf("content cannot be empty")
	}

	// Verify sender and receiver exist (check all user types)
	sender, err := uc.userRepo.GetAnyUserByID(senderID)
	if err != nil {
		log.Printf("[ERROR] SendMessage - Sender not found: %v", err)
		return nil, fmt.Errorf("sender not found")
	}

	receiver, err := uc.userRepo.GetAnyUserByID(receiverID)
	if err != nil {
		log.Printf("[ERROR] SendMessage - Receiver not found: %v", err)
		return nil, fmt.Errorf("receiver not found")
	}

	if sender.ID == receiver.ID {
		return nil, fmt.Errorf("cannot send message to yourself")
	}

	// Create message
	message := &domain.Message{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
		Status:     domain.MessageStatusSent,
		SentAt:     time.Now(),
	}

	// Save to database
	savedMessage, err := uc.messageRepo.CreateMessage(message)
	if err != nil {
		log.Printf("[ERROR] SendMessage - Failed to save message: %v", err)
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	log.Printf("[SUCCESS] SendMessage - Message sent: from=%s, to=%s, message_id=%s", senderID, receiverID, savedMessage.ID)
	return savedMessage, nil
}

// MarkAsDelivered marks a message as delivered to the receiver
func (uc *MessageUseCase) MarkAsDelivered(messageID string) error {
	if messageID == "" {
		return fmt.Errorf("message_id is required")
	}

	if err := uc.messageRepo.MarkAsDelivered(messageID); err != nil {
		log.Printf("[ERROR] MarkAsDelivered - Failed to update message: %v", err)
		return fmt.Errorf("failed to update message: %w", err)
	}

	log.Printf("[SUCCESS] MarkAsDelivered - Message marked as delivered: %s", messageID)
	return nil
}

// MarkAsRead marks a message as read by the receiver
func (uc *MessageUseCase) MarkAsRead(messageID string) error {
	if messageID == "" {
		return fmt.Errorf("message_id is required")
	}

	if err := uc.messageRepo.MarkAsRead(messageID); err != nil {
		log.Printf("[ERROR] MarkAsRead - Failed to update message: %v", err)
		return fmt.Errorf("failed to update message: %w", err)
	}

	log.Printf("[SUCCESS] MarkAsRead - Message marked as read: %s", messageID)
	return nil
}

// GetConversation retrieves conversation between two users
// Returns messages sorted by sent_at descending (newest first)
func (uc *MessageUseCase) GetConversation(userID1, userID2 string, limit, offset int) ([]*domain.Message, error) {
	if userID1 == "" || userID2 == "" {
		return nil, fmt.Errorf("both user_ids are required")
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	if offset < 0 {
		offset = 0
	}

	messages, err := uc.messageRepo.GetConversationMessages(userID1, userID2, limit, offset)
	if err != nil {
		log.Printf("[ERROR] GetConversation - Failed to retrieve conversation: %v", err)
		return nil, fmt.Errorf("failed to retrieve conversation: %w", err)
	}

	return messages, nil
}

// GetUserMessages retrieves all undelivered messages for a specific user
func (uc *MessageUseCase) GetUserMessages(userID string, limit, offset int) ([]*domain.Message, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	if offset < 0 {
		offset = 0
	}

	// Get undelivered messages for the user
	messages, err := uc.messageRepo.GetUndeliveredMessages(userID)
	if err != nil {
		log.Printf("[ERROR] GetUserMessages - Failed to retrieve messages: %v", err)
		return nil, fmt.Errorf("failed to retrieve messages: %w", err)
	}

	// Apply limit and offset to results
	if offset >= len(messages) {
		return []*domain.Message{}, nil
	}

	end := offset + limit
	if end > len(messages) {
		end = len(messages)
	}

	return messages[offset:end], nil
}

// AddAttachment adds a file attachment to a message
func (uc *MessageUseCase) AddAttachment(messageID, fileName, filePath string, fileSize int64, mimeType string) (*domain.MessageAttachment, error) {
	if messageID == "" || fileName == "" || filePath == "" {
		return nil, fmt.Errorf("message_id, file_name, and file_path are required")
	}

	// Verify message exists
	_, err := uc.messageRepo.GetMessageByID(messageID)
	if err != nil {
		log.Printf("[ERROR] AddAttachment - Message not found: %v", err)
		return nil, fmt.Errorf("message not found")
	}

	// Create attachment
	attachment := &domain.MessageAttachment{
		MessageID: messageID,
		FileName:  fileName,
		FilePath:  filePath,
		FileSize:  fileSize,
		MimeType:  mimeType,
		CreatedAt: time.Now(),
	}

	savedAttachment, err := uc.messageAttachmentRepo.CreateAttachment(attachment)
	if err != nil {
		log.Printf("[ERROR] AddAttachment - Failed to save attachment: %v", err)
		return nil, fmt.Errorf("failed to save attachment: %w", err)
	}

	log.Printf("[SUCCESS] AddAttachment - Attachment added: message=%s, file=%s", messageID, fileName)
	return savedAttachment, nil
}

// GetAttachments retrieves all attachments for a message
func (uc *MessageUseCase) GetAttachments(messageID string) ([]*domain.MessageAttachment, error) {
	if messageID == "" {
		return nil, fmt.Errorf("message_id is required")
	}

	attachments, err := uc.messageAttachmentRepo.GetAttachmentsByMessageID(messageID)
	if err != nil {
		log.Printf("[ERROR] GetAttachments - Failed to retrieve attachments: %v", err)
		return nil, fmt.Errorf("failed to retrieve attachments: %w", err)
	}

	return attachments, nil
}

// GetUnreadCount returns count of unread messages for a user
// For now, returns the count of undelivered messages
func (uc *MessageUseCase) GetUnreadCount(userID string) (int, error) {
	if userID == "" {
		return 0, fmt.Errorf("user_id is required")
	}

	messages, err := uc.messageRepo.GetUndeliveredMessages(userID)
	if err != nil {
		log.Printf("[ERROR] GetUnreadCount - Failed to get unread messages: %v", err)
		return 0, fmt.Errorf("failed to get unread messages: %w", err)
	}

	return len(messages), nil
}

// DeleteMessage deletes a message
func (uc *MessageUseCase) DeleteMessage(messageID, userID string) error {
	if messageID == "" || userID == "" {
		return fmt.Errorf("message_id and user_id are required")
	}

	message, err := uc.messageRepo.GetMessageByID(messageID)
	if err != nil {
		log.Printf("[ERROR] DeleteMessage - Message not found: %v", err)
		return fmt.Errorf("message not found")
	}

	// Only sender or receiver can delete
	if message.SenderID != userID && message.ReceiverID != userID {
		return fmt.Errorf("unauthorized: only sender or receiver can delete message")
	}

	// Delete attachments first
	if err := uc.messageAttachmentRepo.DeleteAttachmentsByMessageID(messageID); err != nil {
		log.Printf("[ERROR] DeleteMessage - Failed to delete attachments: %v", err)
		// Continue with message deletion
	}

	// Delete the message
	if err := uc.messageRepo.DeleteMessage(messageID); err != nil {
		log.Printf("[ERROR] DeleteMessage - Failed to delete message: %v", err)
		return fmt.Errorf("failed to delete message: %w", err)
	}

	log.Printf("[SUCCESS] DeleteMessage - Message deleted: %s, by_user=%s", messageID, userID)
	return nil
}
