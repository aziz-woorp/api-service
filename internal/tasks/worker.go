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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/fraiday-org/api-service/internal/config"
	"github.com/fraiday-org/api-service/internal/models"
	"github.com/fraiday-org/api-service/internal/service"
)

const (
	TypeChatWorkflow         = "chat_workflow"
	TypeSuggestionWorkflow   = "suggestion_workflow"
	TypeEventProcessor       = "event_processor"
	TypeProcessEvent         = "process_event"
	TypeDeliverToProcessor   = "deliver_to_processor"
)

// TaskWorker wraps RabbitMQ connection for task processing
type TaskWorker struct {
	conn                      *amqp.Connection
	channel                   *amqp.Channel
	logger                    *zap.Logger
	aiService                 *service.AIService
	databaseService           *service.DatabaseService
	eventPublisherService     *service.EventPublisherService
	processorDispatchService  *service.ProcessorDispatchService
	payloadService            *service.PayloadService
	chatMessageService        *service.ChatMessageService
	taskClient                *TaskClient
	queues                    []string
	concurrency               int
	cfg                       *config.Config
	wg                        sync.WaitGroup
	ctx                       context.Context
	cancel                    context.CancelFunc
}

// NewTaskWorker creates a new task worker
func NewTaskWorker(rabbitMQURL string, logger *zap.Logger, aiURL, aiToken string, databaseService *service.DatabaseService, eventPublisherService *service.EventPublisherService, payloadService *service.PayloadService, chatMessageService *service.ChatMessageService, cfg *config.Config) (*TaskWorker, error) {
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

	// Initialize AI service
	aiService := service.NewAIService(logger, aiURL, aiToken)
	
	// Initialize ProcessorDispatchService
	processorDispatchService := service.NewProcessorDispatchService(logger, conn)
	
	// Initialize TaskClient for enqueueing tasks
	taskClient, err := NewTaskClient(rabbitMQURL, logger, cfg)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create task client: %w", err)
	}

	return &TaskWorker{
		conn:                     conn,
		channel:                  channel,
		logger:                   logger,
		aiService:                aiService,
		databaseService:          databaseService,
		eventPublisherService:    eventPublisherService,
		processorDispatchService: processorDispatchService,
		payloadService:           payloadService,
		chatMessageService:       chatMessageService,
		taskClient:               taskClient,
		queues:                   []string{cfg.CeleryDefaultQueue, cfg.CeleryEventsQueue, "default"},
		concurrency:              10,
		cfg:                      cfg,
		ctx:                      ctx,
		cancel:                   cancel,
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
		
		// Check retry count and handle exponential backoff for delivery tasks
		retries, _ := celeryMsg["retries"].(float64)
		maxRetries := 3
		
		// For deliver_to_processor tasks, use exponential backoff retry logic
		if taskType == TypeDeliverToProcessor && retries < float64(maxRetries) {
			// Calculate countdown for exponential backoff: 60s, 120s, 240s
			countdown := time.Duration(60 * (1 << int(retries))) * time.Second
			
			tw.logger.Info("Scheduling retry with exponential backoff",
				zap.String("task_id", taskID),
				zap.String("task_type", taskType),
				zap.Int("retry", int(retries)+1),
				zap.Int("max_retries", maxRetries),
				zap.Duration("countdown", countdown))
			
			// For exponential backoff, we need to publish a delayed task
			// Since RabbitMQ doesn't natively support delayed messages, we'll use TTL + DLX
			tw.scheduleRetry(msg, taskType, kwargs, int(retries)+1, countdown)
			msg.Ack(false) // Ack the original message
		} else if retries < float64(maxRetries) {
			msg.Nack(false, true) // Requeue for immediate retry for other task types
		} else {
			tw.logger.Error("All retries exhausted, sending to DLQ",
				zap.String("task_id", taskID),
				zap.String("task_type", taskType),
				zap.Int("retries", int(retries)))
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

// scheduleRetry schedules a task for retry with exponential backoff
func (tw *TaskWorker) scheduleRetry(originalMsg amqp.Delivery, taskType string, kwargs map[string]interface{}, retryCount int, countdown time.Duration) {
	// Create retry message with updated retry count
	message := map[string]interface{}{
		"id":      fmt.Sprintf("%d", time.Now().UnixNano()),
		"task":    taskType,
		"args":    []interface{}{},
		"kwargs":  kwargs,
		"retries": retryCount,
		"eta":     nil,
		"expires": nil,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		tw.logger.Error("Failed to marshal retry message", zap.Error(err))
		return
	}

	// Create a temporary queue with TTL for delayed execution
	delayedQueueName := fmt.Sprintf("events_delayed_%d", time.Now().UnixNano())
	
	// Declare temporary queue with TTL and DLX pointing back to events queue
	_, err = tw.channel.QueueDeclare(
		delayedQueueName,
		false, // not durable (temporary)
		true,  // delete when unused
		false, // not exclusive
		false, // no-wait
		amqp.Table{
			"x-message-ttl":             int64(countdown.Milliseconds()),
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": tw.cfg.CeleryEventsQueue,
		},
	)
	if err != nil {
		tw.logger.Error("Failed to declare delayed queue", zap.Error(err))
		return
	}

	// Publish message to delayed queue
	err = tw.channel.Publish(
		"",               // exchange
		delayedQueueName, // routing key (queue name)
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBytes,
		},
	)
	if err != nil {
		tw.logger.Error("Failed to publish retry message", zap.Error(err))
		return
	}

	tw.logger.Info("Scheduled retry message",
		zap.String("queue", delayedQueueName),
		zap.Duration("delay", countdown),
		zap.Int("retry_count", retryCount))
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
	case TypeProcessEvent:
		return tw.HandleProcessEvent(ctx, kwargs)
	case TypeDeliverToProcessor:
		return tw.HandleDeliverToProcessor(ctx, kwargs)
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
	_, err = tw.eventPublisherService.PublishChatMessageEvent(
		ctx,
		models.EventTypeChatWorkflowProcessing,
		payload.MessageID,
		&payload.SessionID,
		map[string]interface{}{
			"status":     "ai_processing_started",
			"session_id": payload.SessionID,
		},
	)
	if err != nil {
		tw.logger.Error("Failed to publish processing event", zap.Error(err))
		// Don't return error, continue with processing
	}
	
	// 2. Process AI request
	tw.logger.Info("Processing AI request",
		zap.String("message_id", payload.MessageID),
		zap.String("session_id", payload.SessionID))
	
	// Get message content and context from database
	message, err := tw.databaseService.GetChatMessage(ctx, payload.MessageID)
	if err != nil {
		tw.logger.Error("Failed to get message from database", zap.Error(err))
		
		// Publish error event with detailed error information (matching Python)
		_, publishErr := tw.eventPublisherService.PublishChatMessageEvent(
			ctx,
			models.EventTypeChatWorkflowError,
			payload.MessageID,
			&payload.SessionID,
			map[string]interface{}{
				"error":      fmt.Sprintf("%+v", err), // Include stack trace
				"session_id": payload.SessionID,
				"stage":      "message_retrieval",
			},
		)
		if publishErr != nil {
			tw.logger.Error("Failed to publish error event", zap.Error(publishErr))
		}
		
		return fmt.Errorf("failed to get message: %w", err)
	}
	
	sessionContext, err := tw.databaseService.GetSessionContext(ctx, payload.SessionID)
	if err != nil {
		tw.logger.Warn("Failed to get session context, using minimal context", zap.Error(err))
		sessionContext = map[string]interface{}{"session_id": payload.SessionID}
	}
	
	var aiResponse *service.AIResponse
	
	if payload.SuggestionMode {
		aiResponse, err = tw.aiService.GenerateSuggestions(ctx, payload.MessageID, payload.SessionID, message.Text, sessionContext)
	} else {
		aiResponse, err = tw.aiService.GenerateChatResponse(ctx, payload.MessageID, payload.SessionID, message.Text, sessionContext)
	}
	
	if err != nil {
		tw.logger.Error("Failed to process AI request", zap.Error(err))
		
		// Publish error event with detailed error information (matching Python)
		_, publishErr := tw.eventPublisherService.PublishChatMessageEvent(
			ctx,
			models.EventTypeChatWorkflowError,
			payload.MessageID,
			&payload.SessionID,
			map[string]interface{}{
				"error":      fmt.Sprintf("%+v", err), // Include stack trace
				"session_id": payload.SessionID,
				"stage":      "ai_processing",
			},
		)
		if publishErr != nil {
			tw.logger.Error("Failed to publish error event", zap.Error(publishErr))
		}
		
		return fmt.Errorf("AI processing failed: %w", err)
	}
	
	tw.logger.Info("AI response received",
		zap.String("message_id", aiResponse.MessageID),
		zap.String("response_length", fmt.Sprintf("%d", len(aiResponse.Response))))
	
	// Save AI response to database (matching Python implementation)
	// Handle different response formats (Slack/Sunshine vs regular AI service)
	var responseText string
	var confidenceScore float64
	var attachments []models.Attachment
	var closeSession bool
	var answerData interface{}

	if aiResponse.Result != nil {
		// Slack/Sunshine format - data is in Result field
		responseText = aiResponse.Result.Text
		confidenceScore = aiResponse.Result.ConfidenceScore
		attachments = convertAIAttachments(aiResponse.Result.Attachments)
		if aiResponse.Result.Metadata != nil {
			if closeSessionVal, ok := aiResponse.Result.Metadata["close_session"].(bool); ok {
				closeSession = closeSessionVal
			}
		}
		answerData = aiResponse.Result.Data
	} else {
		// Regular AI service format - data is in Data field
		responseText = aiResponse.Data.Answer.AnswerText
		confidenceScore = aiResponse.Data.ConfidenceScore
		attachments = convertAIAttachments(aiResponse.Data.Answer.Attachments)
		closeSession = aiResponse.Metadata.CloseSession
		answerData = aiResponse.Data.Answer.AnswerData
	}

	responseMessage := &models.ChatMessage{
		Text:        responseText,                      // Use extracted text
		Sender:      "fraiday-bot",                    // Add sender field (BOT_SENDER_NAME equivalent)
		SenderName:  "fraiday-bot",                    // Add sender name field
		SenderType:  "assistant",
		SessionID:   message.SessionID,
		Category:    models.MessageCategoryMessage,
		Confidence:  confidenceScore,                   // Use extracted confidence score
		Attachments: attachments,                       // Use converted attachments
		Config: map[string]interface{}{
			"ai_response": true,
			"original_message_id": payload.MessageID,
		},
		Data: map[string]interface{}{
			"close_session": closeSession,
			"meta_data":     answerData, // Add metadata like Python
		},
	}
	
	// Use ChatMessageService to create the message (this will publish chat_message_created event)
	if err := tw.chatMessageService.CreateChatMessage(ctx, responseMessage); err != nil {
		tw.logger.Error("Failed to save AI response to database", zap.Error(err))
		// Don't return error here as the AI processing was successful
	}
	
	// 3. Generate response based on message configuration
	// Check if suggestion mode is enabled
	if payload.SuggestionMode {
		// Create suggestion entity
		tw.logger.Info("Creating chat suggestion",
			zap.String("message_id", payload.MessageID))
		
		// Publish suggestion created event with full payload (matching Python)
		suggestionPayload, err := tw.payloadService.CreateChatSuggestionPayload(ctx, responseMessage.ID.Hex())
		if err != nil {
			tw.logger.Error("Failed to create suggestion payload", zap.Error(err))
			suggestionPayload = map[string]interface{}{
				"id":         responseMessage.ID.Hex(),
				"message_id": payload.MessageID,
				"session_id": payload.SessionID,
				"content":    aiResponse.Response,
			}
		}

		_, err = tw.eventPublisherService.PublishChatSuggestionEvent(
			ctx,
			models.EventTypeChatSuggestionCreated,
			responseMessage.ID.Hex(),
			&payload.MessageID,
			suggestionPayload,
		)
		if err != nil {
			tw.logger.Error("Failed to publish suggestion created event", zap.Error(err))
		}
	} else {
		// Create chat message response
		tw.logger.Info("Creating chat message response",
			zap.String("message_id", payload.MessageID))
		
		// Publish workflow completed event with full message payloads (matching Python)
		userMessagePayload, err := tw.payloadService.CreateChatMessagePayload(ctx, payload.MessageID)
		if err != nil {
			tw.logger.Error("Failed to create user message payload", zap.Error(err))
			userMessagePayload = map[string]interface{}{"id": payload.MessageID}
		}

		aiMessagePayload, err := tw.payloadService.CreateChatMessagePayload(ctx, responseMessage.ID.Hex())
		if err != nil {
			tw.logger.Error("Failed to create AI message payload", zap.Error(err))
			aiMessagePayload = map[string]interface{}{"id": responseMessage.ID.Hex()}
		}

		_, err = tw.eventPublisherService.PublishChatMessageEvent(
			ctx,
			models.EventTypeChatWorkflowCompleted,
			responseMessage.ID.Hex(),
			&payload.SessionID,
			map[string]interface{}{
				"user_message": userMessagePayload,
				"ai_message":   aiMessagePayload,
				"session_id":   payload.SessionID,
			},
		)
		if err != nil {
			tw.logger.Error("Failed to publish workflow completed event", zap.Error(err))
		}
		
		// Check for handover scenario (confidence_score = 0)
		if confidenceScore == 0 {
			// Reuse the payloads we already created for consistency
			_, err = tw.eventPublisherService.PublishChatMessageEvent(
				ctx,
				models.EventTypeChatWorkflowHandover,
				responseMessage.ID.Hex(),
				&payload.SessionID,
				map[string]interface{}{
					"user_message": userMessagePayload,
					"ai_message":   aiMessagePayload,
					"session_id":   payload.SessionID,
				},
			)
			if err != nil {
				tw.logger.Error("Failed to publish handover event", zap.Error(err))
			}
		}
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

	// 1. Fetch message from database
	message, err := tw.databaseService.GetChatMessage(ctx, payload.MessageID)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}

	// 2. Get session context
	sessionContext, err := tw.databaseService.GetSessionContext(ctx, payload.SessionID)
	if err != nil {
		return fmt.Errorf("failed to get session context: %w", err)
	}

	// 3. Generate suggestions using AI service
	aiResponse, err := tw.aiService.GenerateSuggestions(ctx, payload.MessageID, payload.SessionID, message.Text, sessionContext)
	if err != nil {
		return fmt.Errorf("failed to generate suggestions: %w", err)
	}

	// 4. Save AI response as a new message
	suggestionMessage := &service.ChatMessage{
		Text:       aiResponse.Response,
		SenderType: "assistant",
		SessionID:  message.SessionID,
		Category:   models.MessageCategoryMessage,
		Config: map[string]interface{}{
			"suggestion_mode": true,
			"original_message_id": payload.MessageID,
		},
		Data: map[string]interface{}{
			"type":        "suggestion",
			"source":      "ai_service",
			"suggestions": aiResponse.Suggestions,
		},
	}

	if err := tw.databaseService.SaveChatMessage(ctx, suggestionMessage); err != nil {
		return fmt.Errorf("failed to save suggestion message: %w", err)
	}

	tw.logger.Info("Completed suggestion workflow task",
		zap.String("message_id", payload.MessageID),
		zap.String("suggestion_id", aiResponse.MessageID+"_suggestion"))

	return nil
}



// HandleEventProcessor handles event processor tasks
// This mirrors the process_event task from Python backend
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

	// Get the event from the database
	event, err := tw.eventPublisherService.EventService.GetEventByID(ctx, payload.EventID)
	if err != nil {
		return fmt.Errorf("failed to get event: %w", err)
	}

	// Process the event using the existing async processing logic
	if err := tw.eventPublisherService.ProcessEventAsync(ctx, event); err != nil {
		return fmt.Errorf("failed to process event: %w", err)
	}

	tw.logger.Info("Completed event processor task",
		zap.String("event_id", payload.EventID))

	return nil
}

// HandleProcessEvent handles process_event tasks (matching Python logic)
// This mirrors the process_event task from Python backend
func (tw *TaskWorker) HandleProcessEvent(ctx context.Context, kwargs map[string]interface{}) error {
	// Parse payload
	payloadBytes, err := json.Marshal(kwargs)
	if err != nil {
		return fmt.Errorf("failed to marshal kwargs: %w", err)
	}

	var payload ProcessEventPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal process event payload: %w", err)
	}

	tw.logger.Info("Processing process_event task",
		zap.String("event_id", payload.EventID),
		zap.String("event_type", payload.EventType),
		zap.String("entity_type", payload.EntityType),
		zap.String("entity_id", payload.EntityID))

	// Get the original event from database
	event, err := tw.eventPublisherService.EventService.GetEventByID(ctx, payload.EventID)
	if err != nil {
		tw.logger.Error("Event not found", zap.String("event_id", payload.EventID), zap.Error(err))
		return fmt.Errorf("event %s not found: %w", payload.EventID, err)
	}

	// Get client_id from the entity
	clientID, err := tw.getClientIDForEntity(ctx, payload.EntityType, payload.EntityID)
	if err != nil {
		tw.logger.Error("Could not determine client_id", 
			zap.String("entity_type", payload.EntityType),
			zap.String("entity_id", payload.EntityID),
			zap.Error(err))
		return fmt.Errorf("client ID not found: %w", err)
	}

	tw.logger.Info("Resolved client for event", 
		zap.String("event_id", payload.EventID),
		zap.String("client_id", clientID))

	// Convert clientID to ObjectID
	clientObjID, err := primitive.ObjectIDFromHex(clientID)
	if err != nil {
		return fmt.Errorf("invalid client ID format: %w", err)
	}

	// Find matching processors for this event
	processors, err := tw.eventPublisherService.EventProcessorConfigService.GetConfigsForEventAndClient(
		ctx, clientObjID, models.EventType(payload.EventType), models.EntityType(payload.EntityType),
	)
	if err != nil {
		return fmt.Errorf("failed to get matching processors: %w", err)
	}

	if len(processors) == 0 {
		tw.logger.Info("No matching processors found",
			zap.String("event_type", payload.EventType),
			zap.String("client_id", clientID))
		return nil // This is not an error - just skip processing
	}

	// Prepare event data for dispatching (matching Python logic)
	dispatchData := map[string]interface{}{
		"event_id":    payload.EventID,
		"event_type":  payload.EventType,
		"entity_type": payload.EntityType,
		"entity_id":   payload.EntityID,
		"parent_id":   payload.ParentID,
		"data":        payload.Data,
		"timestamp":   event.CreatedAt.Format(time.RFC3339),
		"client_id":   clientID,
	}

	// For each processor, create a delivery record and dispatch in a separate task
	deliveryResults := make([]map[string]interface{}, 0, len(processors))
	
	for _, processor := range processors {
		// Convert event ID to ObjectID
		eventObjID, err := primitive.ObjectIDFromHex(payload.EventID)
		if err != nil {
			tw.logger.Error("Invalid event ID format", 
				zap.String("event_id", payload.EventID), 
				zap.Error(err))
			continue
		}

		// Create delivery record
		delivery, err := tw.eventPublisherService.EventDeliveryTrackingService.CreateDeliveryRecord(
			ctx, eventObjID, processor.ID, dispatchData, 3, // Max 3 retries
		)
		if err != nil {
			tw.logger.Error("Failed to create delivery record", 
				zap.String("processor_id", processor.ID.Hex()), 
				zap.Error(err))
			continue
		}

		// Dispatch to processor in a separate task with retry capability
		err = tw.taskClient.EnqueueDeliverToProcessor(
			ctx,
			processor.ID.Hex(),
			dispatchData,
			delivery.ID.Hex(),
		)
		if err != nil {
			tw.logger.Error("Failed to enqueue delivery task", 
				zap.String("processor_id", processor.ID.Hex()), 
				zap.Error(err))
			continue
		}

		deliveryResults = append(deliveryResults, map[string]interface{}{
			"processor_id":   processor.ID.Hex(),
			"processor_name": processor.Name,
			"delivery_id":    delivery.ID.Hex(),
			"status":         "dispatched",
		})
	}

	tw.logger.Info("Dispatched event to processors", 
		zap.String("event_id", payload.EventID),
		zap.String("client_id", clientID),
		zap.Int("processor_count", len(processors)))

	return nil
}

// HandleDeliverToProcessor handles deliver_to_processor tasks (matching Python logic)
// This mirrors the deliver_to_processor task from Python backend
func (tw *TaskWorker) HandleDeliverToProcessor(ctx context.Context, kwargs map[string]interface{}) error {
	// Parse payload
	payloadBytes, err := json.Marshal(kwargs)
	if err != nil {
		return fmt.Errorf("failed to marshal kwargs: %w", err)
	}

	var payload DeliverToProcessorPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal deliver to processor payload: %w", err)
	}

	tw.logger.Info("Processing deliver_to_processor task",
		zap.String("processor_id", payload.ProcessorID),
		zap.String("delivery_id", payload.DeliveryID))

	// Get the processor
	processor, err := tw.eventPublisherService.EventProcessorConfigService.GetProcessorByID(ctx, payload.ProcessorID)
	if err != nil {
		tw.logger.Error("Processor not found", 
			zap.String("processor_id", payload.ProcessorID), 
			zap.Error(err))

		// Record failure
		_, recordErr := tw.eventPublisherService.EventDeliveryTrackingService.RecordAttempt(
			ctx,
			payload.DeliveryID,
			models.AttemptStatusFailure,
			0,
			"",
			map[string]interface{}{"error": fmt.Sprintf("Processor %s not found", payload.ProcessorID)},
		)
		if recordErr != nil {
			tw.logger.Error("Failed to record attempt", zap.Error(recordErr))
		}

		return fmt.Errorf("processor not found: %w", err)
	}

	// Try to dispatch
	result := tw.processorDispatchService.DispatchToProcessor(ctx, processor, payload.EventData)

	// Record the attempt
	attempt, err := tw.eventPublisherService.EventDeliveryTrackingService.RecordAttempt(
		ctx,
		payload.DeliveryID,
		func() models.AttemptStatus {
			if result.Success {
				return models.AttemptStatusSuccess
			}
			return models.AttemptStatusFailure
		}(),
		result.ResponseStatus,
		result.ResponseBody,
		map[string]interface{}{"error": result.ErrorMessage},
	)
	if err != nil {
		tw.logger.Error("Failed to record delivery attempt", zap.Error(err))
		// Continue processing even if we can't record the attempt
	}

	if result.Success {
		tw.logger.Info("Successfully delivered to processor",
			zap.String("processor_id", payload.ProcessorID),
			zap.String("delivery_id", payload.DeliveryID),
			zap.Int("attempt", int(attempt.AttemptNumber)))
		return nil
	}

	// If delivery failed, return error to trigger retry mechanism
	tw.logger.Error("Failed to deliver to processor",
		zap.String("processor_id", payload.ProcessorID),
		zap.String("delivery_id", payload.DeliveryID),
		zap.String("error", result.ErrorMessage),
		zap.Int("response_status", result.ResponseStatus))

	return fmt.Errorf("delivery failed: %s", result.ErrorMessage)
}

// getClientIDForEntity determines client_id for different entity types
// This mirrors the _get_client_id_for_entity function from Python backend
func (tw *TaskWorker) getClientIDForEntity(ctx context.Context, entityType, entityID string) (string, error) {
	switch entityType {
	case string(models.EntityTypeChatMessage):
		// Get message and then get session to find client
		message, err := tw.databaseService.GetChatMessage(ctx, entityID)
		if err != nil {
			tw.logger.Error("Failed to get chat message for client resolution", 
				zap.String("entity_id", entityID), zap.Error(err))
			return "", fmt.Errorf("failed to get chat message: %w", err)
		}
		
		// Get session to find client
		session, err := tw.databaseService.GetChatSessionByID(ctx, message.SessionID.Hex())
		if err != nil {
			tw.logger.Error("Failed to get chat session for client resolution", 
				zap.String("session_id", message.SessionID.Hex()), zap.Error(err))
			return "", fmt.Errorf("failed to get chat session: %w", err)
		}
		
		if session.Client == nil {
			return "", fmt.Errorf("chat session has no client associated")
		}
		
		return session.Client.Hex(), nil

	case string(models.EntityTypeChatSession):
		session, err := tw.databaseService.GetChatSessionByID(ctx, entityID)
		if err != nil {
			tw.logger.Error("Failed to get chat session for client resolution", 
				zap.String("entity_id", entityID), zap.Error(err))
			return "", fmt.Errorf("failed to get chat session: %w", err)
		}
		
		if session.Client == nil {
			return "", fmt.Errorf("chat session has no client associated")
		}
		
		return session.Client.Hex(), nil

	case string(models.EntityTypeChatSuggestion):
		// Get suggestion and then get session to find client
		// For now, we'll need to implement a GetChatSuggestion method in DatabaseService
		// This is a placeholder that returns an error for now
		return "", fmt.Errorf("chat suggestion client resolution not yet implemented")

	case string(models.EntityTypeAIService):
		// For AI service events, we need to find the related chat message via parent_id
		// This assumes parent_id points to a chat message
		events, err := tw.eventPublisherService.EventService.GetEntityEvents(ctx, models.EntityTypeAIService, entityID)
		if err != nil {
			return "", fmt.Errorf("failed to get AI service events: %w", err)
		}
		
		if len(events) > 0 && events[0].ParentID != "" {
			return tw.getClientIDForEntity(ctx, string(models.EntityTypeChatMessage), events[0].ParentID)
		}
		
		return "", fmt.Errorf("could not determine client_id for AI service entity")

	default:
		return "", fmt.Errorf("unsupported entity type: %s", entityType)
	}
}

// convertAIAttachments converts AI service attachments to ChatMessage attachments
func convertAIAttachments(aiAttachments []service.AIAttachment) []models.Attachment {
	if len(aiAttachments) == 0 {
		return nil
	}

	attachments := make([]models.Attachment, len(aiAttachments))
	for i, aiAttachment := range aiAttachments {
		attachment := models.Attachment{
			Type:     aiAttachment.Type,
			FileName: aiAttachment.FileName,
			FileURL:  aiAttachment.FileURL,
			FileType: aiAttachment.FileType,
		}

		// Handle carousel data if present
		if aiAttachment.Type == "carousel" && len(aiAttachment.Carousel.Items) > 0 {
			// Convert carousel data to map[string]interface{} to match expected format
			carouselData := make(map[string]interface{})
			
			// Convert carousel items to interface{} slice
			items := make([]interface{}, len(aiAttachment.Carousel.Items))
			for j, item := range aiAttachment.Carousel.Items {
				itemData := map[string]interface{}{
					"title":       item.Title,
					"description": item.Description,
				}
				
				// Add optional fields if present
				if item.MediaURL != "" {
					itemData["media_url"] = item.MediaURL
				}
				if item.MediaType != "" {
					itemData["media_type"] = item.MediaType
				}
				if item.DefaultActionURL != "" {
					itemData["default_action_url"] = item.DefaultActionURL
				}
				if len(item.Buttons) > 0 {
					itemData["buttons"] = item.Buttons
				}
				
				items[j] = itemData
			}
			
			carouselData["items"] = items
			attachment.Carousel = carouselData
		}

		// Handle top-level buttons if present (for non-carousel attachments)
		if len(aiAttachment.Buttons) > 0 && aiAttachment.Type != "carousel" {
			if attachment.Carousel == nil {
				attachment.Carousel = make(map[string]interface{})
			}
			attachment.Carousel["buttons"] = aiAttachment.Buttons
		}

		attachments[i] = attachment
	}

	return attachments
}

