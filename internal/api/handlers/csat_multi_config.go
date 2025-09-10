// Package handlers provides HTTP handlers for multi-CSAT configuration API endpoints.
package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/fraiday-org/api-service/internal/api/dto"
	"github.com/fraiday-org/api-service/internal/models"
	"github.com/fraiday-org/api-service/internal/utils"
)

// === Multi-CSAT Configuration Handlers ===

// ListCSATConfigurations lists all CSAT configurations for a client and channel.
func (h *CSATHandler) ListCSATConfigurations(c *gin.Context) {
	clientID, err := primitive.ObjectIDFromHex(c.Param("client_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client_id"})
		return
	}

	channelID, err := primitive.ObjectIDFromHex(c.Param("channel_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}

	configs, err := h.CSATService.CSATConfigRepo.GetAllByClientAndChannel(c.Request.Context(), clientID, channelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responses []dto.CSATConfigurationResponse
	for _, config := range configs {
		response := dto.CSATConfigurationResponse{
			ID:                config.ID.Hex(),
			ClientID:          config.Client.Hex(),
			ChannelID:         config.ClientChannel.Hex(),
			Type:              config.Type,
			Enabled:           config.Enabled,
			TriggerConditions: config.TriggerConditions,
			CreatedAt:         config.CreatedAt,
			UpdatedAt:         config.UpdatedAt,
		}
		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, gin.H{"configurations": responses})
}

// CreateCSATConfiguration creates a new CSAT configuration with type in request body.
func (h *CSATHandler) CreateCSATConfiguration(c *gin.Context) {
	clientID, err := primitive.ObjectIDFromHex(c.Param("client_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client_id"})
		return
	}

	channelID, err := primitive.ObjectIDFromHex(c.Param("channel_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}

	var req dto.CSATConfigurationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate CSAT type format
	if err := utils.ValidateCSATType(req.Type); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid type format: %v", err)})
		return
	}

	// Check if configuration already exists for this type
	existing, _ := h.CSATService.CSATConfigRepo.GetByClientChannelAndType(c.Request.Context(), clientID, channelID, req.Type)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("CSAT configuration already exists for type '%s'", req.Type)})
		return
	}

	// Create new configuration
	config := &models.CSATConfiguration{
		Client:            clientID,
		ClientChannel:     channelID,
		Type:              req.Type,
		Enabled:           req.Enabled,
		TriggerConditions: req.TriggerConditions,
	}

	if err := h.CSATService.CSATConfigRepo.Create(c.Request.Context(), config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.CSATConfigurationResponse{
		ID:                config.ID.Hex(),
		ClientID:          config.Client.Hex(),
		ChannelID:         config.ClientChannel.Hex(),
		Type:              config.Type,
		Enabled:           config.Enabled,
		TriggerConditions: config.TriggerConditions,
		CreatedAt:         config.CreatedAt,
		UpdatedAt:         config.UpdatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GetCSATConfigurationByType retrieves a specific CSAT configuration by type.
func (h *CSATHandler) GetCSATConfigurationByType(c *gin.Context) {
	clientID, err := primitive.ObjectIDFromHex(c.Param("client_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client_id"})
		return
	}

	channelID, err := primitive.ObjectIDFromHex(c.Param("channel_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}

	csatType := c.Param("type")
	if err := utils.ValidateCSATType(csatType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid type format: %v", err)})
		return
	}

	config, err := h.CSATService.CSATConfigRepo.GetByClientChannelAndType(c.Request.Context(), clientID, channelID, csatType)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "CSAT configuration not found"})
		return
	}

	response := dto.CSATConfigurationResponse{
		ID:                config.ID.Hex(),
		ClientID:          config.Client.Hex(),
		ChannelID:         config.ClientChannel.Hex(),
		Type:              config.Type,
		Enabled:           config.Enabled,
		TriggerConditions: config.TriggerConditions,
		CreatedAt:         config.CreatedAt,
		UpdatedAt:         config.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateCSATConfigurationByType updates a specific CSAT configuration by type.
func (h *CSATHandler) UpdateCSATConfigurationByType(c *gin.Context) {
	clientID, err := primitive.ObjectIDFromHex(c.Param("client_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client_id"})
		return
	}

	channelID, err := primitive.ObjectIDFromHex(c.Param("channel_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}

	csatType := c.Param("type")
	if err := utils.ValidateCSATType(csatType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid type format: %v", err)})
		return
	}

	var req dto.CSATConfigurationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate that the type in the request matches the URL parameter
	if req.Type != csatType {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type in request body must match type in URL"})
		return
	}

	// Get existing configuration
	config, err := h.CSATService.CSATConfigRepo.GetByClientChannelAndType(c.Request.Context(), clientID, channelID, csatType)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "CSAT configuration not found"})
		return
	}

	// Update configuration
	config.Enabled = req.Enabled
	config.TriggerConditions = req.TriggerConditions

	if err := h.CSATService.CSATConfigRepo.Update(c.Request.Context(), config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.CSATConfigurationResponse{
		ID:                config.ID.Hex(),
		ClientID:          config.Client.Hex(),
		ChannelID:         config.ClientChannel.Hex(),
		Type:              config.Type,
		Enabled:           config.Enabled,
		TriggerConditions: config.TriggerConditions,
		CreatedAt:         config.CreatedAt,
		UpdatedAt:         config.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteCSATConfigurationByType deletes a specific CSAT configuration by type.
func (h *CSATHandler) DeleteCSATConfigurationByType(c *gin.Context) {
	clientID, err := primitive.ObjectIDFromHex(c.Param("client_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client_id"})
		return
	}

	channelID, err := primitive.ObjectIDFromHex(c.Param("channel_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}

	csatType := c.Param("type")
	if err := utils.ValidateCSATType(csatType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid type format: %v", err)})
		return
	}

	// Get existing configuration to verify it exists
	config, err := h.CSATService.CSATConfigRepo.GetByClientChannelAndType(c.Request.Context(), clientID, channelID, csatType)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "CSAT configuration not found"})
		return
	}

	// Delete the configuration
	if err := h.CSATService.CSATConfigRepo.Delete(c.Request.Context(), config.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("CSAT configuration for type '%s' deleted successfully", csatType)})
}

// GetCSATQuestionsByType retrieves CSAT questions for a specific configuration type.
func (h *CSATHandler) GetCSATQuestionsByType(c *gin.Context) {
	clientID, err := primitive.ObjectIDFromHex(c.Param("client_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client_id"})
		return
	}

	channelID, err := primitive.ObjectIDFromHex(c.Param("channel_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}

	csatType := c.Param("type")
	if err := utils.ValidateCSATType(csatType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid type format: %v", err)})
		return
	}

	// Get the configuration first
	config, err := h.CSATService.CSATConfigRepo.GetByClientChannelAndType(c.Request.Context(), clientID, channelID, csatType)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "CSAT configuration not found"})
		return
	}

	// Get questions for this configuration
	questions, err := h.CSATService.CSATQuestionRepo.GetByConfigurationID(c.Request.Context(), config.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responses []dto.CSATQuestionResponse
	for _, question := range questions {
		response := dto.CSATQuestionResponse{
			ID:                   question.ID.Hex(),
			CSATConfigurationID:  question.CSATConfigurationID.Hex(),
			QuestionText:         question.QuestionText,
			Options:              question.Options,
			Order:                question.Order,
			Active:               question.Active,
			CreatedAt:            question.CreatedAt,
			UpdatedAt:            question.UpdatedAt,
		}
		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, gin.H{"questions": responses})
}

// UpdateCSATQuestionsByType updates CSAT questions for a specific configuration type.
func (h *CSATHandler) UpdateCSATQuestionsByType(c *gin.Context) {
	clientID, err := primitive.ObjectIDFromHex(c.Param("client_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client_id"})
		return
	}

	channelID, err := primitive.ObjectIDFromHex(c.Param("channel_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}

	csatType := c.Param("type")
	if err := utils.ValidateCSATType(csatType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid type format: %v", err)})
		return
	}

	var req dto.CSATQuestionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the configuration first
	config, err := h.CSATService.CSATConfigRepo.GetByClientChannelAndType(c.Request.Context(), clientID, channelID, csatType)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "CSAT configuration not found"})
		return
	}

	// Convert request to question models
	var questions []models.CSATQuestionTemplate
	for _, questionReq := range req.Questions {
		question := models.CSATQuestionTemplate{
			CSATConfigurationID: config.ID,
			QuestionText:        questionReq.QuestionText,
			Options:             questionReq.Options,
			Order:               questionReq.Order,
			Active:              questionReq.Active,
		}
		questions = append(questions, question)
	}

	// Update questions for this configuration (transactional)
	if err := h.CSATService.CSATQuestionRepo.UpdateQuestionsForConfiguration(c.Request.Context(), config.ID, questions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get updated questions to return
	updatedQuestions, err := h.CSATService.CSATQuestionRepo.GetByConfigurationID(c.Request.Context(), config.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responses []dto.CSATQuestionResponse
	for _, question := range updatedQuestions {
		response := dto.CSATQuestionResponse{
			ID:                   question.ID.Hex(),
			CSATConfigurationID:  question.CSATConfigurationID.Hex(),
			QuestionText:         question.QuestionText,
			Options:              question.Options,
			Order:                question.Order,
			Active:               question.Active,
			CreatedAt:            question.CreatedAt,
			UpdatedAt:            question.UpdatedAt,
		}
		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, gin.H{"questions": responses})
}
