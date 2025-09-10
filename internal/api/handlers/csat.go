// Package handlers provides HTTP handlers for CSAT API endpoints.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/fraiday-org/api-service/internal/api/dto"
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

	// Trigger CSAT survey using external session_id and type
	session, err := h.CSATService.TriggerCSATSurveyBySessionID(c.Request.Context(), req.SessionID, req.Type)
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

// Legacy GetCSATConfiguration - REMOVED for multi-CSAT configuration support
// Use ListCSATConfigurations or GetCSATConfigurationByType instead

// Legacy UpdateCSATConfiguration - REMOVED for multi-CSAT configuration support
// Use CreateCSATConfiguration, UpdateCSATConfigurationByType instead

// Legacy UpdateCSATQuestions - REMOVED for multi-CSAT configuration support
// Use UpdateCSATQuestionsByType instead

// Legacy GetCSATQuestions - REMOVED for multi-CSAT configuration support
// Use GetCSATQuestionsByType instead

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

	// Convert ObjectIDs to strings for QuestionsSent
	var questionsSentStrings []string
	for _, questionID := range session.QuestionsSent {
		questionsSentStrings = append(questionsSentStrings, questionID.Hex())
	}

	response := dto.CSATSessionResponse{
		ID:                   session.ID.Hex(),
		ChatSessionID:        session.ChatSessionID,
		CSATConfigurationID:  session.CSATConfigurationID.Hex(),
		ClientID:             session.Client.Hex(),
		ChannelID:            session.ClientChannel.Hex(),
		ThreadSessionID:      session.ThreadSessionID,
		ThreadContext:        session.ThreadContext,
		Status:               session.Status,
		TriggeredAt:          session.TriggeredAt,
		CompletedAt:          session.CompletedAt,
		CurrentQuestionIndex: session.CurrentQuestionIndex,
		QuestionsSent:        questionsSentStrings,
		CreatedAt:            session.CreatedAt,
		UpdatedAt:            session.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}
