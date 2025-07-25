// Package repository provides MongoDB access for chat message feedback.
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

type ChatMessageFeedbackRepository struct {
	Collection *mongo.Collection
}

func NewChatMessageFeedbackRepository(db *mongo.Database) *ChatMessageFeedbackRepository {
	return &ChatMessageFeedbackRepository{
		Collection: db.Collection("chat_message_feedback"),
	}
}

func (r *ChatMessageFeedbackRepository) CreateFeedback(ctx context.Context, feedback *models.ChatMessageFeedback) error {
	now := time.Now()
	feedback.ID = primitive.NewObjectID()
	feedback.CreatedAt = now
	feedback.UpdatedAt = now
	_, err := r.Collection.InsertOne(ctx, feedback)
	return err
}

func (r *ChatMessageFeedbackRepository) ListFeedbacksByMessageID(ctx context.Context, messageID primitive.ObjectID) ([]models.ChatMessageFeedback, error) {
	filter := bson.M{"chat_message_id": messageID}
	cur, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var feedbacks []models.ChatMessageFeedback
	for cur.Next(ctx) {
		var fb models.ChatMessageFeedback
		if err := cur.Decode(&fb); err != nil {
			return nil, err
		}
		feedbacks = append(feedbacks, fb)
	}
	return feedbacks, cur.Err()
}

func (r *ChatMessageFeedbackRepository) UpdateFeedback(ctx context.Context, feedbackID primitive.ObjectID, update bson.M) (*models.ChatMessageFeedback, error) {
	update["updated_at"] = time.Now()
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated models.ChatMessageFeedback
	err := r.Collection.FindOneAndUpdate(ctx, bson.M{"_id": feedbackID}, bson.M{"$set": update}, opts).Decode(&updated)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}
