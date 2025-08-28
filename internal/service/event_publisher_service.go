// Package service provides business logic for event publishing.
package service

import (
	"context"
	"fmt"
	"log"

	"github.com/fraiday-org/api-service/internal/models"
)

// EventPublisherService encapsulates business logic for event publishing.
type EventPublisherService struct {
	EventService                  *EventService
	EventProcessorConfigService   *EventProcessorConfigService
	EventDeliveryTrackingService  *EventDeliveryTrackingService
}

// NewEventPublisherService creates a new EventPublisherService.
func NewEventPublisherService(
	eventService *EventService,
	processorConfigService *EventProcessorConfigService,
	deliveryTrackingService *EventDeliveryTrackingService,
) *EventPublisherService {
	return &EventPublisherService{
		EventService:                 eventService,
		EventProcessorConfigService:  processorConfigService,
		EventDeliveryTrackingService: deliveryTrackingService,
	}
}

// PublishEvent creates an event and triggers asynchronous processing.
func (s *EventPublisherService) PublishEvent(
	ctx context.Context,
	eventType models.EventType,
	entityType models.EntityType,
	entityID string,
	parentID *string,
	data map[string]interface{},
) (*models.Event, error) {
	// Create and save the event
	event, err := s.EventService.CreateEvent(
		ctx,
		eventType,
		entityType,
		entityID,
		parentID,
		data,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	// Trigger asynchronous event processing
	go func() {
		if err := s.processEventAsync(context.Background(), event); err != nil {
			log.Printf("Failed to process event %s: %v", event.ID.Hex(), err)
		}
	}()

	return event, nil
}

// processEventAsync handles the asynchronous processing of events.
func (s *EventPublisherService) processEventAsync(ctx context.Context, event *models.Event) error {
	// Find matching processor configurations
	configs, err := s.EventProcessorConfigService.GetConfigsForEvent(
		ctx,
		event.EventType,
		event.EntityType,
	)
	if err != nil {
		return fmt.Errorf("failed to get processor configs: %w", err)
	}

	if len(configs) == 0 {
		log.Printf("No processor configurations found for event %s (type: %s, entity: %s)",
			event.ID.Hex(), event.EventType, event.EntityType)
		return nil
	}

	// Create delivery records for each matching processor
	for _, config := range configs {
		if err := s.createDeliveryRecord(ctx, event, &config); err != nil {
			log.Printf("Failed to create delivery record for event %s and config %s: %v",
				event.ID.Hex(), config.ID.Hex(), err)
			continue
		}
	}

	return nil
}

// createDeliveryRecord creates a delivery record for an event and processor config.
func (s *EventPublisherService) createDeliveryRecord(
	ctx context.Context,
	event *models.Event,
	config *models.EventProcessorConfig,
) error {
	// Prepare the request payload
	requestPayload := map[string]interface{}{
		"event_id":     event.ID.Hex(),
		"event_type":   event.EventType,
		"entity_type":  event.EntityType,
		"entity_id":    event.EntityID,
		"data":         event.Data,
		"created_at":   event.CreatedAt,
	}

	if event.ParentID != "" {
		requestPayload["parent_id"] = event.ParentID
	}

	// Set default max attempts (can be made configurable)
	maxAttempts := 3

	// Create the delivery record
	_, err := s.EventDeliveryTrackingService.CreateDeliveryRecord(
		ctx,
		event.ID,
		config.ID,
		requestPayload,
		maxAttempts,
	)
	if err != nil {
		return fmt.Errorf("failed to create delivery record: %w", err)
	}

	log.Printf("Created delivery record for event %s to processor %s (%s)",
		event.ID.Hex(), config.Name, config.ProcessorType)

	return nil
}

// PublishChatSessionEvent publishes a chat session related event.
func (s *EventPublisherService) PublishChatSessionEvent(
	ctx context.Context,
	eventType models.EventType,
	sessionID string,
	data map[string]interface{},
) (*models.Event, error) {
	return s.PublishEvent(
		ctx,
		eventType,
		models.EntityTypeChatSession,
		sessionID,
		nil,
		data,
	)
}

// PublishChatMessageEvent publishes a chat message related event.
func (s *EventPublisherService) PublishChatMessageEvent(
	ctx context.Context,
	eventType models.EventType,
	messageID string,
	sessionID *string,
	data map[string]interface{},
) (*models.Event, error) {
	return s.PublishEvent(
		ctx,
		eventType,
		models.EntityTypeChatMessage,
		messageID,
		sessionID,
		data,
	)
}

// PublishChatSuggestionEvent publishes a chat suggestion related event.
func (s *EventPublisherService) PublishChatSuggestionEvent(
	ctx context.Context,
	eventType models.EventType,
	suggestionID string,
	messageID *string,
	data map[string]interface{},
) (*models.Event, error) {
	return s.PublishEvent(
		ctx,
		eventType,
		models.EntityTypeChatSuggestion,
		suggestionID,
		messageID,
		data,
	)
}

// PublishAIServiceEvent publishes an AI service related event.
func (s *EventPublisherService) PublishAIServiceEvent(
	ctx context.Context,
	eventType models.EventType,
	serviceID string,
	data map[string]interface{},
) (*models.Event, error) {
	return s.PublishEvent(
		ctx,
		eventType,
		models.EntityTypeAIService,
		serviceID,
		nil,
		data,
	)
}

// GetEventStatus returns the processing status of an event.
func (s *EventPublisherService) GetEventStatus(ctx context.Context, eventID string) (map[string]interface{}, error) {
	// Get the event
	event, err := s.EventService.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	// Get delivery records for the event
	deliveries, err := s.EventDeliveryTrackingService.GetDeliveriesForEvent(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deliveries: %w", err)
	}

	// Prepare status response
	status := map[string]interface{}{
		"event_id":    event.ID.Hex(),
		"event_type":  event.EventType,
		"entity_type": event.EntityType,
		"entity_id":   event.EntityID,
		"created_at":  event.CreatedAt,
		"deliveries":  make([]map[string]interface{}, 0),
	}

	if event.ParentID != "" {
		status["parent_id"] = event.ParentID
	}

	// Add delivery information
	for _, delivery := range deliveries {
		deliveryInfo := map[string]interface{}{
			"delivery_id":              delivery.ID.Hex(),
			"processor_config_id":      delivery.EventProcessorConfigID.Hex(),
			"status":                   delivery.Status,
			"current_attempts":         delivery.CurrentAttempts,
			"max_attempts":             delivery.MaxAttempts,
			"created_at":               delivery.CreatedAt,
			"updated_at":               delivery.UpdatedAt,
		}

		status["deliveries"] = append(status["deliveries"].([]map[string]interface{}), deliveryInfo)
	}

	return status, nil
}