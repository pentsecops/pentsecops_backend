package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/pentsecops/backend/internal/core/domain"
	"github.com/pentsecops/backend/internal/core/usecases"
	"github.com/pentsecops/backend/pkg/middleware"
	"github.com/pentsecops/backend/pkg/services"
)

// ChatHandler handles chat-related HTTP requests
type ChatHandler struct {
	chatUseCase    *usecases.ChatUseCase
	tokenManager   *middleware.TokenManager
	documentParser *services.DocumentParser
	knowledgeBase  *services.KnowledgeBase
}

// NewChatHandler creates a new chat handler
func NewChatHandler(chatUseCase *usecases.ChatUseCase, tokenManager *middleware.TokenManager) *ChatHandler {
	return &ChatHandler{
		chatUseCase:    chatUseCase,
		tokenManager:   tokenManager,
		documentParser: services.NewDocumentParser(),
		knowledgeBase:  services.NewKnowledgeBase(),
	}
}

// ProcessChat handles POST /api/v1/chats requests with optional file attachments
func (ch *ChatHandler) ProcessChat(c *fiber.Ctx) error {
	log.Printf("[API] POST /api/v1/chats - Processing chat request")

	// Extract and validate token
	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		log.Printf("[ERROR] ProcessChat: Missing authentication token")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized: missing token",
		})
	}

	claims, err := ch.tokenManager.VerifyToken(tokenString)
	if err != nil {
		log.Printf("[ERROR] ProcessChat: Token validation failed: %v", err)
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized: invalid token",
		})
	}

	userIDStr := claims.UserID
	userRole := claims.Role

	// Get query from form data
	query := c.FormValue("query", "")
	if query == "" {
		log.Printf("[ERROR] ProcessChat: Empty query")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Query is required",
		})
	}

	log.Printf("[API] Processing chat for user %s (role: %s) with query length: %d", userIDStr, userRole, len(query))

	// Initialize document content
	var documentContent string

	// Parse uploaded files if present
	form, err := c.MultipartForm()
	if err == nil && form != nil {
		files := form.File["documents"]
		if len(files) > 0 {
			log.Printf("[INFO] ProcessChat: Processing %d uploaded documents", len(files))
			var allDocumentText []string

			for idx, file := range files {
				log.Printf("[INFO] ProcessChat: Parsing file %d: %s", idx+1, file.Filename)

				// Validate file size (max 10MB)
				if file.Size > 10*1024*1024 {
					log.Printf("[ERROR] ProcessChat: File %s exceeds 10MB limit", file.Filename)
					return c.Status(http.StatusBadRequest).JSON(fiber.Map{
						"success": false,
						"error":   fmt.Sprintf("File %s exceeds 10MB limit", file.Filename),
					})
				}

				// Parse document
				text, parseErr := ch.documentParser.ParseDocument(file)
				if parseErr != nil {
					log.Printf("[ERROR] ProcessChat: Failed to parse file %s: %v", file.Filename, parseErr)
					return c.Status(http.StatusBadRequest).JSON(fiber.Map{
						"success": false,
						"error":   fmt.Sprintf("Failed to parse file %s: %v", file.Filename, parseErr),
					})
				}

				allDocumentText = append(allDocumentText, fmt.Sprintf("--- Document: %s ---\n%s", file.Filename, text))
			}

			// Combine all document text
			documentContent = fmt.Sprintf("UPLOADED DOCUMENTS CONTEXT:\n\n%s\n\n", fmt.Sprintf("%v", allDocumentText))
			log.Printf("[SUCCESS] ProcessChat: Parsed %d documents with total %d characters", len(files), len(documentContent))
		}
	}

	// Create chat request with document context
	request := &domain.ChatRequest{
		Query:           query,
		DocumentContent: documentContent,
	}

	// Process chat with history and documents
	response, err := ch.chatUseCase.ProcessChatWithHistoryAndDocuments(
		userIDStr, userRole, request, ch.knowledgeBase,
	)
	if err != nil {
		log.Printf("[ERROR] ProcessChat: Use case error: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	if !response.Success {
		log.Printf("[ERROR] ProcessChat: Response indicates failure: %s", response.Error)
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	log.Printf("[SUCCESS] ProcessChat: Chat processed for user %s", userIDStr)
	return c.Status(http.StatusOK).JSON(response)
}

// GetConversationHistory handles GET /api/v1/chats/history requests
func (ch *ChatHandler) GetConversationHistory(c *fiber.Ctx) error {
	log.Printf("[API] GET /api/v1/chats/history - Retrieving conversation history")

	// Extract and validate token
	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		log.Printf("[ERROR] GetConversationHistory: Missing authentication token")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized: missing token",
		})
	}

	claims, err := ch.tokenManager.VerifyToken(tokenString)
	if err != nil {
		log.Printf("[ERROR] GetConversationHistory: Token validation failed: %v", err)
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized: invalid token",
		})
	}

	userIDStr := claims.UserID

	// Get optional limit from query params
	limit := c.QueryInt("limit", 50)
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	log.Printf("[API] Retrieving conversation history for user %s with limit %d", userIDStr, limit)

	// Get conversation history
	response, err := ch.chatUseCase.GetConversationHistory(userIDStr, limit)
	if err != nil {
		log.Printf("[ERROR] GetConversationHistory: Use case error: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve conversation history",
		})
	}

	log.Printf("[SUCCESS] GetConversationHistory: Retrieved %d messages for user %s", len(response.Messages), userIDStr)
	return c.Status(http.StatusOK).JSON(response)
}

// ClearConversationHistory handles DELETE /api/v1/chats/history requests
func (ch *ChatHandler) ClearConversationHistory(c *fiber.Ctx) error {
	log.Printf("[API] DELETE /api/v1/chats/history - Clearing conversation history")

	// Extract and validate token
	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		log.Printf("[ERROR] ClearConversationHistory: Missing authentication token")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized: missing token",
		})
	}

	claims, err := ch.tokenManager.VerifyToken(tokenString)
	if err != nil {
		log.Printf("[ERROR] ClearConversationHistory: Token validation failed: %v", err)
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized: invalid token",
		})
	}

	userIDStr := claims.UserID

	log.Printf("[API] Clearing conversation history for user %s", userIDStr)

	// Clear history
	err = ch.chatUseCase.ClearConversationHistory(userIDStr)
	if err != nil {
		log.Printf("[ERROR] ClearConversationHistory: Use case error: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to clear conversation history",
		})
	}

	log.Printf("[SUCCESS] ClearConversationHistory: Cleared history for user %s", userIDStr)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Conversation history cleared",
	})
}
