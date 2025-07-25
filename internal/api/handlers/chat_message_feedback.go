// Package handlers provides Gin HTTP handlers for chat message feedback.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/fraiday-org/api-service/internal/api/dto"
	"github.com/fraiday-org/api-service/internal/service"
)

// ChatMessageFeedbackHandler provides HTTP handlers for chat message feedback.
type ChatMessageFeedbackHandler struct {
	Service *service.ChatMessageFeedbackService
}

// NewChatMessageFeedbackHandler creates a new ChatMessageFeedbackHandler.
func NewChatMessageFeedbackHandler(svc *service.ChatMessageFeedbackService) *ChatMessageFeedbackHandler {
	return &ChatMessageFeedbackHandler{Service: svc}
}

// CreateFeedback handles POST /messages/:message_id/feedbacks
func (h *ChatMessageFeedbackHandler) CreateFeedback(c *gin.Context) {
	messageID := c.Param("message_id")
	var req dto.ChatMessageFeedbackCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	feedback, err := h.Service.CreateFeedback(c.Request.Context(), messageID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := dto.ChatMessageFeedbackResponse{
		ID:            feedback.ID.Hex(),
		ChatMessageID: feedback.ChatMessageID.Hex(),
		Rating:        feedback.Rating,
		Comment:       feedback.Comment,
		Metadata:      feedback.Metadata,
		CreatedAt:     feedback.CreatedAt,
		UpdatedAt:     feedback.UpdatedAt,
	}
	c.JSON(http.StatusCreated, resp)
}

// ListFeedbacks handles GET /messages/:message_id/feedbacks
func (h *ChatMessageFeedbackHandler) ListFeedbacks(c *gin.Context) {
	messageID := c.Param("message_id")
	feedbacks, err := h.Service.ListFeedbacksByMessageID(c.Request.Context(), messageID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := make([]dto.ChatMessageFeedbackResponse, len(feedbacks))
	for i, fb := range feedbacks {
		resp[i] = dto.ChatMessageFeedbackResponse{
			ID:            fb.ID.Hex(),
			ChatMessageID: fb.ChatMessageID.Hex(),
			Rating:        fb.Rating,
			Comment:       fb.Comment,
			Metadata:      fb.Metadata,
			CreatedAt:     fb.CreatedAt,
			UpdatedAt:     fb.UpdatedAt,
		}
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateFeedback handles PATCH /messages/:message_id/feedbacks/:feedback_id
func (h *ChatMessageFeedbackHandler) UpdateFeedback(c *gin.Context) {
	feedbackID := c.Param("feedback_id")
	var req struct {
		Rating   *int                   `json:"rating,omitempty"`
		Comment  *string                `json:"comment,omitempty"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.Service.UpdateFeedback(c.Request.Context(), feedbackID, req.Rating, req.Comment, req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := dto.ChatMessageFeedbackResponse{
		ID:            updated.ID.Hex(),
		ChatMessageID: updated.ChatMessageID.Hex(),
		Rating:        updated.Rating,
		Comment:       updated.Comment,
		Metadata:      updated.Metadata,
		CreatedAt:     updated.CreatedAt,
		UpdatedAt:     updated.UpdatedAt,
	}
	c.JSON(http.StatusOK, resp)
}
