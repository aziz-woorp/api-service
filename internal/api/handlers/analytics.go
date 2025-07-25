// Package handlers provides Gin HTTP handlers for analytics endpoints.
package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/fraiday-org/api-service/internal/service"
)

// AnalyticsHandler provides HTTP handlers for analytics endpoints.
type AnalyticsHandler struct {
	Service *service.AnalyticsService
}

// NewAnalyticsHandler creates a new AnalyticsHandler.
func NewAnalyticsHandler(svc *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{Service: svc}
}

// GetDashboardMetrics handles GET /analytics/dashboard
func (h *AnalyticsHandler) GetDashboardMetrics(c *gin.Context) {
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid start_time"})
		return
	}
	endTime := time.Now().UTC()
	if endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = t
		}
	}
	resp := h.Service.GetDashboardMetrics(startTime, endTime)
	c.JSON(http.StatusOK, resp)
}

// GetBotEngagementMetrics handles GET /analytics/bot-engagement
func (h *AnalyticsHandler) GetBotEngagementMetrics(c *gin.Context) {
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid start_time"})
		return
	}
	endTime := time.Now().UTC()
	if endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = t
		}
	}
	resp := h.Service.GetBotEngagementMetrics(startTime, endTime)
	c.JSON(http.StatusOK, resp)
}

// GetContainmentRateMetrics handles GET /analytics/containment-rate
func (h *AnalyticsHandler) GetContainmentRateMetrics(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	aggregation := c.DefaultQuery("aggregation", "auto")
	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid start_date"})
		return
	}
	endDate := time.Now().UTC()
	if endDateStr != "" {
		if t, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = t
		}
	}
	resp := h.Service.GetContainmentRateMetrics(startDate, endDate, aggregation)
	c.JSON(http.StatusOK, resp)
}
