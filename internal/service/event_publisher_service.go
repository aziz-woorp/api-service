// Package service provides business logic for event publishing.
package service

import (
	"context"
	"fmt"
	"log"

	"github.com/fraiday-org/api-service/internal/models"
	"github.com/fraiday-org/api-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EventPublisherService encapsulates business logic for event publishing.
type EventPublisherService struct {
	EventService                  *EventService
	EventProcessorConfigService   *EventProcessorConfigService
	EventDeliveryTrackingService  *EventDeliveryTrackingService
	ChatSessionRepo               *repository.ChatSessionRepository
	ChatMessageRepo               *repository.ChatMessageRepository
	TaskClient                    TaskClient // Interface for publishing tasks to RabbitMQ
}

// TaskClient defines the interface for publishing tasks to RabbitMQ
type TaskClient interface {
	PublishEventProcessorTask(ctx context.Context, eventID string, eventType models.EventType, entityType models.EntityType, entityID string, parentID *string, data map[string]interface{}) error
}

// NewEventPublisherService creates a new EventPublisherService.
func NewEventPublisherService(
	eventService *EventService,
	processorConfigService *EventProcessorConfigService,
	deliveryTrackingService *EventDeliveryTrackingService,
	chatSessionRepo *repository.ChatSessionRepository,
	chatMessageRepo *repository.ChatMessageRepository,
	taskClient TaskClient,
) *EventPublisherService {
	return &EventPublisherService{
		EventService:                 eventService,
		EventProcessorConfigService:  processorConfigService,
		EventDeliveryTrackingService: deliveryTrackingService,
		ChatSessionRepo:              chatSessionRepo,
		ChatMessageRepo:              chatMessageRepo,
		TaskClient:                   taskClient,
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

	// Publish event to RabbitMQ for asynchronous processing
	if s.TaskClient != nil {
		err = s.TaskClient.PublishEventProcessorTask(
			ctx,
			event.ID.Hex(),
			event.EventType,
			event.EntityType,
			event.EntityID,
			func() *string {
				if event.ParentID == "" {
					return nil
				}
				return &event.ParentID
			}(),
			event.Data,
		)
		if err != nil {
			log.Printf("Failed to publish event processor task for event %s: %v", event.ID.Hex(), err)
			// Fallback to sync processing
			go func() {
				if err := s.ProcessEventAsync(context.Background(), event); err != nil {
					log.Printf("Failed to process event %s in fallback: %v", event.ID.Hex(), err)
				}
			}()
		}
	} else {
		// Fallback to synchronous processing if no task client available
		go func() {
			if err := s.ProcessEventAsync(context.Background(), event); err != nil {
				log.Printf("Failed to process event %s: %v", event.ID.Hex(), err)
			}
		}()
	}

	return event, nil
}

// ProcessEventAsync handles the asynchronous processing of events.
func (s *EventPublisherService) ProcessEventAsync(ctx context.Context, event *models.Event) error {
	// Get client ID from the entity
	clientID, err := s.getClientIDForEntity(ctx, event.EntityType, event.EntityID)
	if err != nil {
		log.Printf("Could not determine client ID for event %s (type: %s, entity: %s): %v", 
			event.ID.Hex(), event.EventType, event.EntityType, err)
		// Don't fail the task - just skip processing if we can't find the client
		return nil
	}

	if clientID == nil {
		log.Printf("No client ID found for event %s (type: %s, entity: %s)",
			event.ID.Hex(), event.EventType, event.EntityType)
		return nil
	}

	// Find matching processor configurations for this client
	configs, err := s.EventProcessorConfigService.GetConfigsForEventAndClient(
		ctx,
		*clientID,
		event.EventType,
		event.EntityType,
	)
	if err != nil {
		return fmt.Errorf("failed to get processor configs: %w", err)
	}

	if len(configs) == 0 {
		log.Printf("No processor configurations found for event %s (type: %s, entity: %s, client: %s)",
			event.ID.Hex(), event.EventType, event.EntityType, clientID.Hex())
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

// getClientIDForEntity determines the client ID for different entity types.
func (s *EventPublisherService) getClientIDForEntity(ctx context.Context, entityType models.EntityType, entityID string) (*primitive.ObjectID, error) {
	objectID, err := primitive.ObjectIDFromHex(entityID)
	if err != nil {
		return nil, fmt.Errorf("invalid entity ID: %w", err)
	}

	switch entityType {
	case models.EntityTypeChatMessage:
		// Get message and then get session to find client
		message, err := s.ChatMessageRepo.GetByID(ctx, objectID)
		if err != nil {
			return nil, fmt.Errorf("failed to get chat message: %w", err)
		}
		
		session, err := s.ChatSessionRepo.GetByID(ctx, message.SessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to get chat session: %w", err)
		}
		
		if session.Client == nil {
			return nil, fmt.Errorf("chat session has no client ID")
		}
		
		return session.Client, nil

	case models.EntityTypeChatSession:
		// Get session directly to find client
		session, err := s.ChatSessionRepo.GetByID(ctx, objectID)
		if err != nil {
			return nil, fmt.Errorf("failed to get chat session: %w", err)
		}
		
		return session.Client, nil

	case models.EntityTypeChatSuggestion:
		// For suggestions, we might need to implement ChatSuggestionRepository if available
		// For now, return nil to indicate unsupported
		log.Printf("Client ID resolution for chat suggestions not yet implemented")
		return nil, nil

	case models.EntityTypeAIService:
		// For AI service events, we need to find the related entity through parent_id
		// This would require additional event lookup logic
		log.Printf("Client ID resolution for AI service events not yet implemented")
		return nil, nil

	default:
		return nil, fmt.Errorf("unsupported entity type: %s", entityType)
	}
}