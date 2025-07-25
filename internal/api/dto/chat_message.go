// Package dto defines request/response payloads for chat message endpoints.
package dto

import (
	"github.com/fraiday-org/api-service/internal/models"
)

// ChatMessageCreate represents the payload for creating a chat message.
type ChatMessageCreate struct {
	ExternalID  string                 `json:"external_id,omitempty"`
	Sender      string                 `json:"sender" binding:"required"`
	SenderName  string                 `json:"sender_name,omitempty"`
	SenderType  string                 `json:"sender_type" binding:"required"`
	SessionID   string                 `json:"session_id" binding:"required"`
	Text        string                 `json:"text" binding:"required"`
	Attachments []models.Attachment    `json:"attachments,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Category    string                 `json:"category" binding:"required"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// ChatMessageUpdate represents the payload for updating a chat message.
type ChatMessageUpdate struct {
	Text        *string                `json:"text,omitempty"`
	Sender      *string                `json:"sender,omitempty"`
	SenderName  *string                `json:"sender_name,omitempty"`
	Attachments []models.Attachment    `json:"attachments,omitempty"`
	Category    *string                `json:"category,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// BulkChatMessageCreate represents the payload for bulk-creating chat messages.
type BulkChatMessageCreate struct {
	SessionID   string                 `json:"session_id" binding:"required"`
	Messages    []ChatMessageCreate    `json:"messages" binding:"required"`
}
