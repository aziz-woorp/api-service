package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/fraiday-org/api-service/internal/api"
	"github.com/fraiday-org/api-service/internal/config"
	"github.com/fraiday-org/api-service/internal/repository"
	"github.com/fraiday-org/api-service/internal/service"
	"github.com/fraiday-org/api-service/internal/tasks"
)

func main() {
	// Parse command line arguments
	var (
		mode        = flag.String("mode", "server", "Mode to run: server or worker")
		queue       = flag.String("queue", "", "Queue name for worker mode")
		concurrency = flag.Int("concurrency", 1, "Number of concurrent workers")
	)
	flag.Parse()

	// If no mode specified but args exist, use first arg as mode
	if len(os.Args) > 1 && *mode == "server" {
		*mode = os.Args[1]
	}

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

	// Run based on mode
	switch *mode {
	case "server":
		runServer(cfg, logger, mongoClient)
	case "worker":
		runWorker(cfg, logger, mongoClient, *queue, *concurrency)
	default:
		logger.Fatal("Invalid mode", zap.String("mode", *mode))
	}
}

func runServer(cfg *config.Config, logger *zap.Logger, mongoClient *mongo.Client) {
	// Set up Gin engine
	engine := api.SetupRouter(cfg, logger, mongoClient)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.AppPort)
	logger.Info("Starting server", zap.String("addr", addr))
	if err := engine.Run(addr); err != nil {
		logger.Fatal("Server exited with error", zap.Error(err))
	}
}

func runWorker(cfg *config.Config, logger *zap.Logger, mongoClient *mongo.Client, queueName string, concurrency int) {
	if queueName == "" {
		logger.Fatal("Queue name is required for worker mode")
	}

	// Build Redis URL from config
	redisURL := buildRedisURL(cfg)

	// Parse queue names (comma-separated)
	queues := strings.Split(queueName, ",")
	for i, q := range queues {
		queues[i] = strings.TrimSpace(q)
	}

	logger.Info("Starting worker", 
		zap.Strings("queues", queues), 
		zap.Int("concurrency", concurrency),
		zap.String("redis_url", redisURL))

	// Initialize database service
	databaseService := service.NewDatabaseService(logger, mongoClient, "api_service")
	
	// Initialize task worker
	taskWorker := tasks.NewTaskWorker(redisURL, logger, cfg.AIServiceURL, cfg.SlackAIToken, databaseService)

	// Handle shutdown signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logger.Info("Shutdown signal received")
		taskWorker.Stop()
	}()

	// Start the worker
	if err := taskWorker.Start(); err != nil {
		logger.Fatal("Failed to start task worker", zap.Error(err))
	}

	logger.Info("Worker stopped")
}

// buildRedisURL constructs a Redis URL from config components
func buildRedisURL(cfg *config.Config) string {
	// If we have a broker URL (for Celery compatibility), use it
	if cfg.CeleryBrokerURL != "" {
		return cfg.CeleryBrokerURL
	}

	// Build from individual components
	host := cfg.RedisHost
	if host == "" {
		host = "localhost"
	}

	port := cfg.RedisPort
	if port == 0 {
		port = 6379
	}

	db := cfg.RedisDB
	if db == 0 {
		db = 0
	}

	if cfg.RedisPassword != "" {
		return fmt.Sprintf("redis://:%s@%s:%d/%d", cfg.RedisPassword, host, port, db)
	}
	return fmt.Sprintf("redis://%s:%d/%d", host, port, db)
}
