# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Build & Run
- `make build` - Build the Go binary for API/Worker
- `make run` - Run the API server locally (loads .env file)
- `make clean` - Remove build artifacts

### Worker Services
- `make run-chat-workflow-worker` - Run chat workflow worker locally
- `make run-events-worker` - Run events worker locally  
- `make run-default-worker` - Run default worker locally
- `make run-with-workers` - Run API + 2 workers in background

### Docker Operations
- `make docker-build` - Build the Docker image for API
- `make docker-up` - Start all services with Docker Compose
- `make docker-down` - Stop all services
- `ENV_FILE=env/.env.dev PROFILE=dev make docker-up` - Start with specific environment

### Testing
- `go test ./...` - Run all tests
- `go test ./internal/tasks/` - Run specific package tests

## Architecture Overview

This is a **Go REST API service** built with the **Gin web framework** following clean architecture principles:

### Core Structure
```
cmd/api/main.go          # Application entrypoint (server & worker modes)
internal/
├── api/                 # API layer (config, handlers, middleware, routes, utils)
├── config/              # Configuration management
├── models/              # Domain models and enums  
├── repository/          # Data access layer (MongoDB)
├── service/             # Business logic layer
└── tasks/               # Background task processing (RabbitMQ)
```

### Key Technologies
- **Web Framework**: Gin with middleware stack (request ID, logging, recovery, CORS)
- **Database**: MongoDB with repository pattern
- **Message Queue**: RabbitMQ for background task processing
- **Logging**: Zap logger (development/production modes)
- **Configuration**: Environment-based with .env files

### Dual Mode Application
The application runs in two modes via command-line flags:
- **Server mode**: `go run ./cmd/api/main.go` (default)
- **Worker mode**: `go run ./cmd/api/main.go -mode=worker -queue=chat_workflow -concurrency=4`

### Database Layer
- Uses repository pattern with MongoDB
- Connection established in main.go and passed to handlers
- Database service wrapper for common operations

### Task Processing
- RabbitMQ-based background task system
- Multiple queue support (chat_workflow, events, default)
- Configurable concurrency per worker

## Environment Configuration

The application uses environment files:
- `.env` - Local development (gitignored)
- `env/.env.dev` - Development template
- `env/.env.prod` - Production template

Key environment variables:
- `APP_PORT` - Server port (default: 8080)
- `MONGODB_URI` - MongoDB connection string with database name in path
- `MONGODB_DB` - Optional: Override database name extracted from URI
- `CELERY_BROKER_URL` / `RABBITMQ_*` - RabbitMQ connection settings
- `GIN_MODE` - Gin framework mode (debug/release)

## API Structure

### Middleware Stack (applied in order)
1. Request ID generation
2. Request/response logging
3. Recovery from panics
4. CORS handling
5. Error handling

### Standard Endpoints
- `GET /health` - Detailed health check with dependencies
- `GET /ping` - Simple health check
- Authentication endpoints under `/auth`
- CSAT (Customer Satisfaction) endpoints under `/csat`

## Development Patterns

### Adding New Features
1. Create models in `internal/models/`
2. Implement repository in `internal/repository/`
3. Add business logic in `internal/service/`
4. Create handlers in `internal/api/handlers/`
5. Add DTOs in `internal/api/dto/`
6. Register routes in `internal/api/routes/`

### Error Handling
- Use structured error responses
- Middleware handles panic recovery
- Centralized logging with request IDs

### Testing
- Test files follow `*_test.go` naming
- Use testify package for assertions
- Tests located alongside source files

## Current State

The repository is currently on branch `feat-csat` implementing Customer Satisfaction functionality. Recent changes include:
- CSAT models, repositories, handlers, and services
- Migration from previous architecture
- Cleanup of deprecated user-related code
- Addition of comprehensive documentation

## Dependencies

Key Go modules:
- `github.com/gin-gonic/gin` - Web framework
- `go.mongodb.org/mongo-driver` - MongoDB driver
- `github.com/rabbitmq/amqp091-go` - RabbitMQ client
- `go.uber.org/zap` - Logging
- `github.com/stretchr/testify` - Testing utilities