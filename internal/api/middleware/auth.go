package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/fraiday-org/api-service/internal/api/utils"
	"github.com/fraiday-org/api-service/internal/repository"
)

func AuthMiddleware(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	userRepo := repository.NewUserRepository(db)
	return func(c *gin.Context) {
		// Allow unauthenticated access to login, health, ping, docs
		path := c.Request.URL.Path
		if path == "/auth/login" || path == "/health" || path == "/ping" || strings.HasPrefix(path, "/docs") {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Decode token to get username
		username, err := utils.DecodeTokenUsername(token)
		if err != nil {
			logger.Warn("invalid token", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// Lookup user and validate token
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		user, err := userRepo.FindByUsername(ctx, username)
		if err != nil || user == nil || !user.IsActive {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
			return
		}
		// Validate token signature and expiry (1h)
		_, err = utils.ValidateToken(token, user.SecretKey, time.Hour)
		if err != nil {
			logger.Warn("token validation failed", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		// Attach user to context
		c.Set("user", user)
		c.Next()
	}
}
