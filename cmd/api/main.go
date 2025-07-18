package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/fraiday-org/api-service/internal/api"
	"github.com/fraiday-org/api-service/internal/config"
	"github.com/fraiday-org/api-service/internal/repository"
)

func main() {
	// Load config
	cfg := config.LoadConfig()

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Initialize logger
	logger, err := zap.NewProduction()
	if cfg.AppEnv == "development" {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Connect to MongoDB
	mongoClient, err := repository.NewMongoClient(cfg.MongoURI)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			logger.Error("Failed to disconnect from MongoDB", zap.Error(err))
		}
	}()

	// Set up Gin engine
	engine := api.SetupRouter(cfg, logger, mongoClient)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.AppPort)
	logger.Info("Starting server", zap.String("addr", addr))
	if err := engine.Run(addr); err != nil {
		logger.Fatal("Server exited with error", zap.Error(err))
	}
}
