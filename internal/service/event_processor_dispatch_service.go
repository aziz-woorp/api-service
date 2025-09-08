// Package service provides business logic for event processor dispatching.
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/fraiday-org/api-service/internal/models"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// ProcessorDispatchResult represents the result of dispatching to a processor
type ProcessorDispatchResult struct {
	Success        bool
	ResponseStatus int
	ResponseBody   string
	ErrorMessage   string
}

// ProcessorDispatchService handles dispatching events to processors
type ProcessorDispatchService struct {
	logger     *zap.Logger
	httpClient *http.Client
	amqpConn   *amqp.Connection
}

// NewProcessorDispatchService creates a new ProcessorDispatchService
func NewProcessorDispatchService(logger *zap.Logger, amqpConn *amqp.Connection) *ProcessorDispatchService {
	return &ProcessorDispatchService{
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		amqpConn: amqpConn,
	}
}

// DispatchToProcessor dispatches event data to a specific processor
// Returns (success, response_status, response_body, error_message) matching Python logic
func (s *ProcessorDispatchService) DispatchToProcessor(
	ctx context.Context,
	processor *models.EventProcessorConfig,
	eventData map[string]interface{},
) ProcessorDispatchResult {
	switch processor.ProcessorType {
	case models.ProcessorTypeHTTPWebhook:
		return s.dispatchToHTTPWebhook(ctx, processor, eventData)
	case models.ProcessorTypeAMQP:
		return s.dispatchToAMQP(ctx, processor, eventData)
	default:
		return ProcessorDispatchResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("unsupported processor type: %s", processor.ProcessorType),
		}
	}
}

// dispatchToHTTPWebhook dispatches event to HTTP webhook endpoint
func (s *ProcessorDispatchService) dispatchToHTTPWebhook(
	ctx context.Context,
	processor *models.EventProcessorConfig,
	eventData map[string]interface{},
) ProcessorDispatchResult {
	// Get webhook configuration
	config := processor.Config
	url, ok := config["webhook_url"].(string)
	if !ok {
		return ProcessorDispatchResult{
			Success:      false,
			ErrorMessage: "webhook URL not configured",
		}
	}

	// Prepare request payload
	payload, err := json.Marshal(eventData)
	if err != nil {
		return ProcessorDispatchResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to marshal payload: %v", err),
		}
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return ProcessorDispatchResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to create request: %v", err),
		}
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Fraiday-Events/1.0")

	// Add authentication if configured
	if auth, exists := config["auth"]; exists {
		if authMap, ok := auth.(map[string]interface{}); ok {
			if authType, ok := authMap["type"].(string); ok {
				switch authType {
				case "bearer":
					if token, ok := authMap["token"].(string); ok {
						req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
					}
				case "basic":
					if username, ok := authMap["username"].(string); ok {
						if password, ok := authMap["password"].(string); ok {
							req.SetBasicAuth(username, password)
						}
					}
				}
			}
		}
	}

	// Add custom headers if configured
	if headers, exists := config["headers"]; exists {
		if headersMap, ok := headers.(map[string]interface{}); ok {
			for key, value := range headersMap {
				if valueStr, ok := value.(string); ok {
					req.Header.Set(key, valueStr)
				}
			}
		}
	}

	s.logger.Debug("Dispatching to HTTP webhook",
		zap.String("url", url),
		zap.String("processor_id", processor.ID.Hex()))

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return ProcessorDispatchResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("HTTP request failed: %v", err),
		}
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ProcessorDispatchResult{
			Success:        false,
			ResponseStatus: resp.StatusCode,
			ErrorMessage:   fmt.Sprintf("failed to read response body: %v", err),
		}
	}

	// Check if request was successful (2xx status codes)
	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	result := ProcessorDispatchResult{
		Success:        success,
		ResponseStatus: resp.StatusCode,
		ResponseBody:   string(body),
	}

	if !success {
		result.ErrorMessage = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	s.logger.Debug("HTTP webhook response",
		zap.String("processor_id", processor.ID.Hex()),
		zap.Int("status", resp.StatusCode),
		zap.Bool("success", success))

	return result
}

// dispatchToAMQP dispatches event to AMQP queue/exchange
func (s *ProcessorDispatchService) dispatchToAMQP(
	ctx context.Context,
	processor *models.EventProcessorConfig,
	eventData map[string]interface{},
) ProcessorDispatchResult {
	if s.amqpConn == nil {
		return ProcessorDispatchResult{
			Success:      false,
			ErrorMessage: "AMQP connection not available",
		}
	}

	// Create channel
	channel, err := s.amqpConn.Channel()
	if err != nil {
		return ProcessorDispatchResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to create AMQP channel: %v", err),
		}
	}
	defer channel.Close()

	// Get AMQP configuration
	config := processor.Config
	exchange, _ := config["exchange"].(string)
	routingKey, _ := config["routing_key"].(string)
	queue, _ := config["queue"].(string)

	// If queue is specified, ensure it exists
	if queue != "" {
		_, err = channel.QueueDeclare(
			queue,
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			return ProcessorDispatchResult{
				Success:      false,
				ErrorMessage: fmt.Sprintf("failed to declare queue: %v", err),
			}
		}
	}

	// Prepare message payload
	payload, err := json.Marshal(eventData)
	if err != nil {
		return ProcessorDispatchResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to marshal payload: %v", err),
		}
	}

	// Prepare message headers
	headers := amqp.Table{
		"content_type": "application/json",
		"timestamp":    time.Now().Unix(),
	}

	// Add custom headers if configured
	if configHeaders, exists := config["headers"]; exists {
		if headersMap, ok := configHeaders.(map[string]interface{}); ok {
			for key, value := range headersMap {
				headers[key] = value
			}
		}
	}

	s.logger.Debug("Dispatching to AMQP",
		zap.String("exchange", exchange),
		zap.String("routing_key", routingKey),
		zap.String("queue", queue),
		zap.String("processor_id", processor.ID.Hex()))

	// Publish message
	err = channel.PublishWithContext(
		ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        payload,
			Headers:     headers,
			Timestamp:   time.Now(),
		},
	)

	if err != nil {
		return ProcessorDispatchResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to publish message: %v", err),
		}
	}

	s.logger.Debug("AMQP message published",
		zap.String("processor_id", processor.ID.Hex()))

	return ProcessorDispatchResult{
		Success:      true,
		ResponseBody: "Message published successfully",
	}
}