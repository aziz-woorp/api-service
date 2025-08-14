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
	"github.com/fraiday-org/api-service/internal/service"
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

	// Chat Messages
	chatMsgRepo := repository.NewChatMessageRepository(db)
	chatMsgService := service.NewChatMessageService(chatMsgRepo)
	chatMsgHandler := handlers.NewChatMessageHandler(chatMsgService)

	r.POST("/messages", chatMsgHandler.CreateMessage)
	r.GET("/messages", chatMsgHandler.ListMessages)
	r.PUT("/messages/:id", chatMsgHandler.UpdateMessage)
	r.POST("/messages/bulk", chatMsgHandler.BulkCreateMessages)

	// Chat Message Feedback
	chatMsgFeedbackRepo := repository.NewChatMessageFeedbackRepository(db)
	chatMsgFeedbackService := service.NewChatMessageFeedbackService(chatMsgFeedbackRepo)
	chatMsgFeedbackHandler := handlers.NewChatMessageFeedbackHandler(chatMsgFeedbackService)

	r.POST("/messages/:message_id/feedbacks", chatMsgFeedbackHandler.CreateFeedback)
	r.GET("/messages/:message_id/feedbacks", chatMsgFeedbackHandler.ListFeedbacks)
	r.PATCH("/messages/:message_id/feedbacks/:feedback_id", chatMsgFeedbackHandler.UpdateFeedback)

	// Chat Sessions
	chatSessionRepo := repository.NewChatSessionRepository(db)
	chatSessionService := service.NewChatSessionService(chatSessionRepo)
	chatSessionHandler := handlers.NewChatSessionHandler(chatSessionService)

	r.POST("/sessions", chatSessionHandler.CreateSession)
	r.GET("/sessions/:session_id", chatSessionHandler.GetSession)
	r.GET("/sessions", chatSessionHandler.ListSessions)

	// Chat Session Threads
	chatSessionThreadRepo := repository.NewChatSessionThreadRepository(db)
	chatSessionThreadService := service.NewChatSessionThreadService(chatSessionThreadRepo)
	chatSessionThreadHandler := handlers.NewChatSessionThreadHandler(chatSessionThreadService)

	r.POST("/sessions/:session_id/threads", chatSessionThreadHandler.CreateThread)
	r.GET("/sessions/:session_id/threads", chatSessionThreadHandler.ListThreads)
	r.GET("/sessions/:session_id/active_thread", chatSessionThreadHandler.GetActiveThread)
	r.POST("/sessions/:session_id/close_thread", chatSessionThreadHandler.CloseThread)

	// Chat Session Recap
	chatSessionRecapRepo := repository.NewChatSessionRecapRepository(db)
	chatSessionRecapService := service.NewChatSessionRecapService(chatSessionRecapRepo)
	chatSessionRecapHandler := handlers.NewChatSessionRecapHandler(chatSessionRecapService)

	r.POST("/sessions/:session_id/recap", chatSessionRecapHandler.GenerateRecap)
	r.GET("/sessions/:session_id/recap", chatSessionRecapHandler.GetLatestRecap)

	// Analytics
	analyticsService := service.NewAnalyticsService()
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)

	r.GET("/analytics/dashboard", analyticsHandler.GetDashboardMetrics)
	r.GET("/analytics/bot-engagement", analyticsHandler.GetBotEngagementMetrics)
	r.GET("/analytics/containment-rate", analyticsHandler.GetContainmentRateMetrics)

	// Clients
	clientRepo := repository.NewClientRepository(db)
	clientService := service.NewClientService(clientRepo)
	clientHandler := handlers.NewClientHandler(clientService)

	r.POST("/clients", clientHandler.CreateClient)
	r.GET("/clients", clientHandler.ListClients)
	r.PUT("/clients/:client_id", clientHandler.UpdateClient)

	// Client Channels
	clientChannelHandler := handlers.NewClientChannelHandler(logger)
	r.POST("/clients/:client_id/channels", clientChannelHandler.CreateChannel)
	r.GET("/clients/:client_id/channels", clientChannelHandler.ListChannels)
	r.GET("/clients/:client_id/channels/:channel_id", clientChannelHandler.GetChannel)
	r.PUT("/clients/:client_id/channels/:channel_id", clientChannelHandler.UpdateChannel)
	r.DELETE("/clients/:client_id/channels/:channel_id", clientChannelHandler.DeleteChannel)
	r.GET("/clients/:client_id/channels/:channel_id/config", clientChannelHandler.GetChannelConfig)
	r.PUT("/clients/:client_id/channels/:channel_id/config", clientChannelHandler.UpdateChannelConfig)



	// Events
	eventsHandler := handlers.NewEventsHandler(logger)
	r.POST("/events/processor-configs", eventsHandler.CreateEventProcessorConfig)
	r.GET("/events/processor-configs", eventsHandler.ListEventProcessorConfigs)
	r.GET("/events/processor-configs/:config_id", eventsHandler.GetEventProcessorConfig)
	r.PUT("/events/processor-configs/:config_id", eventsHandler.UpdateEventProcessorConfig)
	r.DELETE("/events/processor-configs/:config_id", eventsHandler.DeleteEventProcessorConfig)
	r.POST("/events/process", eventsHandler.ProcessEvent)
	r.GET("/events/:event_id/status", eventsHandler.GetEventStatus)



	// Metrics
	metricsHandler := handlers.NewMetricsHandler(logger)
	r.GET("/metrics", metricsHandler.GetMetrics)
}
