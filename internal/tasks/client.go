package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// TaskClient wraps RabbitMQ connection for task enqueueing
type TaskClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	logger  *zap.Logger
}

// NewTaskClient creates a new task client
func NewTaskClient(rabbitMQURL string, logger *zap.Logger) (*TaskClient, error) {
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	client := &TaskClient{
		conn:    conn,
		channel: channel,
		logger:  logger,
	}

	// Declare queues
	if err := client.declareQueues(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to declare queues: %w", err)
	}

	return client, nil
}

// declareQueues declares all required queues
func (tc *TaskClient) declareQueues() error {
	queues := []string{
		"chat_workflow",
		"events",
		"default",
	}

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
func (tc *TaskClient) Close() error {
	if tc.channel != nil {
		tc.channel.Close()
	}
	if tc.conn != nil {
		return tc.conn.Close()
	}
	return nil
}

// ChatWorkflowPayload represents the payload for chat workflow tasks
type ChatWorkflowPayload struct {
	MessageID      string `json:"message_id"`
	SessionID      string `json:"session_id"`
	SuggestionMode bool   `json:"suggestion_mode,omitempty"`
}

// SuggestionWorkflowPayload represents the payload for suggestion workflow tasks
type SuggestionWorkflowPayload struct {
	MessageID string `json:"message_id"`
	SessionID string `json:"session_id"`
}

// EventProcessorPayload represents the payload for event processor tasks
type EventProcessorPayload struct {
	EventID    string                 `json:"event_id"`
	EventType  string                 `json:"event_type"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	ParentID   string                 `json:"parent_id"`
	Data       map[string]interface{} `json:"data"`
}

// publishTask publishes a task to the specified queue
func (tc *TaskClient) publishTask(ctx context.Context, queueName, taskType string, payload interface{}) error {
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

// EnqueueChatWorkflow enqueues a chat workflow task
func (tc *TaskClient) EnqueueChatWorkflow(ctx context.Context, messageID, sessionID string) error {
	payload := ChatWorkflowPayload{
		MessageID: messageID,
		SessionID: sessionID,
	}

	return tc.publishTask(ctx, "chat_workflow", TypeChatWorkflow, payload)
}

// EnqueueSuggestionWorkflow enqueues a suggestion workflow task
func (tc *TaskClient) EnqueueSuggestionWorkflow(ctx context.Context, messageID, sessionID string) error {
	payload := SuggestionWorkflowPayload{
		MessageID: messageID,
		SessionID: sessionID,
	}

	return tc.publishTask(ctx, "chat_workflow", TypeSuggestionWorkflow, payload)
}

// EnqueueEventProcessor enqueues an event processor task
func (tc *TaskClient) EnqueueEventProcessor(ctx context.Context, eventID, eventType, entityType, entityID string, data map[string]interface{}) error {
	payload := EventProcessorPayload{
		EventID:    eventID,
		EventType:  eventType,
		EntityType: entityType,
		EntityID:   entityID,
		Data:       data,
	}

	return tc.publishTask(ctx, "events", TypeEventProcessor, payload)
}
