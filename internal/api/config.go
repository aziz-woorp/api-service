package api

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/fraiday-org/api-service/internal/api/middleware"
	"github.com/fraiday-org/api-service/internal/api/routes"
	"github.com/fraiday-org/api-service/internal/config"
	"github.com/fraiday-org/api-service/internal/service"
)

func SetupRouter(cfg *config.Config, logger *zap.Logger, mongoClient *mongo.Client) *gin.Engine {
	engine := gin.New()

	// Initialize metrics service
	metricsService := service.NewMetricsService(logger)

	// Middleware
	engine.Use(middleware.RequestID())
	engine.Use(middleware.Logger(logger))
	engine.Use(middleware.Recovery(logger))
	engine.Use(middleware.CORS())
	engine.Use(middleware.ErrorHandler())
	engine.Use(middleware.MetricsMiddleware(metricsService))

	// Register routes
	routes.Register(engine, cfg, logger, mongoClient)

	return engine
}
