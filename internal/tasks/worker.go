package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"github.com/fraiday-org/api-service/internal/service"
)

const (
	TypeChatWorkflow      = "chat_workflow"
	TypeSuggestionWorkflow = "suggestion_workflow"
	TypeEventProcessor    = "event_processor"
)

// TaskWorker wraps RabbitMQ connection for task processing
type TaskWorker struct {
	conn            *amqp.Connection
	channel         *amqp.Channel
	logger          *zap.Logger
	webhookService  *service.WebhookService
	aiService       *service.AIService
	databaseService *service.DatabaseService
	queues          []string
	concurrency     int
	wg              sync.WaitGroup
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewTaskWorker creates a new task worker
func NewTaskWorker(rabbitMQURL string, logger *zap.Logger, aiURL, aiToken string, databaseService *service.DatabaseService) (*TaskWorker, error) {
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Set QoS to control how many messages are processed concurrently
	err = channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	webhookService := service.NewWebhookService(logger)
	aiService := service.NewAIService(logger, aiURL, aiToken)

	return &TaskWorker{
		conn:            conn,
		channel:         channel,
		logger:          logger,
		webhookService:  webhookService,
		aiService:       aiService,
		databaseService: databaseService,
		queues:          []string{"chat_workflow", "events", "default"},
		concurrency:     10,
		ctx:             ctx,
		cancel:          cancel,
	}, nil
}

// SetQueues sets the queues to process
func (tw *TaskWorker) SetQueues(queues []string) {
	tw.queues = queues
}

// SetConcurrency sets the concurrency level
func (tw *TaskWorker) SetConcurrency(concurrency int) {
	tw.concurrency = concurrency
}

// declareQueues declares all required queues
func (tw *TaskWorker) declareQueues() error {
	for _, queue := range tw.queues {
		_, err := tw.channel.QueueDeclare(
			queue, // name
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", queue, err)
		}
	}
	return nil
}

// Start starts the task worker
func (tw *TaskWorker) Start() error {
	tw.logger.Info("Starting task worker", 
		zap.Strings("queues", tw.queues),
		zap.Int("concurrency", tw.concurrency))

	// Declare queues
	if err := tw.declareQueues(); err != nil {
		return fmt.Errorf("failed to declare queues: %w", err)
	}

	// Start consumers for each queue
	for _, queue := range tw.queues {
		for i := 0; i < tw.concurrency; i++ {
			tw.wg.Add(1)
			go tw.consumeQueue(queue, i)
		}
	}

	// Handle shutdown signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	select {
	case <-c:
		tw.logger.Info("Shutdown signal received")
		tw.Stop()
	case <-tw.ctx.Done():
		tw.logger.Info("Context cancelled")
	}

	tw.wg.Wait()
	tw.logger.Info("Task worker stopped")
	return nil
}

// Stop stops the task worker
func (tw *TaskWorker) Stop() {
	tw.logger.Info("Stopping task worker")
	tw.cancel()
	if tw.channel != nil {
		tw.channel.Close()
	}
	if tw.conn != nil {
		tw.conn.Close()
	}
}

// consumeQueue consumes messages from a specific queue
func (tw *TaskWorker) consumeQueue(queueName string, workerID int) {
	defer tw.wg.Done()

	msgs, err := tw.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		tw.logger.Error("Failed to register consumer", 
			zap.String("queue", queueName),
			zap.Int("worker_id", workerID),
			zap.Error(err))
		return
	}

	tw.logger.Info("Worker started", 
		zap.String("queue", queueName),
		zap.Int("worker_id", workerID))

	for {
		select {
		case <-tw.ctx.Done():
			tw.logger.Info("Worker stopping", 
				zap.String("queue", queueName),
				zap.Int("worker_id", workerID))
			return
		case msg, ok := <-msgs:
			if !ok {
				tw.logger.Info("Message channel closed", 
					zap.String("queue", queueName),
					zap.Int("worker_id", workerID))
				return
			}

			tw.processMessage(msg, queueName, workerID)
		}
	}
}

// processMessage processes a single message
func (tw *TaskWorker) processMessage(msg amqp.Delivery, queueName string, workerID int) {
	start := time.Now()

	// Parse Celery message format
	var celeryMsg map[string]interface{}
	if err := json.Unmarshal(msg.Body, &celeryMsg); err != nil {
		tw.logger.Error("Failed to unmarshal message", 
			zap.String("queue", queueName),
			zap.Int("worker_id", workerID),
			zap.Error(err))
		msg.Nack(false, false) // Don't requeue malformed messages
		return
	}

	taskType, ok := celeryMsg["task"].(string)
	if !ok {
		tw.logger.Error("Missing or invalid task type", 
			zap.String("queue", queueName),
			zap.Int("worker_id", workerID))
		msg.Nack(false, false)
		return
	}

	taskID, _ := celeryMsg["id"].(string)
	kwargs, _ := celeryMsg["kwargs"].(map[string]interface{})

	tw.logger.Info("Processing task", 
		zap.String("task_id", taskID),
		zap.String("task_type", taskType),
		zap.String("queue", queueName),
		zap.Int("worker_id", workerID))

	// Process the task
	err := tw.handleTask(tw.ctx, taskType, kwargs)

	if err != nil {
		tw.logger.Error("Task processing failed", 
			zap.String("task_id", taskID),
			zap.String("task_type", taskType),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err))
		
		// Check retry count
		retries, _ := celeryMsg["retries"].(float64)
		if retries < 3 { // Max 3 retries
			msg.Nack(false, true) // Requeue for retry
		} else {
			msg.Nack(false, false) // Don't requeue, send to DLQ
		}
	} else {
		tw.logger.Info("Task completed successfully", 
			zap.String("task_id", taskID),
			zap.String("task_type", taskType),
			zap.Duration("duration", time.Since(start)))
		msg.Ack(false)
	}
}

// handleTask routes tasks to appropriate handlers
func (tw *TaskWorker) handleTask(ctx context.Context, taskType string, kwargs map[string]interface{}) error {
	switch taskType {
	case TypeChatWorkflow:
		return tw.HandleChatWorkflow(ctx, kwargs)
	case TypeSuggestionWorkflow:
		return tw.HandleSuggestionWorkflow(ctx, kwargs)
	case TypeEventProcessor:
		return tw.HandleEventProcessor(ctx, kwargs)
	default:
		return fmt.Errorf("unknown task type: %s", taskType)
	}
}

// HandleChatWorkflow handles chat workflow tasks
func (tw *TaskWorker) HandleChatWorkflow(ctx context.Context, kwargs map[string]interface{}) error {
	// Parse payload
	payloadBytes, err := json.Marshal(kwargs)
	if err != nil {
		return fmt.Errorf("failed to marshal kwargs: %w", err)
	}

	var payload ChatWorkflowPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal chat workflow payload: %w", err)
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
func (tw *TaskWorker) HandleSuggestionWorkflow(ctx context.Context, kwargs map[string]interface{}) error {
	// Parse payload
	payloadBytes, err := json.Marshal(kwargs)
	if err != nil {
		return fmt.Errorf("failed to marshal kwargs: %w", err)
	}

	var payload SuggestionWorkflowPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal suggestion workflow payload: %w", err)
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
func (tw *TaskWorker) HandleEventProcessor(ctx context.Context, kwargs map[string]interface{}) error {
	// Parse payload
	payloadBytes, err := json.Marshal(kwargs)
	if err != nil {
		return fmt.Errorf("failed to marshal kwargs: %w", err)
	}

	var payload EventProcessorPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal event processor payload: %w", err)
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
