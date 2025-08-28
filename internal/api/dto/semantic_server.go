// Package dto defines request/response payloads for semantic server endpoints.
package dto

import (
	"time"
)

// SemanticServerCreate represents the payload for creating a semantic server.
type SemanticServerCreate struct {
	ServerName string `json:"server_name" binding:"required"`
	IsDefault  *bool  `json:"is_default,omitempty"`
}

// SemanticServerUpdate represents the payload for updating a semantic server.
type SemanticServerUpdate struct {
	ServerName *string `json:"server_name,omitempty"`
	IsActive   *bool   `json:"is_active,omitempty"`
	IsDefault  *bool   `json:"is_default,omitempty"`
}

// SemanticServerResponse represents the response payload for a semantic server.
type SemanticServerResponse struct {
	ID         string    `json:"id"`
	ServerName string    `json:"server_name"`
	ClientID   *string   `json:"client_id,omitempty"`
	IsActive   bool      `json:"is_active"`
	IsDefault  bool      `json:"is_default"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// SemanticServerListResponse represents the response for listing semantic servers.
type SemanticServerListResponse struct {
	Servers []SemanticServerResponse `json:"servers"`
	Total   int                      `json:"total"`
}