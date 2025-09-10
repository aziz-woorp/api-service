// Package repository provides data access layer for CSAT responses.
package repository

import (
	"context"
	"fmt"

	"github.com/fraiday-org/api-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// CSATResponseRepository encapsulates database operations for CSAT responses.
type CSATResponseRepository struct {
	collection *mongo.Collection
}

// NewCSATResponseRepository creates a new CSATResponseRepository.
func NewCSATResponseRepository(db *mongo.Database) *CSATResponseRepository {
	return &CSATResponseRepository{
		collection: db.Collection("csat_responses"),
	}
}

// Create creates a new CSAT response.
func (r *CSATResponseRepository) Create(ctx context.Context, response *models.CSATResponse) error {
	response.BeforeCreate()
	_, err := r.collection.InsertOne(ctx, response)
	if err != nil {
		return fmt.Errorf("failed to create CSAT response: %w", err)
	}
	return nil
}

// GetByID retrieves a CSAT response by ID.
func (r *CSATResponseRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.CSATResponse, error) {
	var response models.CSATResponse
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&response)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("CSAT response not found")
		}
		return nil, fmt.Errorf("failed to get CSAT response: %w", err)
	}
	return &response, nil
}

// GetBySessionID retrieves all CSAT responses for a session.
func (r *CSATResponseRepository) GetBySessionID(ctx context.Context, sessionID primitive.ObjectID) ([]models.CSATResponse, error) {
	var responses []models.CSATResponse
	cursor, err := r.collection.Find(ctx, bson.M{"csat_session": sessionID})
	if err != nil {
		return nil, fmt.Errorf("failed to get CSAT responses: %w", err)
	}
	defer cursor.Close(ctx)
	
	if err = cursor.All(ctx, &responses); err != nil {
		return nil, fmt.Errorf("failed to decode CSAT responses: %w", err)
	}
	
	return responses, nil
}

// GetBySessionAndQuestion retrieves a CSAT response for a specific session and question.
func (r *CSATResponseRepository) GetBySessionAndQuestion(ctx context.Context, sessionID, questionID primitive.ObjectID) (*models.CSATResponse, error) {
	var response models.CSATResponse
	filter := bson.M{
		"csat_session":      sessionID,
		"question_template": questionID,
	}
	
	err := r.collection.FindOne(ctx, filter).Decode(&response)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("CSAT response not found for session and question")
		}
		return nil, fmt.Errorf("failed to get CSAT response: %w", err)
	}
	return &response, nil
}

// Update updates a CSAT response.
func (r *CSATResponseRepository) Update(ctx context.Context, response *models.CSATResponse) error {
	response.BeforeUpdate()
	filter := bson.M{"_id": response.ID}
	update := bson.M{"$set": response}
	
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update CSAT response: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("CSAT response not found")
	}
	return nil
}

// Delete deletes a CSAT response.
func (r *CSATResponseRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete CSAT response: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("CSAT response not found")
	}
	return nil
}

// List retrieves CSAT responses based on filter criteria.
func (r *CSATResponseRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]models.CSATResponse, error) {
	var responses []models.CSATResponse
	
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list CSAT responses: %w", err)
	}
	defer cursor.Close(ctx)
	
	if err = cursor.All(ctx, &responses); err != nil {
		return nil, fmt.Errorf("failed to decode CSAT responses: %w", err)
	}
	
	return responses, nil
}
