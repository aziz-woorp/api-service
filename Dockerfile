# syntax=docker/dockerfile:1

# --- Build Stage ---
ARG SERVICE_NAME=api-service
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git for go mod
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the binary for the specified service
ARG SERVICE_NAME=api-service
RUN CGO_ENABLED=0 GOOS=linux go build -o ${SERVICE_NAME} ./cmd/api/main.go

# --- Final Stage ---
FROM alpine:3.19

WORKDIR /app

# Create non-root user
RUN adduser -D -g '' appuser

# Copy binary from builder
ARG SERVICE_NAME=api-service
COPY --from=builder /app/${SERVICE_NAME} .

# Copy env file (default to .env.dev, can be overridden at build time)
ARG ENV_FILE=env/.env.dev
COPY ${ENV_FILE} .env

# Expose port (default 8080, can be overridden by env)
EXPOSE 8080

USER appuser

# Run the binary
ENTRYPOINT ["./api-service"]
