// Package dto defines request/response payloads for chat session thread endpoints.
package dto

import (
	"time"
)

// ThreadResponse is the response for a single thread.
type ThreadResponse struct {
	ThreadID        string    `json:"thread_id"`
	ThreadSessionID string    `json:"thread_session_id"`
	ParentSessionID string    `json:"parent_session_id"`
	ChatSessionID   string    `json:"chat_session_id"`
	Active          bool      `json:"active"`
	LastActivity    time.Time `json:"last_activity"`
}

// ThreadListResponse is the response for a list of threads.
type ThreadListResponse struct {
	Threads []ThreadResponse `json:"threads"`
	Total   int              `json:"total"`
}
