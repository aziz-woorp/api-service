// Package service provides business logic for analytics endpoints.
package service

import (
	"time"

	"github.com/fraiday-org/api-service/internal/api/dto"
)

type AnalyticsService struct{}

func NewAnalyticsService() *AnalyticsService {
	return &AnalyticsService{}
}

func (s *AnalyticsService) GetDashboardMetrics(startTime, endTime time.Time) *dto.DashboardMetricsResponse {
	// Stubbed data
	data := map[string]interface{}{
		"total_conversations": 123,
		"handoff_rate":        12.5,
		"containment_rate":    87.5,
		"conversations_by_time": []map[string]interface{}{
			{"time": "2025-07-25T00:00:00Z", "count": 10},
			{"time": "2025-07-25T01:00:00Z", "count": 15},
		},
		"sessions_by_hour": []map[string]interface{}{
			{"hour": "00", "count": 5},
			{"hour": "01", "count": 8},
		},
		"last_updated": time.Now().UTC(),
	}
	return &dto.DashboardMetricsResponse{
		Success: true,
		Data:    data,
	}
}

func (s *AnalyticsService) GetBotEngagementMetrics(startTime, endTime time.Time) *dto.BotEngagementMetricsResponse {
	// Stubbed data
	data := map[string]interface{}{
		"avg_session_duration":           300,
		"avg_messages_per_session":       7.2,
		"avg_resolution_time":            250,
		"first_response_time":            5,
		"response_rate":                  98.5,
		"first_response_time_distribution": []map[string]interface{}{
			{"range": "0-5s", "count": 20},
			{"range": "5-10s", "count": 5},
		},
		"session_duration_distribution": []map[string]interface{}{
			{"range": "0-1m", "count": 10},
			{"range": "1-3m", "count": 15},
		},
		"messages_per_session_distribution": []map[string]interface{}{
			{"count": "1-2", "sessions": 8},
			{"count": "3-5", "sessions": 12},
		},
		"last_updated": time.Now().UTC(),
	}
	return &dto.BotEngagementMetricsResponse{
		Success: true,
		Data:    data,
	}
}

func (s *AnalyticsService) GetContainmentRateMetrics(startTime, endTime time.Time, aggregation string) *dto.ContainmentRateResponse {
	// Stubbed data
	data := []interface{}{
		map[string]interface{}{
			"time":       "2025-07-25T00:00:00Z",
			"time_label": "00:00",
			"value":      85.0,
			"unit":       "percent",
		},
		map[string]interface{}{
			"time":       "2025-07-25T01:00:00Z",
			"time_label": "01:00",
			"value":      90.0,
			"unit":       "percent",
		},
	}
	metadata := map[string]interface{}{
		"aggregation": aggregation,
		"start_time":  startTime,
		"end_time":    endTime,
	}
	return &dto.ContainmentRateResponse{
		Success:  true,
		Data:     data,
		Metadata: metadata,
	}
}
