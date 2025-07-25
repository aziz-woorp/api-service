// Package dto defines request/response payloads for chat session recap endpoints.
package dto

import (
	"time"
)

// ChatSessionRecapResponse is the response for a chat session recap.
type ChatSessionRecapResponse struct {
	ID        string                 `json:"id"`
	SessionID string                 `json:"session_id"`
	RecapData map[string]interface{} `json:"recap_data"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}
