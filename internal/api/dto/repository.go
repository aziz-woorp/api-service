// Package dto defines request/response payloads for repository endpoints.
package dto

import (
	"time"
)

// RepositoryCreate represents the payload for creating a repository.
type RepositoryCreate struct {
	RepositoryConfig RepositoryConfig `json:"repository_config" binding:"required"`
	IsDefault        *bool            `json:"is_default,omitempty"`
}

// RepositoryConfig represents repository configuration.
type RepositoryConfig struct {
	RepoURL  string `json:"repo_url" binding:"required"`
	Branch   string `json:"branch" binding:"required"`
	APIKey   string `json:"api_key" binding:"required"`
	BasePath string `json:"base_path" binding:"required"`
}

// RepositoryUpdate represents the payload for updating a repository.
type RepositoryUpdate struct {
	RepositoryConfig *RepositoryConfig `json:"repository_config,omitempty"`
	IsActive         *bool             `json:"is_active,omitempty"`
	IsDefault        *bool             `json:"is_default,omitempty"`
}

// RepositoryResponse represents the response payload for a repository.
type RepositoryResponse struct {
	ID               string           `json:"id"`
	RepositoryConfig RepositoryConfig `json:"repository_config"`
	ClientID         *string          `json:"client_id,omitempty"`
	IsActive         bool             `json:"is_active"`
	IsDefault        bool             `json:"is_default"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

// RepositoryListResponse represents the response for listing repositories.
type RepositoryListResponse struct {
	Repositories []RepositoryResponse `json:"repositories"`
	Total        int                  `json:"total"`
}