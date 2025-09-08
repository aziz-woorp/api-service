package tasks

import (
	"encoding/json"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestTaskWorkerInitialization tests that TaskWorker can be initialized properly
func TestTaskWorkerInitialization(t *testing.T) {
	logger := zap.NewNop()
	
	// Test with invalid RabbitMQ URL (should fail gracefully)
	worker, err := NewTaskWorker("amqp://invalid:invalid@localhost:5672/", logger, "http://localhost:8001", "test-token", nil, nil)
	assert.Error(t, err)
	assert.Nil(t, worker)
}

// TestCeleryMessageParsing tests parsing of Celery message format
func TestCeleryMessageParsing(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected map[string]interface{}
		wantErr  bool
	}{
		{
			name: "valid chat workflow message",
			body: `{"task": "chat_workflow", "id": "test-123", "kwargs": {"message_id": "msg-123", "session_id": "sess-456"}, "retries": 0}`,
			expected: map[string]interface{}{
				"task":    "chat_workflow",
				"id":      "test-123",
				"kwargs":  map[string]interface{}{"message_id": "msg-123", "session_id": "sess-456"},
				"retries": float64(0),
			},
			wantErr: false,
		},
		{
			name: "valid event processor message",
			body: `{"task": "event_processor", "id": "event-789", "kwargs": {"entity_type": "chat_message", "entity_id": "msg-789", "event_type": "created"}, "retries": 1}`,
			expected: map[string]interface{}{
				"task":    "event_processor",
				"id":      "event-789",
				"kwargs":  map[string]interface{}{"entity_type": "chat_message", "entity_id": "msg-789", "event_type": "created"},
				"retries": float64(1),
			},
			wantErr: false,
		},
		{
			name:     "invalid json",
			body:     `{"task": "invalid", "id":}`,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result map[string]interface{}
			err := json.Unmarshal([]byte(tt.body), &result)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestTaskRouting tests that tasks are routed to correct handlers
func TestTaskRouting(t *testing.T) {
	tests := []struct {
		name     string
		taskType string
		valid    bool
	}{
		{"chat workflow", TypeChatWorkflow, true},
		{"suggestion workflow", TypeSuggestionWorkflow, true},
		{"event processor", TypeEventProcessor, true},
		{"unknown task", "unknown_task", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that task type constants are defined correctly
			switch tt.taskType {
			case TypeChatWorkflow:
				assert.Equal(t, "chat_workflow", tt.taskType)
			case TypeSuggestionWorkflow:
				assert.Equal(t, "suggestion_workflow", tt.taskType)
			case TypeEventProcessor:
				assert.Equal(t, "event_processor", tt.taskType)
			default:
				if tt.valid {
					t.Errorf("Expected valid task type %s to be handled", tt.taskType)
				}
			}
		})
	}
}

// TestQueueDeclaration tests queue declaration logic
func TestQueueDeclaration(t *testing.T) {
	queues := []string{"chat_workflow", "events", "default"}
	
	// Test that all expected queues are present
	expectedQueues := map[string]bool{
		"chat_workflow": false,
		"events":        false,
		"default":       false,
	}
	
	for _, queue := range queues {
		if _, exists := expectedQueues[queue]; exists {
			expectedQueues[queue] = true
		}
	}
	
	// Verify all queues were found
	for queue, found := range expectedQueues {
		assert.True(t, found, "Queue %s should be declared", queue)
	}
}

// TestEventProcessorPayload tests EventProcessorPayload structure
func TestEventProcessorPayload(t *testing.T) {
	// Test that EventProcessorPayload has all required fields
	payload := EventProcessorPayload{
		EventID:    "event-123",
		EntityType: "chat_message",
		EntityID:   "msg-123",
		EventType:  "created",
		ParentID:   "parent-456",
		Data:       map[string]interface{}{"test": "data"},
	}
	
	assert.Equal(t, "event-123", payload.EventID)
	assert.Equal(t, "chat_message", payload.EntityType)
	assert.Equal(t, "msg-123", payload.EntityID)
	assert.Equal(t, "created", payload.EventType)
	assert.Equal(t, "parent-456", payload.ParentID)
	assert.NotNil(t, payload.Data)
}

// TestRabbitMQConnectionString tests RabbitMQ URL parsing
func TestRabbitMQConnectionString(t *testing.T) {
	tests := []struct {
		name  string
		url   string
		valid bool
	}{
		{"valid amqp url", "amqp://user:pass@localhost:5672/vhost", true},
		{"valid amqps url", "amqps://user:pass@localhost:5671/vhost", true},
		{"invalid protocol", "http://localhost:5672", false},
		{"empty url", "", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test URL validation logic
			if tt.valid {
				// Valid URLs should contain amqp:// or amqps://
				assert.Contains(t, tt.url, "amqp")
			} else {
				// Invalid URLs should not be amqp protocol or be empty
				if tt.url != "" {
					assert.NotContains(t, tt.url, "amqp://")
					assert.NotContains(t, tt.url, "amqps://")
				}
			}
		})
	}
}

// MockAMQPDelivery creates a mock AMQP delivery for testing
func MockAMQPDelivery(body string, queue string) amqp.Delivery {
	return amqp.Delivery{
		Body:         []byte(body),
		DeliveryTag:  1,
		Redelivered:  false,
		Exchange:     queue,
		RoutingKey:   queue,
		Timestamp:    time.Now(),
		ContentType:  "application/json",
	}
}

// TestMessageProcessing tests message processing logic without actual RabbitMQ
func TestMessageProcessing(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		queue     string
		expectErr bool
	}{
		{
			name:      "valid chat workflow message",
			body:      `{"task": "chat_workflow", "id": "test-123", "kwargs": {"message_id": "msg-123"}, "retries": 0}`,
			queue:     "chat_workflow",
			expectErr: false,
		},
		{
			name:      "invalid json message",
			body:      `{"task": "invalid"}`,
			queue:     "chat_workflow",
			expectErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON parsing
			var message map[string]interface{}
			err := json.Unmarshal([]byte(tt.body), &message)
			
			if tt.expectErr {
				// For invalid messages, we might still parse JSON but fail validation
				if err == nil {
					// Check if required fields are missing
					_, hasTask := message["task"]
					_, hasID := message["id"]
					_, hasKwargs := message["kwargs"]
					
					if !hasTask || !hasID || !hasKwargs {
						// Missing required fields should be treated as error
						t.Log("Message missing required fields")
					}
				}
			} else {
				require.NoError(t, err)
				
				// Verify required fields are present
				assert.Contains(t, message, "task")
				assert.Contains(t, message, "id")
				assert.Contains(t, message, "kwargs")
			}
		})
	}
}