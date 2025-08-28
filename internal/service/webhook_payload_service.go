// Package service provides webhook payload creation strategies.
package service

import (
	"context"
	"fmt"

	"github.com/fraiday-org/api-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WebhookPayloadStrategy defines the interface for creating webhook payloads.
type WebhookPayloadStrategy interface {
	CreatePayload(ctx context.Context, entityID string) (map[string]interface{}, error)
	HandleResponse(ctx context.Context, entityID string, responseData map[string]interface{}) error
	GetEntityType() models.EntityType
}

// MessagePayloadStrategy handles webhook payloads for chat messages.
type MessagePayloadStrategy struct {
	MessageService *ChatMessageService
	SessionService *ChatSessionService
}

// NewMessagePayloadStrategy creates a new MessagePayloadStrategy.
func NewMessagePayloadStrategy(
	messageService *ChatMessageService,
	sessionService *ChatSessionService,
) *MessagePayloadStrategy {
	return &MessagePayloadStrategy{
		MessageService: messageService,
		SessionService: sessionService,
	}
}

// CreatePayload creates a webhook payload for a chat message.
func (s *MessagePayloadStrategy) CreatePayload(ctx context.Context, entityID string) (map[string]interface{}, error) {
	// Parse entity ID to ObjectID
	objID, err := primitive.ObjectIDFromHex(entityID)
	if err != nil {
		return nil, fmt.Errorf("invalid entity ID: %w", err)
	}

	// Get the message
	message, err := s.MessageService.GetChatMessageByID(ctx, objID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	// Get the session
	session, err := s.SessionService.GetSession(ctx, message.SessionID.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Create the payload
	payload := map[string]interface{}{
		"id":         message.ID.Hex(),
		"created_at": message.CreatedAt,
		"updated_at": message.UpdatedAt,
		"session_id": session.ID,
		"text":       message.Text,
		"sender":     message.Sender,
		"sender_type": message.SenderType,
		"data":       message.Data,
	}

	// Add attachments if present
	if len(message.Attachments) > 0 {
		attachments := make([]map[string]interface{}, len(message.Attachments))
		for i, attachment := range message.Attachments {
			attachments[i] = map[string]interface{}{
				"type":      attachment.Type,
				"file_url":  attachment.FileURL,
				"file_name": attachment.FileName,
				"file_size": attachment.FileSize,
				"file_type": attachment.FileType,
				"carousel":  attachment.Carousel,
			}
		}
		payload["attachments"] = attachments
	}

	// Add confidence score if present
	if message.Confidence > 0 {
		payload["confidence_score"] = message.Confidence
	}

	return payload, nil
}

// HandleResponse handles the webhook response for a chat message.
func (s *MessagePayloadStrategy) HandleResponse(ctx context.Context, entityID string, responseData map[string]interface{}) error {
	// Check if response contains an external ID
	if externalID, ok := responseData["id"].(string); ok && externalID != "" {
		// Parse entity ID to ObjectID
		objID, err := primitive.ObjectIDFromHex(entityID)
		if err != nil {
			return fmt.Errorf("invalid entity ID: %w", err)
		}

		// Update the message with external ID in data field
		update := bson.M{
			"$set": bson.M{
				"data.external_id": externalID,
			},
		}

		// Update the message
		if err := s.MessageService.UpdateChatMessage(ctx, objID, update); err != nil {
			return fmt.Errorf("failed to update message with external ID: %w", err)
		}
	}

	return nil
}

// GetEntityType returns the entity type for messages.
func (s *MessagePayloadStrategy) GetEntityType() models.EntityType {
	return models.EntityTypeChatMessage
}

// SuggestionPayloadStrategy handles webhook payloads for chat suggestions.
type SuggestionPayloadStrategy struct {
	SuggestionService *ChatMessageSuggestionService
	MessageService    *ChatMessageService
	SessionService    *ChatSessionService
}

// NewSuggestionPayloadStrategy creates a new SuggestionPayloadStrategy.
func NewSuggestionPayloadStrategy(
	suggestionService *ChatMessageSuggestionService,
	messageService *ChatMessageService,
	sessionService *ChatSessionService,
) *SuggestionPayloadStrategy {
	return &SuggestionPayloadStrategy{
		SuggestionService: suggestionService,
		MessageService:    messageService,
		SessionService:     sessionService,
	}
}

// CreatePayload creates a webhook payload for a chat suggestion.
func (s *SuggestionPayloadStrategy) CreatePayload(ctx context.Context, entityID string) (map[string]interface{}, error) {
	// Get the suggestion
	suggestion, err := s.SuggestionService.GetSuggestion(ctx, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get suggestion: %w", err)
	}
	if suggestion == nil {
		return nil, fmt.Errorf("suggestion not found")
	}

	// Note: ChatMessageSuggestion doesn't have ChatMessageID field in the current model
	// We'll use the session to get recent messages or skip this for now
	// For now, we'll create a minimal message payload
	messagePayload := map[string]interface{}{
		"session_id": suggestion.ChatSessionID.Hex(),
		"text":       "Related to suggestion", // Placeholder
	}

	// Get the session
	session, err := s.SessionService.GetSession(ctx, suggestion.ChatSessionID.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Create the suggestion payload
	payload := map[string]interface{}{
		"id":           suggestion.ID.Hex(),
		"created_at":   suggestion.CreatedAt,
		"updated_at":   suggestion.UpdatedAt,
		"chat_message": messagePayload,
		"session_id":      session.ID,
		"suggestion_text": suggestion.SuggestionText,
		"suggestion_type": suggestion.SuggestionType,
		"confidence_score": suggestion.ConfidenceScore,
		"metadata":        suggestion.Metadata,
		"is_used":         suggestion.IsUsed,
	}

	// ChatMessageSuggestion doesn't have attachments in the current model

	return payload, nil
}

// HandleResponse handles the webhook response for a chat suggestion.
func (s *SuggestionPayloadStrategy) HandleResponse(ctx context.Context, entityID string, responseData map[string]interface{}) error {
	// Check if response contains an external ID
	if externalID, ok := responseData["id"].(string); ok && externalID != "" {
		// Update the suggestion with external ID using MarkSuggestionAsUsed
		// Note: The current service doesn't have a generic update method,
		// so we'll use the available method or implement a custom update
		if err := s.SuggestionService.MarkSuggestionAsUsed(ctx, entityID); err != nil {
			return fmt.Errorf("failed to mark suggestion as used: %w", err)
		}
		// TODO: Add a proper update method to ChatMessageSuggestionService
		// to handle external_id updates
	}

	return nil
}

// GetEntityType returns the entity type for suggestions.
func (s *SuggestionPayloadStrategy) GetEntityType() models.EntityType {
	return models.EntityTypeChatSuggestion
}

// WebhookPayloadService manages webhook payload creation.
type WebhookPayloadService struct {
	strategies map[models.EntityType]WebhookPayloadStrategy
}

// NewWebhookPayloadService creates a new WebhookPayloadService.
func NewWebhookPayloadService(
	messageStrategy *MessagePayloadStrategy,
	suggestionStrategy *SuggestionPayloadStrategy,
) *WebhookPayloadService {
	strategies := make(map[models.EntityType]WebhookPayloadStrategy)
	strategies[models.EntityTypeChatMessage] = messageStrategy
	strategies[models.EntityTypeChatSuggestion] = suggestionStrategy

	return &WebhookPayloadService{
		strategies: strategies,
	}
}

// CreatePayload creates a webhook payload for the given entity.
func (s *WebhookPayloadService) CreatePayload(ctx context.Context, entityType models.EntityType, entityID string) (map[string]interface{}, error) {
	strategy, exists := s.strategies[entityType]
	if !exists {
		return nil, fmt.Errorf("no strategy found for entity type: %s", entityType)
	}

	return strategy.CreatePayload(ctx, entityID)
}

// HandleResponse handles a webhook response for the given entity.
func (s *WebhookPayloadService) HandleResponse(ctx context.Context, entityType models.EntityType, entityID string, responseData map[string]interface{}) error {
	strategy, exists := s.strategies[entityType]
	if !exists {
		return fmt.Errorf("no strategy found for entity type: %s", entityType)
	}

	return strategy.HandleResponse(ctx, entityID, responseData)
}