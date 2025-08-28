// Package handlers provides HTTP handlers for event processor config endpoints.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/fraiday-org/api-service/internal/api/dto"
	"github.com/fraiday-org/api-service/internal/service"
)

// EventProcessorConfigHandler handles event processor config related HTTP requests.
type EventProcessorConfigHandler struct {
	processorConfigService *service.EventProcessorConfigService
}

// NewEventProcessorConfigHandler creates a new EventProcessorConfigHandler.
func NewEventProcessorConfigHandler(processorConfigService *service.EventProcessorConfigService) *EventProcessorConfigHandler {
	return &EventProcessorConfigHandler{
		processorConfigService: processorConfigService,
	}
}

// CreateProcessorConfig handles POST /api/v1/clients/{client_id}/processor-configs
func (h *EventProcessorConfigHandler) CreateProcessorConfig(c *gin.Context) {
	clientID := c.Param("client_id")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id is required"})
		return
	}

	var req dto.ProcessorConfigCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ClientID = clientID

	// TODO: Convert DTO to service parameters and call appropriate create method
	// This is a placeholder - needs proper implementation based on processor type
	c.JSON(http.StatusNotImplemented, gin.H{"error": "CreateProcessorConfig not yet implemented"})
}

// GetProcessorConfig handles GET /api/v1/clients/{client_id}/processor-configs/{config_id}
func (h *EventProcessorConfigHandler) GetProcessorConfig(c *gin.Context) {
	clientID := c.Param("client_id")
	configID := c.Param("config_id")

	if clientID == "" || configID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id and config_id are required"})
		return
	}

	config, err := h.processorConfigService.GetConfigByID(c.Request.Context(), configID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// TODO: Convert model to DTO response
	c.JSON(http.StatusOK, config)
}

// UpdateProcessorConfig handles PUT /api/v1/clients/{client_id}/processor-configs/{config_id}
func (h *EventProcessorConfigHandler) UpdateProcessorConfig(c *gin.Context) {
	clientID := c.Param("client_id")
	configID := c.Param("config_id")

	if clientID == "" || configID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id and config_id are required"})
		return
	}

	var req dto.ProcessorConfigUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Convert DTO to update map and call UpdateConfig
	err := h.processorConfigService.UpdateConfig(c.Request.Context(), configID, map[string]interface{}{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Config updated successfully"})
}

// DeleteProcessorConfig handles DELETE /api/v1/clients/{client_id}/processor-configs/{config_id}
func (h *EventProcessorConfigHandler) DeleteProcessorConfig(c *gin.Context) {
	clientID := c.Param("client_id")
	configID := c.Param("config_id")

	if clientID == "" || configID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id and config_id are required"})
		return
	}

	err := h.processorConfigService.DeleteConfig(c.Request.Context(), configID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListProcessorConfigs handles GET /api/v1/clients/{client_id}/processor-configs
func (h *EventProcessorConfigHandler) ListProcessorConfigs(c *gin.Context) {
	clientID := c.Param("client_id")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id is required"})
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset parameter"})
		return
	}

	// TODO: Convert clientID string to ObjectID and call ListConfigs
	configs, err := h.processorConfigService.ListConfigs(c.Request.Context(), nil, nil, nil, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// TODO: Convert models to DTO response
	c.JSON(http.StatusOK, gin.H{"configs": configs, "total": len(configs)})
}