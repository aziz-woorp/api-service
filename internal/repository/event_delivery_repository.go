// Package repository provides data access layer for event delivery tracking.
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

// EventDeliveryRepository handles database operations for event deliveries.
type EventDeliveryRepository struct {
	collection *mongo.Collection
}

// NewEventDeliveryRepository creates a new EventDeliveryRepository.
func NewEventDeliveryRepository(db *mongo.Database) *EventDeliveryRepository {
	return &EventDeliveryRepository{
		collection: db.Collection("event_deliveries"),
	}
}

// Create inserts a new event delivery record into the database.
func (r *EventDeliveryRepository) Create(ctx context.Context, delivery *models.EventDelivery) error {
	delivery.ID = primitive.NewObjectID()
	delivery.CreatedAt = time.Now()
	delivery.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, delivery)
	if err != nil {
		return fmt.Errorf("failed to insert event delivery: %w", err)
	}

	return nil
}

// GetByID retrieves an event delivery by its ID.
func (r *EventDeliveryRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.EventDelivery, error) {
	var delivery models.EventDelivery
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&delivery)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("event delivery not found")
		}
		return nil, fmt.Errorf("failed to find event delivery: %w", err)
	}

	return &delivery, nil
}

// List retrieves event deliveries based on filter criteria with pagination.
func (r *EventDeliveryRepository) List(
	ctx context.Context,
	filter map[string]interface{},
	limit int,
	offset int,
) ([]models.EventDelivery, error) {
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
		return nil, fmt.Errorf("failed to find event deliveries: %w", err)
	}
	defer cursor.Close(ctx)

	var deliveries []models.EventDelivery
	if err = cursor.All(ctx, &deliveries); err != nil {
		return nil, fmt.Errorf("failed to decode event deliveries: %w", err)
	}

	return deliveries, nil
}

// Update modifies an existing event delivery.
func (r *EventDeliveryRepository) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	update["updated_at"] = time.Now()

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": update},
	)
	if err != nil {
		return fmt.Errorf("failed to update event delivery: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("event delivery not found")
	}

	return nil
}

// Delete removes an event delivery from the database.
func (r *EventDeliveryRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete event delivery: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("event delivery not found")
	}

	return nil
}

// GetByEventID retrieves event deliveries for a specific event.
func (r *EventDeliveryRepository) GetByEventID(ctx context.Context, eventID primitive.ObjectID) ([]models.EventDelivery, error) {
	filter := bson.M{"event_id": eventID}
	return r.List(ctx, filter, 0, 0)
}

// GetByProcessorConfigID retrieves event deliveries for a specific processor configuration.
func (r *EventDeliveryRepository) GetByProcessorConfigID(ctx context.Context, configID primitive.ObjectID) ([]models.EventDelivery, error) {
	filter := bson.M{"event_processor_config_id": configID}
	return r.List(ctx, filter, 0, 0)
}

// GetByStatus retrieves event deliveries with a specific status.
func (r *EventDeliveryRepository) GetByStatus(ctx context.Context, status models.DeliveryStatus) ([]models.EventDelivery, error) {
	filter := bson.M{"status": status}
	return r.List(ctx, filter, 0, 0)
}

// GetPendingDeliveries retrieves event deliveries that are pending or need retry.
func (r *EventDeliveryRepository) GetPendingDeliveries(ctx context.Context) ([]models.EventDelivery, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"status": models.DeliveryStatusPending},
			{
				"status":           models.DeliveryStatusFailed,
				"current_attempts": bson.M{"$lt": "$max_attempts"},
			},
		},
	}
	return r.List(ctx, filter, 0, 0)
}

// IncrementAttempts increments the current attempts count for a delivery.
func (r *EventDeliveryRepository) IncrementAttempts(ctx context.Context, id primitive.ObjectID) error {
	update := bson.M{
		"$inc": bson.M{"current_attempts": 1},
		"$set": bson.M{"updated_at": time.Now()},
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		update,
	)
	if err != nil {
		return fmt.Errorf("failed to increment attempts: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("event delivery not found")
	}

	return nil
}

// UpdateStatus updates the status of an event delivery.
func (r *EventDeliveryRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status models.DeliveryStatus) error {
	update := bson.M{
		"status":     status,
		"updated_at": time.Now(),
	}

	if status == models.DeliveryStatusCompleted {
		update["delivered_at"] = time.Now()
	}

	return r.Update(ctx, id, update)
}

// Count returns the total number of event deliveries matching the filter.
func (r *EventDeliveryRepository) Count(ctx context.Context, filter map[string]interface{}) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count event deliveries: %w", err)
	}

	return count, nil
}