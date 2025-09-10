package routes

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/fraiday-org/api-service/internal/api/handlers"
	"github.com/fraiday-org/api-service/internal/api/middleware"
	"github.com/fraiday-org/api-service/internal/config"
	"github.com/fraiday-org/api-service/internal/repository"
	"github.com/fraiday-org/api-service/internal/service"
	"github.com/fraiday-org/api-service/internal/tasks"
)

func Register(r *gin.Engine, cfg *config.Config, logger *zap.Logger, mongoClient *mongo.Client) {
	db := mongoClient.Database(cfg.MongoDB)


	// Auth middleware (protects all except /auth/login, /health, /ping, /docs)
	r.Use(middleware.AuthMiddleware(logger, db))


	// Health and Monitoring
	healthHandler := handlers.NewHealthHandler(cfg, logger, mongoClient)
	r.GET("/api/v1/health", healthHandler.Health)
	r.GET("/api/v1/ping", healthHandler.Ping)
	r.GET("/api/v1/readiness", healthHandler.Readiness)
	r.GET("/api/v1/healthz", healthHandler.Healthz)
	
	// Metrics
	metricsHandler := handlers.NewMetricsHandler(logger)
	r.GET("/api/v1/metrics", metricsHandler.GetMetrics)

	// Clients (moved up for use in message creation)
	clientRepo := repository.NewClientRepository(db)
	clientService := service.NewClientService(clientRepo)
	clientHandler := handlers.NewClientHandler(clientService)
	
	// Client Channels
	clientChannelRepo := repository.NewClientChannelRepository(db)
	clientChannelService := service.NewClientChannelService(clientChannelRepo, clientRepo)
	clientChannelHandler := handlers.NewClientChannelHandler(logger)

	// Chat Sessions
	chatSessionRepo := repository.NewChatSessionRepository(db)
	chatSessionService := service.NewChatSessionService(chatSessionRepo)
	chatSessionHandler := handlers.NewChatSessionHandler(chatSessionService)

	// Initialize event services for chat message events
	eventRepo := repository.NewEventRepository(db)
	eventService := service.NewEventService(eventRepo)
	eventProcessorConfigRepo := repository.NewEventProcessorConfigRepository(db)
	eventProcessorConfigService := service.NewEventProcessorConfigService(eventProcessorConfigRepo)
	eventDeliveryRepo := repository.NewEventDeliveryRepository(db)
	eventDeliveryAttemptRepo := repository.NewEventDeliveryAttemptRepository(db)
	eventDeliveryTrackingService := service.NewEventDeliveryTrackingService(eventDeliveryRepo, eventDeliveryAttemptRepo)
	
	// Chat Messages
	chatMsgRepo := repository.NewChatMessageRepository(db)
	
	// Initialize task client for event publishing to RabbitMQ
	rabbitMQURL := cfg.GetRabbitMQURL()
	taskClient, err := tasks.NewTaskClient(rabbitMQURL, logger, cfg)
	if err != nil {
		logger.Warn("Failed to create task client for API server, events will be processed directly", zap.Error(err))
		taskClient = nil
	}
	
	// Initialize PayloadService with ThreadManagerService from ChatSessionService first
	payloadService := service.NewPayloadService(nil, chatSessionService, chatSessionService.ThreadManager) // ChatMessageService will be set later
	
	// Initialize EventPublisherService with PayloadService
	eventPublisherService := service.NewEventPublisherService(eventService, eventProcessorConfigService, eventDeliveryTrackingService, chatSessionRepo, chatMsgRepo, nil, nil, payloadService, taskClient)
	
	chatMsgService := service.NewChatMessageService(chatMsgRepo, eventPublisherService, payloadService)
	
	// Update PayloadService with ChatMessageService
	payloadService.ChatMessageService = chatMsgService
	
	chatMsgHandler := handlers.NewChatMessageHandler(chatMsgService, chatSessionService, clientService, clientChannelService)

	r.POST("/api/v1/messages", chatMsgHandler.CreateMessage)
	r.GET("/api/v1/messages", chatMsgHandler.ListMessages)
	r.PUT("/api/v1/messages/:id", chatMsgHandler.UpdateMessage)
	r.POST("/api/v1/messages/bulk", chatMsgHandler.BulkCreateMessages)

	// Chat Message Feedback
	chatMsgFeedbackRepo := repository.NewChatMessageFeedbackRepository(db)
	chatMsgFeedbackService := service.NewChatMessageFeedbackService(chatMsgFeedbackRepo)
	chatMsgFeedbackHandler := handlers.NewChatMessageFeedbackHandler(chatMsgFeedbackService)

	r.POST("/api/v1/messages/:message_id/feedbacks", chatMsgFeedbackHandler.CreateFeedback)
	r.GET("/api/v1/messages/:message_id/feedbacks", chatMsgFeedbackHandler.ListFeedbacks)
	r.PATCH("/api/v1/messages/:message_id/feedbacks/:feedback_id", chatMsgFeedbackHandler.UpdateFeedback)

	r.POST("/api/v1/sessions", chatSessionHandler.CreateSession)
	r.GET("/api/v1/sessions/:session_id", chatSessionHandler.GetSession)
	r.GET("/api/v1/sessions", chatSessionHandler.ListSessions)

	// Chat Session Threads
	chatSessionThreadRepo := repository.NewChatSessionThreadRepository(db)
	chatSessionThreadService := service.NewChatSessionThreadService(chatSessionThreadRepo)
	chatSessionThreadHandler := handlers.NewChatSessionThreadHandler(chatSessionThreadService)

	r.POST("/api/v1/sessions/:session_id/threads", chatSessionThreadHandler.CreateThread)
	r.GET("/api/v1/sessions/:session_id/threads", chatSessionThreadHandler.ListThreads)
	r.GET("/api/v1/sessions/:session_id/active_thread", chatSessionThreadHandler.GetActiveThread)
	r.POST("/api/v1/sessions/:session_id/close_thread", chatSessionThreadHandler.CloseThread)

	// Chat Session Recap
	chatSessionRecapRepo := repository.NewChatSessionRecapRepository(db)
	chatSessionRecapService := service.NewChatSessionRecapService(chatSessionRecapRepo)
	chatSessionRecapHandler := handlers.NewChatSessionRecapHandler(chatSessionRecapService)

	r.POST("/api/v1/sessions/:session_id/recap", chatSessionRecapHandler.GenerateRecap)
	r.GET("/api/v1/sessions/:session_id/recap", chatSessionRecapHandler.GetLatestRecap)

	// Analytics
	analyticsService := service.NewAnalyticsService()
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)

	r.GET("/api/v1/analytics/dashboard", analyticsHandler.GetDashboardMetrics)
	r.GET("/api/v1/analytics/bot-engagement", analyticsHandler.GetBotEngagementMetrics)
	r.GET("/api/v1/analytics/containment-rate", analyticsHandler.GetContainmentRateMetrics)

	// Client endpoints (using services defined earlier)
	r.POST("/api/v1/clients", clientHandler.CreateClient)
	r.GET("/api/v1/clients", clientHandler.ListClients)
	r.PUT("/api/v1/clients/:client_id", clientHandler.UpdateClient)

	// Client Channel endpoints (using handler defined earlier)
	r.POST("/api/v1/clients/:client_id/channels", clientChannelHandler.CreateChannel)
	r.GET("/api/v1/clients/:client_id/channels", clientChannelHandler.ListChannels)
	r.GET("/api/v1/clients/:client_id/channels/:channel_id", clientChannelHandler.GetChannel)
	r.PUT("/api/v1/clients/:client_id/channels/:channel_id", clientChannelHandler.UpdateChannel)
	r.DELETE("/api/v1/clients/:client_id/channels/:channel_id", clientChannelHandler.DeleteChannel)
	r.GET("/api/v1/clients/:client_id/channels/:channel_id/config", clientChannelHandler.GetChannelConfig)
	r.PUT("/api/v1/clients/:client_id/channels/:channel_id/config", clientChannelHandler.UpdateChannelConfig)

	// Events
	eventsHandler := handlers.NewEventsHandler(logger)
	r.POST("/api/v1/events/processor-configs", eventsHandler.CreateEventProcessorConfig)
	r.GET("/api/v1/events/processor-configs", eventsHandler.ListEventProcessorConfigs)
	r.GET("/api/v1/events/processor-configs/:config_id", eventsHandler.GetEventProcessorConfig)
	r.PUT("/api/v1/events/processor-configs/:config_id", eventsHandler.UpdateEventProcessorConfig)
	r.DELETE("/api/v1/events/processor-configs/:config_id", eventsHandler.DeleteEventProcessorConfig)
	r.POST("/api/v1/events/process", eventsHandler.ProcessEvent)
	r.GET("/api/v1/events/:event_id/status", eventsHandler.GetEventStatus)

	// Event Processor Configs (Client-specific) - reuse existing services
	eventProcessorConfigHandler := handlers.NewEventProcessorConfigHandler(eventProcessorConfigService)

	r.POST("/api/v1/clients/:client_id/processor-configs", eventProcessorConfigHandler.CreateProcessorConfig)
	r.GET("/api/v1/clients/:client_id/processor-configs", eventProcessorConfigHandler.ListProcessorConfigs)
	r.GET("/api/v1/clients/:client_id/processor-configs/:config_id", eventProcessorConfigHandler.GetProcessorConfig)
	r.PUT("/api/v1/clients/:client_id/processor-configs/:config_id", eventProcessorConfigHandler.UpdateProcessorConfig)
	r.DELETE("/api/v1/clients/:client_id/processor-configs/:config_id", eventProcessorConfigHandler.DeleteProcessorConfig)


	// CSAT (Customer Satisfaction)
	csatConfigRepo := repository.NewCSATConfigurationRepository(db)
	csatQuestionRepo := repository.NewCSATQuestionTemplateRepository(db)
	csatSessionRepo := repository.NewCSATSessionRepository(db)
	csatResponseRepo := repository.NewCSATResponseRepository(db)
	
	// CSAT Event Publisher Service - with CSAT repositories for proper client resolution
	csatEventPublisherService := service.NewEventPublisherService(eventService, eventProcessorConfigService, eventDeliveryTrackingService, chatSessionRepo, chatMsgRepo, csatSessionRepo, csatQuestionRepo, payloadService, taskClient)
	
	csatService := service.NewCSATService(
		csatConfigRepo,
		csatQuestionRepo,
		csatSessionRepo,
		csatResponseRepo,
		chatMsgRepo,
		chatSessionRepo,
		chatSessionThreadService,
		csatEventPublisherService,
		payloadService,
	)
	csatHandler := handlers.NewCSATHandler(csatService)

	// CSAT API endpoints
	r.POST("/api/v1/csat/trigger", csatHandler.TriggerCSAT)
	r.POST("/api/v1/csat/respond", csatHandler.RespondToCSAT)
	r.GET("/api/v1/csat/sessions/:session_id", csatHandler.GetCSATSession)
	
	// CSAT configuration and questions (client-specific)
	r.GET("/api/v1/clients/:client_id/channels/:channel_id/csat/config", csatHandler.GetCSATConfiguration)
	r.PUT("/api/v1/clients/:client_id/channels/:channel_id/csat/config", csatHandler.UpdateCSATConfiguration)
	r.GET("/api/v1/clients/:client_id/channels/:channel_id/csat/questions", csatHandler.GetCSATQuestions)
	r.PUT("/api/v1/clients/:client_id/channels/:channel_id/csat/questions", csatHandler.UpdateCSATQuestions)
}
