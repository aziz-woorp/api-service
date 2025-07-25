// Package handlers provides Gin HTTP handlers for chat session recaps.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/fraiday-org/api-service/internal/service"
)

// ChatSessionRecapHandler provides HTTP handlers for chat session recaps.
type ChatSessionRecapHandler struct {
	Service *service.ChatSessionRecapService
}

// NewChatSessionRecapHandler creates a new ChatSessionRecapHandler.
func NewChatSessionRecapHandler(svc *service.ChatSessionRecapService) *ChatSessionRecapHandler {
	return &ChatSessionRecapHandler{Service: svc}
}

// GenerateRecap handles POST /sessions/:session_id/recap
func (h *ChatSessionRecapHandler) GenerateRecap(c *gin.Context) {
	sessionID := c.Param("session_id")
	var recapData map[string]interface{}
	if err := c.ShouldBindJSON(&recapData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.Service.GenerateRecap(c.Request.Context(), sessionID, recapData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetLatestRecap handles GET /sessions/:session_id/recap
func (h *ChatSessionRecapHandler) GetLatestRecap(c *gin.Context) {
	sessionID := c.Param("session_id")
	resp, err := h.Service.GetLatestRecap(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
