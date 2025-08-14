package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// TaskWorker wraps asynq.Server for task processing
type TaskWorker struct {
	server *asynq.Server
	mux    *asynq.ServeMux
	logger *zap.Logger
}

// NewTaskWorker creates a new task worker
func NewTaskWorker(redisAddr string, logger *zap.Logger) *TaskWorker {
	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"chat_workflow": 6,
				"events":        3,
				"default":       1,
			},
			RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
				return time.Duration(n) * time.Second
			},
		},
	)

	mux := asynq.NewServeMux()

	return &TaskWorker{
		server: server,
		mux:    mux,
		logger: logger,
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

	// TODO: Implement actual chat workflow logic
	// This would include:
	// 1. Fetch message from database
	// 2. Call AI service
	// 3. Generate response
	// 4. Save response to database
	// 5. Send webhook notifications

	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

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

	// TODO: Implement actual event processing logic
	// This would include:
	// 1. Validate event
	// 2. Process based on event type
	// 3. Update database
	// 4. Send notifications
	// 5. Trigger downstream events

	// Simulate processing time
	time.Sleep(25 * time.Millisecond)

	tw.logger.Info("Completed event processor task",
		zap.String("event_id", payload.EventID))

	return nil
}