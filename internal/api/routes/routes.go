package routes

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/fraiday-org/api-service/internal/api/handlers"
	"github.com/fraiday-org/api-service/internal/api/middleware"
	"github.com/fraiday-org/api-service/internal/config"
	"github.com/fraiday-org/api-service/internal/repository"
)

func Register(r *gin.Engine, cfg *config.Config, logger *zap.Logger, mongoClient *mongo.Client) {
	db := mongoClient.Database("api_service_dev") // or from config/env

	// Ensure demo user exists (for dev/demo)
	userRepo := repository.NewUserRepository(db)
	_ = userRepo.EnsureDemoUser(context.Background())

	// Auth middleware (protects all except /auth/login, /health, /ping, /docs)
	r.Use(middleware.AuthMiddleware(logger, db))

	// Auth
	authHandler := handlers.NewAuthHandler(logger, db)
	r.POST("/auth/login", authHandler.Login)

	// Health
	healthHandler := handlers.NewHealthHandler(cfg, logger, mongoClient)
	r.GET("/health", healthHandler.Health)
	r.GET("/ping", healthHandler.Ping)
}
