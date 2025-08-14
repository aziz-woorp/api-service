# Migration Guide: Python Celery to Go Asynq

This document outlines the migration from Python Celery-based task processing to Go Asynq-based task processing in the Fraiday API Service.

## Overview

The migration replaces the Python Celery worker system with a Go-based Asynq task processing system, providing better performance, type safety, and integration with the existing Go codebase.

## Architecture Changes

### Before (Python Celery)
- **Task Queue**: Redis with Celery
- **Worker Process**: Python-based Celery workers
- **Task Types**: `send_chat_workflow`, `process_event`, `trigger_sync_job`
- **Queues**: `chat_workflow`, `events`, `semantic_layer`

### After (Go Asynq)
- **Task Queue**: Redis with Asynq
- **Worker Process**: Go-based Asynq workers
- **Task Types**: `TypeChatWorkflow`, `TypeSuggestionWorkflow`, `TypeEventProcessor`, `TypeSemanticLayerSync`
- **Queues**: `chat_workflow`, `events`, `default`

## Task Mapping

| Python Celery Task | Go Asynq Task | Handler Function |
|-------------------|---------------|------------------|
| `send_chat_workflow` | `TypeChatWorkflow` | `HandleChatWorkflow` |
| `send_suggestion_workflow` | `TypeSuggestionWorkflow` | `HandleSuggestionWorkflow` |
| `process_event` | `TypeEventProcessor` | `HandleEventProcessor` |
| `trigger_sync_job` | `TypeSemanticLayerSync` | `HandleSemanticLayerSync` |

## New Components

### 1. Task Client (`internal/tasks/client.go`)

Handles task enqueueing with the following methods:
- `EnqueueChatWorkflow(messageID, sessionID string, suggestionMode bool)`
- `EnqueueSuggestionWorkflow(messageID, sessionID string)`
- `EnqueueEventProcessor(eventID, clientID, eventType, entityType string, payload map[string]interface{})`
- `EnqueueSemanticLayerSync(clientID, syncType string)`

### 2. Task Worker (`internal/tasks/worker.go`)

Processes tasks with integrated services:
- **AI Service**: Handles AI processing requests
- **Webhook Service**: Sends HTTP webhook notifications
- **Database Service**: Manages MongoDB operations

### 3. Service Layer

#### AI Service (`internal/service/ai_service.go`)
- Replaces Python AI service integration
- Supports both chat responses and suggestions
- Configurable AI endpoint and authentication

#### Webhook Service (`internal/service/webhook_service.go`)
- Handles HTTP webhook delivery
- Supports multiple webhook types (chat, suggestions, events)
- Retry logic and error handling

#### Database Service (`internal/service/database_service.go`)
- MongoDB operations for chat messages and sessions
- Event delivery record management
- Session context retrieval

## Configuration Changes

### Environment Variables

The following environment variables are now used:

```bash
# AI Service Configuration
AI_SERVICE_URL=https://ai-service.example.com
SLACK_AI_TOKEN=your-ai-token

# Redis Configuration (existing)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_DB=0
REDIS_PASSWORD=

# MongoDB Configuration (existing)
MONGODB_URI=mongodb://localhost:27017/api_service
```

### Deployment Configuration

Updated `deploy-go.yaml` includes:

```yaml
workers:
  chatWorkflow:
    replicaCount: 3
    concurrency: 10
    queues: ["chat_workflow"]
  
  eventProcessor:
    replicaCount: 2
    concurrency: 5
    queues: ["events"]
  
  semanticLayer:
    replicaCount: 1
    concurrency: 3
    queues: ["default"]
```

## Task Payload Structures

### ChatWorkflowPayload
```go
type ChatWorkflowPayload struct {
    MessageID      string `json:"message_id"`
    SessionID      string `json:"session_id"`
    SuggestionMode bool   `json:"suggestion_mode"`
}
```

### EventProcessorPayload
```go
type EventProcessorPayload struct {
    EventID    string                 `json:"event_id"`
    ClientID   string                 `json:"client_id"`
    EventType  string                 `json:"event_type"`
    EntityType string                 `json:"entity_type"`
    Payload    map[string]interface{} `json:"payload"`
}
```

### SemanticLayerSyncPayload
```go
type SemanticLayerSyncPayload struct {
    ClientID string `json:"client_id"`
    SyncType string `json:"sync_type"`
}
```

## Migration Steps

### 1. Code Migration
- ✅ Implemented Go Asynq task system
- ✅ Created service layer for AI, webhook, and database operations
- ✅ Updated deployment configuration
- ✅ Added comprehensive error handling and logging

### 2. Testing
- [ ] Unit tests for task handlers
- [ ] Integration tests with Redis and MongoDB
- [ ] Load testing for task processing

### 3. Deployment
- [ ] Deploy Go workers alongside Python workers
- [ ] Gradually migrate task enqueueing to Go client
- [ ] Monitor task processing metrics
- [ ] Decommission Python Celery workers

## Performance Improvements

### Expected Benefits
1. **Memory Usage**: Reduced memory footprint compared to Python
2. **Startup Time**: Faster worker startup and task processing
3. **Type Safety**: Compile-time type checking for task payloads
4. **Concurrency**: Better handling of concurrent tasks
5. **Monitoring**: Integrated metrics and logging

### Queue Priorities
- `chat_workflow`: Priority 6 (highest)
- `events`: Priority 3 (medium)
- `default`: Priority 2 (lowest)

## Monitoring and Observability

### Logging
Structured logging with Zap includes:
- Task execution start/completion
- AI service request/response times
- Webhook delivery status
- Database operation results
- Error details with context

### Metrics
- Task processing duration
- Queue depth and processing rates
- AI service response times
- Webhook delivery success rates
- Database operation latency

## Troubleshooting

### Common Issues

1. **Task Not Processing**
   - Check Redis connection
   - Verify queue configuration
   - Check worker logs for errors

2. **AI Service Errors**
   - Verify AI_SERVICE_URL configuration
   - Check authentication token
   - Monitor AI service availability

3. **Database Connection Issues**
   - Verify MongoDB URI
   - Check database permissions
   - Monitor connection pool status

4. **Webhook Delivery Failures**
   - Check webhook endpoint availability
   - Verify payload format
   - Monitor retry attempts

### Debug Commands

```bash
# Check task queue status
redis-cli -h localhost -p 6379 LLEN asynq:queues:chat_workflow

# Monitor worker logs
kubectl logs -f deployment/api-service-chat-workflow-worker

# Check database connectivity
mongo mongodb://localhost:27017/api_service --eval "db.runCommand('ping')"
```

## Rollback Plan

If issues arise during migration:

1. **Immediate Rollback**
   - Scale down Go workers
   - Scale up Python Celery workers
   - Redirect task enqueueing to Python system

2. **Data Consistency**
   - Ensure task completion before rollback
   - Verify database state consistency
   - Check for pending tasks in queues

## Future Enhancements

1. **Task Scheduling**: Implement cron-like task scheduling
2. **Dead Letter Queues**: Handle failed tasks with retry limits
3. **Task Priorities**: Fine-grained task priority management
4. **Distributed Tracing**: Add OpenTelemetry integration
5. **Auto-scaling**: Dynamic worker scaling based on queue depth

## References

- [Asynq Documentation](https://github.com/hibiken/asynq)
- [Go MongoDB Driver](https://pkg.go.dev/go.mongodb.org/mongo-driver)
- [Zap Logging](https://pkg.go.dev/go.uber.org/zap)
- [Gin Web Framework](https://gin-gonic.com/)