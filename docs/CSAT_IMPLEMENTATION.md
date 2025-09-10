# CSAT (Customer Satisfaction) Survey System

## Overview

The CSAT system allows you to send sequential customer satisfaction surveys as chat messages, collect user responses, and store CSAT data separately from the main chat history. The system leverages the existing event system for message delivery and follows the current codebase architecture.

## Features

- **Event-driven message delivery** using existing webhook system
- **Sequential question flow** with proper session management
- **Button attachments** for rating and multiple-choice questions
- **Client and channel-specific** configurations
- **Authentication** via existing AuthMiddleware
- **MongoDB collections** with snake_case naming conventions
- **No database pollution** - CSAT messages sent as event payloads, not stored as chat messages

## Architecture

### Models

1. **CSATConfiguration** (`csat_configurations`)
   - Enables/disables CSAT for client/channel combinations
   - Stores trigger conditions

2. **CSATQuestionTemplate** (`csat_question_templates`)
   - Defines survey questions with types (rating, text, multiple_choice)
   - Supports ordered question sequences

3. **CSATSession** (`csat_sessions`)
   - Tracks survey progress for each chat session
   - Manages current question index and completion status

4. **CSATResponse** (`csat_responses`)
   - Stores user responses to individual questions

### Event Types

- `csat_triggered` - CSAT survey initiated (entity_type: `csat_session`)
- `csat_message_sent` - CSAT question sent as structured payload (entity_type: `csat_question`)
- `csat_completed` - CSAT survey completed (entity_type: `csat_session`)

### Event Structure

**CSAT Message Sent Event:**
```json
{
  "event_type": "csat_message_sent",
  "entity_type": "csat_question",
  "entity_id": "question_template_id",
  "data": {
    "csat_session_id": "session_id",
    "question_id": "question_template_id",
    "chat_session_id": "chat_session_id",
    "message_type": "question",
    "chat_message": {
      "id": "temp_generated_id",
      "sender": "system",
      "sender_name": "CSAT Survey",
      "text": "How would you rate your experience?",
      "attachments": [...],
      "data": {...}
    }
  }
}
```

**Benefits:**
- ✅ No database chat message records created
- ✅ Proper entity type routing for task workers
- ✅ Complete chat message structure for upstream processing
- ✅ Prevents session lookup failures

## API Endpoints

### Core CSAT Operations

```http
POST /csat/trigger
POST /csat/respond
GET /csat/sessions/:session_id
```

### Configuration Management

```http
GET /clients/:client_id/channels/:channel_id/csat/config
PUT /clients/:client_id/channels/:channel_id/csat/config
GET /clients/:client_id/channels/:channel_id/csat/questions
PUT /clients/:client_id/channels/:channel_id/csat/questions
```

## Enhanced Response Behavior

### Smart Question Flow Management

The CSAT system implements intelligent response handling to prevent duplicate questions:

**New Response Behavior:**
- Creates new `CSATResponse` record
- Advances `CurrentQuestionIndex` in CSAT session
- Automatically sends next question in sequence (by `order` field)
- Progresses survey flow naturally

**Update Response Behavior:**
- Updates existing `CSATResponse.response_value`
- Does NOT advance question index
- Does NOT send duplicate questions
- Allows users to change answers without spam

**Implementation Logic:**
```go
// Check if response exists
existingResponse := GetBySessionAndQuestion(csatSessionID, questionID)
if existingResponse != nil {
    // UPDATE: Just update response value, no next question
    existingResponse.ResponseValue = newValue
    Update(existingResponse)
    return responseID // Exit early
} else {
    // NEW: Create response AND send next question
    CreateResponse(newResponse)
    AdvanceQuestionIndex()
    SendNextQuestion() // Only for new responses
}
```

**Benefits:**
- ✅ Prevents question spam when users change answers
- ✅ Maintains proper survey progression for new responses
- ✅ Supports multiple calls to same question safely
- ✅ Preserves question ordering by `order` field

## Usage Examples

### 1. Configure CSAT

```json
PUT /clients/{client_id}/channels/{channel_id}/csat/config
{
  "enabled": true,
  "trigger_conditions": {
    "trigger_after_messages": 10,
    "trigger_on_session_end": true
  }
}
```

### 2. Set Up Questions

```json
PUT /clients/{client_id}/channels/{channel_id}/csat/questions
{
  "questions": [
    {
      "question_text": "How would you rate your overall experience?",
      "options": ["5", "4", "3", "2", "1"],
      "order": 1,
      "active": true
    },
    {
      "question_text": "What could we improve?",
      "options": ["Great", "Good", "Average", "Poor"],
      "order": 2,
      "active": true
    }
  ]
}
```

### 3. Trigger Survey

```json
POST /csat/trigger
{
  "chat_session_id": "session_123",
  "client_id": "60d5ec49f1b2c8b1f8e4e123",
  "channel_id": "60d5ec49f1b2c8b1f8e4e456"
}
```

### 4. Submit Response

```json
POST /csat/respond
{
  "csat_session_id": "60d5ec49f1b2c8b1f8e4e789",
  "question_id": "60d5ec49f1b2c8b1f8e4e012",
  "response_value": "5"
}
```

## Message Format

CSAT questions are sent as chat messages with the following structure:

```json
{
  "sender": "system",
  "sender_name": "CSAT Survey",
  "sender_type": "system",
  "text": "How would you rate your overall experience?",
  "attachments": [
    {
      "type": "carousel",
      "carousel": {
        "type": "buttons",
        "buttons": [
          {"text": "5", "value": "5", "type": "button"},
          {"text": "4", "value": "4", "type": "button"},
          {"text": "3", "value": "3", "type": "button"},
          {"text": "2", "value": "2", "type": "button"},
          {"text": "1", "value": "1", "type": "button"}
        ]
      }
    }
  ],
  "category": "info",
  "data": {
    "csat_message": true,
    "csat_session_id": "60d5ec49f1b2c8b1f8e4e789",
    "question_id": "60d5ec49f1b2c8b1f8e4e012",
    "question_type": "rating",
    "options": ["5", "4", "3", "2", "1"]
  }
}
```

## Event Flow

1. **Survey Triggered**
   - `csat.triggered` event published
   - CSAT session created with status "pending"

2. **Questions Sent**
   - For each question: `csat.message.sent` event published
   - Chat message created with question and buttons
   - Session status updated to "in_progress"

3. **Responses Processed**
   - User responses stored in `csat_responses`
   - Next question sent automatically
   - Session `current_question_index` incremented

4. **Survey Completed**
   - `csat.completed` event published
   - Thank you message sent
   - Session status updated to "completed"

## Integration with Existing Systems

### Event System
- Uses existing `EventPublisherService` for message delivery
- Integrates with `event_processor_config` for webhook delivery
- Events processed asynchronously via existing event handlers

### Authentication
- All endpoints protected by existing `AuthMiddleware`
- Client and channel access controlled via existing patterns

### Chat System
- CSAT messages stored in existing `chat_messages` collection
- Messages marked with `csat_message: true` in data field
- Follows existing message structure and attachment format

## Database Collections

All collections follow snake_case naming:

- `csat_configurations`
- `csat_question_templates`
- `csat_sessions`
- `csat_responses`

## Testing

See `examples/csat_usage_example.go` for complete usage examples and workflow demonstrations.

## Next Steps

1. **Unit Tests**: Add comprehensive test coverage
2. **Integration Tests**: Test event flow end-to-end
3. **Analytics**: Add CSAT analytics and reporting
4. **Webhooks**: Verify webhook delivery for CSAT events
5. **UI Components**: Create frontend components for CSAT management
