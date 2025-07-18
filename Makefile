APP_NAME=genie
APP_PORT?=8080
ENV_FILE?=env/.env.dev
PROFILE?=dev

.PHONY: help build run docker-build docker-up docker-down clean

help:
	@echo "Usage:"
	@echo "  make build             Build the Go binary for API"
	@echo "  make run               Run the API locally"
	@echo "  make docker-build      Build the Docker image for API"
	@echo "  make docker-up         Start all services with Docker Compose"
	@echo "  make docker-down       Stop all services"
	@echo "  make clean             Remove build artifacts"
	@echo "  make build-worker      Build the Go binary for worker (future)"
	@echo "  make run-worker        Run the worker locally (future)"

build:
	go build -o bin/$(APP_NAME) ./cmd/api/main.go

run:
	APP_PORT=$(APP_PORT) APP_ENV=development GIN_MODE=debug LOG_LEVEL=INFO MONGO_URI=mongodb://localhost:27017/api_service_dev go run ./cmd/api/main.go

docker-build:
	docker build --build-arg SERVICE_NAME=$(APP_NAME) --build-arg ENV_FILE=$(ENV_FILE) -t $(APP_NAME):latest .

docker-up:
	ENV_FILE=$(ENV_FILE) APP_PORT=$(APP_PORT) docker compose --profile $(PROFILE) up --build

docker-down:
	docker compose down

clean:
	rm -rf bin/

# Worker service (future microservice example)
build-worker:
	go build -o bin/worker-service ./cmd/worker/main.go

run-worker:
	go run ./cmd/worker/main.go
