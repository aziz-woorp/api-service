// Package repository provides data access layer for CSAT sessions.
package repository

import (
	"context"
	"fmt"

	"github.com/fraiday-org/api-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CSATSessionRepository encapsulates database operations for CSAT sessions.
type CSATSessionRepository struct {
	collection *mongo.Collection
}

// NewCSATSessionRepository creates a new CSATSessionRepository.
func NewCSATSessionRepository(db *mongo.Database) *CSATSessionRepository {
	return &CSATSessionRepository{
		collection: db.Collection("csat_sessions"),
	}
}

// Create creates a new CSAT session.
func (r *CSATSessionRepository) Create(ctx context.Context, session *models.CSATSession) error {
	session.BeforeCreate()
	_, err := r.collection.InsertOne(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to create CSAT session: %w", err)
	}
	return nil
}

// GetByID retrieves a CSAT session by ID.
func (r *CSATSessionRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.CSATSession, error) {
	var session models.CSATSession
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("CSAT session not found")
		}
		return nil, fmt.Errorf("failed to get CSAT session: %w", err)
	}
	return &session, nil
}

// GetByChatSessionID retrieves a CSAT session by chat session ID.
func (r *CSATSessionRepository) GetByChatSessionID(ctx context.Context, chatSessionID string) (*models.CSATSession, error) {
	var session models.CSATSession
	err := r.collection.FindOne(ctx, bson.M{"chat_session_id": chatSessionID}).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("CSAT session not found")
		}
		return nil, fmt.Errorf("failed to get CSAT session: %w", err)
	}
	return &session, nil
}

// Update updates a CSAT session.
func (r *CSATSessionRepository) Update(ctx context.Context, session *models.CSATSession) error {
	session.BeforeUpdate()
	filter := bson.M{"_id": session.ID}
	update := bson.M{"$set": session}
	
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update CSAT session: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("CSAT session not found")
	}
	return nil
}

// Delete deletes a CSAT session.
func (r *CSATSessionRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete CSAT session: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("CSAT session not found")
	}
	return nil
}

// List retrieves CSAT sessions based on filter criteria.
func (r *CSATSessionRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]models.CSATSession, error) {
	var sessions []models.CSATSession
	
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list CSAT sessions: %w", err)
	}
	defer cursor.Close(ctx)
	
	if err = cursor.All(ctx, &sessions); err != nil {
		return nil, fmt.Errorf("failed to decode CSAT sessions: %w", err)
	}
	
	return sessions, nil
}

// GetActiveByChatSessionID retrieves an active CSAT session by chat session ID.
func (r *CSATSessionRepository) GetActiveByChatSessionID(ctx context.Context, chatSessionID string) (*models.CSATSession, error) {
	var session models.CSATSession
	filter := bson.M{
		"chat_session_id": chatSessionID,
		"status":          bson.M{"$in": []string{"pending", "in_progress"}},
	}
	err := r.collection.FindOne(ctx, filter).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("active CSAT session not found")
		}
		return nil, fmt.Errorf("failed to get active CSAT session: %w", err)
	}
	return &session, nil
}

// GetActiveByBaseSessionID retrieves an active CSAT session by base session ID.
// This handles threaded sessions where the stored chat_session_id might be "base#thread"
// but the lookup is done with just "base".
func (r *CSATSessionRepository) GetActiveByBaseSessionID(ctx context.Context, baseSessionID string) (*models.CSATSession, error) {
	var session models.CSATSession
	
	// Use regex to find sessions where chat_session_id starts with baseSessionID
	// This handles both exact matches and threaded sessions (base#thread)
	filter := bson.M{
		"chat_session_id": bson.M{
			"$regex":   "^" + baseSessionID + "(#.*)?$",
			"$options": "i", // case insensitive
		},
		"status": bson.M{"$in": []string{"pending", "in_progress"}},
	}
	
	// Sort by created_at descending to get the most recent session
	opts := options.FindOne().SetSort(bson.D{{"created_at", -1}})
	
	err := r.collection.FindOne(ctx, filter, opts).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("active CSAT session not found for base session %s", baseSessionID)
		}
		return nil, fmt.Errorf("failed to get active CSAT session by base session ID: %w", err)
	}
	return &session, nil
}
