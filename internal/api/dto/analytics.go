// Package dto defines request/response payloads for analytics endpoints.
package dto

// DashboardMetricsResponse is the response for dashboard analytics metrics.
type DashboardMetricsResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   *string                `json:"error,omitempty"`
}

// BotEngagementMetricsResponse is the response for bot engagement metrics.
type BotEngagementMetricsResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   *string                `json:"error,omitempty"`
}

// ContainmentRateResponse is the response for containment rate analytics.
type ContainmentRateResponse struct {
	Success  bool                   `json:"success"`
	Data     []interface{}          `json:"data,omitempty"`
	Error    *string                `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
