package service

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/fraiday-org/api-service/internal/models"
)

// ChatMessageSuggestionService handles chat message suggestion operations
type ChatMessageSuggestionService struct {
	collection *mongo.Collection
}

// NewChatMessageSuggestionService creates a new ChatMessageSuggestionService
func NewChatMessageSuggestionService(db *mongo.Database) *ChatMessageSuggestionService {
	return &ChatMessageSuggestionService{
		collection: db.Collection("chat_message_suggestions"),
	}
}

// GetSuggestionsForSession retrieves recent suggestions for a chat session
func (s *ChatMessageSuggestionService) GetSuggestionsForSession(ctx context.Context, chatSessionID string, limit int) ([]*models.ChatMessageSuggestion, error) {
	sessionObjID, err := primitive.ObjectIDFromHex(chatSessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid chat session ID: %w", err)
	}

	filter := bson.M{"chat_session": sessionObjID}
	opts := options.Find().SetSort(bson.D{{"created_at", -1}}).SetLimit(int64(limit))

	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find suggestions: %w", err)
	}
	defer cursor.Close(ctx)

	var suggestions []*models.ChatMessageSuggestion
	if err = cursor.All(ctx, &suggestions); err != nil {
		return nil, fmt.Errorf("failed to decode suggestions: %w", err)
	}

	return suggestions, nil
}

// GetSuggestion retrieves a specific suggestion by ID
func (s *ChatMessageSuggestionService) GetSuggestion(ctx context.Context, suggestionID string) (*models.ChatMessageSuggestion, error) {
	objID, err := primitive.ObjectIDFromHex(suggestionID)
	if err != nil {
		return nil, fmt.Errorf("invalid suggestion ID: %w", err)
	}

	var suggestion models.ChatMessageSuggestion
	err = s.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&suggestion)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find suggestion: %w", err)
	}

	return &suggestion, nil
}

// CreateSuggestion creates a new chat message suggestion
func (s *ChatMessageSuggestionService) CreateSuggestion(ctx context.Context, suggestion *models.ChatMessageSuggestion) error {
	suggestion.BeforeCreate()

	_, err := s.collection.InsertOne(ctx, suggestion)
	if err != nil {
		return fmt.Errorf("failed to create suggestion: %w", err)
	}

	return nil
}

// MarkSuggestionAsUsed marks a suggestion as used
func (s *ChatMessageSuggestionService) MarkSuggestionAsUsed(ctx context.Context, suggestionID string) error {
	objID, err := primitive.ObjectIDFromHex(suggestionID)
	if err != nil {
		return fmt.Errorf("invalid suggestion ID: %w", err)
	}

	now := time.Now().UTC()
	update := bson.M{
		"$set": bson.M{
			"is_used":    true,
			"used_at":    now,
			"updated_at": now,
		},
	}

	_, err = s.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return fmt.Errorf("failed to mark suggestion as used: %w", err)
	}

	return nil
}

// GetActiveSuggestionsForClient retrieves active suggestions for a client
func (s *ChatMessageSuggestionService) GetActiveSuggestionsForClient(ctx context.Context, clientID string, limit int) ([]*models.ChatMessageSuggestion, error) {
	clientObjID, err := primitive.ObjectIDFromHex(clientID)
	if err != nil {
		return nil, fmt.Errorf("invalid client ID: %w", err)
	}

	filter := bson.M{
		"client":  clientObjID,
		"is_used": false,
	}
	opts := options.Find().SetSort(bson.D{{"created_at", -1}}).SetLimit(int64(limit))

	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find active suggestions: %w", err)
	}
	defer cursor.Close(ctx)

	var suggestions []*models.ChatMessageSuggestion
	if err = cursor.All(ctx, &suggestions); err != nil {
		return nil, fmt.Errorf("failed to decode suggestions: %w", err)
	}

	return suggestions, nil
}

// DeleteSuggestion deletes a suggestion by ID
func (s *ChatMessageSuggestionService) DeleteSuggestion(ctx context.Context, suggestionID string) error {
	objID, err := primitive.ObjectIDFromHex(suggestionID)
	if err != nil {
		return fmt.Errorf("invalid suggestion ID: %w", err)
	}

	_, err = s.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return fmt.Errorf("failed to delete suggestion: %w", err)
	}

	return nil
}