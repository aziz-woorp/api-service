// Package dto defines request/response payloads for semantic layer endpoints.
package dto

import (
	"time"
)

// SemanticLayerCreate represents the payload for creating a semantic layer.
type SemanticLayerCreate struct {
	SemanticServerID string `json:"semantic_server_id" binding:"required"`
	RepositoryID     string `json:"repository_id" binding:"required"`
}

// SemanticLayerResponse represents the response payload for a semantic layer.
type SemanticLayerResponse struct {
	ID                   string                    `json:"id"`
	Client               ClientInlineResponse      `json:"client"`
	ClientRepository     RepositoryInlineResponse  `json:"client_repository"`
	ClientSemanticServer SemanticServerInlineResponse `json:"client_semantic_server"`
	RepositoryFolder     string                    `json:"repository_folder"`
	IsActive             bool                      `json:"is_active"`
	CreatedAt            time.Time                 `json:"created_at"`
	UpdatedAt            time.Time                 `json:"updated_at"`
}

// ClientInlineResponse represents inline client data.
type ClientInlineResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ClientID string `json:"client_id"`
	IsActive bool   `json:"is_active"`
}

// RepositoryInlineResponse represents inline repository data.
type RepositoryInlineResponse struct {
	ID               string                    `json:"id"`
	RepositoryConfig RepositoryConfigResponse  `json:"repository_config"`
	ClientID         *string                   `json:"client_id,omitempty"`
	IsActive         bool                      `json:"is_active"`
	IsDefault        bool                      `json:"is_default"`
	CreatedAt        time.Time                 `json:"created_at"`
	UpdatedAt        time.Time                 `json:"updated_at"`
}

// RepositoryConfigResponse represents repository configuration.
type RepositoryConfigResponse struct {
	RepoURL   string `json:"repo_url"`
	Branch    string `json:"branch"`
	APIKey    string `json:"api_key"`
	BasePath  string `json:"base_path"`
}

// SemanticServerInlineResponse represents inline semantic server data.
type SemanticServerInlineResponse struct {
	ID         string    `json:"id"`
	ServerName string    `json:"server_name"`
	IsActive   bool      `json:"is_active"`
	IsDefault  bool      `json:"is_default"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	ClientID   *string   `json:"client_id,omitempty"`
}

// AddOrRemoveDataStoreRequest represents the payload for adding/removing data stores.
type AddOrRemoveDataStoreRequest struct {
	DataStoreID string `json:"data_store_id" binding:"required"`
}