// Package handlers provides Gin HTTP handlers for client endpoints.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/fraiday-org/api-service/internal/api/dto"
	"github.com/fraiday-org/api-service/internal/service"
)

// ClientHandler provides HTTP handlers for client endpoints.
type ClientHandler struct {
	Service *service.ClientService
}

// NewClientHandler creates a new ClientHandler.
func NewClientHandler(svc *service.ClientService) *ClientHandler {
	return &ClientHandler{Service: svc}
}

// CreateClient handles POST /clients
func (h *ClientHandler) CreateClient(c *gin.Context) {
	var req dto.ClientCreateOrUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.Service.CreateClient(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListClients handles GET /clients
func (h *ClientHandler) ListClients(c *gin.Context) {
	resp, err := h.Service.ListClients(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateClient handles PUT /clients/:client_id
func (h *ClientHandler) UpdateClient(c *gin.Context) {
	clientID := c.Param("client_id")
	var req dto.ClientCreateOrUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.Service.UpdateClient(c.Request.Context(), clientID, &req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
