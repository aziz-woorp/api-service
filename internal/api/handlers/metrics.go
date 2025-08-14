package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// MetricsHandler handles Prometheus metrics endpoints
type MetricsHandler struct {
	logger *zap.Logger
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(logger *zap.Logger) *MetricsHandler {
	return &MetricsHandler{
		logger: logger,
	}
}

// GetMetrics returns Prometheus metrics
func (h *MetricsHandler) GetMetrics(c *gin.Context) {
	h.logger.Debug("Metrics endpoint called")
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}