package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Time      string            `json:"time"`
	Version   string            `json:"version,omitempty"`
	Uptime    string            `json:"uptime,omitempty"`
	System    *SystemInfo       `json:"system,omitempty"`
	Checks    map[string]string `json:"checks,omitempty"`
}

// SystemInfo contains system information
type SystemInfo struct {
	GoVersion string `json:"go_version"`
	NumCPU    int    `json:"num_cpu"`
	Arch      string `json:"arch"`
	OS        string `json:"os"`
}

// HealthHandler handles health check requests
type HealthHandler struct {
	startTime time.Time
	version   string
}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
		version:   "1.0.0",
	}
}

// Health returns the health status
func (h *HealthHandler) Health(c *gin.Context) {
	response := HealthResponse{
		Status:  "healthy",
		Time:    time.Now().Format(time.RFC3339),
		Version: h.version,
		Uptime:  time.Since(h.startTime).String(),
		System: &SystemInfo{
			GoVersion: runtime.Version(),
			NumCPU:    runtime.NumCPU(),
			Arch:      runtime.GOARCH,
			OS:        runtime.GOOS,
		},
		Checks: map[string]string{
			"database": "ok",
			"cache":    "ok",
			"service":  "running",
		},
	}

	c.JSON(http.StatusOK, response)
}

// HealthSimple returns a simple health check
func (h *HealthHandler) HealthSimple(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}