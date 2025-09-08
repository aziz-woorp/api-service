// Package handlers provides Gin HTTP handlers for chat messages.
package handlers

import (
	"net/http"

	"strconv"

	"github.com/fraiday-org/api-service/internal/api/dto"
	"github.com/fraiday-org/api-service/internal/models"
	"github.com/fraiday-org/api-service/internal/service"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatMessageHandler provides HTTP handlers for chat messages.
type ChatMessageHandler struct {
	Service              *service.ChatMessageService
	SessionService       *service.ChatSessionService
	ClientService        *service.ClientService
	ClientChannelService *service.ClientChannelService
}

// NewChatMessageHandler creates a new ChatMessageHandler.
func NewChatMessageHandler(svc *service.ChatMessageService, sessionSvc *service.ChatSessionService, clientSvc *service.ClientService, clientChannelSvc *service.ClientChannelService) *ChatMessageHandler {
	return &ChatMessageHandler{
		Service:              svc,
		SessionService:       sessionSvc,
		ClientService:        clientSvc,
		ClientChannelService: clientChannelSvc,
	}
}

// CreateMessage handles POST /messages
func (h *ChatMessageHandler) CreateMessage(c *gin.Context) {
	var req dto.ChatMessageCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate sender type
	if err := service.ValidateSenderType(req.SenderType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Step 1: Client validation (matching Python logic)
	client, err := h.ClientService.GetClient(c.Request.Context(), req.ClientID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "client not found"})
		return
	}
	if !client.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client is not active"})
		return
	}

	// Step 2: Client channel resolution (matching Python logic)
	clientChannel, err := h.ClientChannelService.GetChannelByType(c.Request.Context(), req.ClientID, req.ClientChannelType)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "client channel not found"})
		return
	}
	if !clientChannel.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client channel is not active"})
		return
	}

	// Step 3: Get or create session with client/channel association and threading support (matching Python logic)
	session, effectiveSessionID, err := h.SessionService.GetOrCreateSessionBySessionID(c.Request.Context(), req.SessionID, client, clientChannel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get or create session"})
		return
	}

	msg := &models.ChatMessage{
		ExternalID:  req.ExternalID,
		Sender:      req.Sender,
		SenderName:  req.SenderName,
		SenderType:  req.SenderType,
		SessionID:   session.ID, // Use the session's MongoDB _id
		Text:        req.Text,
		Attachments: req.Attachments,
		Data:        req.Data,
		Category:    models.MessageCategory(req.Category),
		Config:      req.Config,
	}

	if err := h.Service.CreateChatMessage(c.Request.Context(), msg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Background workflow triggers (AI chat/suggestion) - AFTER message is saved
	// Use effective session ID (which includes thread info if threading is enabled)
	aiEnabled, aiOk := msg.Config["ai_enabled"].(bool)
	suggestionMode, suggestionOk := msg.Config["suggestion_mode"].(bool)
	if aiOk && aiEnabled && (!suggestionOk || !suggestionMode) {
		// AI chat workflow - message should now have ID assigned by database
		messageID := msg.ID.Hex() // msg.ID is now populated after successful creation
		service.TriggerChatWorkflow(c.Request.Context(), messageID, effectiveSessionID)
	} else if suggestionOk && suggestionMode && (!aiOk || !aiEnabled) {
		// Suggestion workflow - message should now have ID assigned by database
		messageID := msg.ID.Hex() // msg.ID is now populated after successful creation
		service.TriggerSuggestionWorkflow(c.Request.Context(), messageID, effectiveSessionID)
	}

	c.JSON(http.StatusCreated, msg)
}

// ListMessages handles GET /messages
func (h *ChatMessageHandler) ListMessages(c *gin.Context) {
	sessionIDStr := c.Query("session_id")
	userID := c.Query("user_id")
	lastN := int64(0)
	if n := c.Query("last_n"); n != "" {
		// Parse last_n as int64
		if parsed, err := strconv.ParseInt(n, 10, 64); err == nil {
			lastN = parsed
		}
	}

	var sessionID *primitive.ObjectID
	if sessionIDStr != "" {
		sessionID = service.ParseObjectID(sessionIDStr)
	}

	var userIDPtr *string
	if userID != "" {
		userIDPtr = &userID
	}

	messages, err := h.Service.ListMessages(c.Request.Context(), sessionID, userIDPtr, lastN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, messages)
}

// UpdateMessage handles PUT /messages/:id
func (h *ChatMessageHandler) UpdateMessage(c *gin.Context) {
	idStr := c.Param("id")
	id := service.ParseObjectID(idStr)
	if id == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message id"})
		return
	}

	var req dto.ChatMessageUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update := bson.M{}
	if req.Text != nil {
		update["text"] = *req.Text
	}
	if req.Sender != nil {
		update["sender"] = *req.Sender
	}
	if req.SenderName != nil {
		update["sender_name"] = *req.SenderName
	}
	if req.Attachments != nil {
		update["attachments"] = req.Attachments
	}
	if req.Category != nil {
		update["category"] = *req.Category
	}
	if req.Config != nil {
		update["config"] = req.Config
	}
	if req.Data != nil {
		update["data"] = req.Data
	}

	if err := h.Service.UpdateChatMessage(c.Request.Context(), *id, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// BulkCreateMessages handles POST /messages/bulk
func (h *ChatMessageHandler) BulkCreateMessages(c *gin.Context) {
	var req dto.BulkChatMessageCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessionID := service.ParseObjectID(req.SessionID)
	if sessionID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_id"})
		return
	}

	msgs := make([]models.ChatMessage, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = models.ChatMessage{
			ExternalID:  m.ExternalID,
			Sender:      m.Sender,
			SenderName:  m.SenderName,
			SenderType:  m.SenderType,
			SessionID:   *sessionID,
			Text:        m.Text,
			Attachments: m.Attachments,
			Data:        m.Data,
			Category:    models.MessageCategory(m.Category),
			Config:      m.Config,
		}
	}

	if err := h.Service.BulkCreateChatMessages(c.Request.Context(), msgs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Trigger workflow for the latest message - AFTER bulk create
	if len(msgs) > 0 {
		latestIdx := len(msgs) - 1
		latest := msgs[latestIdx]
		aiEnabled, aiOk := latest.Config["ai_enabled"].(bool)
		suggestionMode, suggestionOk := latest.Config["suggestion_mode"].(bool)
		// latest.ID is now populated after successful bulk creation
		messageID := latest.ID.Hex()
		sessionID := latest.SessionID.Hex()
		if aiOk && aiEnabled && (!suggestionOk || !suggestionMode) {
			service.TriggerChatWorkflow(c.Request.Context(), messageID, sessionID)
		} else if suggestionOk && suggestionMode && (!aiOk || !aiEnabled) {
			service.TriggerSuggestionWorkflow(c.Request.Context(), messageID, sessionID)
		}
	}

	c.Status(http.StatusCreated)
}
