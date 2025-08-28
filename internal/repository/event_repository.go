// Package repository provides data access layer for events.
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

// EventRepository handles database operations for events.
type EventRepository struct {
	collection *mongo.Collection
}

// NewEventRepository creates a new EventRepository.
func NewEventRepository(db *mongo.Database) *EventRepository {
	return &EventRepository{
		collection: db.Collection("events"),
	}
}

// Create inserts a new event into the database.
func (r *EventRepository) Create(ctx context.Context, event *models.Event) error {
	event.ID = primitive.NewObjectID()
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	return nil
}

// GetByID retrieves an event by its ID.
func (r *EventRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Event, error) {
	var event models.Event
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&event)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("event not found")
		}
		return nil, fmt.Errorf("failed to find event: %w", err)
	}

	return &event, nil
}

// List retrieves events based on filter criteria with pagination.
func (r *EventRepository) List(
	ctx context.Context,
	filter map[string]interface{},
	limit int,
	offset int,
) ([]models.Event, error) {
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
		return nil, fmt.Errorf("failed to find events: %w", err)
	}
	defer cursor.Close(ctx)

	var events []models.Event
	if err = cursor.All(ctx, &events); err != nil {
		return nil, fmt.Errorf("failed to decode events: %w", err)
	}

	return events, nil
}

// Update modifies an existing event.
func (r *EventRepository) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	update["updated_at"] = time.Now()

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": update},
	)
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("event not found")
	}

	return nil
}

// Delete removes an event from the database.
func (r *EventRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("event not found")
	}

	return nil
}

// Count returns the total number of events matching the filter.
func (r *EventRepository) Count(ctx context.Context, filter map[string]interface{}) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count events: %w", err)
	}

	return count, nil
}