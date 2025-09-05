APP_NAME=genie
APP_PORT?=8080
ENV_FILE?=env/.env.dev
PROFILE?=dev

.PHONY: help build run docker-build docker-up docker-down clean run-chat-workflow-worker run-events-worker run-default-worker run-with-workers

help:
	@echo "Usage:"
	@echo "  make build                    Build the Go binary for API/Worker"
	@echo "  make run                      Run the API server locally"
	@echo "  make run-chat-workflow-worker Run chat workflow worker locally"
	@echo "  make run-events-worker        Run events worker locally"
	@echo "  make run-default-worker       Run default worker locally"
	@echo "  make run-with-workers         Run API + 2 workers in background"
	@echo "  make docker-build             Build the Docker image for API"
	@echo "  make docker-up                Start all services with Docker Compose"
	@echo "  make docker-down              Stop all services"
	@echo "  make clean                    Remove build artifacts"
	@echo ""
	@echo "All commands load environment variables from .env file"

build:
	go build -o bin/$(APP_NAME) ./cmd/api/main.go

run:
	bash -c 'set -a && source .env && set +a && go run ./cmd/api/main.go'

docker-build:
	docker build --build-arg SERVICE_NAME=$(APP_NAME) --build-arg ENV_FILE=$(ENV_FILE) -t $(APP_NAME):latest .

docker-up:
	ENV_FILE=$(ENV_FILE) APP_PORT=$(APP_PORT) docker compose --profile $(PROFILE) up --build

docker-down:
	docker compose down

clean:
	rm -rf bin/

# Worker services - matching deployment.yaml pattern
run-chat-workflow-worker:
	bash -c 'set -a && source .env && set +a && go run ./cmd/api/main.go -mode=worker -queue=chat_workflow -concurrency=4'

run-events-worker:
	bash -c 'set -a && source .env && set +a && go run ./cmd/api/main.go -mode=worker -queue=events -concurrency=2'

run-default-worker:
	bash -c 'set -a && source .env && set +a && go run ./cmd/api/main.go -mode=worker -queue=default -concurrency=2'

# Run API server and 2 workers (chat-workflow + events) in background
run-with-workers:
	@echo "Starting API server and workers..."
	@echo "Use 'pkill -f \"go run\"' to stop all processes"
	bash -c 'set -a && source .env && set +a && go run ./cmd/api/main.go' &
	bash -c 'set -a && source .env && set +a && go run ./cmd/api/main.go -mode=worker -queue=chat_workflow -concurrency=4' &
	bash -c 'set -a && source .env && set +a && go run ./cmd/api/main.go -mode=worker -queue=events -concurrency=2' &
	wait
