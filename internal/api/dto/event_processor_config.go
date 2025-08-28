// Package dto defines request/response payloads for event processor config endpoints.
package dto

import (
	"time"

	"github.com/fraiday-org/api-service/internal/models"
)

// ProcessorConfigCreate represents the payload for creating an event processor config.
type ProcessorConfigCreate struct {
	Name         string                 `json:"name" binding:"required"`
	ClientID     string                 `json:"client_id" binding:"required"`
	ProcessorType models.ProcessorType   `json:"processor_type" binding:"required"`
	Config       map[string]interface{} `json:"config" binding:"required"`
	EventTypes   []models.EventType     `json:"event_types" binding:"required"`
	EntityTypes  []models.EntityType    `json:"entity_types" binding:"required"`
	Description  *string                `json:"description,omitempty"`
	IsActive     *bool                  `json:"is_active,omitempty"`
}

// ProcessorConfigUpdate represents the payload for updating an event processor config.
type ProcessorConfigUpdate struct {
	Name         *string                `json:"name,omitempty"`
	ProcessorType *models.ProcessorType  `json:"processor_type,omitempty"`
	Config       map[string]interface{} `json:"config,omitempty"`
	EventTypes   []models.EventType     `json:"event_types,omitempty"`
	EntityTypes  []models.EntityType    `json:"entity_types,omitempty"`
	Description  *string                `json:"description,omitempty"`
	IsActive     *bool                  `json:"is_active,omitempty"`
}

// ProcessorConfigResponse represents the response payload for an event processor config.
type ProcessorConfigResponse struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	ClientID     string                 `json:"client_id"`
	ProcessorType string                `json:"processor_type"`
	Config       map[string]interface{} `json:"config"`
	EventTypes   []string               `json:"event_types"`
	EntityTypes  []string               `json:"entity_types"`
	Description  *string                `json:"description,omitempty"`
	IsActive     bool                   `json:"is_active"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// ProcessorConfigListResponse represents the response for listing event processor configs.
type ProcessorConfigListResponse struct {
	Configs []ProcessorConfigResponse `json:"configs"`
	Total   int                       `json:"total"`
}