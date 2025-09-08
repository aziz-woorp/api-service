// Package repository provides MongoDB access for chat sessions.
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

type ChatSessionRepository struct {
	Collection *mongo.Collection
}

func NewChatSessionRepository(db *mongo.Database) *ChatSessionRepository {
	return &ChatSessionRepository{
		Collection: db.Collection("chat_sessions"),
	}
}

func (r *ChatSessionRepository) Create(ctx context.Context, session *models.ChatSession) error {
	now := time.Now()
	session.ID = primitive.NewObjectID()
	session.CreatedAt = now
	session.UpdatedAt = now
	session.Active = true
	_, err := r.Collection.InsertOne(ctx, session)
	return err
}

func (r *ChatSessionRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.ChatSession, error) {
	var session models.ChatSession
	err := r.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *ChatSessionRepository) GetBySessionID(ctx context.Context, sessionID string) (*models.ChatSession, error) {
	var session models.ChatSession
	err := r.Collection.FindOne(ctx, bson.M{"session_id": sessionID}).Decode(&session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// ListWithFilters implements basic filtering and pagination. Advanced aggregation (handover, etc.) can be added as needed.
func (r *ChatSessionRepository) ListWithFilters(ctx context.Context, filter bson.M, skip, limit int64, sort bson.D) ([]models.ChatSession, int64, error) {
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(sort)
	cur, err := r.Collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)

	var sessions []models.ChatSession
	for cur.Next(ctx) {
		var s models.ChatSession
		if err := cur.Decode(&s); err != nil {
			return nil, 0, err
		}
		sessions = append(sessions, s)
	}
	count, err := r.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return sessions, count, nil
}
