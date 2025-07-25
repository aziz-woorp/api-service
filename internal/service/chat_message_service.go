// Package service provides business logic for chat messages.
package service

import (
	"context"
	"errors"

	"github.com/fraiday-org/api-service/internal/models"
	"github.com/fraiday-org/api-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatMessageService encapsulates business logic for chat messages.
type ChatMessageService struct {
	Repo *repository.ChatMessageRepository
}

// NewChatMessageService creates a new ChatMessageService.
func NewChatMessageService(repo *repository.ChatMessageRepository) *ChatMessageService {
	return &ChatMessageService{Repo: repo}
}

// CreateChatMessage creates a new chat message.
func (s *ChatMessageService) CreateChatMessage(ctx context.Context, msg *models.ChatMessage) error {
	// TODO: Add session/thread management, sender type validation, event publishing as needed.
	return s.Repo.Create(ctx, msg)
}

// ListMessages retrieves chat messages by session, user, or other filters.
func (s *ChatMessageService) ListMessages(ctx context.Context, sessionID *primitive.ObjectID, userID *string, lastN int64) ([]models.ChatMessage, error) {
	filter := bson.M{}
	if sessionID != nil {
		filter["session"] = *sessionID
	}
	if userID != nil {
		filter["sender"] = *userID
	}
	return s.Repo.List(ctx, filter, lastN)
}

// UpdateChatMessage updates an existing chat message by ID.
func (s *ChatMessageService) UpdateChatMessage(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	return s.Repo.Update(ctx, id, update)
}

// BulkCreateChatMessages creates multiple chat messages at once.
func (s *ChatMessageService) BulkCreateChatMessages(ctx context.Context, msgs []models.ChatMessage) error {
	return s.Repo.BulkCreate(ctx, msgs)
}

// GetChatMessageByID retrieves a chat message by its ObjectID.
func (s *ChatMessageService) GetChatMessageByID(ctx context.Context, id primitive.ObjectID) (*models.ChatMessage, error) {
	return s.Repo.GetByID(ctx, id)
}

// Helper to parse string to ObjectID, returns nil if invalid.
func ParseObjectID(idStr string) *primitive.ObjectID {
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return nil
	}
	return &id
}

// ValidateSenderType checks if the sender type is valid (user, assistant, system, or client:...).
func ValidateSenderType(senderType string) error {
	if senderType == string(models.SenderTypeUser) ||
		senderType == string(models.SenderTypeAssistant) ||
		senderType == string(models.SenderTypeSystem) ||
		len(senderType) > 7 && senderType[:7] == "client:" {
		return nil
	}
	return errors.New("invalid sender type")
}
