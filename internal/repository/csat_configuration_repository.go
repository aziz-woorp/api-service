// Package repository provides data access layer for CSAT configurations.
package repository

import (
	"context"
	"fmt"

	"github.com/fraiday-org/api-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// CSATConfigurationRepository encapsulates database operations for CSAT configurations.
type CSATConfigurationRepository struct {
	collection *mongo.Collection
}

// NewCSATConfigurationRepository creates a new CSATConfigurationRepository.
func NewCSATConfigurationRepository(db *mongo.Database) *CSATConfigurationRepository {
	return &CSATConfigurationRepository{
		collection: db.Collection("csat_configurations"),
	}
}

// Create creates a new CSAT configuration.
func (r *CSATConfigurationRepository) Create(ctx context.Context, config *models.CSATConfiguration) error {
	config.BeforeCreate()
	_, err := r.collection.InsertOne(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create CSAT configuration: %w", err)
	}
	return nil
}

// GetByID retrieves a CSAT configuration by ID.
func (r *CSATConfigurationRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.CSATConfiguration, error) {
	var config models.CSATConfiguration
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("CSAT configuration not found")
		}
		return nil, fmt.Errorf("failed to get CSAT configuration: %w", err)
	}
	return &config, nil
}

// GetByClientAndChannel retrieves a CSAT configuration by client and channel.
func (r *CSATConfigurationRepository) GetByClientAndChannel(ctx context.Context, clientID, channelID primitive.ObjectID) (*models.CSATConfiguration, error) {
	var config models.CSATConfiguration
	filter := bson.M{
		"client":         clientID,
		"client_channel": channelID,
	}
	err := r.collection.FindOne(ctx, filter).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("CSAT configuration not found")
		}
		return nil, fmt.Errorf("failed to get CSAT configuration: %w", err)
	}
	return &config, nil
}

// Update updates a CSAT configuration.
func (r *CSATConfigurationRepository) Update(ctx context.Context, config *models.CSATConfiguration) error {
	config.BeforeUpdate()
	filter := bson.M{"_id": config.ID}
	update := bson.M{"$set": config}
	
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update CSAT configuration: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("CSAT configuration not found")
	}
	return nil
}

// Delete deletes a CSAT configuration.
func (r *CSATConfigurationRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete CSAT configuration: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("CSAT configuration not found")
	}
	return nil
}

// List retrieves CSAT configurations based on filter criteria.
func (r *CSATConfigurationRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]models.CSATConfiguration, error) {
	var configs []models.CSATConfiguration
	
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list CSAT configurations: %w", err)
	}
	defer cursor.Close(ctx)
	
	if err = cursor.All(ctx, &configs); err != nil {
		return nil, fmt.Errorf("failed to decode CSAT configurations: %w", err)
	}
	
	return configs, nil
}
