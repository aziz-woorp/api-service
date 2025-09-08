// Package service provides business logic for events.
package service

import (
	"context"
	"fmt"

	"github.com/fraiday-org/api-service/internal/models"
	"github.com/fraiday-org/api-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EventService encapsulates business logic for events.
type EventService struct {
	Repo *repository.EventRepository
}

// NewEventService creates a new EventService.
func NewEventService(repo *repository.EventRepository) *EventService {
	return &EventService{
		Repo: repo,
	}
}

// CreateEvent creates and saves a new event.
func (s *EventService) CreateEvent(
	ctx context.Context,
	eventType models.EventType,
	entityType models.EntityType,
	entityID string,
	parentID *string,
	data map[string]interface{},
) (*models.Event, error) {
	event := &models.Event{
		EventType:  eventType,
		EntityType: entityType,
		EntityID:   entityID,
		Data:       data,
	}

	if parentID != nil {
		event.ParentID = *parentID
	}

	if data == nil {
		event.Data = make(map[string]interface{})
	}

	if err := s.Repo.Create(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return event, nil
}

// GetEventByID retrieves an event by its ID.
func (s *EventService) GetEventByID(ctx context.Context, eventID string) (*models.Event, error) {
	id, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		return nil, fmt.Errorf("invalid event ID: %w", err)
	}

	event, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return event, nil
}

// ListEvents retrieves events based on filter criteria.
func (s *EventService) ListEvents(
	ctx context.Context,
	eventType *models.EventType,
	entityType *models.EntityType,
	entityID *string,
	parentID *string,
	limit int,
	offset int,
) ([]models.Event, error) {
	filter := make(map[string]interface{})

	if eventType != nil {
		filter["event_type"] = *eventType
	}
	if entityType != nil {
		filter["entity_type"] = *entityType
	}
	if entityID != nil {
		filter["entity_id"] = *entityID
	}
	if parentID != nil {
		filter["parent_id"] = *parentID
	}

	events, err := s.Repo.List(ctx, filter, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	return events, nil
}

// GetEventsByEntityID retrieves all events for a specific entity.
func (s *EventService) GetEventsByEntityID(
	ctx context.Context,
	entityType models.EntityType,
	entityID string,
) ([]models.Event, error) {
	filter := map[string]interface{}{
		"entity_type": entityType,
		"entity_id":   entityID,
	}

	events, err := s.Repo.List(ctx, filter, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get events for entity: %w", err)
	}

	return events, nil
}

// GetEntityEvents retrieves events for a specific entity type and ID (alias for GetEventsByEntityID)
func (s *EventService) GetEntityEvents(ctx context.Context, entityType models.EntityType, entityID string) ([]*models.Event, error) {
	events, err := s.GetEventsByEntityID(ctx, entityType, entityID)
	if err != nil {
		return nil, err
	}

	// Convert to pointer slice to match expected signature
	result := make([]*models.Event, len(events))
	for i := range events {
		result[i] = &events[i]
	}

	return result, nil
}

// GetChildEvents retrieves all child events for a parent event.
func (s *EventService) GetChildEvents(ctx context.Context, parentID string) ([]models.Event, error) {
	filter := map[string]interface{}{
		"parent_id": parentID,
	}

	events, err := s.Repo.List(ctx, filter, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get child events: %w", err)
	}

	return events, nil
}