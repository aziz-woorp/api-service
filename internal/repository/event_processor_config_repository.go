// Package repository provides data access layer for event processor configurations.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/fraiday-org/api-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EventProcessorConfigRepository handles database operations for event processor configurations.
type EventProcessorConfigRepository struct {
	collection *mongo.Collection
}

// NewEventProcessorConfigRepository creates a new EventProcessorConfigRepository.
func NewEventProcessorConfigRepository(db *mongo.Database) *EventProcessorConfigRepository {
	return &EventProcessorConfigRepository{
		collection: db.Collection("event_processor_configs"),
	}
}

// Create inserts a new event processor configuration into the database.
func (r *EventProcessorConfigRepository) Create(ctx context.Context, config *models.EventProcessorConfig) error {
	config.BeforeCreate()

	_, err := r.collection.InsertOne(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to insert event processor config: %w", err)
	}

	return nil
}

// GetByID retrieves an event processor configuration by its ID.
func (r *EventProcessorConfigRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.EventProcessorConfig, error) {
	var config models.EventProcessorConfig
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("event processor config not found")
		}
		return nil, fmt.Errorf("failed to find event processor config: %w", err)
	}

	return &config, nil
}

// List retrieves event processor configurations based on filter criteria with pagination.
func (r *EventProcessorConfigRepository) List(
	ctx context.Context,
	filter map[string]interface{},
	limit int,
	offset int,
) ([]models.EventProcessorConfig, error) {
	opts := options.Find()

	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	if offset > 0 {
		opts.SetSkip(int64(offset))
	}

	// Sort by creation time descending (newest first)
	opts.SetSort(bson.D{{"created_at", -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find event processor configs: %w", err)
	}
	defer cursor.Close(ctx)

	var configs []models.EventProcessorConfig
	if err = cursor.All(ctx, &configs); err != nil {
		return nil, fmt.Errorf("failed to decode event processor configs: %w", err)
	}

	return configs, nil
}

// Update modifies an existing event processor configuration.
func (r *EventProcessorConfigRepository) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	update["updated_at"] = time.Now().UTC()

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": update},
	)
	if err != nil {
		return fmt.Errorf("failed to update event processor config: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("event processor config not found")
	}

	return nil
}

// Delete removes an event processor configuration from the database.
func (r *EventProcessorConfigRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete event processor config: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("event processor config not found")
	}

	return nil
}

// GetByClientID retrieves event processor configurations for a specific client.
func (r *EventProcessorConfigRepository) GetByClientID(ctx context.Context, clientID primitive.ObjectID) ([]models.EventProcessorConfig, error) {
	filter := bson.M{"client": clientID}
	return r.List(ctx, filter, 0, 0)
}

// GetActiveConfigs retrieves all active event processor configurations.
func (r *EventProcessorConfigRepository) GetActiveConfigs(ctx context.Context) ([]models.EventProcessorConfig, error) {
	filter := bson.M{"is_active": true}
	return r.List(ctx, filter, 0, 0)
}

// GetConfigsForEvent retrieves configurations that should process a specific event.
func (r *EventProcessorConfigRepository) GetConfigsForEvent(
	ctx context.Context,
	eventType models.EventType,
	entityType models.EntityType,
) ([]models.EventProcessorConfig, error) {
	filter := bson.M{
		"is_active": true,
		"$and": []bson.M{
			{
				"$or": []bson.M{
					{"event_types": bson.M{"$in": []models.EventType{eventType}}},
					{"event_types": bson.M{"$size": 0}}, // Empty array means all events
				},
			},
			{
				"$or": []bson.M{
					{"entity_types": bson.M{"$in": []models.EntityType{entityType}}},
					{"entity_types": bson.M{"$size": 0}}, // Empty array means all entities
				},
			},
		},
	}

	return r.List(ctx, filter, 0, 0)
}

// Count returns the total number of event processor configurations matching the filter.
func (r *EventProcessorConfigRepository) Count(ctx context.Context, filter map[string]interface{}) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count event processor configs: %w", err)
	}

	return count, nil
}