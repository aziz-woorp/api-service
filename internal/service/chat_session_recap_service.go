// Package service provides business logic for chat session recaps.
package service

import (
	"context"
	"errors"

	"github.com/fraiday-org/api-service/internal/api/dto"
	"github.com/fraiday-org/api-service/internal/models"
	"github.com/fraiday-org/api-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatSessionRecapService struct {
	Repo *repository.ChatSessionRecapRepository
}

func NewChatSessionRecapService(repo *repository.ChatSessionRecapRepository) *ChatSessionRecapService {
	return &ChatSessionRecapService{Repo: repo}
}

// GenerateRecap simulates AI-based recap generation and stores it.
func (s *ChatSessionRecapService) GenerateRecap(ctx context.Context, sessionID string, recapData map[string]interface{}) (*dto.ChatSessionRecapResponse, error) {
	sid, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return nil, errors.New("invalid session_id")
	}
	recap := &models.ChatSessionRecap{
		SessionID: sid,
		RecapData: recapData,
	}
	if err := s.Repo.Create(ctx, recap); err != nil {
		return nil, err
	}
	return &dto.ChatSessionRecapResponse{
		ID:        recap.ID.Hex(),
		SessionID: recap.SessionID.Hex(),
		RecapData: recap.RecapData,
		CreatedAt: recap.CreatedAt,
		UpdatedAt: recap.UpdatedAt,
	}, nil
}

// GetLatestRecap retrieves the latest recap for a session.
func (s *ChatSessionRecapService) GetLatestRecap(ctx context.Context, sessionID string) (*dto.ChatSessionRecapResponse, error) {
	sid, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return nil, errors.New("invalid session_id")
	}
	recap, err := s.Repo.GetLatestBySessionID(ctx, sid)
	if err != nil {
		return nil, err
	}
	return &dto.ChatSessionRecapResponse{
		ID:        recap.ID.Hex(),
		SessionID: recap.SessionID.Hex(),
		RecapData: recap.RecapData,
		CreatedAt: recap.CreatedAt,
		UpdatedAt: recap.UpdatedAt,
	}, nil
}
