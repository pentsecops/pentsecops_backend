package usecases

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/pentsecops/backend/internal/adapter/repositories"
	"github.com/pentsecops/backend/internal/core/domain"
	"github.com/pentsecops/backend/pkg/services"
)

// ChatUseCase handles chat business logic
type ChatUseCase struct {
	chatRepo    *repositories.ChatRepository
	groqService *services.GroqService
	userRepo    domain.UserRepository
}

// NewChatUseCase creates a new chat use case
func NewChatUseCase(chatRepo *repositories.ChatRepository, groqService *services.GroqService, userRepo domain.UserRepository) *ChatUseCase {
	return &ChatUseCase{
		chatRepo:    chatRepo,
		groqService: groqService,
		userRepo:    userRepo,
	}
}

// ProcessChat sends a user query to Groq and returns the response
func (cu *ChatUseCase) ProcessChat(userID string, userRole string, request *domain.ChatRequest) (*domain.ChatResponse, error) {
	if userID == "" {
		log.Printf("[ERROR] ProcessChat: User ID is required")
		return &domain.ChatResponse{
			Success: false,
			Error:   "User ID is required",
		}, fmt.Errorf("user ID is required")
	}

	if userRole == "" {
		log.Printf("[ERROR] ProcessChat: User role is required")
		return &domain.ChatResponse{
			Success: false,
			Error:   "User role is required",
		}, fmt.Errorf("user role is required")
	}

	// Verify user exists - for non-admin users, check users table. For admin, no check needed
	if userRole != "admin" {
		_, err := cu.userRepo.GetUserByID(userID)
		if err != nil {
			log.Printf("[ERROR] ProcessChat: User not found: %v", err)
			return &domain.ChatResponse{
				Success: false,
				Error:   "User not found",
			}, fmt.Errorf("user not found: %w", err)
		}
	}

	if request.Query == "" {
		log.Printf("[ERROR] ProcessChat: Query is required")
		return &domain.ChatResponse{
			Success: false,
			Error:   "Query is required",
		}, fmt.Errorf("query is required")
	}

	// Create user message
	userMessage := &domain.ChatMessage{
		ID:        uuid.New().String(),
		UserID:    userID,
		UserType:  userRole,
		Role:      "user",
		Content:   request.Query,
		CreatedAt: time.Now(),
	}

	// Save user message to database
	saveErr := cu.chatRepo.SaveMessage(userMessage)
	if saveErr != nil {
		log.Printf("[ERROR] ProcessChat: Failed to save user message: %v", saveErr)
		return &domain.ChatResponse{
			Success: false,
			Error:   "Failed to save message",
		}, saveErr
	}

	// Send to Groq API
	log.Printf("[API] Sending query to Groq API for user %s", userID)
	assistantResponse, err := cu.groqService.SendMessage(request.Query)
	if err != nil {
		log.Printf("[ERROR] ProcessChat: Groq API error: %v", err)
		return &domain.ChatResponse{
			Success: false,
			Error:   fmt.Sprintf("AI service error: %v", err),
		}, err
	}

	// Create assistant message
	assistantMessage := &domain.ChatMessage{
		ID:        uuid.New().String(),
		UserID:    userID,
		UserType:  userRole,
		Role:      "assistant",
		Content:   assistantResponse,
		CreatedAt: time.Now(),
	}

	// Save assistant message to database
	err = cu.chatRepo.SaveMessage(assistantMessage)
	if err != nil {
		log.Printf("[ERROR] ProcessChat: Failed to save assistant message: %v", err)
		return &domain.ChatResponse{
			Success: false,
			Error:   "Failed to save response",
		}, err
	}

	// Return success response
	response := &domain.ChatResponse{
		Messages: []domain.ChatMessage{
			*userMessage,
			*assistantMessage,
		},
		Success: true,
	}

	log.Printf("[SUCCESS] ProcessChat: Query processed successfully for user %s", userID)
	return response, nil
}

// GetConversationHistory retrieves recent chat messages for a user
func (cu *ChatUseCase) GetConversationHistory(userID string, limit int) (*domain.ChatResponse, error) {
	if userID == "" {
		log.Printf("[ERROR] GetConversationHistory: User ID is required")
		return &domain.ChatResponse{
			Success: false,
			Error:   "User ID is required",
		}, fmt.Errorf("user ID is required")
	}

	if limit <= 0 {
		limit = 50
	}

	messages, err := cu.chatRepo.GetChatHistory(userID, limit)
	if err != nil {
		log.Printf("[ERROR] GetConversationHistory: Failed to retrieve history: %v", err)
		return &domain.ChatResponse{
			Success: false,
			Error:   "Failed to retrieve conversation history",
		}, err
	}

	response := &domain.ChatResponse{
		Messages: messages,
		Success:  true,
	}

	log.Printf("[SUCCESS] GetConversationHistory: Retrieved %d messages for user %s", len(messages), userID)
	return response, nil
}

// ClearConversationHistory deletes all messages for a user
func (cu *ChatUseCase) ClearConversationHistory(userID string) error {
	if userID == "" {
		log.Printf("[ERROR] ClearConversationHistory: User ID is required")
		return fmt.Errorf("user ID is required")
	}

	err := cu.chatRepo.DeleteUserMessages(userID)
	if err != nil {
		log.Printf("[ERROR] ClearConversationHistory: Failed to clear history: %v", err)
		return err
	}

	log.Printf("[SUCCESS] ClearConversationHistory: Cleared all messages for user %s", userID)
	return nil
}

// ProcessChatWithHistory sends a message with conversation history to Groq
func (cu *ChatUseCase) ProcessChatWithHistory(userID string, userRole string, request *domain.ChatRequest) (*domain.ChatResponse, error) {
	if userID == "" {
		log.Printf("[ERROR] ProcessChatWithHistory: User ID is required")
		return &domain.ChatResponse{
			Success: false,
			Error:   "User ID is required",
		}, fmt.Errorf("user ID is required")
	}

	if userRole == "" {
		log.Printf("[ERROR] ProcessChatWithHistory: User role is required")
		return &domain.ChatResponse{
			Success: false,
			Error:   "User role is required",
		}, fmt.Errorf("user role is required")
	}

	// Verify user exists - for non-admin users, check users table. For admin, no check needed
	if userRole != "admin" {
		_, err := cu.userRepo.GetUserByID(userID)
		if err != nil {
			log.Printf("[ERROR] ProcessChatWithHistory: User not found: %v", err)
			return &domain.ChatResponse{
				Success: false,
				Error:   "User not found",
			}, fmt.Errorf("user not found: %w", err)
		}
	}

	if request.Query == "" {
		log.Printf("[ERROR] ProcessChatWithHistory: Query is required")
		return &domain.ChatResponse{
			Success: false,
			Error:   "Query is required",
		}, fmt.Errorf("query is required")
	}

	// Retrieve conversation history (last 10 messages for context)
	history, histErr := cu.chatRepo.GetChatHistory(userID, 10)
	if histErr != nil {
		log.Printf("[ERROR] ProcessChatWithHistory: Failed to retrieve history: %v", histErr)
		// Continue without history - not a fatal error
		history = []domain.ChatMessage{}
	}

	// Build message array for Groq (reverse order since we got DESC)
	var groqMessages []services.Message
	for i := len(history) - 1; i >= 0; i-- {
		groqMessages = append(groqMessages, services.Message{
			Role:    history[i].Role,
			Content: history[i].Content,
		})
	}

	// Add current user message
	groqMessages = append(groqMessages, services.Message{
		Role:    "user",
		Content: request.Query,
	})

	// Create user message record
	userMessage := &domain.ChatMessage{
		ID:        uuid.New().String(),
		UserID:    userID,
		UserType:  userRole,
		Role:      "user",
		Content:   request.Query,
		CreatedAt: time.Now(),
	}

	// Save user message to database
	saveErr := cu.chatRepo.SaveMessage(userMessage)
	if saveErr != nil {
		log.Printf("[ERROR] ProcessChatWithHistory: Failed to save user message: %v", saveErr)
		return &domain.ChatResponse{
			Success: false,
			Error:   "Failed to save message",
		}, saveErr
	}

	// Send to Groq API with history
	log.Printf("[API] Sending query to Groq API with history for user %s", userID)
	assistantResponse, groqErr := cu.groqService.SendMessageWithHistory(groqMessages)
	if groqErr != nil {
		log.Printf("[ERROR] ProcessChatWithHistory: Groq API error: %v", groqErr)
		return &domain.ChatResponse{
			Success: false,
			Error:   fmt.Sprintf("AI service error: %v", groqErr),
		}, groqErr
	}

	// Create assistant message
	assistantMessage := &domain.ChatMessage{
		ID:        uuid.New().String(),
		UserID:    userID,
		UserType:  userRole,
		Role:      "assistant",
		Content:   assistantResponse,
		CreatedAt: time.Now(),
	}

	// Save assistant message to database
	assistantSaveErr := cu.chatRepo.SaveMessage(assistantMessage)
	if assistantSaveErr != nil {
		log.Printf("[ERROR] ProcessChatWithHistory: Failed to save assistant message: %v", assistantSaveErr)
		return &domain.ChatResponse{
			Success: false,
			Error:   "Failed to save response",
		}, assistantSaveErr
	}

	// Return success response with user and assistant messages only
	response := &domain.ChatResponse{
		Messages: []domain.ChatMessage{
			*userMessage,
			*assistantMessage,
		},
		Success: true,
	}

	log.Printf("[SUCCESS] ProcessChatWithHistory: Query processed successfully for user %s", userID)
	return response, nil
}

// ProcessChatWithHistoryAndDocuments processes chat with document context and checks knowledge base
func (cu *ChatUseCase) ProcessChatWithHistoryAndDocuments(
	userID string,
	userRole string,
	request *domain.ChatRequest,
	kb *services.KnowledgeBase,
) (*domain.ChatResponse, error) {
	if userID == "" {
		log.Printf("[ERROR] ProcessChatWithHistoryAndDocuments: User ID is required")
		return &domain.ChatResponse{
			Success: false,
			Error:   "User ID is required",
		}, fmt.Errorf("user ID is required")
	}

	if userRole == "" {
		log.Printf("[ERROR] ProcessChatWithHistoryAndDocuments: User role is required")
		return &domain.ChatResponse{
			Success: false,
			Error:   "User role is required",
		}, fmt.Errorf("user role is required")
	}

	// Verify user exists - for non-admin users, check users table
	if userRole != "admin" {
		_, err := cu.userRepo.GetUserByID(userID)
		if err != nil {
			log.Printf("[ERROR] ProcessChatWithHistoryAndDocuments: User not found: %v", err)
			return &domain.ChatResponse{
				Success: false,
				Error:   "User not found",
			}, fmt.Errorf("user not found: %w", err)
		}
	}

	if request.Query == "" {
		log.Printf("[ERROR] ProcessChatWithHistoryAndDocuments: Query is required")
		return &domain.ChatResponse{
			Success: false,
			Error:   "Query is required",
		}, fmt.Errorf("query is required")
	}

	// Create user message record
	userMessage := &domain.ChatMessage{
		ID:        uuid.New().String(),
		UserID:    userID,
		UserType:  userRole,
		Role:      "user",
		Content:   request.Query,
		CreatedAt: time.Now(),
	}

	// Save user message to database
	saveErr := cu.chatRepo.SaveMessage(userMessage)
	if saveErr != nil {
		log.Printf("[ERROR] ProcessChatWithHistoryAndDocuments: Failed to save user message: %v", saveErr)
		return &domain.ChatResponse{
			Success: false,
			Error:   "Failed to save message",
		}, saveErr
	}

	// Check if question is about Pentsecops platform
	var assistantResponse string
	isPentsecopsQuestion := false

	if kb != nil {
		isPentsecopsQuestion = kb.IsPentsecopsQuestion(request.Query)
		if isPentsecopsQuestion {
			log.Printf("[INFO] ProcessChatWithHistoryAndDocuments: Detected Pentsecops platform question")
			answer, found := kb.SearchKnowledgeBase(request.Query)
			if found {
				assistantResponse = answer
				log.Printf("[INFO] ProcessChatWithHistoryAndDocuments: Using knowledge base answer for Pentsecops question")
			}
		}
	}

	// If not a Pentsecops question or knowledge base search failed, use Groq API
	if !isPentsecopsQuestion || assistantResponse == "" {
		// Retrieve conversation history (last 10 messages for context)
		history, histErr := cu.chatRepo.GetChatHistory(userID, 10)
		if histErr != nil {
			log.Printf("[ERROR] ProcessChatWithHistoryAndDocuments: Failed to retrieve history: %v", histErr)
			history = []domain.ChatMessage{}
		}

		// Build message array for Groq (reverse order since we got DESC)
		var groqMessages []services.Message
		for i := len(history) - 1; i >= 0; i-- {
			groqMessages = append(groqMessages, services.Message{
				Role:    history[i].Role,
				Content: history[i].Content,
			})
		}

		// Prepare query with document context if available
		finalQuery := request.Query
		if request.DocumentContent != "" {
			finalQuery = fmt.Sprintf("%s\n\n%s", request.DocumentContent, request.Query)
			log.Printf("[INFO] ProcessChatWithHistoryAndDocuments: Appended document context to query")
		}

		// Add current user message with document context
		groqMessages = append(groqMessages, services.Message{
			Role:    "user",
			Content: finalQuery,
		})

		// Send to Groq API with history and document context
		log.Printf("[API] Sending query to Groq API with documents for user %s", userID)
		var groqErr error
		assistantResponse, groqErr = cu.groqService.SendMessageWithHistory(groqMessages)
		if groqErr != nil {
			log.Printf("[ERROR] ProcessChatWithHistoryAndDocuments: Groq API error: %v", groqErr)
			return &domain.ChatResponse{
				Success: false,
				Error:   fmt.Sprintf("AI service error: %v", groqErr),
			}, groqErr
		}
	}

	// Create assistant message
	assistantMessage := &domain.ChatMessage{
		ID:        uuid.New().String(),
		UserID:    userID,
		UserType:  userRole,
		Role:      "assistant",
		Content:   assistantResponse,
		CreatedAt: time.Now(),
	}

	// Save assistant message to database
	assistantSaveErr := cu.chatRepo.SaveMessage(assistantMessage)
	if assistantSaveErr != nil {
		log.Printf("[ERROR] ProcessChatWithHistoryAndDocuments: Failed to save assistant message: %v", assistantSaveErr)
		return &domain.ChatResponse{
			Success: false,
			Error:   "Failed to save response",
		}, assistantSaveErr
	}

	// Return success response
	response := &domain.ChatResponse{
		Messages: []domain.ChatMessage{
			*userMessage,
			*assistantMessage,
		},
		Success: true,
	}

	if isPentsecopsQuestion {
		log.Printf("[SUCCESS] ProcessChatWithHistoryAndDocuments: Pentsecops question answered from knowledge base for user %s", userID)
	} else {
		log.Printf("[SUCCESS] ProcessChatWithHistoryAndDocuments: Query processed via Groq API for user %s", userID)
	}
	return response, nil
}
