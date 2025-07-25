// Package dto defines request/response payloads for client endpoints.
package dto

// ClientCreateOrUpdateRequest is the payload for creating or updating a client.
type ClientCreateOrUpdateRequest struct {
	Name     string  `json:"name" binding:"required"`
	ClientID *string `json:"client_id,omitempty"`
	Email    *string `json:"email,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// ClientResponse is the response payload for a client.
type ClientResponse struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Email    *string `json:"email,omitempty"`
	ClientID string  `json:"client_id"`
	IsActive bool    `json:"is_active"`
}
