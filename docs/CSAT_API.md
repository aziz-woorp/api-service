# CSAT API Documentation

## Overview

The CSAT (Customer Satisfaction) API provides functionality to trigger and manage customer satisfaction surveys within chat sessions. The API automatically handles session resolution, threading, and client/channel identification.

## Key Features

- **Simplified Trigger API**: Only requires external `session_id`
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

Triggers a CSAT survey for a chat session.

**Endpoint:** `POST /api/v1/csat/trigger`

**Request Body:**
```json
{
  "session_id": "external_session_123"
}
```

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
1. **Session ID Parsing**: Parses session_id to extract base session and potential thread info
   - `session_123` → base: `session_123`, thread: none
   - `session_123_thread_456` → base: `session_123`, thread: `thread_456`
2. **Session Lookup**: Finds chat session using base session_id (supports startsWith matching)
3. **Client/Channel Extraction**: Extracts client and channel from chat session
4. **Threading Determination**: 
   - If session_id contains thread info, use that directly
   - Otherwise, check for active threads (30-minute inactivity threshold)
   - Fall back to main session if no threading available
5. **Context Selection**: Uses appropriate session context (thread or main)
6. **CSAT Creation**: Creates CSAT session with resolved context

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

## Configuration Endpoints

### Get CSAT Configuration

**Endpoint:** `GET /api/v1/clients/{client_id}/channels/{channel_id}/csat/config`

### Update CSAT Configuration

**Endpoint:** `PUT /api/v1/clients/{client_id}/channels/{channel_id}/csat/config`

**Request Body:**
```json
{
  "enabled": true,
  "trigger_conditions": {
    "trigger_after_messages": 10,
    "trigger_on_session_end": true
  }
}
```

### Get CSAT Questions

**Endpoint:** `GET /api/v1/clients/{client_id}/channels/{channel_id}/csat/questions`

### Update CSAT Questions

**Endpoint:** `PUT /api/v1/clients/{client_id}/channels/{channel_id}/csat/questions`

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

### Basic CSAT Trigger

**Simple Session ID:**
```bash
curl -X POST http://localhost:8000/api/v1/csat/trigger \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "session_id": "my_external_session_123"
  }'
```

**Thread-Appended Session ID:**
```bash
curl -X POST http://localhost:8000/api/v1/csat/trigger \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "session_id": "my_external_session_123_thread_456"
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

### CSAT Session

```go
type CSATSession struct {
    ID                   string     `json:"id"`
    ChatSessionID        string     `json:"chat_session_id"`
    Client               string     `json:"client"`
    ClientChannel        string     `json:"client_channel"`
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

## Best Practices

1. **Session ID Management**: Ensure external session IDs are unique and consistent
2. **Error Handling**: Always handle potential errors (session not found, CSAT disabled)
3. **Event Processing**: Subscribe to CSAT events for integration workflows
4. **Thread Awareness**: Design UI to handle both threaded and non-threaded contexts
5. **Configuration**: Set up CSAT configuration and questions before triggering surveys

## Migration Notes

### Changes from Previous API

- **Simplified Request**: No longer requires `client_id` and `channel_id`
- **Automatic Resolution**: Client and channel resolved from session
- **Thread Support**: Added automatic thread detection and context
- **Enhanced Events**: Events now include thread information

### Backward Compatibility

This API is not backward compatible with previous versions. The old API has been completely replaced with the new simplified interface.