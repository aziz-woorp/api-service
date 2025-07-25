// Package handlers provides Gin HTTP handlers for chat sessions.
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/fraiday-org/api-service/internal/api/dto"
	"github.com/fraiday-org/api-service/internal/service"
)

// ChatSessionHandler provides HTTP handlers for chat sessions.
type ChatSessionHandler struct {
	Service *service.ChatSessionService
}

// NewChatSessionHandler creates a new ChatSessionHandler.
func NewChatSessionHandler(svc *service.ChatSessionService) *ChatSessionHandler {
	return &ChatSessionHandler{Service: svc}
}

// CreateSession handles POST /sessions
func (h *ChatSessionHandler) CreateSession(c *gin.Context) {
	sessionID, err := h.Service.CreateSession(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ChatSessionCreateResponse{SessionID: sessionID})
}

// GetSession handles GET /sessions/:session_id
func (h *ChatSessionHandler) GetSession(c *gin.Context) {
	id := c.Param("session_id")
	resp, err := h.Service.GetSession(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListSessions handles GET /sessions
func (h *ChatSessionHandler) ListSessions(c *gin.Context) {
	var (
		clientID      *string
		clientChannel *string
		userID        *string
		sessionID     *string
		active        *bool
		startDate     *time.Time
		endDate       *time.Time
	)
	if v := c.Query("client_id"); v != "" {
		clientID = &v
	}
	if v := c.Query("client_channel"); v != "" {
		clientChannel = &v
	}
	if v := c.Query("user_id"); v != "" {
		userID = &v
	}
	if v := c.Query("session_id"); v != "" {
		sessionID = &v
	}
	if v := c.Query("active"); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			active = &b
		}
	}
	if v := c.Query("start_date"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			startDate = &t
		}
	}
	if v := c.Query("end_date"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			endDate = &t
		}
	}
	skip := int64(0)
	limit := int64(10)
	if v := c.Query("skip"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			skip = n
		}
	}
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			limit = n
		}
	}
	params := service.ListSessionsParams{
		ClientID:      clientID,
		ClientChannel: clientChannel,
		UserID:        userID,
		SessionID:     sessionID,
		Active:        active,
		StartDate:     startDate,
		EndDate:       endDate,
		Skip:          skip,
		Limit:         limit,
	}
	resp, err := h.Service.ListSessions(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
