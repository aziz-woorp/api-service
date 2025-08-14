package service

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

// MetricsService handles application metrics
type MetricsService struct {
	logger *zap.Logger

	// HTTP metrics
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight prometheus.Gauge

	// Application metrics
	appInfo              *prometheus.GaugeVec
	chatMessagesTotal    prometheus.Counter
	chatSessionsTotal    prometheus.Counter
	aiResponseTime       prometheus.Histogram
	taskQueueSize        *prometheus.GaugeVec
	taskProcessingTime   *prometheus.HistogramVec
}

// NewMetricsService creates a new metrics service
func NewMetricsService(logger *zap.Logger) *MetricsService {
	return &MetricsService{
		logger: logger,
		httpRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		httpRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		httpRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
		),
		appInfo: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "app_info",
				Help: "Application information",
			},
			[]string{"version", "name"},
		),
		chatMessagesTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "chat_messages_total",
				Help: "Total number of chat messages processed",
			},
		),
		chatSessionsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "chat_sessions_total",
				Help: "Total number of chat sessions created",
			},
		),
		aiResponseTime: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "ai_response_time_seconds",
				Help:    "AI response time in seconds",
				Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0},
			},
		),
		taskQueueSize: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "task_queue_size",
				Help: "Number of tasks in queue",
			},
			[]string{"queue"},
		),
		taskProcessingTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "task_processing_time_seconds",
				Help:    "Task processing time in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"task_type"},
		),
	}
}

// InitAppInfo initializes application information metrics
func (ms *MetricsService) InitAppInfo(version, name string) {
	ms.appInfo.WithLabelValues(version, name).Set(1)
	ms.logger.Info("Initialized app info metrics", 
		zap.String("version", version),
		zap.String("name", name))
}

// TrackRequestStart tracks the start of an HTTP request
func (ms *MetricsService) TrackRequestStart(method, path string) time.Time {
	ms.httpRequestsInFlight.Inc()
	return time.Now()
}

// TrackRequestEnd tracks the end of an HTTP request
func (ms *MetricsService) TrackRequestEnd(startTime time.Time, method, path string, statusCode int) {
	ms.httpRequestsInFlight.Dec()
	ms.httpRequestsTotal.WithLabelValues(method, path, fmt.Sprintf("%d", statusCode)).Inc()
	ms.httpRequestDuration.WithLabelValues(method, path).Observe(time.Since(startTime).Seconds())
}

// IncrementChatMessages increments the chat messages counter
func (ms *MetricsService) IncrementChatMessages() {
	ms.chatMessagesTotal.Inc()
}

// IncrementChatSessions increments the chat sessions counter
func (ms *MetricsService) IncrementChatSessions() {
	ms.chatSessionsTotal.Inc()
}

// ObserveAIResponseTime observes AI response time
func (ms *MetricsService) ObserveAIResponseTime(duration time.Duration) {
	ms.aiResponseTime.Observe(duration.Seconds())
}

// SetTaskQueueSize sets the task queue size
func (ms *MetricsService) SetTaskQueueSize(queue string, size float64) {
	ms.taskQueueSize.WithLabelValues(queue).Set(size)
}

// ObserveTaskProcessingTime observes task processing time
func (ms *MetricsService) ObserveTaskProcessingTime(taskType string, duration time.Duration) {
	ms.taskProcessingTime.WithLabelValues(taskType).Observe(duration.Seconds())
}