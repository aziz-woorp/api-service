// Package service provides business logic for event processing.
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/fraiday-org/api-service/internal/models"
	"github.com/rabbitmq/amqp091-go"
)

// EventProcessorService handles the actual delivery of events to processors.
type EventProcessorService struct {
	EventDeliveryTrackingService *EventDeliveryTrackingService
	EventProcessorConfigService  *EventProcessorConfigService
	WebhookPayloadService        *WebhookPayloadService
	httpClient                   *http.Client
}

// NewEventProcessorService creates a new EventProcessorService.
func NewEventProcessorService(
	deliveryTrackingService *EventDeliveryTrackingService,
	processorConfigService *EventProcessorConfigService,
	webhookPayloadService *WebhookPayloadService,
) *EventProcessorService {
	return &EventProcessorService{
		EventDeliveryTrackingService: deliveryTrackingService,
		EventProcessorConfigService:  processorConfigService,
		WebhookPayloadService:        webhookPayloadService,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProcessPendingDeliveries processes all pending event deliveries.
func (s *EventProcessorService) ProcessPendingDeliveries(ctx context.Context) error {
	// Get all pending deliveries
	deliveries, err := s.EventDeliveryTrackingService.GetPendingDeliveries(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pending deliveries: %w", err)
	}

	log.Printf("Processing %d pending deliveries", len(deliveries))

	// Process each delivery
	for _, delivery := range deliveries {
		if err := s.ProcessDelivery(ctx, &delivery); err != nil {
			log.Printf("Failed to process delivery %s: %v", delivery.ID.Hex(), err)
		}
	}

	return nil
}

// ProcessDelivery processes a single event delivery.
func (s *EventProcessorService) ProcessDelivery(ctx context.Context, delivery *models.EventDelivery) error {
	// Check if delivery can be retried
	if !delivery.CanRetry() {
		log.Printf("Delivery %s has exceeded max attempts (%d)", delivery.ID.Hex(), delivery.MaxAttempts)
		return nil
	}

	// Get the processor configuration
	config, err := s.EventProcessorConfigService.GetConfigByID(ctx, delivery.EventProcessorConfigID.Hex())
	if err != nil {
		return fmt.Errorf("failed to get processor config: %w", err)
	}

	// Check if config is active
	if !config.IsActive {
		log.Printf("Processor config %s is inactive, skipping delivery %s", config.ID.Hex(), delivery.ID.Hex())
		return nil
	}

	// Process based on processor type
	switch config.ProcessorType {
	case models.ProcessorTypeHTTPWebhook:
		return s.processHTTPWebhook(ctx, delivery, config)
	case models.ProcessorTypeAMQP:
		return s.processAMQP(ctx, delivery, config)
	default:
		return fmt.Errorf("unsupported processor type: %s", config.ProcessorType)
	}
}

// processHTTPWebhook processes an HTTP webhook delivery.
func (s *EventProcessorService) processHTTPWebhook(
	ctx context.Context,
	delivery *models.EventDelivery,
	config *models.EventProcessorConfig,
) error {
	// Get HTTP webhook configuration
	webhookConfig, err := config.GetHttpWebhookConfig()
	if err != nil {
		return fmt.Errorf("failed to get HTTP webhook config: %w", err)
	}

	// Record attempt start
	startTime := time.Now()

	// Prepare request payload - use existing RequestPayload for now
	// TODO: Integrate with WebhookPayloadService when Event model is available
	payloadBytes, err := json.Marshal(delivery.RequestPayload)
	if err != nil {
		return s.recordFailedAttempt(ctx, delivery, 0, fmt.Sprintf("Failed to marshal payload: %v", err), startTime)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", webhookConfig.WebhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return s.recordFailedAttempt(ctx, delivery, 0, fmt.Sprintf("Failed to create request: %v", err), startTime)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for key, value := range webhookConfig.Headers {
		req.Header.Set(key, value)
	}

	// Set timeout if specified
	if webhookConfig.Timeout > 0 {
		ctx, cancel := context.WithTimeout(ctx, time.Duration(webhookConfig.Timeout)*time.Second)
		defer cancel()
		req = req.WithContext(ctx)
	}

	// Make the request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return s.recordFailedAttempt(ctx, delivery, 0, fmt.Sprintf("HTTP request failed: %v", err), startTime)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		respBody = []byte("Failed to read response")
	}

	// TODO: Handle webhook response when Event model integration is complete
	// For now, we'll skip response handling

	// Check if request was successful
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return s.recordSuccessfulAttempt(ctx, delivery, resp.StatusCode, string(respBody), startTime)
	}

	return s.recordFailedAttempt(ctx, delivery, resp.StatusCode, string(respBody), startTime)
}

// processAMQP processes an AMQP delivery.
func (s *EventProcessorService) processAMQP(
	ctx context.Context,
	delivery *models.EventDelivery,
	config *models.EventProcessorConfig,
) error {
	// Get AMQP configuration
	amqpConfig, err := config.GetAmqpConfig()
	if err != nil {
		return fmt.Errorf("failed to get AMQP config: %w", err)
	}

	// Record attempt start
	startTime := time.Now()

	// Prepare connection URL
	var username, password string
	if amqpConfig.Username != nil {
		username = *amqpConfig.Username
	}
	if amqpConfig.Password != nil {
		password = *amqpConfig.Password
	}
	
	connURL := fmt.Sprintf("amqp://%s:%s@%s:%d%s",
		username,
		password,
		amqpConfig.Host,
		amqpConfig.Port,
		amqpConfig.VHost,
	)

	// Connect to AMQP
	conn, err := amqp091.Dial(connURL)
	if err != nil {
		return s.recordFailedAttempt(ctx, delivery, 0, fmt.Sprintf("Failed to connect to AMQP: %v", err), startTime)
	}
	defer conn.Close()

	// Create channel
	ch, err := conn.Channel()
	if err != nil {
		return s.recordFailedAttempt(ctx, delivery, 0, fmt.Sprintf("Failed to create AMQP channel: %v", err), startTime)
	}
	defer ch.Close()

	// Prepare message payload
	payloadBytes, err := json.Marshal(delivery.RequestPayload)
	if err != nil {
		return s.recordFailedAttempt(ctx, delivery, 0, fmt.Sprintf("Failed to marshal payload: %v", err), startTime)
	}

	// Publish message
	err = ch.Publish(
		amqpConfig.Exchange,
		amqpConfig.RoutingKey,
		false, // mandatory
		false, // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        payloadBytes,
		},
	)
	if err != nil {
		return s.recordFailedAttempt(ctx, delivery, 0, fmt.Sprintf("Failed to publish message: %v", err), startTime)
	}

	return s.recordSuccessfulAttempt(ctx, delivery, 200, "Message published successfully", startTime)
}

// recordSuccessfulAttempt records a successful delivery attempt.
func (s *EventProcessorService) recordSuccessfulAttempt(
	ctx context.Context,
	delivery *models.EventDelivery,
	statusCode int,
	responsePayload string,
	startTime time.Time,
) error {
	_, err := s.EventDeliveryTrackingService.RecordDeliveryAttempt(
		ctx,
		delivery.ID,
		delivery.CurrentAttempts+1,
		models.DeliveryStatusCompleted,
		&statusCode,
		"", // no error message
		delivery.RequestPayload,
		map[string]interface{}{"response": responsePayload},
	)
	if err != nil {
		log.Printf("Failed to record successful attempt for delivery %s: %v", delivery.ID.Hex(), err)
	}

	log.Printf("Successfully delivered event to processor %s (delivery: %s)",
		delivery.EventProcessorConfigID.Hex(), delivery.ID.Hex())

	return nil
}

// recordFailedAttempt records a failed delivery attempt.
func (s *EventProcessorService) recordFailedAttempt(
	ctx context.Context,
	delivery *models.EventDelivery,
	statusCode int,
	errorMessage string,
	startTime time.Time,
) error {
	_, err := s.EventDeliveryTrackingService.RecordDeliveryAttempt(
		ctx,
		delivery.ID,
		delivery.CurrentAttempts+1,
		models.DeliveryStatusFailed,
		&statusCode,
		errorMessage,
		delivery.RequestPayload,
		map[string]interface{}{"error": errorMessage},
	)
	if err != nil {
		log.Printf("Failed to record failed attempt for delivery %s: %v", delivery.ID.Hex(), err)
	}

	log.Printf("Failed to deliver event to processor %s (delivery: %s): %s",
		delivery.EventProcessorConfigID.Hex(), delivery.ID.Hex(), errorMessage)

	return fmt.Errorf("delivery failed: %s", errorMessage)
}

// RetryFailedDeliveries retries all failed deliveries that can be retried.
func (s *EventProcessorService) RetryFailedDeliveries(ctx context.Context) error {
	// This would typically get failed deliveries that haven't exceeded max attempts
	// For now, we'll use the existing GetPendingDeliveries method
	deliveries, err := s.EventDeliveryTrackingService.GetPendingDeliveries(ctx)
	if err != nil {
		return fmt.Errorf("failed to get deliveries for retry: %w", err)
	}

	retryCount := 0
	for _, delivery := range deliveries {
		if delivery.Status == models.DeliveryStatusFailed && delivery.CanRetry() {
			if err := s.ProcessDelivery(ctx, &delivery); err != nil {
				log.Printf("Failed to retry delivery %s: %v", delivery.ID.Hex(), err)
			} else {
				retryCount++
			}
		}
	}

	log.Printf("Retried %d failed deliveries", retryCount)
	return nil
}