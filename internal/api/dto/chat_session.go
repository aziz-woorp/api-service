// Package dto defines request/response payloads for chat session endpoints.
package dto

import (
	"time"
)

// ChatSessionCreateResponse is the response for creating a session.
type ChatSessionCreateResponse struct {
	SessionID string `json:"session_id"`
}

// ChatSessionResponse is the response for getting a session.
type ChatSessionResponse struct {
	ID        string     `json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Active    bool       `json:"active"`
}

// ChatSessionListItem is an item in the session list.
type ChatSessionListItem struct {
	ID            string     `json:"id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	SessionID     string     `json:"session_id"`
	Active        bool       `json:"active"`
	Client        *string    `json:"client,omitempty"`
	ClientChannel *string    `json:"client_channel,omitempty"`
	Participants  []string   `json:"participants,omitempty"`
	Handover      bool       `json:"handover"`
}

// ChatSessionListResponse is the response for listing sessions.
type ChatSessionListResponse struct {
	Sessions []ChatSessionListItem `json:"sessions"`
	Total    int                   `json:"total"`
}
