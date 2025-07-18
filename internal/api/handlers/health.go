package handlers

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"

	"github.com/fraiday-org/api-service/internal/config"
)

var startTime = time.Now()

type HealthHandler struct {
	cfg         *config.Config
	logger      *zap.Logger
	mongoClient *mongo.Client
}

func NewHealthHandler(cfg *config.Config, logger *zap.Logger, mongoClient *mongo.Client) *HealthHandler {
	return &HealthHandler{
		cfg:         cfg,
		logger:      logger,
		mongoClient: mongoClient,
	}
}

func (h *HealthHandler) Health(c *gin.Context) {
	now := time.Now().UTC()
	uptime := now.Sub(startTime).Truncate(time.Second).String()

	// MongoDB check
	dbStatus := "ok"
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := h.mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		dbStatus = "error"
	}

	// Cache check (stub)
	cacheStatus := "ok"

	resp := gin.H{
		"status":  "healthy",
		"time":    now.Format(time.RFC3339),
		"version": "1.0.0",
		"uptime":  uptime,
		"system": gin.H{
			"go_version": runtime.Version(),
			"num_cpu":    runtime.NumCPU(),
			"arch":       runtime.GOARCH,
			"os":         runtime.GOOS,
		},
		"checks": gin.H{
			"database": dbStatus,
			"cache":    cacheStatus,
			"service":  "running",
		},
	}

	c.JSON(http.StatusOK, resp)
}

func (h *HealthHandler) Ping(c *gin.Context) {
	now := time.Now().UTC()
	resp := gin.H{
		"status": "ok",
		"time":   now.Format(time.RFC3339),
	}
	c.JSON(http.StatusOK, resp)
}
