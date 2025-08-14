package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// AIService handles AI processing requests
type AIService struct {
	logger     *zap.Logger
	httpClient *http.Client
	aiURL      string
	aiToken    string
}

// NewAIService creates a new AI service
func NewAIService(logger *zap.Logger, aiURL, aiToken string) *AIService {
	return &AIService{
		logger: logger,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		aiURL:   aiURL,
		aiToken: aiToken,
	}
}

// AIRequest represents the request structure for AI processing
type AIRequest struct {
	MessageID   string                 `json:"message_id"`
	SessionID   string                 `json:"session_id"`
	Message     string                 `json:"message"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Suggestion  bool                   `json:"suggestion,omitempty"`
	Attachments []map[string]interface{} `json:"attachments,omitempty"`
}

// AIResponse represents the response structure from AI processing
type AIResponse struct {
	MessageID   string                 `json:"message_id"`
	SessionID   string                 `json:"session_id"`
	Response    string                 `json:"response"`
	Suggestions []string               `json:"suggestions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// ProcessAIRequest sends a request to the AI service and returns the response
func (ai *AIService) ProcessAIRequest(ctx context.Context, request AIRequest) (*AIResponse, error) {
	ai.logger.Info("Processing AI request",
		zap.String("message_id", request.MessageID),
		zap.String("session_id", request.SessionID),
		zap.Bool("suggestion", request.Suggestion))

	// Marshal request to JSON
	requestBytes, err := json.Marshal(request)
	if err != nil {
		ai.logger.Error("Failed to marshal AI request", zap.Error(err))
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", ai.aiURL, bytes.NewBuffer(requestBytes))
	if err != nil {
		ai.logger.Error("Failed to create AI request", zap.Error(err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ai.aiToken))
	req.Header.Set("User-Agent", "Fraiday-AI-Client/1.0")

	// Send request
	resp, err := ai.httpClient.Do(req)
	if err != nil {
		ai.logger.Error("Failed to send AI request", zap.Error(err))
		return nil, fmt.Errorf("failed to send AI request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		ai.logger.Warn("AI service returned non-success status",
			zap.Int("status_code", resp.StatusCode))
		return nil, fmt.Errorf("AI service returned status %d", resp.StatusCode)
	}

	// Parse response
	var aiResponse AIResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResponse); err != nil {
		ai.logger.Error("Failed to decode AI response", zap.Error(err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	ai.logger.Info("AI request processed successfully",
		zap.String("message_id", request.MessageID),
		zap.String("response_length", fmt.Sprintf("%d", len(aiResponse.Response))))

	return &aiResponse, nil
}

// GenerateChatResponse generates a chat response using AI
func (ai *AIService) GenerateChatResponse(ctx context.Context, messageID, sessionID, message string, context map[string]interface{}) (*AIResponse, error) {
	request := AIRequest{
		MessageID: messageID,
		SessionID: sessionID,
		Message:   message,
		Context:   context,
		Suggestion: false,
	}

	return ai.ProcessAIRequest(ctx, request)
}

// GenerateSuggestions generates suggestions using AI
func (ai *AIService) GenerateSuggestions(ctx context.Context, messageID, sessionID, message string, context map[string]interface{}) (*AIResponse, error) {
	request := AIRequest{
		MessageID: messageID,
		SessionID: sessionID,
		Message:   message,
		Context:   context,
		Suggestion: true,
	}

	return ai.ProcessAIRequest(ctx, request)
}

// HealthCheck checks if the AI service is available
func (ai *AIService) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", ai.aiURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ai.aiToken))

	resp, err := ai.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send health check request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("AI service health check failed with status %d", resp.StatusCode)
	}

	return nil
}