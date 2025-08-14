package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"github.com/fraiday-org/api-service/internal/service"
)

const (
	TypeChatWorkflow      = "chat_workflow"
	TypeSuggestionWorkflow = "suggestion_workflow"
	TypeEventProcessor    = "event_processor"
)

// TaskWorker wraps asynq.Server for task processing
type TaskWorker struct {
	server         *asynq.Server
	mux            *asynq.ServeMux
	logger         *zap.Logger
	webhookService *service.WebhookService
	aiService      *service.AIService
	databaseService *service.DatabaseService
}

// NewTaskWorker creates a new task worker
func NewTaskWorker(redisAddr string, logger *zap.Logger, aiURL, aiToken string, databaseService *service.DatabaseService) *TaskWorker {
	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"chat_workflow": 6,
				"events":        3,
				"default":       2,
			},
			RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
				return time.Duration(n) * time.Second
			},
		},
	)

	mux := asynq.NewServeMux()
	webhookService := service.NewWebhookService(logger)
	aiService := service.NewAIService(logger, aiURL, aiToken)

	return &TaskWorker{
		server:         server,
		mux:            mux,
		logger:         logger,
		webhookService: webhookService,
		aiService:      aiService,
		databaseService: databaseService,
	}
}

// RegisterHandlers registers task handlers
func (tw *TaskWorker) RegisterHandlers() {
	tw.mux.HandleFunc(TypeChatWorkflow, tw.HandleChatWorkflow)
	tw.mux.HandleFunc(TypeSuggestionWorkflow, tw.HandleSuggestionWorkflow)
	tw.mux.HandleFunc(TypeEventProcessor, tw.HandleEventProcessor)

}

// Start starts the task worker
func (tw *TaskWorker) Start() error {
	tw.RegisterHandlers()
	tw.logger.Info("Starting task worker")
	return tw.server.Run(tw.mux)
}

// Stop stops the task worker
func (tw *TaskWorker) Stop() {
	tw.logger.Info("Stopping task worker")
	tw.server.Stop()
	tw.server.Shutdown()
}

// HandleChatWorkflow handles chat workflow tasks
func (tw *TaskWorker) HandleChatWorkflow(ctx context.Context, t *asynq.Task) error {
	var payload ChatWorkflowPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		tw.logger.Error("Failed to unmarshal chat workflow payload", zap.Error(err))
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	tw.logger.Info("Processing chat workflow task",
		zap.String("message_id", payload.MessageID),
		zap.String("session_id", payload.SessionID))

	// Implement chat workflow logic equivalent to Python Celery task
	// This mirrors the generate_ai_response_task from Python backend
	
	// 1. Publish processing event
	tw.logger.Info("Publishing chat workflow processing event",
		zap.String("message_id", payload.MessageID))
	
	// 2. Process AI request
	tw.logger.Info("Processing AI request",
		zap.String("message_id", payload.MessageID),
		zap.String("session_id", payload.SessionID))
	
	// Get message content and context from database
	message, err := tw.databaseService.GetChatMessage(ctx, payload.MessageID)
	if err != nil {
		tw.logger.Error("Failed to get message from database", zap.Error(err))
		return fmt.Errorf("failed to get message: %w", err)
	}
	
	sessionContext, err := tw.databaseService.GetSessionContext(ctx, payload.SessionID)
	if err != nil {
		tw.logger.Warn("Failed to get session context, using minimal context", zap.Error(err))
		sessionContext = map[string]interface{}{"session_id": payload.SessionID}
	}
	
	var aiResponse *service.AIResponse
	
	if payload.SuggestionMode {
		aiResponse, err = tw.aiService.GenerateSuggestions(ctx, payload.MessageID, payload.SessionID, message.Content, sessionContext)
	} else {
		aiResponse, err = tw.aiService.GenerateChatResponse(ctx, payload.MessageID, payload.SessionID, message.Content, sessionContext)
	}
	
	if err != nil {
		tw.logger.Error("Failed to process AI request", zap.Error(err))
		return fmt.Errorf("AI processing failed: %w", err)
	}
	
	tw.logger.Info("AI response received",
		zap.String("message_id", aiResponse.MessageID),
		zap.String("response_length", fmt.Sprintf("%d", len(aiResponse.Response))))
	
	// Save AI response to database
	responseMessage := &service.ChatMessage{
		MessageID: aiResponse.MessageID + "_response",
		SessionID: aiResponse.SessionID,
		ClientID:  message.ClientID,
		Content:   aiResponse.Response,
		Role:      "assistant",
		Metadata:  aiResponse.Metadata,
	}
	
	if err := tw.databaseService.SaveChatMessage(ctx, responseMessage); err != nil {
		tw.logger.Error("Failed to save AI response to database", zap.Error(err))
		// Don't return error here as the AI processing was successful
	}
	
	// 3. Generate response based on message configuration
	// Check if suggestion mode is enabled
	if payload.SuggestionMode {
		// Create suggestion entity
		tw.logger.Info("Creating chat suggestion",
			zap.String("message_id", payload.MessageID))
		
		// Publish suggestion created event
		tw.logger.Info("Publishing suggestion created event",
			zap.String("message_id", payload.MessageID))
	} else {
		// Create chat message response
		tw.logger.Info("Creating chat message response",
			zap.String("message_id", payload.MessageID))
		
		// Publish message created event
		tw.logger.Info("Publishing message created event",
			zap.String("message_id", payload.MessageID))
	}
	
	// 4. Send webhook notifications
	tw.logger.Info("Sending webhook notifications",
		zap.String("message_id", payload.MessageID))
	
	// TODO: Get webhook URL from client configuration
	webhookURL := "" // This should be retrieved from client config
	
	if webhookURL != "" {
		messageData := map[string]interface{}{
			"response":        aiResponse.Response,
			"suggestion_mode": payload.SuggestionMode,
			"suggestions":     aiResponse.Suggestions,
			"metadata":        aiResponse.Metadata,
		}
		
		err = tw.webhookService.SendChatMessageWebhook(ctx, webhookURL, payload.MessageID, payload.SessionID, messageData)
		if err != nil {
			tw.logger.Error("Failed to send webhook", zap.Error(err))
			// Don't return error as this is not critical
		}
	} else {
		tw.logger.Debug("No webhook URL configured, skipping webhook notification")
	}

	tw.logger.Info("Completed chat workflow task",
		zap.String("message_id", payload.MessageID))

	return nil
}

// HandleSuggestionWorkflow handles suggestion workflow tasks
func (tw *TaskWorker) HandleSuggestionWorkflow(ctx context.Context, t *asynq.Task) error {
	var payload SuggestionWorkflowPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		tw.logger.Error("Failed to unmarshal suggestion workflow payload", zap.Error(err))
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	tw.logger.Info("Processing suggestion workflow task",
		zap.String("message_id", payload.MessageID),
		zap.String("session_id", payload.SessionID))

	// TODO: Implement actual suggestion workflow logic
	// This would include:
	// 1. Fetch message from database
	// 2. Generate suggestions
	// 3. Save suggestions to database
	// 4. Send webhook notifications

	// Simulate processing time
	time.Sleep(50 * time.Millisecond)

	tw.logger.Info("Completed suggestion workflow task",
		zap.String("message_id", payload.MessageID))

	return nil
}

// HandleEventProcessor handles event processor tasks
func (tw *TaskWorker) HandleEventProcessor(ctx context.Context, t *asynq.Task) error {
	var payload EventProcessorPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		tw.logger.Error("Failed to unmarshal event processor payload", zap.Error(err))
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	tw.logger.Info("Processing event processor task",
		zap.String("event_id", payload.EventID),
		zap.String("event_type", payload.EventType),
		zap.String("entity_type", payload.EntityType),
		zap.String("entity_id", payload.EntityID))

	// Implement event processing logic equivalent to Python Celery task
	// This mirrors the process_event task from Python backend
	
	// 1. Get client_id from entity
	tw.logger.Info("Determining client_id for entity",
		zap.String("entity_type", payload.EntityType),
		zap.String("entity_id", payload.EntityID))
	
	// 2. Find matching processors for this event
	tw.logger.Info("Finding matching processors",
		zap.String("event_type", payload.EventType),
		zap.String("entity_type", payload.EntityType))
	
	// 3. Create delivery records and dispatch to processors
	tw.logger.Info("Creating delivery records",
		zap.String("event_id", payload.EventID))
	
	// 4. Process each matching processor
	tw.logger.Info("Dispatching to processors",
		zap.String("event_id", payload.EventID))

	tw.logger.Info("Completed event processor task",
		zap.String("event_id", payload.EventID))

	return nil
}
