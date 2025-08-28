// Package dto defines request/response payloads for client channel endpoints.
package dto

import "github.com/fraiday-org/api-service/internal/models"

// ClientChannelCreateOrUpdateRequest is the payload for creating or updating a client channel.
type ClientChannelCreateOrUpdateRequest struct {
	ChannelType   models.ChannelType         `json:"channel_type" binding:"required"`
	ChannelConfig map[string]interface{}     `json:"channel_config" binding:"required"`
	IsActive      *bool                      `json:"is_active,omitempty"`
}

// ClientChannelResponse is the response payload for a client channel.
type ClientChannelResponse struct {
	ID            string                     `json:"id"`
	ChannelType   models.ChannelType         `json:"channel_type"`
	ChannelConfig map[string]interface{}     `json:"channel_config"`
	IsActive      bool                       `json:"is_active"`
}