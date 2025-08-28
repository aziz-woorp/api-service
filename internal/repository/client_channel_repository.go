// Package repository provides MongoDB access for client channels.
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

type ClientChannelRepository struct {
	Collection *mongo.Collection
}

func NewClientChannelRepository(db *mongo.Database) *ClientChannelRepository {
	return &ClientChannelRepository{
		Collection: db.Collection("client_channels"),
	}
}

func (r *ClientChannelRepository) Create(ctx context.Context, channel *models.ClientChannel) error {
	now := time.Now().UTC()
	channel.ID = primitive.NewObjectID()
	channel.CreatedAt = now
	channel.UpdatedAt = now
	_, err := r.Collection.InsertOne(ctx, channel)
	return err
}

func (r *ClientChannelRepository) List(ctx context.Context, filter bson.M) ([]models.ClientChannel, error) {
	cur, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var channels []models.ClientChannel
	for cur.Next(ctx) {
		var c models.ClientChannel
		if err := cur.Decode(&c); err != nil {
			return nil, err
		}
		channels = append(channels, c)
	}
	return channels, cur.Err()
}

func (r *ClientChannelRepository) GetByFilter(ctx context.Context, filter bson.M) (*models.ClientChannel, error) {
	var channel models.ClientChannel
	err := r.Collection.FindOne(ctx, filter).Decode(&channel)
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

func (r *ClientChannelRepository) Update(ctx context.Context, id primitive.ObjectID, update bson.M) (*models.ClientChannel, error) {
	update["updated_at"] = time.Now().UTC()
	filter := bson.M{"_id": id}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated models.ClientChannel
	err := r.Collection.FindOneAndUpdate(ctx, filter, bson.M{"$set": update}, opts).Decode(&updated)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *ClientChannelRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.ClientChannel, error) {
	var channel models.ClientChannel
	err := r.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&channel)
	if err != nil {
		return nil, err
	}
	return &channel, nil
}