// Package repository provides MongoDB access for clients.
package repository

import (
	"context"

	"github.com/fraiday-org/api-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ClientRepository struct {
	Collection *mongo.Collection
}

func NewClientRepository(db *mongo.Database) *ClientRepository {
	return &ClientRepository{
		Collection: db.Collection("clients"),
	}
}

func (r *ClientRepository) Create(ctx context.Context, client *models.Client) error {
	client.ID = primitive.NewObjectID()
	_, err := r.Collection.InsertOne(ctx, client)
	return err
}

func (r *ClientRepository) List(ctx context.Context) ([]models.Client, error) {
	filter := bson.M{"is_active": true}
	cur, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var clients []models.Client
	for cur.Next(ctx) {
		var c models.Client
		if err := cur.Decode(&c); err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}
	return clients, cur.Err()
}

func (r *ClientRepository) Update(ctx context.Context, clientID string, update bson.M) (*models.Client, error) {
	filter := bson.M{"client_id": clientID}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated models.Client
	err := r.Collection.FindOneAndUpdate(ctx, filter, bson.M{"$set": update}, opts).Decode(&updated)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *ClientRepository) GetByClientID(ctx context.Context, clientID string) (*models.Client, error) {
	var client models.Client
	err := r.Collection.FindOne(ctx, bson.M{"client_id": clientID}).Decode(&client)
	if err != nil {
		return nil, err
	}
	return &client, nil
}
