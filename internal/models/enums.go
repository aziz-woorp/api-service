package models

// DatabaseType represents the type of database
type DatabaseType string

const (
	DatabaseTypeClickhouse DatabaseType = "clickhouse"
	DatabaseTypePostgres   DatabaseType = "postgres"
	DatabaseTypeQdrant     DatabaseType = "qdrant"
	DatabaseTypeWeaviate   DatabaseType = "weaviate"
)

// EngineType represents the type of engine
type EngineType string

const (
	EngineTypeStructured   EngineType = "structured"
	EngineTypeUnstructured EngineType = "unstructured"
)

// ExecutionStatus represents the status of execution
type ExecutionStatus string

const (
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
)

// ChannelType represents the type of client channel
type ChannelType string

const (
	ChannelTypeWebhook  ChannelType = "webhook"
	ChannelTypeSlack    ChannelType = "slack"
	ChannelTypeSunshine ChannelType = "sunshine"
)

// EventType represents the type of system event
type EventType string

const (
	// Chat Session Events
	EventTypeChatSessionCreated  EventType = "chat_session_created"
	EventTypeChatSessionInactive EventType = "chat_session_inactive"

	// Chat Message Events
	EventTypeChatMessageCreated EventType = "chat_message_created"

	// Chat Workflow Events
	EventTypeChatWorkflowProcessing EventType = "chat_workflow_processing"
	EventTypeChatWorkflowCompleted  EventType = "chat_workflow_completed"
	EventTypeChatWorkflowError      EventType = "chat_workflow_error"
	EventTypeChatWorkflowHandover   EventType = "chat_workflow_handover"

	// Chat Message Suggestion Events
	EventTypeChatSuggestionCreated EventType = "chat_suggestion_created"

	// AI Service Events
	EventTypeAIRequestSent     EventType = "ai_request_sent"
	EventTypeAIResponseReceived EventType = "ai_response_received"

	// CSAT Events
	EventTypeCSATTriggered    EventType = "csat.triggered"
	EventTypeCSATMessageSent  EventType = "csat.message.sent"
	EventTypeCSATCompleted    EventType = "csat.completed"
)

// EntityType represents the type of entity in events
type EntityType string

const (
	EntityTypeChatSession   EntityType = "chat_session"
	EntityTypeChatMessage   EntityType = "chat_message"
	EntityTypeChatSuggestion EntityType = "chat_suggestion"
	EntityTypeAIService     EntityType = "ai_service"
	EntityTypeCSATSession   EntityType = "csat_session"
	EntityTypeCSATQuestion  EntityType = "csat_question"
	EntityTypeCSATResponse  EntityType = "csat_response"
)

// DeliveryStatus represents the status of event delivery
type DeliveryStatus string

const (
	DeliveryStatusPending    DeliveryStatus = "pending"
	DeliveryStatusInProgress DeliveryStatus = "in_progress"
	DeliveryStatusCompleted  DeliveryStatus = "completed"
	DeliveryStatusFailed     DeliveryStatus = "failed"
)

// ProcessorType represents the type of event processor
type ProcessorType string

const (
	ProcessorTypeHTTPWebhook ProcessorType = "http_webhook"
	ProcessorTypeAMQP        ProcessorType = "amqp"
)

// AttemptStatus represents the status of a delivery attempt
type AttemptStatus string

const (
	AttemptStatusSuccess AttemptStatus = "success"
	AttemptStatusFailure AttemptStatus = "failure"
)