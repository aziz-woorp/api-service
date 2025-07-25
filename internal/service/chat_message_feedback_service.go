// Package service provides business logic for chat message feedback.
package service

import (
	"context"
	"errors"

	"github.com/fraiday-org/api-service/internal/api/dto"
	"github.com/fraiday-org/api-service/internal/models"
	"github.com/fraiday-org/api-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatMessageFeedbackService struct {
	Repo *repository.ChatMessageFeedbackRepository
}

func NewChatMessageFeedbackService(repo *repository.ChatMessageFeedbackRepository) *ChatMessageFeedbackService {
	return &ChatMessageFeedbackService{Repo: repo}
}

func (s *ChatMessageFeedbackService) CreateFeedback(ctx context.Context, messageID string, req *dto.ChatMessageFeedbackCreate) (*models.ChatMessageFeedback, error) {
	msgID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return nil, errors.New("invalid message_id")
	}
	feedback := &models.ChatMessageFeedback{
		ChatMessageID: msgID,
		Rating:        req.Rating,
		Comment:       req.Comment,
		Metadata:      req.Metadata,
	}
	if err := s.Repo.CreateFeedback(ctx, feedback); err != nil {
		return nil, err
	}
	return feedback, nil
}

func (s *ChatMessageFeedbackService) ListFeedbacksByMessageID(ctx context.Context, messageID string) ([]models.ChatMessageFeedback, error) {
	msgID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return nil, errors.New("invalid message_id")
	}
	return s.Repo.ListFeedbacksByMessageID(ctx, msgID)
}

func (s *ChatMessageFeedbackService) UpdateFeedback(ctx context.Context, feedbackID string, rating *int, comment *string, metadata map[string]interface{}) (*models.ChatMessageFeedback, error) {
	fbID, err := primitive.ObjectIDFromHex(feedbackID)
	if err != nil {
		return nil, errors.New("invalid feedback_id")
	}
	update := bson.M{}
	if rating != nil {
		update["rating"] = *rating
	}
	if comment != nil {
		update["comment"] = *comment
	}
	if metadata != nil {
		update["metadata"] = metadata
	}
	return s.Repo.UpdateFeedback(ctx, fbID, update)
}
