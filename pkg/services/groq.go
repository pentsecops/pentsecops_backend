package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// GroqService handles communication with Groq API
type GroqService struct {
	apiKey  string
	apiURL  string
	enabled bool
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GroqRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float32   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

type GroqResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// NewGroqService creates a new Groq service
func NewGroqService() (*GroqService, error) {
	apiKey := os.Getenv("GROQ_API_KEY")

	if apiKey == "" {
		log.Printf("[WARN] Groq service disabled: missing GROQ_API_KEY")
		return &GroqService{enabled: false}, nil
	}

	service := &GroqService{
		apiKey:  apiKey,
		apiURL:  "https://api.groq.com/openai/v1/chat/completions",
		enabled: true,
	}

	log.Printf("[INFO] Groq AI service initialized successfully")
	return service, nil
}

// IsEnabled returns true if Groq service is properly configured
func (gs *GroqService) IsEnabled() bool {
	return gs.enabled
}

// SendMessage sends a message to Groq API and receives a response
func (gs *GroqService) SendMessage(userQuery string) (string, error) {
	if !gs.enabled {
		return "", fmt.Errorf("Groq service is not enabled")
	}

	// Prepare the request
	payload := GroqRequest{
		Model: "llama-3.3-70b-versatile", // Using Groq's high-performance model
		Messages: []Message{
			{
				Role:    "user",
				Content: userQuery,
			},
		},
		Temperature: 0.7,
		MaxTokens:   2048,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal Groq request: %v", err)
		return "", fmt.Errorf("failed to prepare request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", gs.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[ERROR] Failed to create HTTP request: %v", err)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", gs.apiKey))

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ERROR] Failed to call Groq API: %v", err)
		return "", fmt.Errorf("failed to call Groq API: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read Groq response: %v", err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] Groq API returned status %d: %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("Groq API error: status %d", resp.StatusCode)
	}

	// Parse response
	var groqResp GroqResponse
	err = json.Unmarshal(body, &groqResp)
	if err != nil {
		log.Printf("[ERROR] Failed to parse Groq response: %v", err)
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if groqResp.Error != nil {
		log.Printf("[ERROR] Groq API error: %s", groqResp.Error.Message)
		return "", fmt.Errorf("Groq API error: %s", groqResp.Error.Message)
	}

	// Check if we got a response
	if len(groqResp.Choices) == 0 {
		log.Printf("[ERROR] Groq API returned no choices")
		return "", fmt.Errorf("Groq API returned no response")
	}

	responseText := groqResp.Choices[0].Message.Content
	log.Printf("[SUCCESS] Groq API response received successfully")
	return responseText, nil
}

// SendMessageWithHistory sends a message with conversation history to Groq API
func (gs *GroqService) SendMessageWithHistory(messages []Message) (string, error) {
	if !gs.enabled {
		return "", fmt.Errorf("Groq service is not enabled")
	}

	// Prepare the request
	payload := GroqRequest{
		Model:       "llama-3.3-70b-versatile",
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   2048,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal Groq request: %v", err)
		return "", fmt.Errorf("failed to prepare request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", gs.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[ERROR] Failed to create HTTP request: %v", err)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", gs.apiKey))

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ERROR] Failed to call Groq API: %v", err)
		return "", fmt.Errorf("failed to call Groq API: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read Groq response: %v", err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] Groq API returned status %d: %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("Groq API error: status %d", resp.StatusCode)
	}

	// Parse response
	var groqResp GroqResponse
	err = json.Unmarshal(body, &groqResp)
	if err != nil {
		log.Printf("[ERROR] Failed to parse Groq response: %v", err)
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if groqResp.Error != nil {
		log.Printf("[ERROR] Groq API error: %s", groqResp.Error.Message)
		return "", fmt.Errorf("Groq API error: %s", groqResp.Error.Message)
	}

	// Check if we got a response
	if len(groqResp.Choices) == 0 {
		log.Printf("[ERROR] Groq API returned no choices")
		return "", fmt.Errorf("Groq API returned no response")
	}

	responseText := groqResp.Choices[0].Message.Content
	log.Printf("[SUCCESS] Groq API response received successfully")
	return responseText, nil
}
