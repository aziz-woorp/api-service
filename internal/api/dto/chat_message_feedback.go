// Package dto defines request/response payloads for chat message feedback endpoints.
package dto

import (
	"time"
)

// ChatMessageFeedbackCreate represents the payload for creating feedback.
type ChatMessageFeedbackCreate struct {
	Rating   int                    `json:"rating" binding:"required"`
	Comment  *string                `json:"comment,omitempty"`
	Metadata map[string]interface{} `json:"metadata" binding:"required"`
}

// ChatMessageFeedbackResponse represents the response payload for feedback.
type ChatMessageFeedbackResponse struct {
	ID            string                 `json:"id"`
	ChatMessageID string                 `json:"chat_message_id"`
	Rating        int                    `json:"rating"`
	Comment       *string                `json:"comment,omitempty"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}
