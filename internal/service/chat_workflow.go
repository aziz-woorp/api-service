// Package service provides background workflow triggers for chat messages.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"github.com/fraiday-org/api-service/internal/config"
)

// Simple task client to avoid circular imports
type simpleTaskClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	logger  *zap.Logger
	cfg     *config.Config
}

var (
	taskClient *simpleTaskClient
	taskClientOnce sync.Once
	taskClientLogger *zap.Logger
	taskClientConfig *config.Config
)

// initTaskClient initializes the global task client once
func initTaskClient() {
	if taskClientLogger == nil {
		// Use a minimal logger if none provided
		logger, _ := zap.NewProduction()
		taskClientLogger = logger
	}

	if taskClientConfig == nil {
		// Load config if not provided
		taskClientConfig = config.LoadConfig()
	}

	// Get RabbitMQ URL from config
	rabbitMQURL := taskClientConfig.GetRabbitMQURL()
	
	client, err := newSimpleTaskClient(rabbitMQURL, taskClientLogger, taskClientConfig)
	if err != nil {
		taskClientLogger.Error("Failed to create task client", 
			zap.Error(err),
			zap.String("rabbitmq_url", rabbitMQURL))
		return
	}
	taskClient = client
	taskClientLogger.Info("Task client initialized successfully", 
		zap.String("rabbitmq_url", rabbitMQURL))
}

// newSimpleTaskClient creates a simple task client
func newSimpleTaskClient(rabbitMQURL string, logger *zap.Logger, cfg *config.Config) (*simpleTaskClient, error) {
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	client := &simpleTaskClient{
		conn:    conn,
		channel: channel,
		logger:  logger,
		cfg:     cfg,
	}

	// Declare queues
	if err := client.declareQueues(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to declare queues: %w", err)
	}

	return client, nil
}

// declareQueues declares all required queues
func (tc *simpleTaskClient) declareQueues() error {
	queues := []string{tc.cfg.CeleryDefaultQueue, tc.cfg.CeleryEventsQueue, "default"}

	for _, queue := range queues {
		_, err := tc.channel.QueueDeclare(
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

// Close closes the task client
func (tc *simpleTaskClient) Close() error {
	if tc.channel != nil {
		tc.channel.Close()
	}
	if tc.conn != nil {
		return tc.conn.Close()
	}
	return nil
}

// publishTask publishes a task to the specified queue
func (tc *simpleTaskClient) publishTask(ctx context.Context, queueName, taskType string, payload interface{}) error {
	// Create message with Celery-compatible format
	message := map[string]interface{}{
		"id":      fmt.Sprintf("%d", time.Now().UnixNano()),
		"task":    taskType,
		"args":    []interface{}{},
		"kwargs":  payload,
		"retries": 0,
		"eta":     nil,
		"expires": nil,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = tc.channel.PublishWithContext(
		ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // make message persistent
			Body:         messageBytes,
			Headers: amqp.Table{
				"task": taskType,
				"id":   message["id"],
			},
		},
	)

	if err != nil {
		tc.logger.Error("Failed to publish task", 
			zap.String("queue", queueName),
			zap.String("task_type", taskType),
			zap.Error(err))
		return fmt.Errorf("failed to publish task: %w", err)
	}

	tc.logger.Info("Published task",
		zap.String("task_id", message["id"].(string)),
		zap.String("queue", queueName),
		zap.String("task_type", taskType))

	return nil
}

// InitializeWorkflowService initializes the workflow service with config and logger
func InitializeWorkflowService(cfg *config.Config, logger *zap.Logger) {
	taskClientConfig = cfg
	taskClientLogger = logger
}

// TriggerChatWorkflow triggers an AI chat workflow in the background via RabbitMQ.
func TriggerChatWorkflow(ctx context.Context, messageID string, sessionID string) {
	taskClientOnce.Do(initTaskClient)
	
	if taskClient == nil {
		if taskClientLogger != nil {
			taskClientLogger.Error("Task client not initialized, cannot trigger chat workflow",
				zap.String("message_id", messageID),
				zap.String("session_id", sessionID))
		}
		return
	}

	go func() {
		payload := map[string]interface{}{
			"message_id": messageID,
			"session_id": sessionID,
		}
		
		if err := taskClient.publishTask(ctx, taskClient.cfg.CeleryDefaultQueue, "chat_workflow", payload); err != nil {
			taskClientLogger.Error("Failed to enqueue chat workflow task", 
				zap.String("message_id", messageID),
				zap.String("session_id", sessionID),
				zap.Error(err))
		} else {
			taskClientLogger.Info("Successfully enqueued chat workflow task",
				zap.String("message_id", messageID),
				zap.String("session_id", sessionID))
		}
	}()
}

// TriggerSuggestionWorkflow triggers a suggestion workflow in the background via RabbitMQ.
func TriggerSuggestionWorkflow(ctx context.Context, messageID string, sessionID string) {
	taskClientOnce.Do(initTaskClient)
	
	if taskClient == nil {
		if taskClientLogger != nil {
			taskClientLogger.Error("Task client not initialized, cannot trigger suggestion workflow",
				zap.String("message_id", messageID),
				zap.String("session_id", sessionID))
		}
		return
	}

	go func() {
		payload := map[string]interface{}{
			"message_id": messageID,
			"session_id": sessionID,
		}
		
		if err := taskClient.publishTask(ctx, taskClient.cfg.CeleryDefaultQueue, "suggestion_workflow", payload); err != nil {
			taskClientLogger.Error("Failed to enqueue suggestion workflow task",
				zap.String("message_id", messageID), 
				zap.String("session_id", sessionID),
				zap.Error(err))
		} else {
			taskClientLogger.Info("Successfully enqueued suggestion workflow task",
				zap.String("message_id", messageID),
				zap.String("session_id", sessionID))
		}
	}()
}
