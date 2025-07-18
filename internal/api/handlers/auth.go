package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/fraiday-org/api-service/internal/api/utils"
	"github.com/fraiday-org/api-service/internal/repository"
)

type AuthHandler struct {
	logger         *zap.Logger
	userRepository *repository.UserRepository
}

func NewAuthHandler(logger *zap.Logger, db *mongo.Database) *AuthHandler {
	return &AuthHandler{
		logger:         logger,
		userRepository: repository.NewUserRepository(db),
	}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid login request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	user, err := h.userRepository.ValidateUser(ctx, req.Username, req.Password)
	if err != nil {
		h.logger.Warn("login failed", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := utils.GenerateToken(user.Username, user.SecretKey)
	if err != nil {
		h.logger.Error("token generation failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token})
}
