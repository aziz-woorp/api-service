package middleware

import (
	"encoding/base64"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func AuthMiddleware(logger *zap.Logger, db *mongo.Database) gin.HandlerFunc {
	adminAPIKey := os.Getenv("ADMIN_API_KEY")
	if adminAPIKey == "" {
		adminAPIKey = "sample-api-key" // fallback
	}

	return func(c *gin.Context) {
		// Allow unauthenticated access to health endpoints
		path := c.Request.URL.Path
		if path == "/health" || path == "/ping" || path == "/readiness" || path == "/healthz" || path == "/metrics" || strings.HasPrefix(path, "/docs") {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}

		// Check for API key (Bearer token)
		if strings.HasPrefix(authHeader, "Bearer ") {
			apiKey := strings.TrimPrefix(authHeader, "Bearer ")
			if apiKey == adminAPIKey {
				c.Set("auth_type", "api_key")
				c.Next()
				return
			}
		}

		// Check for Basic Auth (for AI service communication)
		if strings.HasPrefix(authHeader, "Basic ") {
			basicAuth := strings.TrimPrefix(authHeader, "Basic ")
			decoded, err := base64.StdEncoding.DecodeString(basicAuth)
			if err == nil {
				credentials := string(decoded)
				// Expected format: "username:password" or just validate the token
				if credentials == adminAPIKey || strings.Contains(credentials, adminAPIKey) {
					c.Set("auth_type", "basic")
					c.Next()
					return
				}
			}
		}

		logger.Warn("authentication failed", zap.String("path", path))
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
	}
}
