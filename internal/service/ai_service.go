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
	logger        *zap.Logger
	httpClient    *http.Client
	aiURL         string
	aiToken       string
	slackAIURL    string
	slackAIToken  string
	slackWorkflowID string
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

// NewAIServiceWithSlack creates a new AI service with Slack configuration
func NewAIServiceWithSlack(logger *zap.Logger, aiURL, aiToken, slackAIURL, slackAIToken, slackWorkflowID string) *AIService {
	return &AIService{
		logger: logger,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		aiURL:           aiURL,
		aiToken:         aiToken,
		slackAIURL:      slackAIURL,
		slackAIToken:    slackAIToken,
		slackWorkflowID: slackWorkflowID,
	}
}

// AIRequest represents the request structure for AI processing
type AIRequest struct {
	MessageID         string                 `json:"message_id"`
	SessionID         string                 `json:"session_id"`
	CurrentMessage    string                 `json:"current_message"`
	ChatHistory       []interface{}          `json:"chat_history,omitempty"`
	SenderID          string                 `json:"sender_id"`
	SenderType        string                 `json:"sender_type"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	CurrentMessageID  string                 `json:"current_message_id"`
	Context           map[string]interface{} `json:"context,omitempty"`
	Suggestion        bool                   `json:"suggestion,omitempty"`
	Attachments       []map[string]interface{} `json:"attachments,omitempty"`
}

// AIAttachment represents an attachment in AI response
type AIAttachment struct {
	Type     string                 `json:"type"`
	FileName string                 `json:"file_name,omitempty"`
	FileURL  string                 `json:"file_url,omitempty"`
	FileType string                 `json:"file_type,omitempty"`
	Carousel []map[string]interface{} `json:"carousel,omitempty"`
	Buttons  []map[string]interface{} `json:"buttons,omitempty"`
}

// AIAnswer represents the answer data in AI response
type AIAnswer struct {
	AnswerText string                 `json:"answer_text"`
	AnswerData interface{}            `json:"answer_data"`
	AnswerURL  string                 `json:"answer_url"`
	Attachments []AIAttachment        `json:"attachments,omitempty"`
}

// AIData represents the data section in AI response
type AIData struct {
	Answer          AIAnswer `json:"answer"`
	ConfidenceScore float64  `json:"confidence_score"`
}

// AIMetadata represents metadata in AI response
type AIMetadata struct {
	CloseSession bool `json:"close_session,omitempty"`
}

// AIResponse represents the response structure from AI processing
type AIResponse struct {
	Status      string                 `json:"status"`
	Message     string                 `json:"message"`
	Data        AIData                 `json:"data"`
	Metadata    AIMetadata             `json:"metadata,omitempty"`
	Error       string                 `json:"error,omitempty"`
	// Legacy fields for backward compatibility
	MessageID   string                 `json:"message_id,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	Response    string                 `json:"response,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
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
		MessageID:        messageID,
		SessionID:        sessionID,
		CurrentMessage:   message,
		CurrentMessageID: messageID,
		Context:          context,
		Suggestion:       false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	return ai.ProcessAIRequest(ctx, request)
}

// GenerateSuggestions generates suggestions using AI
func (ai *AIService) GenerateSuggestions(ctx context.Context, messageID, sessionID, message string, context map[string]interface{}) (*AIResponse, error) {
	request := AIRequest{
		MessageID:        messageID,
		SessionID:        sessionID,
		CurrentMessage:   message,
		CurrentMessageID: messageID,
		Context:          context,
		Suggestion:       true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
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

// ProcessAIRequestWithHistory processes AI request with chat history
func (ai *AIService) ProcessAIRequestWithHistory(ctx context.Context, messageID, sessionID, message, senderID, senderType string, chatHistory []interface{}, context map[string]interface{}) (*AIResponse, error) {
	request := AIRequest{
		MessageID:        messageID,
		SessionID:        sessionID,
		CurrentMessage:   message,
		CurrentMessageID: messageID,
		ChatHistory:      chatHistory,
		SenderID:         senderID,
		SenderType:       senderType,
		Context:          context,
		Suggestion:       false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	return ai.ProcessAIRequest(ctx, request)
}

// ProcessSlackAIRequest processes AI request for Slack channels
func (ai *AIService) ProcessSlackAIRequest(ctx context.Context, clientID, userID, message, sessionID string, metadata map[string]interface{}) (*AIResponse, error) {
	if ai.slackAIURL == "" || ai.slackAIToken == "" || ai.slackWorkflowID == "" {
		return nil, fmt.Errorf("Slack AI configuration not provided")
	}

	ai.logger.Info("Processing Slack AI request",
		zap.String("client_id", clientID),
		zap.String("user_id", userID),
		zap.String("session_id", sessionID))

	// Prepare Slack AI request payload
	payload := map[string]interface{}{
		"id": ai.slackWorkflowID,
		"input_args": map[string]interface{}{
			"client_id":  clientID,
			"user_id":    userID,
			"human_msg":  message,
			"session_id": sessionID,
			"metadata":   metadata,
		},
	}

	// Marshal request to JSON
	requestBytes, err := json.Marshal(payload)
	if err != nil {
		ai.logger.Error("Failed to marshal Slack AI request", zap.Error(err))
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", ai.slackAIURL, bytes.NewBuffer(requestBytes))
	if err != nil {
		ai.logger.Error("Failed to create Slack AI request", zap.Error(err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", ai.slackAIToken))

	// Send request
	resp, err := ai.httpClient.Do(req)
	if err != nil {
		ai.logger.Error("Failed to send Slack AI request", zap.Error(err))
		return nil, fmt.Errorf("failed to send Slack AI request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		ai.logger.Warn("Slack AI service returned non-success status",
			zap.Int("status_code", resp.StatusCode))
		return nil, fmt.Errorf("Slack AI service returned status %d", resp.StatusCode)
	}

	// Parse response
	var slackResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&slackResponse); err != nil {
		ai.logger.Error("Failed to decode Slack AI response", zap.Error(err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert Slack response to AIResponse format
	return ai.convertSlackResponse(slackResponse)
}

// convertSlackResponse converts Slack AI response to standard AIResponse format
func (ai *AIService) convertSlackResponse(slackResponse map[string]interface{}) (*AIResponse, error) {
	result, ok := slackResponse["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid Slack response format")
	}

	// Extract metadata
	metadata := AIMetadata{}
	if metaDict, exists := result["metadata"].(map[string]interface{}); exists {
		if closeSession, ok := metaDict["close_session"].(bool); ok {
			metadata.CloseSession = closeSession
		}
	}

	// Parse attachments
	var attachments []AIAttachment
	if attachmentList, exists := result["attachments"].([]interface{}); exists {
		for _, att := range attachmentList {
			if attMap, ok := att.(map[string]interface{}); ok {
				attachment := ai.parseAttachment(attMap)
				attachments = append(attachments, attachment)
			}
		}
	}

	// Extract confidence score
	confidenceScore := 0.9 // default
	if score, exists := result["confidence_score"].(float64); exists {
		confidenceScore = score
	}

	// Build response
	response := &AIResponse{
		Status:  slackResponse["status"].(string),
		Message: "",
		Data: AIData{
			Answer: AIAnswer{
				AnswerText:  result["text"].(string),
				AnswerData:  result["data"],
				AnswerURL:   "www.example.com",
				Attachments: attachments,
			},
			ConfidenceScore: confidenceScore,
		},
		Metadata: metadata,
	}

	return response, nil
}

// parseAttachment parses attachment data from AI response
func (ai *AIService) parseAttachment(attachment map[string]interface{}) AIAttachment {
	attachmentType := "file" // default
	if attType, exists := attachment["type"].(string); exists {
		attachmentType = attType
	}

	result := AIAttachment{
		Type: attachmentType,
	}

	if fileName, exists := attachment["file_name"].(string); exists {
		result.FileName = fileName
	}
	if fileURL, exists := attachment["file_url"].(string); exists {
		result.FileURL = fileURL
	}
	if fileType, exists := attachment["file_type"].(string); exists {
		result.FileType = fileType
	}

	// Add carousel data if present
	if attachmentType == "carousel" {
		if carousel, exists := attachment["carousel"].([]interface{}); exists {
			carouselData := make([]map[string]interface{}, len(carousel))
			for i, item := range carousel {
				if itemMap, ok := item.(map[string]interface{}); ok {
					carouselData[i] = itemMap
				}
			}
			result.Carousel = carouselData
		}
	}

	// Add buttons data if present
	if attachmentType == "buttons" {
		if buttons, exists := attachment["buttons"].([]interface{}); exists {
			buttonsData := make([]map[string]interface{}, len(buttons))
			for i, button := range buttons {
				if buttonMap, ok := button.(map[string]interface{}); ok {
					buttonsData[i] = buttonMap
				}
			}
			result.Buttons = buttonsData
		}
	}

	return result
}