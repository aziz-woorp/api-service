// Package service provides business logic for event processor configurations.
package service

import (
	"context"
	"fmt"

	"github.com/fraiday-org/api-service/internal/models"
	"github.com/fraiday-org/api-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EventProcessorConfigService encapsulates business logic for event processor configurations.
type EventProcessorConfigService struct {
	Repo *repository.EventProcessorConfigRepository
}

// NewEventProcessorConfigService creates a new EventProcessorConfigService.
func NewEventProcessorConfigService(repo *repository.EventProcessorConfigRepository) *EventProcessorConfigService {
	return &EventProcessorConfigService{
		Repo: repo,
	}
}

// CreateConfig creates and validates a new event processor configuration.
func (s *EventProcessorConfigService) CreateConfig(
	ctx context.Context,
	name string,
	description *string,
	clientID primitive.ObjectID,
	processorType models.ProcessorType,
	config map[string]interface{},
	eventTypes []models.EventType,
	entityTypes []models.EntityType,
) (*models.EventProcessorConfig, error) {
	processorConfig := &models.EventProcessorConfig{
		Name:          name,
		ClientID:      clientID,
		ProcessorType: processorType,
		Config:        config,
		EventTypes:    eventTypes,
		EntityTypes:   entityTypes,
	}

	if description != nil {
		processorConfig.Description = *description
	}

	// Validate the configuration
	if err := processorConfig.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid processor configuration: %w", err)
	}

	if err := s.Repo.Create(ctx, processorConfig); err != nil {
		return nil, fmt.Errorf("failed to create processor config: %w", err)
	}

	return processorConfig, nil
}

// CreateHTTPWebhookConfig creates a new HTTP webhook processor configuration.
func (s *EventProcessorConfigService) CreateHTTPWebhookConfig(
	ctx context.Context,
	name string,
	description *string,
	clientID primitive.ObjectID,
	webhookURL string,
	headers map[string]string,
	timeout int,
	eventTypes []models.EventType,
	entityTypes []models.EntityType,
) (*models.EventProcessorConfig, error) {
	config := map[string]interface{}{
		"webhook_url": webhookURL,
		"headers":     headers,
		"timeout":     timeout,
	}

	return s.CreateConfig(
		ctx,
		name,
		description,
		clientID,
		models.ProcessorTypeHTTPWebhook,
		config,
		eventTypes,
		entityTypes,
	)
}

// CreateAMQPConfig creates a new AMQP processor configuration.
func (s *EventProcessorConfigService) CreateAMQPConfig(
	ctx context.Context,
	name string,
	description *string,
	clientID primitive.ObjectID,
	host string,
	port int,
	vhost string,
	exchange string,
	routingKey string,
	username *string,
	password *string,
	eventTypes []models.EventType,
	entityTypes []models.EntityType,
) (*models.EventProcessorConfig, error) {
	config := map[string]interface{}{
		"host":        host,
		"port":        port,
		"vhost":       vhost,
		"exchange":    exchange,
		"routing_key": routingKey,
	}

	if username != nil {
		config["username"] = *username
	}
	if password != nil {
		config["password"] = *password
	}

	return s.CreateConfig(
		ctx,
		name,
		description,
		clientID,
		models.ProcessorTypeAMQP,
		config,
		eventTypes,
		entityTypes,
	)
}

// GetConfigByID retrieves an event processor configuration by its ID.
func (s *EventProcessorConfigService) GetConfigByID(ctx context.Context, configID string) (*models.EventProcessorConfig, error) {
	id, err := primitive.ObjectIDFromHex(configID)
	if err != nil {
		return nil, fmt.Errorf("invalid config ID: %w", err)
	}

	config, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get processor config: %w", err)
	}

	return config, nil
}

// ListConfigs retrieves event processor configurations with pagination.
func (s *EventProcessorConfigService) ListConfigs(
	ctx context.Context,
	clientID *primitive.ObjectID,
	processorType *models.ProcessorType,
	isActive *bool,
	limit int,
	offset int,
) ([]models.EventProcessorConfig, error) {
	filter := make(map[string]interface{})

	if clientID != nil {
		filter["client"] = *clientID
	}
	if processorType != nil {
		filter["processor_type"] = *processorType
	}
	if isActive != nil {
		filter["is_active"] = *isActive
	}

	configs, err := s.Repo.List(ctx, filter, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list processor configs: %w", err)
	}

	return configs, nil
}

// UpdateConfig updates an existing event processor configuration.
func (s *EventProcessorConfigService) UpdateConfig(
	ctx context.Context,
	configID string,
	updates map[string]interface{},
) error {
	id, err := primitive.ObjectIDFromHex(configID)
	if err != nil {
		return fmt.Errorf("invalid config ID: %w", err)
	}

	// If config is being updated, validate it
	if newConfig, ok := updates["config"]; ok {
		// Get the current config to check processor type
		currentConfig, err := s.Repo.GetByID(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to get current config: %w", err)
		}

		// Create a temporary config for validation
		tempConfig := *currentConfig
		tempConfig.Config = newConfig.(map[string]interface{})

		if err := tempConfig.ValidateConfig(); err != nil {
			return fmt.Errorf("invalid processor configuration: %w", err)
		}
	}

	if err := s.Repo.Update(ctx, id, updates); err != nil {
		return fmt.Errorf("failed to update processor config: %w", err)
	}

	return nil
}

// DeleteConfig removes an event processor configuration.
func (s *EventProcessorConfigService) DeleteConfig(ctx context.Context, configID string) error {
	id, err := primitive.ObjectIDFromHex(configID)
	if err != nil {
		return fmt.Errorf("invalid config ID: %w", err)
	}

	if err := s.Repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete processor config: %w", err)
	}

	return nil
}

// GetConfigsForEvent retrieves configurations that should process a specific event.
func (s *EventProcessorConfigService) GetConfigsForEvent(
	ctx context.Context,
	eventType models.EventType,
	entityType models.EntityType,
) ([]models.EventProcessorConfig, error) {
	configs, err := s.Repo.GetConfigsForEvent(ctx, eventType, entityType)
	if err != nil {
		return nil, fmt.Errorf("failed to get configs for event: %w", err)
	}

	return configs, nil
}

// GetConfigsForEventAndClient retrieves configurations for a specific client that should process a specific event.
func (s *EventProcessorConfigService) GetConfigsForEventAndClient(
	ctx context.Context,
	clientID primitive.ObjectID,
	eventType models.EventType,
	entityType models.EntityType,
) ([]models.EventProcessorConfig, error) {
	configs, err := s.Repo.GetConfigsForEventAndClient(ctx, clientID, eventType, entityType)
	if err != nil {
		return nil, fmt.Errorf("failed to get configs for event and client: %w", err)
	}

	return configs, nil
}

// ToggleConfigStatus toggles the active status of a processor configuration.
func (s *EventProcessorConfigService) ToggleConfigStatus(ctx context.Context, configID string) error {
	id, err := primitive.ObjectIDFromHex(configID)
	if err != nil {
		return fmt.Errorf("invalid config ID: %w", err)
	}

	// Get current config to toggle status
	currentConfig, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get current config: %w", err)
	}

	updates := map[string]interface{}{
		"is_active": !currentConfig.IsActive,
	}

	if err := s.Repo.Update(ctx, id, updates); err != nil {
		return fmt.Errorf("failed to toggle config status: %w", err)
	}

	return nil
}