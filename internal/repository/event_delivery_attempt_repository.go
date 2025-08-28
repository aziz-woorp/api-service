// Package repository provides data access layer for event delivery attempts.
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

// EventDeliveryAttemptRepository handles database operations for event delivery attempts.
type EventDeliveryAttemptRepository struct {
	collection *mongo.Collection
}

// NewEventDeliveryAttemptRepository creates a new EventDeliveryAttemptRepository.
func NewEventDeliveryAttemptRepository(db *mongo.Database) *EventDeliveryAttemptRepository {
	return &EventDeliveryAttemptRepository{
		collection: db.Collection("event_delivery_attempts"),
	}
}

// Create inserts a new event delivery attempt record into the database.
func (r *EventDeliveryAttemptRepository) Create(ctx context.Context, attempt *models.EventDeliveryAttempt) error {
	attempt.ID = primitive.NewObjectID()
	attempt.CreatedAt = time.Now()
	attempt.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, attempt)
	if err != nil {
		return fmt.Errorf("failed to insert event delivery attempt: %w", err)
	}

	return nil
}

// GetByID retrieves an event delivery attempt by its ID.
func (r *EventDeliveryAttemptRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.EventDeliveryAttempt, error) {
	var attempt models.EventDeliveryAttempt
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&attempt)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("event delivery attempt not found")
		}
		return nil, fmt.Errorf("failed to find event delivery attempt: %w", err)
	}

	return &attempt, nil
}

// List retrieves event delivery attempts based on filter criteria with pagination.
func (r *EventDeliveryAttemptRepository) List(
	ctx context.Context,
	filter map[string]interface{},
	limit int,
	offset int,
) ([]models.EventDeliveryAttempt, error) {
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
		return nil, fmt.Errorf("failed to find event delivery attempts: %w", err)
	}
	defer cursor.Close(ctx)

	var attempts []models.EventDeliveryAttempt
	if err = cursor.All(ctx, &attempts); err != nil {
		return nil, fmt.Errorf("failed to decode event delivery attempts: %w", err)
	}

	return attempts, nil
}

// Update modifies an existing event delivery attempt.
func (r *EventDeliveryAttemptRepository) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	update["updated_at"] = time.Now()

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": update},
	)
	if err != nil {
		return fmt.Errorf("failed to update event delivery attempt: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("event delivery attempt not found")
	}

	return nil
}

// Delete removes an event delivery attempt from the database.
func (r *EventDeliveryAttemptRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete event delivery attempt: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("event delivery attempt not found")
	}

	return nil
}

// GetByDeliveryID retrieves event delivery attempts for a specific delivery.
func (r *EventDeliveryAttemptRepository) GetByDeliveryID(ctx context.Context, deliveryID primitive.ObjectID) ([]models.EventDeliveryAttempt, error) {
	filter := bson.M{"event_delivery_id": deliveryID}
	return r.List(ctx, filter, 0, 0)
}

// GetByStatus retrieves event delivery attempts with a specific status.
func (r *EventDeliveryAttemptRepository) GetByStatus(ctx context.Context, status models.AttemptStatus) ([]models.EventDeliveryAttempt, error) {
	filter := bson.M{"status": status}
	return r.List(ctx, filter, 0, 0)
}

// GetLatestAttempt retrieves the most recent attempt for a specific delivery.
func (r *EventDeliveryAttemptRepository) GetLatestAttempt(ctx context.Context, deliveryID primitive.ObjectID) (*models.EventDeliveryAttempt, error) {
	filter := bson.M{"event_delivery_id": deliveryID}
	opts := options.FindOne().SetSort(bson.D{{"created_at", -1}})

	var attempt models.EventDeliveryAttempt
	err := r.collection.FindOne(ctx, filter, opts).Decode(&attempt)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no attempts found for delivery")
		}
		return nil, fmt.Errorf("failed to find latest attempt: %w", err)
	}

	return &attempt, nil
}

// GetSuccessfulAttempts retrieves all successful attempts for a specific delivery.
func (r *EventDeliveryAttemptRepository) GetSuccessfulAttempts(ctx context.Context, deliveryID primitive.ObjectID) ([]models.EventDeliveryAttempt, error) {
	filter := bson.M{
		"event_delivery_id": deliveryID,
		"status":            models.AttemptStatusSuccess,
	}
	return r.List(ctx, filter, 0, 0)
}

// GetFailedAttempts retrieves all failed attempts for a specific delivery.
func (r *EventDeliveryAttemptRepository) GetFailedAttempts(ctx context.Context, deliveryID primitive.ObjectID) ([]models.EventDeliveryAttempt, error) {
	filter := bson.M{
		"event_delivery_id": deliveryID,
		"status":            models.AttemptStatusFailure,
	}
	return r.List(ctx, filter, 0, 0)
}

// Count returns the total number of event delivery attempts matching the filter.
func (r *EventDeliveryAttemptRepository) Count(ctx context.Context, filter map[string]interface{}) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count event delivery attempts: %w", err)
	}

	return count, nil
}

// CountByDeliveryID returns the number of attempts for a specific delivery.
func (r *EventDeliveryAttemptRepository) CountByDeliveryID(ctx context.Context, deliveryID primitive.ObjectID) (int64, error) {
	filter := map[string]interface{}{"event_delivery_id": deliveryID}
	return r.Count(ctx, filter)
}