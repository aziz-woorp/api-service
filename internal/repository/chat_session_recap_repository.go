// Package repository provides MongoDB access for chat session recaps.
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

type ChatSessionRecapRepository struct {
	Collection *mongo.Collection
}

func NewChatSessionRecapRepository(db *mongo.Database) *ChatSessionRecapRepository {
	return &ChatSessionRecapRepository{
		Collection: db.Collection("chat_session_recaps"),
	}
}

func (r *ChatSessionRecapRepository) Create(ctx context.Context, recap *models.ChatSessionRecap) error {
	now := time.Now()
	recap.ID = primitive.NewObjectID()
	recap.CreatedAt = now
	recap.UpdatedAt = now
	_, err := r.Collection.InsertOne(ctx, recap)
	return err
}

func (r *ChatSessionRecapRepository) GetLatestBySessionID(ctx context.Context, sessionID primitive.ObjectID) (*models.ChatSessionRecap, error) {
	filter := bson.M{"session_id": sessionID}
	opts := options.FindOne().SetSort(bson.D{{"created_at", -1}})
	var recap models.ChatSessionRecap
	err := r.Collection.FindOne(ctx, filter, opts).Decode(&recap)
	if err != nil {
		return nil, err
	}
	return &recap, nil
}
