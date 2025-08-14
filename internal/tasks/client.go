package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// Task types
const (
	TypeChatWorkflow       = "chat:workflow"
	TypeSuggestionWorkflow = "chat:suggestion"
	TypeEventProcessor     = "event:processor"
	TypeSemanticLayer      = "semantic:layer"
)

// TaskClient wraps asynq.Client for task enqueueing
type TaskClient struct {
	client *asynq.Client
	logger *zap.Logger
}

// NewTaskClient creates a new task client
func NewTaskClient(redisAddr string, logger *zap.Logger) *TaskClient {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return &TaskClient{
		client: client,
		logger: logger,
	}
}

// Close closes the task client
func (tc *TaskClient) Close() error {
	return tc.client.Close()
}

// ChatWorkflowPayload represents the payload for chat workflow tasks
type ChatWorkflowPayload struct {
	MessageID string `json:"message_id"`
	SessionID string `json:"session_id"`
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
	Data       map[string]interface{} `json:"data"`
}

// EnqueueChatWorkflow enqueues a chat workflow task
func (tc *TaskClient) EnqueueChatWorkflow(ctx context.Context, messageID, sessionID string) error {
	payload := ChatWorkflowPayload{
		MessageID: messageID,
		SessionID: sessionID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal chat workflow payload: %w", err)
	}

	task := asynq.NewTask(TypeChatWorkflow, payloadBytes)
	info, err := tc.client.Enqueue(task, asynq.Queue("chat_workflow"))
	if err != nil {
		tc.logger.Error("Failed to enqueue chat workflow task", zap.Error(err))
		return err
	}

	tc.logger.Info("Enqueued chat workflow task", 
		zap.String("task_id", info.ID),
		zap.String("message_id", messageID),
		zap.String("session_id", sessionID))
	return nil
}

// EnqueueSuggestionWorkflow enqueues a suggestion workflow task
func (tc *TaskClient) EnqueueSuggestionWorkflow(ctx context.Context, messageID, sessionID string) error {
	payload := SuggestionWorkflowPayload{
		MessageID: messageID,
		SessionID: sessionID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal suggestion workflow payload: %w", err)
	}

	task := asynq.NewTask(TypeSuggestionWorkflow, payloadBytes)
	info, err := tc.client.Enqueue(task, asynq.Queue("chat_workflow"))
	if err != nil {
		tc.logger.Error("Failed to enqueue suggestion workflow task", zap.Error(err))
		return err
	}

	tc.logger.Info("Enqueued suggestion workflow task", 
		zap.String("task_id", info.ID),
		zap.String("message_id", messageID),
		zap.String("session_id", sessionID))
	return nil
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

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event processor payload: %w", err)
	}

	task := asynq.NewTask(TypeEventProcessor, payloadBytes)
	info, err := tc.client.Enqueue(task, asynq.Queue("events"))
	if err != nil {
		tc.logger.Error("Failed to enqueue event processor task", zap.Error(err))
		return err
	}

	tc.logger.Info("Enqueued event processor task", 
		zap.String("task_id", info.ID),
		zap.String("event_id", eventID),
		zap.String("event_type", eventType))
	return nil
}