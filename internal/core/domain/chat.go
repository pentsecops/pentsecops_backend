package domain

import "time"

// ChatMessage represents a message in the chat
type ChatMessage struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	UserType  string    `json:"user_type"` // admin, pentester, or stakeholder
	Role      string    `json:"role"`      // user or assistant
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ChatRequest represents a user's chat request with optional document content
type ChatRequest struct {
	Query           string
	Attachments     []string
	DocumentContent string
}

// ChatResponse represents the response from Groq API
type ChatResponse struct {
	Messages []ChatMessage `json:"messages"`
	Success  bool          `json:"success"`
	Error    string        `json:"error,omitempty"`
}
