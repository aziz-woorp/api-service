// Package repository provides data access layer for client user types.
package repository

import (
	"context"

	"github.com/fraiday-org/api-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ClientUserTypeRepository handles database operations for client user types.
type ClientUserTypeRepository struct {
	Collection *mongo.Collection
}

// NewClientUserTypeRepository creates a new ClientUserTypeRepository.
func NewClientUserTypeRepository(db *mongo.Database) *ClientUserTypeRepository {
	return &ClientUserTypeRepository{
		Collection: db.Collection("client_user_types"),
	}
}

// Create creates a new client user type.
func (r *ClientUserTypeRepository) Create(ctx context.Context, userType *models.ClientUserType) error {
	_, err := r.Collection.InsertOne(ctx, userType)
	return err
}

// List retrieves client user types based on filter.
func (r *ClientUserTypeRepository) List(ctx context.Context, filter bson.M) ([]models.ClientUserType, error) {
	cursor, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var userTypes []models.ClientUserType
	if err = cursor.All(ctx, &userTypes); err != nil {
		return nil, err
	}

	return userTypes, nil
}

// GetByFilter retrieves a single client user type based on filter.
func (r *ClientUserTypeRepository) GetByFilter(ctx context.Context, filter bson.M) (*models.ClientUserType, error) {
	var userType models.ClientUserType
	err := r.Collection.FindOne(ctx, filter).Decode(&userType)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &userType, nil
}

// GetByClientAndTypeID retrieves a client user type by client ID and type ID.
func (r *ClientUserTypeRepository) GetByClientAndTypeID(ctx context.Context, clientID primitive.ObjectID, typeID string) (*models.ClientUserType, error) {
	filter := bson.M{
		"client_id": clientID,
		"type_id":   typeID,
	}
	return r.GetByFilter(ctx, filter)
}

// GetByID retrieves a client user type by its ID.
func (r *ClientUserTypeRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.ClientUserType, error) {
	filter := bson.M{"_id": id}
	return r.GetByFilter(ctx, filter)
}

// Update updates a client user type.
func (r *ClientUserTypeRepository) Update(ctx context.Context, id primitive.ObjectID, update bson.M) (*models.ClientUserType, error) {
	filter := bson.M{"_id": id}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated models.ClientUserType
	err := r.Collection.FindOneAndUpdate(ctx, filter, bson.M{"$set": update}, opts).Decode(&updated)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}