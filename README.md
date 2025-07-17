# API Service

A Go REST API service built with the Gin framework following clean architecture principles.

## Requirements

- Go 1.23+
- Docker (optional)
- Docker Compose (optional)

## Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod tidy
   ```

## Running the Application

### Local Development

```bash
# Using environment variable
APP_PORT=8001 go run cmd/api/main.go

# Or export the variable
export APP_PORT=8001
go run cmd/api/main.go
```

### Using Docker Compose

```bash
# Build and run
docker-compose up --build

# Run in detached mode
docker-compose up -d --build

# Watch mode (auto-rebuild on changes)
docker-compose up --watch

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

## Environment Variables

Configuration is managed through environment variables. Create a `.env` file in the `env/` directory:

```bash
APP_PORT=8001
APP_ENV=development
GIN_MODE=debug
```

### Available Environment Variables

- `APP_PORT`: Server port (default: 8080)
- `APP_ENV`: Application environment (default: development)
- `GIN_MODE`: Gin mode (debug/release/test)
- `LOG_LEVEL`: Log level (default: INFO)
- `APP_TRANSLATIONS_DIR`: Translations directory (default: translation)

## Available Endpoints

- `GET /health` - Detailed health check with system information
- `GET /ping` - Simple health check
- `GET /test/ping` - Test endpoint returning "pong"

## Project Structure

```
cmd/
├── api/
│   └── main.go              # Application entry point
internal/
├── api/                     # API layer
│   ├── handlers/            # HTTP handlers
│   ├── middleware/          # HTTP middleware
│   ├── routes/              # Route definitions
│   ├── dto/                 # Data transfer objects
│   └── config.go            # API server setup
├── service/                 # Business logic layer
├── repository/              # Data access layer
├── models/                  # Data models/entities
├── config/                  # Configuration
│   └── api_config.go        # API configuration
└── errors/                  # Error handling
env/
└── .env                     # Environment variables
docker-compose.yml           # Docker Compose configuration
Dockerfile                   # Docker build configuration
```

## Development Guidelines

### Adding New Features

1. **Handlers**: Create new handlers in `internal/api/handlers/`
2. **Services**: Add business logic in `internal/service/`
3. **Repositories**: Implement data access in `internal/repository/`
4. **Models**: Define data structures in `internal/models/`
5. **Routes**: Register new endpoints in `internal/api/config.go`

### Clean Architecture Principles

- **Separation of Concerns**: Each layer has a specific responsibility
- **Dependency Rule**: Dependencies point inward (handlers → services → repositories)
- **Interface-based Design**: Use interfaces for loose coupling
- **Testability**: Each layer can be tested independently

## Example Health Response

```json
{
  "status": "healthy",
  "time": "2024-01-01T12:00:00Z",
  "version": "1.0.0",
  "uptime": "5m30s",
  "system": {
    "go_version": "go1.23.7",
    "num_cpu": 8,
    "arch": "amd64",
    "os": "linux"
  },
  "checks": {
    "database": "ok",
    "cache": "ok",
    "service": "running"
  }
}
```