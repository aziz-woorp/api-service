# CSAT API Documentation

## Overview

The CSAT (Customer Satisfaction) API provides functionality to trigger and manage customer satisfaction surveys within chat sessions. The API supports **multiple CSAT configurations per client-channel** differentiated by **type**, allowing for different survey types like `ai_bot`, `post_chat`, `issue_resolution`, etc.

## Key Features

- **Multi-CSAT Configuration**: Support for multiple CSAT types per client-channel combination
- **Type-Specific Questions**: Each CSAT type has its own set of question templates
- **Type Validation**: Enforces snake_case naming convention (lowercase, underscores, no spaces)
- **Simplified Trigger API**: Requires `session_id` and `type`
- **Flexible Session ID Format**: Supports both simple and thread-appended session IDs
- **Automatic Resolution**: Resolves client and channel from session
- **Thread-aware**: Automatically detects and uses active threads when available
- **Event-driven**: Publishes events for integration with external systems
- **Postback buttons**: Questions sent with interactive buttons containing CSAT payloads

## Session ID Formats

The CSAT API supports flexible session ID formats to handle both simple sessions and threaded conversations:

### Simple Session ID
```
session_123
```
- Looks up chat session with exact or prefix match
- Uses main session context

### Thread-Appended Session ID  
```
session_123_thread_456
```
- Parses to extract: base=`session_123`, thread=`thread_456`
- Looks up chat session using base session ID
- Uses specified thread context directly
- Bypasses automatic thread detection

### Session Lookup Behavior
1. **Exact Match**: First attempts exact session_id match
2. **Prefix Match**: If exact fails, uses startsWith query (`^session_id`)
3. **Most Recent**: If multiple matches, returns most recently updated session

## API Endpoints

### 1. Trigger CSAT Survey

Triggers a CSAT survey for a chat session with a specific CSAT type.

**Endpoint:** `POST /api/v1/csat/trigger`

**Request Body:**
```json
{
  "session_id": "external_session_123",
  "type": "ai_bot"
}
```

**Parameters:**
- `session_id` (required): External chat session identifier
- `type` (required): CSAT configuration type (must be snake_case: lowercase letters, numbers, underscores only)

**Response:**
```json
{
  "csat_session_id": "507f1f77bcf86cd799439011",
  "status": "pending",
  "triggered_at": "2024-01-15T10:30:00Z",
  "message": "CSAT survey triggered successfully"
}
```

**Resolution Logic:**
1. **Type Validation**: Validates CSAT type format (snake_case: lowercase, numbers, underscores only)
2. **Session ID Parsing**: Parses session_id to extract base session and potential thread info
   - `session_123` → base: `session_123`, thread: none
   - `session_123_thread_456` → base: `session_123`, thread: `thread_456`
3. **Session Lookup**: Finds chat session using base session_id (supports startsWith matching)
4. **Client/Channel Extraction**: Extracts client and channel from chat session
5. **Configuration Lookup**: Gets type-specific CSAT configuration for client+channel+type
6. **Threading Determination**: 
   - If session_id contains thread info, use that directly
   - Otherwise, check for active threads (30-minute inactivity threshold)
   - Fall back to main session if no threading available
7. **Context Selection**: Uses appropriate session context (thread or main)
8. **CSAT Creation**: Creates CSAT session with resolved context and configuration reference

### 2. Respond to CSAT

Submit a response to a CSAT question.

**Endpoint:** `POST /api/v1/csat/respond`

**Request Body:**
```json
{
  "session_id": "external_session_123",
  "csat_question_id": "507f1f77bcf86cd799439012",
  "response_value": "5"
}
```

**Response:**
```json
{
  "response_id": "507f1f77bcf86cd799439018",
  "status": "success",
  "message": "Response recorded successfully"
}
```

**Enhanced Response Logic:**
- **New Response**: Creates response and automatically sends next question in sequence
- **Update Response**: Updates existing response value without sending duplicate questions
- **Multiple Calls**: Safe to call multiple times for same question - will update response value
- **Question Flow**: Questions are sent in order based on the `order` field in question templates

### 3. Get CSAT Session

Retrieve details of a CSAT session.

**Endpoint:** `GET /api/v1/csat/sessions/{session_id}`

**Response:**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "chat_session_id": "external_session_123",
  "client_id": "507f1f77bcf86cd799439013",
  "channel_id": "507f1f77bcf86cd799439014",
  "thread_session_id": "thread_session_456",
  "thread_context": true,
  "status": "completed",
  "triggered_at": "2024-01-15T10:30:00Z",
  "completed_at": "2024-01-15T10:35:00Z",
  "current_question_index": 3,
  "questions_sent": [
    "507f1f77bcf86cd799439015",
    "507f1f77bcf86cd799439016",
    "507f1f77bcf86cd799439017"
  ],
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:35:00Z"
}
```

## Multi-CSAT Configuration Endpoints

### List All CSAT Configurations

Retrieves all CSAT configurations for a client and channel.

**Endpoint:** `GET /api/v1/clients/{client_id}/channels/{channel_id}/csat/configs`

**Response:**
```json
{
  "configurations": [
    {
      "id": "507f1f77bcf86cd799439011",
      "client_id": "507f1f77bcf86cd799439013",
      "channel_id": "507f1f77bcf86cd799439014",
      "type": "ai_bot",
      "enabled": true,
      "trigger_conditions": {
        "trigger_after": "conversation_end"
      },
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": "507f1f77bcf86cd799439012",
      "client_id": "507f1f77bcf86cd799439013",
      "channel_id": "507f1f77bcf86cd799439014",
      "type": "post_chat",
      "enabled": true,
      "trigger_conditions": {
        "trigger_after": "agent_handover"
      },
      "created_at": "2024-01-15T10:35:00Z",
      "updated_at": "2024-01-15T10:35:00Z"
    }
  ]
}
```

### Create CSAT Configuration

Creates a new CSAT configuration with type specified in request body.

**Endpoint:** `POST /api/v1/clients/{client_id}/channels/{channel_id}/csat/configs`

**Request Body:**
```json
{
  "type": "ai_bot",
  "enabled": true,
  "trigger_conditions": {
    "trigger_after": "conversation_end",
    "min_messages": 3
  }
}
```

**Response:**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "client_id": "507f1f77bcf86cd799439013",
  "channel_id": "507f1f77bcf86cd799439014",
  "type": "ai_bot",
  "enabled": true,
  "trigger_conditions": {
    "trigger_after": "conversation_end",
    "min_messages": 3
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### Get CSAT Configuration by Type

Retrieves a specific CSAT configuration by type.

**Endpoint:** `GET /api/v1/clients/{client_id}/channels/{channel_id}/csat/configs/{type}`

**Response:**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "client_id": "507f1f77bcf86cd799439013",
  "channel_id": "507f1f77bcf86cd799439014",
  "type": "ai_bot",
  "enabled": true,
  "trigger_conditions": {
    "trigger_after": "conversation_end"
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### Update CSAT Configuration by Type

Updates a specific CSAT configuration by type.

**Endpoint:** `PUT /api/v1/clients/{client_id}/channels/{channel_id}/csat/configs/{type}`

**Request Body:**
```json
{
  "type": "ai_bot",
  "enabled": false,
  "trigger_conditions": {
    "trigger_after": "conversation_end",
    "min_messages": 5
  }
}
```

**Response:**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "client_id": "507f1f77bcf86cd799439013",
  "channel_id": "507f1f77bcf86cd799439014",
  "type": "ai_bot",
  "enabled": false,
  "trigger_conditions": {
    "trigger_after": "conversation_end",
    "min_messages": 5
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:35:00Z"
}
```

### Delete CSAT Configuration by Type

Deletes a specific CSAT configuration by type.

**Endpoint:** `DELETE /api/v1/clients/{client_id}/channels/{channel_id}/csat/configs/{type}`

**Response:**
```json
{
  "message": "CSAT configuration for type 'ai_bot' deleted successfully"
}
```

## Type-Specific Question Management

### Get CSAT Questions by Type

Retrieves all CSAT questions for a specific configuration type.

**Endpoint:** `GET /api/v1/clients/{client_id}/channels/{channel_id}/csat/configs/{type}/questions`

**Response:**
```json
{
  "questions": [
    {
      "id": "507f1f77bcf86cd799439015",
      "csat_configuration_id": "507f1f77bcf86cd799439011",
      "question_text": "How satisfied are you with the AI assistance?",
      "options": ["5 - Excellent", "4 - Good", "3 - Average", "2 - Poor", "1 - Very Poor"],
      "order": 1,
      "active": true,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": "507f1f77bcf86cd799439016",
      "csat_configuration_id": "507f1f77bcf86cd799439011",
      "question_text": "How likely are you to recommend our AI service?",
      "options": ["10", "9", "8", "7", "6", "5", "4", "3", "2", "1"],
      "order": 2,
      "active": true,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ]
}
```

### Update CSAT Questions by Type

Updates all CSAT questions for a specific configuration type (replaces all existing questions).

**Note**: This operation deletes all existing questions for the configuration and creates new ones. While not atomic, it works with standalone MongoDB instances.

**Endpoint:** `PUT /api/v1/clients/{client_id}/channels/{channel_id}/csat/configs/{type}/questions`

**Request Body:**
```json
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

## Threading Support

The CSAT system automatically handles threaded conversations:

### Thread Detection
- Checks for active threads in the chat session
- Uses 30-minute inactivity threshold
- Selects the most recently active thread

### Thread Context
- **`thread_context: true`**: CSAT triggered within a thread
- **`thread_context: false`**: CSAT triggered in main session
- **`thread_session_id`**: ID of the specific thread (if applicable)

### Fallback Behavior
- If no active threads found, uses main session context
- If threading service unavailable, uses main session context
- Graceful degradation ensures CSAT always works

## Error Handling

### Common Error Responses

**Session Not Found:**
```json
{
  "error": "failed to find chat session with session_id external_session_123: not found"
}
```

**Missing Client Information:**
```json
{
  "error": "chat session external_session_123 missing client information"
}
```

**CSAT Not Enabled:**
```json
{
  "error": "CSAT is not enabled for this client and channel"
}
```

**Active Session Exists:**
```json
{
  "error": "CSAT session already active for this chat session"
}
```

## Event System

The CSAT system publishes events for integration:

### Events Published

1. **`csat_triggered`** - When CSAT survey is initiated
2. **`csat_message_sent`** - When each question is sent
3. **`csat_completed`** - When survey is completed

### Event Data Structure

```json
{
  "event_type": "csat_triggered",
  "entity_type": "csat_session",
  "entity_id": "507f1f77bcf86cd799439011",
  "data": {
    "csat_session_id": "507f1f77bcf86cd799439011",
    "chat_session_id": "external_session_123",
    "client_id": "507f1f77bcf86cd799439013",
    "channel_id": "507f1f77bcf86cd799439014",
    "thread_context": true,
    "thread_session_id": "thread_session_456"
  }
}
```

## Button Attachments

### Interactive CSAT Buttons

CSAT questions are sent with interactive postback buttons that contain structured payloads for automatic response processing.

**Button Structure:**
```json
{
  "type": "buttons",
  "buttons": [
    {
      "type": "postback",
      "text": "5 - Excellent",
      "payload": "csat:507f1f77bcf86cd799439012:5"
    },
    {
      "type": "postback",
      "text": "4 - Good",
      "payload": "csat:507f1f77bcf86cd799439012:4"
    },
    {
      "type": "postback",
      "text": "3 - Average",
      "payload": "csat:507f1f77bcf86cd799439012:3"
    }
  ]
}
```

### Payload Format

**Structure:** `csat:<question_id>:<value>`

- **`csat:`** - Prefix identifying this as a CSAT postback
- **`<question_id>`** - CSAT question template ID (ObjectID)
- **`<value>`** - User's selected option value

**Example:** `csat:507f1f77bcf86cd799439012:5`

### Upstream Processing

Upstream services can:
1. **Detect CSAT payloads** by checking for `csat:` prefix
2. **Parse payload** to extract question_id and value
3. **Auto-submit response** by calling:
   ```bash
   POST /api/v1/csat/respond
   {
     "session_id": "extracted_from_message_context",
     "csat_question_id": "507f1f77bcf86cd799439012",
     "response_value": "5"
   }
   ```

## Usage Examples

### Multi-CSAT Configuration Setup

**Create AI Bot CSAT Configuration:**
```bash
curl -X POST http://localhost:8000/api/v1/clients/507f1f77bcf86cd799439013/channels/507f1f77bcf86cd799439014/csat/configs \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "type": "ai_bot",
    "enabled": true,
    "trigger_conditions": {
      "trigger_after": "conversation_end",
      "min_messages": 3
    }
  }'
```

**Create Post-Chat CSAT Configuration:**
```bash
curl -X POST http://localhost:8000/api/v1/clients/507f1f77bcf86cd799439013/channels/507f1f77bcf86cd799439014/csat/configs \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "type": "post_chat",
    "enabled": true,
    "trigger_conditions": {
      "trigger_after": "agent_handover"
    }
  }'
```

**Setup Questions for AI Bot CSAT:**
```bash
curl -X PUT http://localhost:8000/api/v1/clients/507f1f77bcf86cd799439013/channels/507f1f77bcf86cd799439014/csat/configs/ai_bot/questions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "questions": [
      {
        "question_text": "How satisfied are you with the AI assistance?",
        "options": ["5 - Excellent", "4 - Good", "3 - Average", "2 - Poor", "1 - Very Poor"],
        "order": 1,
        "active": true
      },
      {
        "question_text": "How likely are you to recommend our AI service?",
        "options": ["10", "9", "8", "7", "6", "5", "4", "3", "2", "1"],
        "order": 2,
        "active": true
      }
    ]
  }'
```

### CSAT Survey Triggering

**Trigger AI Bot CSAT:**
```bash
curl -X POST http://localhost:8000/api/v1/csat/trigger \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "session_id": "my_external_session_123",
    "type": "ai_bot"
  }'
```

**Trigger Post-Chat CSAT:**
```bash
curl -X POST http://localhost:8000/api/v1/csat/trigger \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "session_id": "my_external_session_123",
    "type": "post_chat"
  }'
```

**Trigger with Thread-Appended Session ID:**
```bash
curl -X POST http://localhost:8000/api/v1/csat/trigger \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "session_id": "my_external_session_123_thread_456",
    "type": "ai_bot"
  }'
```

### Check Session Status

```bash
curl -X GET http://localhost:8000/api/v1/csat/sessions/507f1f77bcf86cd799439011 \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### Submit Response

```bash
curl -X POST http://localhost:8000/api/v1/csat/respond \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "csat_session_id": "507f1f77bcf86cd799439011",
    "question_id": "507f1f77bcf86cd799439015",
    "response_value": "5"
  }'
```

## Data Models

### CSAT Configuration

```go
type CSATConfiguration struct {
    ID                string                 `json:"id"`
    ClientID          string                 `json:"client_id"`
    ChannelID         string                 `json:"channel_id"`
    Type              string                 `json:"type"`
    Enabled           bool                   `json:"enabled"`
    TriggerConditions map[string]interface{} `json:"trigger_conditions,omitempty"`
    CreatedAt         time.Time              `json:"created_at"`
    UpdatedAt         time.Time              `json:"updated_at"`
}
```

### CSAT Question Template

```go
type CSATQuestionTemplate struct {
    ID                   string    `json:"id"`
    CSATConfigurationID  string    `json:"csat_configuration_id"`
    QuestionText         string    `json:"question_text"`
    Options              []string  `json:"options"`
    Order                int       `json:"order"`
    Active               bool      `json:"active"`
    CreatedAt            time.Time `json:"created_at"`
    UpdatedAt            time.Time `json:"updated_at"`
}
```

### CSAT Session

```go
type CSATSession struct {
    ID                   string     `json:"id"`
    ChatSessionID        string     `json:"chat_session_id"`
    CSATConfigurationID  string     `json:"csat_configuration_id"`
    ClientID             string     `json:"client_id"`
    ChannelID            string     `json:"channel_id"`
    ThreadSessionID      *string    `json:"thread_session_id,omitempty"`
    ThreadContext        bool       `json:"thread_context"`
    Status               string     `json:"status"`
    TriggeredAt          time.Time  `json:"triggered_at"`
    CompletedAt          *time.Time `json:"completed_at,omitempty"`
    CurrentQuestionIndex int        `json:"current_question_index"`
    QuestionsSent        []string   `json:"questions_sent"`
    CreatedAt            time.Time  `json:"created_at"`
    UpdatedAt            time.Time  `json:"updated_at"`
}
```

### Status Values

- **`pending`**: Survey created but not started
- **`in_progress`**: User is responding to questions
- **`completed`**: All questions answered
- **`abandoned`**: User stopped responding

## CSAT Type Naming Convention

### Type Format Rules

- **snake_case only**: Lowercase letters, numbers, and underscores
- **No spaces**: Spaces are not allowed in type names
- **No uppercase**: All letters must be lowercase
- **Valid examples**: `ai_bot`, `post_chat`, `issue_resolution`, `agent_performance`
- **Invalid examples**: `AI Bot`, `post-chat`, `PostChat`, `ai bot`

### Common CSAT Types

- **`ai_bot`**: Default type for AI assistance satisfaction
- **`post_chat`**: Post-conversation satisfaction survey
- **`issue_resolution`**: Issue resolution satisfaction
- **`agent_performance`**: Agent performance evaluation
- **`product_feedback`**: Product-specific feedback
- **`support_quality`**: Support quality assessment

## Best Practices

1. **Type Strategy**: Plan your CSAT types based on different touchpoints in your customer journey
2. **Session ID Management**: Ensure external session IDs are unique and consistent
3. **Type Validation**: Always validate CSAT type format before API calls
4. **Configuration Setup**: Create configurations and questions before triggering surveys
5. **Error Handling**: Handle type-specific errors (configuration not found, type disabled)
6. **Event Processing**: Subscribe to CSAT events for integration workflows
7. **Thread Awareness**: Design UI to handle both threaded and non-threaded contexts
8. **Question Management**: Question updates replace all existing questions for a configuration (non-atomic operation)

## Migration Notes

### Breaking Changes from Previous API

- **Type Required**: All trigger requests now require a `type` parameter
- **Multi-Configuration**: Support for multiple CSAT configurations per client-channel
- **New Endpoints**: Configuration management moved to type-specific endpoints
- **Question Association**: Questions now linked to configurations, not client-channel directly
- **Enhanced Data Models**: Added CSATConfigurationID references throughout

### Migration Steps

1. **Update Trigger Calls**: Add `type` parameter to all CSAT trigger requests
2. **Create Configurations**: Set up CSAT configurations for each desired type
3. **Migrate Questions**: Associate existing questions with appropriate configurations
4. **Update Integration**: Handle new event structures and API responses
5. **Test Type Validation**: Ensure all type names follow snake_case convention

### Backward Compatibility

This API introduces **breaking changes** and is **not backward compatible** with previous versions. The previous single-configuration API has been replaced with the new multi-configuration system.