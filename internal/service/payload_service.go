package service

import (
	"context"
	"fmt"
	"time"

	"github.com/fraiday-org/api-service/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PayloadService handles creation of structured event payloads
type PayloadService struct {
	ChatMessageService *ChatMessageService
	ChatSessionService *ChatSessionService
	ThreadManagerService *ThreadManagerService
}

// NewPayloadService creates a new PayloadService instance
func NewPayloadService(chatMessageService *ChatMessageService, chatSessionService *ChatSessionService, threadManagerService *ThreadManagerService) *PayloadService {
	return &PayloadService{
		ChatMessageService: chatMessageService,
		ChatSessionService: chatSessionService,
		ThreadManagerService: threadManagerService,
	}
}

// ChatMessagePayload represents the structured payload for chat message events
type ChatMessagePayload struct {
	ID           string                 `json:"id"`
	ExternalID   *string                `json:"external_id,omitempty"`
	Sender       string                 `json:"sender"`
	SenderName   *string                `json:"sender_name,omitempty"`
	SenderType   string                 `json:"sender_type"`
	SessionID    string                 `json:"session_id"`
	Text         string                 `json:"text"`
	Attachments  []models.Attachment    `json:"attachments,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
	Category     string                 `json:"category"`
	Config       map[string]interface{} `json:"config,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Confidence   *float64               `json:"confidence,omitempty"`
}

// ChatSuggestionPayload represents the structured payload for chat suggestion events
type ChatSuggestionPayload struct {
	ID            string                 `json:"id"`
	ChatSessionID string                 `json:"chat_session_id"`
	ChatMessageID string                 `json:"chat_message_id"`
	Text          string                 `json:"text"`
	Data          map[string]interface{} `json:"data,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// CreateChatMessagePayload creates a structured payload for a chat message
func (ps *PayloadService) CreateChatMessagePayload(ctx context.Context, messageID string) (map[string]interface{}, error) {
	if ps.ChatMessageService == nil {
		return nil, fmt.Errorf("ChatMessageService is nil in PayloadService")
	}
	
	// Parse message ID
	objID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return nil, fmt.Errorf("invalid message ID: %w", err)
	}

	// Get message from database
	message, err := ps.ChatMessageService.GetChatMessage(ctx, objID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	// Get session information
	session, err := ps.ChatSessionService.GetSessionByID(ctx, message.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Handle optional string fields
	var externalID *string
	if message.ExternalID != "" {
		externalID = &message.ExternalID
	}

	var senderName *string
	if message.SenderName != "" {
		senderName = &message.SenderName
	}

	var confidence *float64
	if message.Confidence != 0 {
		confidence = &message.Confidence
	}

	// Create structured payload
	payload := ChatMessagePayload{
		ID:          message.ID.Hex(),
		ExternalID:  externalID,
		Sender:      message.Sender,
		SenderName:  senderName,
		SenderType:  message.SenderType,
		SessionID:   ps.normalizeSessionID(session.SessionID),
		Text:        message.Text,
		Attachments: message.Attachments,
		Data:        message.Data,
		Category:    string(message.Category),
		Config:      message.Config,
		CreatedAt:   message.CreatedAt,
		UpdatedAt:   message.UpdatedAt,
		Confidence:  confidence,
	}

	// Convert to map for consistency with existing code
	result := map[string]interface{}{
		"id":           payload.ID,
		"sender":       payload.Sender,
		"sender_type":  payload.SenderType,
		"session_id":   payload.SessionID,
		"text":         payload.Text,
		"category":     payload.Category,
		"created_at":   payload.CreatedAt.Format(time.RFC3339),
		"updated_at":   payload.UpdatedAt.Format(time.RFC3339),
	}

	// Add optional fields if they exist
	if payload.ExternalID != nil {
		result["external_id"] = *payload.ExternalID
	}
	if payload.SenderName != nil {
		result["sender_name"] = *payload.SenderName
	}
	if len(payload.Attachments) > 0 {
		result["attachments"] = payload.Attachments
	}
	if payload.Data != nil {
		result["data"] = payload.Data
	}
	if payload.Config != nil {
		result["config"] = payload.Config
	}
	if payload.Confidence != nil {
		result["confidence"] = *payload.Confidence
	}

	return result, nil
}

// CreateChatSuggestionPayload creates a structured payload for a chat suggestion
func (ps *PayloadService) CreateChatSuggestionPayload(ctx context.Context, suggestionID string) (map[string]interface{}, error) {
	// For now, return a basic structure since we don't have a full suggestion model yet
	// This should be expanded when the suggestion model is implemented
	return map[string]interface{}{
		"id":              suggestionID,
		"created_at":      time.Now().Format(time.RFC3339),
		"suggestion_type": "ai_generated",
	}, nil
}

// normalizeSessionID normalizes session ID for threading support
// Matches Python PayloadService.normalize_session_id() implementation
func (ps *PayloadService) normalizeSessionID(sessionID string) string {
	if sessionID == "" {
		return sessionID
	}

	// Use ThreadManagerService to parse session ID and get base session ID
	if ps.ThreadManagerService != nil {
		// Parse session ID to check if it has thread component
		baseSessionID, threadID := ps.ThreadManagerService.ParseSessionID(sessionID)
		
		// If there's a thread component, we need to verify threading is enabled
		if threadID != "" {
			// Check if threading is enabled for this session
			// Note: In a production environment, you might want to cache this lookup
			if threadingEnabled, err := ps.ThreadManagerService.IsThreadingEnabledForSession(context.Background(), baseSessionID); err == nil && threadingEnabled {
				// Threading is enabled and session has thread component, return normalized ID
				return baseSessionID
			}
		}
		
		// If no thread component or threading not enabled, return base session ID anyway
		// This ensures consistency - we always return the base session ID for events
		return baseSessionID
	}

	// Fallback: if no ThreadManagerService available, return original session ID
	return sessionID
}

// PrepareEventData prepares event data with normalized session IDs
// Recursively normalizes session_id and chat_session_id fields in nested objects
func (ps *PayloadService) PrepareEventData(data map[string]interface{}) map[string]interface{} {
	result := ps.normalizeEventDataRecursive(data)
	if normalized, ok := result.(map[string]interface{}); ok {
		return normalized
	}
	// Fallback to original data if type assertion fails
	return data
}

// normalizeEventDataRecursive recursively normalizes session IDs in event data
func (ps *PayloadService) normalizeEventDataRecursive(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, val := range v {
			if k == "session_id" || k == "chat_session_id" {
				if sessionID, ok := val.(string); ok {
					result[k] = ps.normalizeSessionID(sessionID)
				} else {
					result[k] = val
				}
			} else {
				// Recursively process nested objects
				result[k] = ps.normalizeEventDataRecursive(val)
			}
		}
		return result
	case []interface{}:
		// Handle arrays by processing each element
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = ps.normalizeEventDataRecursive(item)
		}
		return result
	default:
		// Return primitive values as-is
		return data
	}
}
