// Package repository provides data access layer for CSAT question templates.
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

// CSATQuestionTemplateRepository encapsulates database operations for CSAT question templates.
type CSATQuestionTemplateRepository struct {
	collection *mongo.Collection
}

// NewCSATQuestionTemplateRepository creates a new CSATQuestionTemplateRepository.
func NewCSATQuestionTemplateRepository(db *mongo.Database) *CSATQuestionTemplateRepository {
	return &CSATQuestionTemplateRepository{
		collection: db.Collection("csat_question_templates"),
	}
}

// Create creates a new CSAT question template.
func (r *CSATQuestionTemplateRepository) Create(ctx context.Context, template *models.CSATQuestionTemplate) error {
	template.BeforeCreate()
	_, err := r.collection.InsertOne(ctx, template)
	if err != nil {
		return fmt.Errorf("failed to create CSAT question template: %w", err)
	}
	return nil
}

// GetByID retrieves a CSAT question template by ID.
func (r *CSATQuestionTemplateRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.CSATQuestionTemplate, error) {
	var template models.CSATQuestionTemplate
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&template)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("CSAT question template not found")
		}
		return nil, fmt.Errorf("failed to get CSAT question template: %w", err)
	}
	return &template, nil
}

// GetByClientAndChannel retrieves CSAT question templates by client and channel, ordered by order field.
func (r *CSATQuestionTemplateRepository) GetByClientAndChannel(ctx context.Context, clientID, channelID primitive.ObjectID) ([]models.CSATQuestionTemplate, error) {
	var templates []models.CSATQuestionTemplate
	filter := bson.M{
		"client":         clientID,
		"client_channel": channelID,
		"active":         true,
	}
	
	opts := options.Find().SetSort(bson.D{{Key: "order", Value: 1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get CSAT question templates: %w", err)
	}
	defer cursor.Close(ctx)
	
	if err = cursor.All(ctx, &templates); err != nil {
		return nil, fmt.Errorf("failed to decode CSAT question templates: %w", err)
	}
	
	return templates, nil
}

// Update updates a CSAT question template.
func (r *CSATQuestionTemplateRepository) Update(ctx context.Context, template *models.CSATQuestionTemplate) error {
	template.BeforeUpdate()
	filter := bson.M{"_id": template.ID}
	update := bson.M{"$set": template}
	
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update CSAT question template: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("CSAT question template not found")
	}
	return nil
}

// Delete deletes a CSAT question template.
func (r *CSATQuestionTemplateRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete CSAT question template: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("CSAT question template not found")
	}
	return nil
}

// List retrieves CSAT question templates based on filter criteria.
func (r *CSATQuestionTemplateRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]models.CSATQuestionTemplate, error) {
	var templates []models.CSATQuestionTemplate
	
	opts := options.Find().SetSort(bson.D{{Key: "order", Value: 1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	if offset > 0 {
		opts.SetSkip(int64(offset))
	}
	
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list CSAT question templates: %w", err)
	}
	defer cursor.Close(ctx)
	
	if err = cursor.All(ctx, &templates); err != nil {
		return nil, fmt.Errorf("failed to decode CSAT question templates: %w", err)
	}
	
	return templates, nil
}
