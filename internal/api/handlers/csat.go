// Package handlers provides HTTP handlers for CSAT API endpoints.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/fraiday-org/api-service/internal/api/dto"
	"github.com/fraiday-org/api-service/internal/models"
	"github.com/fraiday-org/api-service/internal/service"
)

// CSATHandler handles CSAT-related HTTP requests.
type CSATHandler struct {
	CSATService *service.CSATService
}

// NewCSATHandler creates a new CSATHandler.
func NewCSATHandler(csatService *service.CSATService) *CSATHandler {
	return &CSATHandler{
		CSATService: csatService,
	}
}

// TriggerCSAT triggers a CSAT survey for a chat session.
func (h *CSATHandler) TriggerCSAT(c *gin.Context) {
	var req dto.CSATTriggerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Trigger CSAT survey using external session_id
	session, err := h.CSATService.TriggerCSATSurveyBySessionID(c.Request.Context(), req.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.CSATTriggerResponse{
		CSATSessionID: session.ID.Hex(),
		Status:        session.Status,
		TriggeredAt:   session.TriggeredAt,
		Message:       "CSAT survey triggered successfully",
	}

	c.JSON(http.StatusOK, response)
}

// RespondToCSAT handles a user response to a CSAT question.
func (h *CSATHandler) RespondToCSAT(c *gin.Context) {
	var req dto.CSATResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Process response using external session_id
	responseID, err := h.CSATService.ProcessResponseBySessionID(c.Request.Context(), req.SessionID, req.CSATQuestionID, req.ResponseValue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.CSATResponseResponse{
		ResponseID: responseID,
		Status:     "success",
		Message:    "Response recorded successfully",
	}

	c.JSON(http.StatusOK, response)
}

// GetCSATConfiguration retrieves CSAT configuration for a client and channel.
func (h *CSATHandler) GetCSATConfiguration(c *gin.Context) {
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

	config, err := h.CSATService.CSATConfigRepo.GetByClientAndChannel(c.Request.Context(), clientID, channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "CSAT configuration not found"})
		return
	}

	response := dto.CSATConfigurationResponse{
		ID:                config.ID.Hex(),
		ClientID:          config.Client.Hex(),
		ChannelID:         config.ClientChannel.Hex(),
		Enabled:           config.Enabled,
		TriggerConditions: config.TriggerConditions,
		CreatedAt:         config.CreatedAt,
		UpdatedAt:         config.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateCSATConfiguration updates CSAT configuration for a client and channel.
func (h *CSATHandler) UpdateCSATConfiguration(c *gin.Context) {
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

	// Try to get existing configuration
	config, err := h.CSATService.CSATConfigRepo.GetByClientAndChannel(c.Request.Context(), clientID, channelID)
	if err != nil {
		// Create new configuration if it doesn't exist
		config = &models.CSATConfiguration{
			Client:            clientID,
			ClientChannel:     channelID,
			Enabled:           req.Enabled,
			TriggerConditions: req.TriggerConditions,
		}
		if err := h.CSATService.CSATConfigRepo.Create(c.Request.Context(), config); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		// Update existing configuration
		config.Enabled = req.Enabled
		config.TriggerConditions = req.TriggerConditions
		if err := h.CSATService.CSATConfigRepo.Update(c.Request.Context(), config); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	response := dto.CSATConfigurationResponse{
		ID:                config.ID.Hex(),
		ClientID:          config.Client.Hex(),
		ChannelID:         config.ClientChannel.Hex(),
		Enabled:           config.Enabled,
		TriggerConditions: config.TriggerConditions,
		CreatedAt:         config.CreatedAt,
		UpdatedAt:         config.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateCSATQuestions updates CSAT questions for a client and channel.
func (h *CSATHandler) UpdateCSATQuestions(c *gin.Context) {
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

	var req dto.CSATQuestionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var responses []dto.CSATQuestionResponse

	// Create or update each question
	for _, questionReq := range req.Questions {
		question := &models.CSATQuestionTemplate{
			Client:        clientID,
			ClientChannel: channelID,
			QuestionText:  questionReq.QuestionText,
			Options:       questionReq.Options,
			Order:         questionReq.Order,
			Active:        questionReq.Active,
		}

		if err := h.CSATService.CSATQuestionRepo.Create(c.Request.Context(), question); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		response := dto.CSATQuestionResponse{
			ID:           question.ID.Hex(),
			ClientID:     question.Client.Hex(),
			ChannelID:    question.ClientChannel.Hex(),
			QuestionText: question.QuestionText,
			Options:      question.Options,
			Order:        question.Order,
			Active:       question.Active,
			CreatedAt:    question.CreatedAt,
			UpdatedAt:    question.UpdatedAt,
		}

		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, gin.H{"questions": responses})
}

// GetCSATQuestions retrieves CSAT questions for a client and channel.
func (h *CSATHandler) GetCSATQuestions(c *gin.Context) {
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

	questions, err := h.CSATService.CSATQuestionRepo.GetByClientAndChannel(c.Request.Context(), clientID, channelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responses []dto.CSATQuestionResponse
	for _, question := range questions {
		response := dto.CSATQuestionResponse{
			ID:           question.ID.Hex(),
			ClientID:     question.Client.Hex(),
			ChannelID:    question.ClientChannel.Hex(),
			QuestionText: question.QuestionText,
			Options:      question.Options,
			Order:        question.Order,
			Active:       question.Active,
			CreatedAt:    question.CreatedAt,
			UpdatedAt:    question.UpdatedAt,
		}
		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, gin.H{"questions": responses})
}

// GetCSATSession retrieves a CSAT session by ID.
func (h *CSATHandler) GetCSATSession(c *gin.Context) {
	sessionID, err := primitive.ObjectIDFromHex(c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_id"})
		return
	}

	session, err := h.CSATService.CSATSessionRepo.GetByID(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "CSAT session not found"})
		return
	}

	response := dto.CSATSessionResponse{
		ID:                   session.ID.Hex(),
		ChatSessionID:        session.ChatSessionID,
		ClientID:             session.Client.Hex(),
		ChannelID:            session.ClientChannel.Hex(),
		ThreadSessionID:      session.ThreadSessionID,
		ThreadContext:        session.ThreadContext,
		Status:               session.Status,
		TriggeredAt:          session.TriggeredAt,
		CompletedAt:          session.CompletedAt,
		CurrentQuestionIndex: session.CurrentQuestionIndex,
		QuestionsSent:        session.QuestionsSent,
		CreatedAt:            session.CreatedAt,
		UpdatedAt:            session.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}
