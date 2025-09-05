// Package repository provides data access for MongoDB collections.
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/fraiday-org/api-service/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ChatMessageRepository handles CRUD operations for chat messages.
type ChatMessageRepository struct {
	Collection *mongo.Collection
}

// NewChatMessageRepository creates a new repository for chat messages.
func NewChatMessageRepository(db *mongo.Database) *ChatMessageRepository {
	return &ChatMessageRepository{
		Collection: db.Collection("chat_messages"),
	}
}

// Create inserts a new chat message into MongoDB.
func (r *ChatMessageRepository) Create(ctx context.Context, msg *models.ChatMessage) error {
	now := time.Now().UTC()
	msg.CreatedAt = now
	msg.UpdatedAt = now
	
	result, err := r.Collection.InsertOne(ctx, msg)
	if err != nil {
		return err
	}
	
	// Set the generated ObjectID back to the message
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		msg.ID = oid
	}
	
	return nil
}

// List retrieves chat messages by session, user, or other filters.
func (r *ChatMessageRepository) List(ctx context.Context, filter bson.M, limit int64) ([]models.ChatMessage, error) {
	opts := options.Find().SetSort(bson.D{{"created_at", -1}})
	if limit > 0 {
		opts.SetLimit(limit)
	}
	cursor, err := r.Collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []models.ChatMessage
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}
	return messages, nil
}

// Update modifies an existing chat message by ID.
func (r *ChatMessageRepository) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	update["updated_at"] = time.Now().UTC()
	res, err := r.Collection.UpdateByID(ctx, id, bson.M{"$set": update})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("chat message not found")
	}
	return nil
}

// BulkCreate inserts multiple chat messages at once.
func (r *ChatMessageRepository) BulkCreate(ctx context.Context, msgs []models.ChatMessage) error {
	now := time.Now().UTC()
	docs := make([]interface{}, len(msgs))
	for i := range msgs {
		msgs[i].CreatedAt = now
		msgs[i].UpdatedAt = now
		docs[i] = msgs[i]
	}
	
	result, err := r.Collection.InsertMany(ctx, docs)
	if err != nil {
		return err
	}
	
	// Set the generated ObjectIDs back to the messages
	for i, insertedID := range result.InsertedIDs {
		if oid, ok := insertedID.(primitive.ObjectID); ok && i < len(msgs) {
			msgs[i].ID = oid
		}
	}
	
	return nil
}

// GetByID retrieves a chat message by its ObjectID.
func (r *ChatMessageRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.ChatMessage, error) {
	var msg models.ChatMessage
	err := r.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
