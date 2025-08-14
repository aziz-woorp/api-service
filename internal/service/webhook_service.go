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

// WebhookService handles webhook notifications
type WebhookService struct {
	logger     *zap.Logger
	httpClient *http.Client
}

// NewWebhookService creates a new webhook service
func NewWebhookService(logger *zap.Logger) *WebhookService {
	return &WebhookService{
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// WebhookPayload represents the payload structure for webhooks
type WebhookPayload struct {
	EventType  string                 `json:"event_type"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	Data       map[string]interface{} `json:"data"`
	Timestamp  time.Time              `json:"timestamp"`
}

// SendWebhook sends a webhook notification to the specified URL
func (ws *WebhookService) SendWebhook(ctx context.Context, webhookURL string, payload WebhookPayload) error {
	ws.logger.Info("Sending webhook notification",
		zap.String("webhook_url", webhookURL),
		zap.String("event_type", payload.EventType),
		zap.String("entity_type", payload.EntityType),
		zap.String("entity_id", payload.EntityID))

	// Set timestamp if not provided
	if payload.Timestamp.IsZero() {
		payload.Timestamp = time.Now().UTC()
	}

	// Marshal payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		ws.logger.Error("Failed to marshal webhook payload", zap.Error(err))
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		ws.logger.Error("Failed to create webhook request", zap.Error(err))
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Fraiday-Webhook/1.0")

	// Send request
	resp, err := ws.httpClient.Do(req)
	if err != nil {
		ws.logger.Error("Failed to send webhook", 
			zap.String("webhook_url", webhookURL),
			zap.Error(err))
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		ws.logger.Warn("Webhook returned non-success status",
			zap.String("webhook_url", webhookURL),
			zap.Int("status_code", resp.StatusCode))
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	ws.logger.Info("Webhook sent successfully",
		zap.String("webhook_url", webhookURL),
		zap.Int("status_code", resp.StatusCode))

	return nil
}

// SendChatMessageWebhook sends a webhook for chat message events
func (ws *WebhookService) SendChatMessageWebhook(ctx context.Context, webhookURL, messageID, sessionID string, messageData map[string]interface{}) error {
	payload := WebhookPayload{
		EventType:  "chat_message_created",
		EntityType: "chat_message",
		EntityID:   messageID,
		Data: map[string]interface{}{
			"message_id": messageID,
			"session_id": sessionID,
			"message":    messageData,
		},
	}

	return ws.SendWebhook(ctx, webhookURL, payload)
}

// SendSuggestionWebhook sends a webhook for suggestion events
func (ws *WebhookService) SendSuggestionWebhook(ctx context.Context, webhookURL, suggestionID, sessionID string, suggestionData map[string]interface{}) error {
	payload := WebhookPayload{
		EventType:  "suggestion_created",
		EntityType: "chat_suggestion",
		EntityID:   suggestionID,
		Data: map[string]interface{}{
			"suggestion_id": suggestionID,
			"session_id":    sessionID,
			"suggestion":    suggestionData,
		},
	}

	return ws.SendWebhook(ctx, webhookURL, payload)
}

// SendEventWebhook sends a webhook for generic events
func (ws *WebhookService) SendEventWebhook(ctx context.Context, webhookURL, eventType, entityType, entityID string, eventData map[string]interface{}) error {
	payload := WebhookPayload{
		EventType:  eventType,
		EntityType: entityType,
		EntityID:   entityID,
		Data:       eventData,
	}

	return ws.SendWebhook(ctx, webhookURL, payload)
}