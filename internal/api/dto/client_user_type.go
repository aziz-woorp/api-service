// Package dto provides data transfer objects for client user types.
package dto

// ClientUserTypeCreateRequest represents the request to create a client user type.
type ClientUserTypeCreateRequest struct {
	TypeID      string                 `json:"type_id" binding:"required"`
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ClientUserTypeUpdateRequest represents the request to update a client user type.
type ClientUserTypeUpdateRequest struct {
	Name        *string                `json:"name,omitempty"`
	Description *string                `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	IsActive    *bool                  `json:"is_active,omitempty"`
}

// ClientUserTypeResponse represents the response for client user type operations.
type ClientUserTypeResponse struct {
	ID          string                 `json:"id"`
	ClientID    string                 `json:"client_id"`
	TypeID      string                 `json:"type_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
	IsActive    bool                   `json:"is_active"`
}