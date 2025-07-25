// Package repository provides MongoDB access for chat session threads.
package repository

import (
	"context"
	"time"

	"github.com/fraiday-org/api-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ChatSessionThreadRepository struct {
	Collection *mongo.Collection
}

func NewChatSessionThreadRepository(db *mongo.Database) *ChatSessionThreadRepository {
	return &ChatSessionThreadRepository{
		Collection: db.Collection("chat_session_threads"),
	}
}

func (r *ChatSessionThreadRepository) Create(ctx context.Context, thread *models.ChatSessionThread) error {
	thread.ID = primitive.NewObjectID()
	thread.LastActivity = time.Now()
	thread.Active = true
	_, err := r.Collection.InsertOne(ctx, thread)
	return err
}

func (r *ChatSessionThreadRepository) ListBySessionID(ctx context.Context, sessionID primitive.ObjectID, includeInactive bool) ([]models.ChatSessionThread, error) {
	filter := bson.M{"chat_session_id": sessionID}
	if !includeInactive {
		filter["active"] = true
	}
	cur, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var threads []models.ChatSessionThread
	for cur.Next(ctx) {
		var t models.ChatSessionThread
		if err := cur.Decode(&t); err != nil {
			return nil, err
		}
		threads = append(threads, t)
	}
	return threads, cur.Err()
}

func (r *ChatSessionThreadRepository) GetActiveThread(ctx context.Context, sessionID primitive.ObjectID, inactivityMinutes int) (*models.ChatSessionThread, error) {
	filter := bson.M{
		"chat_session_id": sessionID,
		"active":          true,
	}
	opts := options.FindOne().SetSort(bson.D{{"last_activity", -1}})
	var thread models.ChatSessionThread
	err := r.Collection.FindOne(ctx, filter, opts).Decode(&thread)
	if err != nil {
		return nil, err
	}
	// Check inactivity
	if inactivityMinutes > 0 && time.Since(thread.LastActivity) > time.Duration(inactivityMinutes)*time.Minute {
		return nil, mongo.ErrNoDocuments
	}
	return &thread, nil
}

func (r *ChatSessionThreadRepository) CloseThread(ctx context.Context, sessionID primitive.ObjectID, threadID *string) (bool, error) {
	filter := bson.M{"chat_session_id": sessionID, "active": true}
	if threadID != nil {
		filter["thread_id"] = *threadID
	}
	update := bson.M{"$set": bson.M{"active": false}}
	res, err := r.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, err
	}
	return res.ModifiedCount > 0, nil
}
