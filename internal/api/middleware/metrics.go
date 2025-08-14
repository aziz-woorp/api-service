package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/fraiday-org/api-service/internal/service"
)

// MetricsMiddleware creates a middleware for tracking HTTP metrics
func MetricsMiddleware(metricsService *service.MetricsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Track request start
		startTime := metricsService.TrackRequestStart(c.Request.Method, c.FullPath())

		// Process request
		c.Next()

		// Track request end
		metricsService.TrackRequestEnd(startTime, c.Request.Method, c.FullPath(), c.Writer.Status())
	}
}