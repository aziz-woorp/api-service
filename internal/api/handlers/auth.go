package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type AuthHandler struct {
	logger *zap.Logger
}

func NewAuthHandler(logger *zap.Logger, db *mongo.Database) *AuthHandler {
	return &AuthHandler{
		logger: logger,
	}
}

// Health check endpoint - no authentication needed
func (h *AuthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
