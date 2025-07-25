// Package handlers provides Gin HTTP handlers for chat session threads.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/fraiday-org/api-service/internal/service"
)

// ChatSessionThreadHandler provides HTTP handlers for chat session threads.
type ChatSessionThreadHandler struct {
	Service *service.ChatSessionThreadService
}

// NewChatSessionThreadHandler creates a new ChatSessionThreadHandler.
func NewChatSessionThreadHandler(svc *service.ChatSessionThreadService) *ChatSessionThreadHandler {
	return &ChatSessionThreadHandler{Service: svc}
}

// CreateThread handles POST /sessions/:session_id/threads
func (h *ChatSessionThreadHandler) CreateThread(c *gin.Context) {
	sessionID := c.Param("session_id")
	resp, err := h.Service.CreateThread(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListThreads handles GET /sessions/:session_id/threads
func (h *ChatSessionThreadHandler) ListThreads(c *gin.Context) {
	sessionID := c.Param("session_id")
	includeInactive := true
	if v := c.Query("include_inactive"); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			includeInactive = b
		}
	}
	resp, err := h.Service.ListThreads(c.Request.Context(), sessionID, includeInactive)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetActiveThread handles GET /sessions/:session_id/active_thread
func (h *ChatSessionThreadHandler) GetActiveThread(c *gin.Context) {
	sessionID := c.Param("session_id")
	inactivityMinutes := 0
	if v := c.Query("inactivity_minutes"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			inactivityMinutes = n
		}
	}
	resp, err := h.Service.GetActiveThread(c.Request.Context(), sessionID, inactivityMinutes)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CloseThread handles POST /sessions/:session_id/close_thread
func (h *ChatSessionThreadHandler) CloseThread(c *gin.Context) {
	sessionID := c.Param("session_id")
	var threadID *string
	if v := c.Query("thread_id"); v != "" {
		threadID = &v
	}
	ok, err := h.Service.CloseThread(c.Request.Context(), sessionID, threadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ok {
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Thread closed successfully"})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "info", "message": "No active thread found to close"})
	}
}
