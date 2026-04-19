package handlers

import (
	"io"
	"log"
	"mime/multipart"
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pentsecops/backend/internal/adapter/repositories"
	"github.com/pentsecops/backend/internal/core/domain"
	"github.com/pentsecops/backend/internal/core/usecases"
	"github.com/pentsecops/backend/pkg/services"
)

// MessageHandler handles message-related HTTP requests
type MessageHandler struct {
	messageUseCase        *usecases.MessageUseCase
	userRepository        domain.UserRepository
	messageRepository     *repositories.MessageRepository
	messageAttachmentRepo *repositories.MessageAttachmentRepository
	fileService           *services.FileService
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(messageUseCase *usecases.MessageUseCase) *MessageHandler {
	return &MessageHandler{
		messageUseCase: messageUseCase,
	}
}

// NewMessageHandlerWithRepo creates a new message handler with user repository
func NewMessageHandlerWithRepo(messageUseCase *usecases.MessageUseCase, userRepository domain.UserRepository) *MessageHandler {
	return &MessageHandler{
		messageUseCase: messageUseCase,
		userRepository: userRepository,
	}
}

// NewMessageHandlerWithRepos creates a new message handler with repositories
func NewMessageHandlerWithRepos(messageUseCase *usecases.MessageUseCase, userRepository domain.UserRepository, messageRepository *repositories.MessageRepository) *MessageHandler {
	return &MessageHandler{
		messageUseCase:    messageUseCase,
		userRepository:    userRepository,
		messageRepository: messageRepository,
	}
}

// NewMessageHandlerWithAttachments creates a new message handler with full dependencies including attachments
func NewMessageHandlerWithAttachments(
	messageUseCase *usecases.MessageUseCase,
	userRepository domain.UserRepository,
	messageRepository *repositories.MessageRepository,
	messageAttachmentRepo *repositories.MessageAttachmentRepository,
	fileService *services.FileService,
) *MessageHandler {
	return &MessageHandler{
		messageUseCase:        messageUseCase,
		userRepository:        userRepository,
		messageRepository:     messageRepository,
		messageAttachmentRepo: messageAttachmentRepo,
		fileService:           fileService,
	}
}

// SendMessageRequest represents the request payload for sending a message
type SendMessageRequest struct {
	ReceiverID string `json:"receiver_id"`
	Content    string `json:"content"`
}

// MessageAttachmentResponse represents an attachment in the message response
type MessageAttachmentResponse struct {
	ID       string `json:"id"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	MimeType string `json:"mime_type"`
}

// MessageResponse represents the response payload for a message
type MessageResponse struct {
	ID          string                      `json:"id"`
	SenderID    string                      `json:"sender_id"`
	ReceiverID  string                      `json:"receiver_id"`
	Content     string                      `json:"content"`
	Status      string                      `json:"status"`
	SentAt      string                      `json:"sent_at"`
	DeliveredAt *string                     `json:"delivered_at,omitempty"`
	ReadAt      *string                     `json:"read_at,omitempty"`
	Attachments []MessageAttachmentResponse `json:"attachments,omitempty"`
}

// UserListResponse represents a user in the users list (for messaging)
type UserListResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// SendMessage handles POST /api/v1/messages/send
// Sends a 1-to-1 message from authenticated user to another user
// Supports multipart/form-data with optional file attachments
func (h *MessageHandler) SendMessage(c *fiber.Ctx) error {
	// Get authenticated user ID from context
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		log.Printf("[ERROR] SendMessage - User not authenticated")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Check if content-type is multipart/form-data
	contentType := c.Get("Content-Type")
	var receiverID, content string
	var files []*multipart.FileHeader

	if strings.HasPrefix(contentType, "application/json") {
		// Handle JSON request (backward compatibility)
		req := new(SendMessageRequest)
		if err := c.BodyParser(req); err != nil {
			log.Printf("[ERROR] SendMessage - Failed to parse request: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request format",
			})
		}
		receiverID = req.ReceiverID
		content = req.Content
	} else if strings.HasPrefix(contentType, "multipart/form-data") {
		// Handle multipart/form-data request - parse form and files together
		form, err := c.MultipartForm()
		if err != nil {
			log.Printf("[ERROR] SendMessage - Failed to parse multipart form: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid multipart form",
			})
		}

		// Extract form values
		receiverIDArray := form.Value["receiver_id"]
		contentArray := form.Value["content"]

		if len(receiverIDArray) > 0 {
			receiverID = receiverIDArray[0]
		}
		if len(contentArray) > 0 {
			content = contentArray[0]
		}

		// Debug: log all form fields and files
		log.Printf("[DEBUG] SendMessage - Form values keys: %v", len(form.Value))
		log.Printf("[DEBUG] SendMessage - Form file keys: %v", len(form.File))
		for key := range form.File {
			log.Printf("[DEBUG] SendMessage - Found file field: %s with %d files", key, len(form.File[key]))
		}

		// Extract file attachments - check multiple possible field names
		if form.File["attachments"] != nil {
			files = form.File["attachments"]
			log.Printf("[DEBUG] SendMessage - Found attachments field with %d files", len(files))
		} else if form.File["files"] != nil {
			files = form.File["files"]
			log.Printf("[DEBUG] SendMessage - Found files field with %d files", len(files))
		} else {
			// Try to get the first file field if it has any files
			for key, fileList := range form.File {
				if len(fileList) > 0 {
					files = fileList
					log.Printf("[DEBUG] SendMessage - Using field '%s' with %d files", key, len(files))
					break
				}
			}
		}
	} else {
		// Handle other content types as form data
		receiverID = c.FormValue("receiver_id")
		content = c.FormValue("content")
	}

	// Validate request
	if receiverID == "" || content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "receiver_id and content are required",
		})
	}

	// Send message using use case
	message, err := h.messageUseCase.SendMessage(userID, receiverID, content)
	if err != nil {
		log.Printf("[ERROR] SendMessage - Failed to send message: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Handle file attachments if present
	if len(files) > 0 && h.fileService != nil && h.messageAttachmentRepo != nil {
		log.Printf("[DEBUG] SendMessage - Processing %d files, fileService=%v, repo=%v", len(files), h.fileService != nil, h.messageAttachmentRepo != nil)
		for _, fileHeader := range files {
			log.Printf("[DEBUG] SendMessage - Processing file: %s (size: %d bytes)", fileHeader.Filename, fileHeader.Size)
			// Validate file size (max 10MB per attachment)
			if fileHeader.Size > 10*1024*1024 {
				log.Printf("[WARNING] SendMessage - File %s exceeds 10MB size limit", fileHeader.Filename)
				continue
			}

			// Open uploaded file
			src, err := fileHeader.Open()
			if err != nil {
				log.Printf("[ERROR] SendMessage - Failed to open file: %v", err)
				continue
			}

			// Create temporary file
			tmpFile, err := os.CreateTemp("", "message-attachment-*")
			if err != nil {
				log.Printf("[ERROR] SendMessage - Failed to create temp file: %v", err)
				src.Close()
				continue
			}
			tmpFilePath := tmpFile.Name()

			// Copy uploaded file to temp file
			if _, err := io.Copy(tmpFile, src); err != nil {
				log.Printf("[ERROR] SendMessage - Failed to copy file: %v", err)
				src.Close()
				tmpFile.Close()
				os.Remove(tmpFilePath)
				continue
			}
			src.Close()

			// Reset file pointer to beginning before saving
			if _, err := tmpFile.Seek(0, 0); err != nil {
				log.Printf("[ERROR] SendMessage - Failed to seek file: %v", err)
				tmpFile.Close()
				os.Remove(tmpFilePath)
				continue
			}

			// Save file to storage (FileService will copy from tmpFile to storage/attachments/)
			filePath, err := h.fileService.UploadAttachment(tmpFile, fileHeader.Filename)
			tmpFile.Close()
			os.Remove(tmpFilePath)

			if err != nil {
				log.Printf("[ERROR] SendMessage - Failed to save file: %v", err)
				continue
			}

			log.Printf("[DEBUG] SendMessage - File saved to: %s", filePath)

			// Create attachment record in database
			attachment := &domain.MessageAttachment{
				MessageID: message.ID,
				FileName:  fileHeader.Filename,
				FilePath:  filePath,
				FileSize:  fileHeader.Size,
				MimeType:  fileHeader.Header.Get("Content-Type"),
			}

			savedAtt, err := h.messageAttachmentRepo.CreateAttachment(attachment)
			if err != nil {
				log.Printf("[ERROR] SendMessage - Failed to save attachment record: %v", err)
				h.fileService.DeleteAttachment(filePath)
				continue
			}
			log.Printf("[SUCCESS] SendMessage - Attachment saved: id=%s, file=%s", savedAtt.ID, fileHeader.Filename)
		}
	} else {
		log.Printf("[DEBUG] SendMessage - No attachments to process or dependencies missing. Files=%d, fileService=%v, repo=%v", len(files), h.fileService != nil, h.messageAttachmentRepo != nil)
	}

	// Build response - this will load attachments from database
	resp := h.messageToResponse(message)

	log.Printf("[SUCCESS] SendMessage - Message sent: from=%s, to=%s", userID, receiverID)
	return c.Status(fiber.StatusCreated).JSON(resp)
}

// GetConversation handles GET /api/v1/messages/conversation/:user_id
// Retrieves conversation between authenticated user and another user
func (h *MessageHandler) GetConversation(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Get other user ID from path parameter
	otherUserID := c.Params("user_id")
	if otherUserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user_id parameter is required",
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

	// Get conversation
	messages, err := h.messageUseCase.GetConversation(userID, otherUserID, limit, offset)
	if err != nil {
		log.Printf("[ERROR] GetConversation - Failed to retrieve conversation: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Convert to responses
	responses := make([]MessageResponse, len(messages))
	for i, msg := range messages {
		responses[i] = h.messageToResponse(msg)
	}

	return c.JSON(fiber.Map{
		"messages": responses,
		"count":    len(responses),
	})
}

// MarkMessageAsRead handles POST /api/v1/messages/:id/read
// Marks a message as read by the receiver
func (h *MessageHandler) MarkMessageAsRead(c *fiber.Ctx) error {
	messageID := c.Params("id")
	if messageID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "message id is required",
		})
	}

	// Mark as read
	if err := h.messageUseCase.MarkAsRead(messageID); err != nil {
		log.Printf("[ERROR] MarkMessageAsRead - Failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	log.Printf("[SUCCESS] MarkMessageAsRead - Message marked as read: %s", messageID)
	return c.JSON(fiber.Map{
		"success": true,
		"message": "message marked as read",
	})
}

// MarkMessageAsDelivered handles POST /api/v1/messages/:id/delivered
// Marks a message as delivered (used when client receives message via WebSocket)
func (h *MessageHandler) MarkMessageAsDelivered(c *fiber.Ctx) error {
	messageID := c.Params("id")
	if messageID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "message id is required",
		})
	}

	// Mark as delivered
	if err := h.messageUseCase.MarkAsDelivered(messageID); err != nil {
		log.Printf("[ERROR] MarkMessageAsDelivered - Failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	log.Printf("[SUCCESS] MarkMessageAsDelivered - Message marked as delivered: %s", messageID)
	return c.JSON(fiber.Map{
		"success": true,
		"message": "message marked as delivered",
	})
}

// GetUnreadCount handles GET /api/v1/messages/unread/count
// Returns count of unread messages for authenticated user
func (h *MessageHandler) GetUnreadCount(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Get unread count
	count, err := h.messageUseCase.GetUnreadCount(userID)
	if err != nil {
		log.Printf("[ERROR] GetUnreadCount - Failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"user_id":      userID,
		"unread_count": count,
	})
}

// DeleteMessage handles DELETE /api/v1/messages/:id
// Deletes a message (only sender or receiver can delete)
func (h *MessageHandler) DeleteMessage(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	messageID := c.Params("id")
	if messageID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "message id is required",
		})
	}

	// Delete message
	if err := h.messageUseCase.DeleteMessage(messageID, userID); err != nil {
		log.Printf("[ERROR] DeleteMessage - Failed: %v", err)
		if err.Error() == "unauthorized: only sender or receiver can delete message" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	log.Printf("[SUCCESS] DeleteMessage - Message deleted: %s", messageID)
	return c.JSON(fiber.Map{
		"success": true,
		"message": "message deleted",
	})
}

// Helper function to convert domain.Message to MessageResponse
func (h *MessageHandler) messageToResponse(msg *domain.Message) MessageResponse {
	resp := MessageResponse{
		ID:         msg.ID,
		SenderID:   msg.SenderID,
		ReceiverID: msg.ReceiverID,
		Content:    msg.Content,
		Status:     string(msg.Status),
		SentAt:     msg.SentAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if msg.DeliveredAt != nil {
		deliveredAt := msg.DeliveredAt.Format("2006-01-02T15:04:05Z07:00")
		resp.DeliveredAt = &deliveredAt
	}

	if msg.ReadAt != nil {
		readAt := msg.ReadAt.Format("2006-01-02T15:04:05Z07:00")
		resp.ReadAt = &readAt
	}

	// Load attachments if repository is available
	if h.messageAttachmentRepo != nil {
		attachments, err := h.messageAttachmentRepo.GetAttachmentsByMessageID(msg.ID)
		if err != nil {
			log.Printf("[DEBUG] messageToResponse - Error loading attachments: %v", err)
		} else {
			log.Printf("[DEBUG] messageToResponse - Found %d attachments for message %s", len(attachments), msg.ID)
			if len(attachments) > 0 {
				resp.Attachments = make([]MessageAttachmentResponse, len(attachments))
				for i, att := range attachments {
					resp.Attachments[i] = MessageAttachmentResponse{
						ID:       att.ID,
						FileName: att.FileName,
						FileSize: att.FileSize,
						MimeType: att.MimeType,
					}
				}
			}
		}
	} else {
		log.Printf("[DEBUG] messageToResponse - messageAttachmentRepo is nil")
	}

	return resp
}

// GetUsers handles GET /api/v1/messages/users
// Returns all users (admin, pentester, stakeholder) with uuid, email, role, first_name, last_name
func (h *MessageHandler) GetUsers(c *fiber.Ctx) error {
	// Get authenticated user ID from context
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		log.Printf("[ERROR] GetUsers - User not authenticated")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Get user repository - try from context first, then from handler
	var userRepo domain.UserRepository
	if repo, ok := c.Locals("userRepository").(domain.UserRepository); ok && repo != nil {
		userRepo = repo
	} else if h.userRepository != nil {
		userRepo = h.userRepository
	} else {
		log.Printf("[ERROR] GetUsers - User repository not initialized")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "user repository not available",
		})
	}

	// Type assert to get the implementation
	var allUsers []*repositories.UserWithName
	var err error

	if userRepoImpl, ok := userRepo.(*repositories.UserRepositoryImpl); ok {
		allUsers, err = userRepoImpl.GetAllUsersWithNames()
	} else {
		log.Printf("[WARN] GetUsers - Could not cast to UserRepositoryImpl")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "repository method not available",
		})
	}

	if err != nil {
		log.Printf("[ERROR] GetUsers - Failed to get users: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch users",
		})
	}

	// Convert to response format
	var response []UserListResponse
	for _, user := range allUsers {
		response = append(response, UserListResponse{
			ID:        user.ID,
			Email:     user.Email,
			Role:      user.Role,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		})
	}

	log.Printf("[SUCCESS] GetUsers - Retrieved %d users", len(response))
	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
		"count":   len(response),
	})
}

// GetChats handles GET /api/v1/messages
// Retrieves all chats (conversations) for the authenticated user
func (h *MessageHandler) GetChats(c *fiber.Ctx) error {
	// Get authenticated user ID from context
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		log.Printf("[ERROR] GetChats - User not authenticated")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Get chats from message repository
	chats, err := h.messageRepository.GetUserChats(userID)
	if err != nil {
		log.Printf("[ERROR] GetChats - Failed to get chats: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch chats",
		})
	}

	// Handle empty chats
	if chats == nil {
		chats = []*repositories.ChatInfo{}
	}

	log.Printf("[SUCCESS] GetChats - Retrieved %d chats for user %s", len(chats), userID)
	return c.JSON(fiber.Map{
		"success": true,
		"data":    chats,
		"count":   len(chats),
	})
}

// DownloadMessageAttachment handles GET /api/v1/messages/attachments/:attachment_id
// Allows downloading an attachment from a message
func (h *MessageHandler) DownloadMessageAttachment(c *fiber.Ctx) error {
	// Get authenticated user ID from context
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		log.Printf("[ERROR] DownloadMessageAttachment - User not authenticated")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	attachmentID := c.Params("attachment_id")
	if attachmentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "attachment_id is required",
		})
	}

	// Check if attachment repository is available
	if h.messageAttachmentRepo == nil || h.fileService == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "attachment service not available",
		})
	}

	// Get attachment from database
	attachment, err := h.messageAttachmentRepo.GetAttachmentByID(attachmentID)
	if err != nil {
		log.Printf("[ERROR] DownloadMessageAttachment - Attachment not found: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "attachment not found",
		})
	}

	// Read file from storage
	fileData, err := h.fileService.ReadFile(attachment.FilePath)
	if err != nil {
		log.Printf("[ERROR] DownloadMessageAttachment - Failed to read file: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to read file",
		})
	}

	// Set response headers
	c.Set("Content-Disposition", "attachment; filename="+attachment.FileName)
	c.Set("Content-Type", attachment.MimeType)
	c.Set("Content-Length", string(rune(len(fileData))))

	log.Printf("[SUCCESS] DownloadMessageAttachment - Downloaded: %s by user %s", attachment.FileName, userID)
	return c.Send(fileData)
}
