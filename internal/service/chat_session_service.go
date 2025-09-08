// Package service provides business logic for chat sessions.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/fraiday-org/api-service/internal/api/dto"
	"github.com/fraiday-org/api-service/internal/models"
	"github.com/fraiday-org/api-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatSessionService struct {
	Repo *repository.ChatSessionRepository
}

func NewChatSessionService(repo *repository.ChatSessionRepository) *ChatSessionService {
	return &ChatSessionService{Repo: repo}
}

func (s *ChatSessionService) CreateSession(ctx context.Context) (string, error) {
	session := &models.ChatSession{
		SessionID: primitive.NewObjectID().Hex(),
	}
	if err := s.Repo.Create(ctx, session); err != nil {
		return "", err
	}
	// Return the session_id that client will use for future message creation
	return session.SessionID, nil
}

func (s *ChatSessionService) GetOrCreateSessionBySessionID(ctx context.Context, sessionID string) (*models.ChatSession, error) {
	// Try to find existing session by session_id
	session, err := s.Repo.GetBySessionID(ctx, sessionID)
	if err == nil {
		return session, nil
	}
	
	// If not found, create new session with this session_id
	session = &models.ChatSession{
		SessionID: sessionID,
		Active:    true,
	}
	if err := s.Repo.Create(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *ChatSessionService) GetSession(ctx context.Context, id string) (*dto.ChatSessionResponse, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid session id")
	}
	session, err := s.Repo.GetByID(ctx, objID)
	if err != nil {
		return nil, err
	}
	return &dto.ChatSessionResponse{
		ID:        session.ID.Hex(),
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
		Active:    session.Active,
	}, nil
}

type ListSessionsParams struct {
	ClientID      *string
	ClientChannel *string
	UserID        *string
	SessionID     *string
	Active        *bool
	StartDate     *time.Time
	EndDate       *time.Time
	Skip          int64
	Limit         int64
}

func (s *ChatSessionService) ListSessions(ctx context.Context, params ListSessionsParams) (*dto.ChatSessionListResponse, error) {
	filter := bson.M{}
	if params.ClientID != nil {
		if objID, err := primitive.ObjectIDFromHex(*params.ClientID); err == nil {
			filter["client"] = objID
		}
	}
	if params.ClientChannel != nil {
		if objID, err := primitive.ObjectIDFromHex(*params.ClientChannel); err == nil {
			filter["client_channel"] = objID
		}
	}
	if params.Active != nil {
		filter["active"] = *params.Active
	}
	if params.StartDate != nil && params.EndDate != nil {
		filter["updated_at"] = bson.M{"$gte": *params.StartDate, "$lte": *params.EndDate}
	} else if params.StartDate != nil {
		filter["updated_at"] = bson.M{"$gte": *params.StartDate}
	} else if params.EndDate != nil {
		filter["updated_at"] = bson.M{"$lte": *params.EndDate}
	}
	if params.SessionID != nil {
		filter["session_id"] = bson.M{"$regex": *params.SessionID, "$options": "i"}
	}
	// UserID filtering would require a lookup in the messages collection, which is not implemented here.

	sessions, total, err := s.Repo.ListWithFilters(ctx, filter, params.Skip, params.Limit, bson.D{{"updated_at", -1}})
	if err != nil {
		return nil, err
	}
	resp := &dto.ChatSessionListResponse{
		Sessions: make([]dto.ChatSessionListItem, len(sessions)),
		Total:    int(total),
	}
	for i, s := range sessions {
		var client, channel *string
		if s.Client != nil {
			str := s.Client.Hex()
			client = &str
		}
		if s.ClientChannel != nil {
			str := s.ClientChannel.Hex()
			channel = &str
		}
		resp.Sessions[i] = dto.ChatSessionListItem{
			ID:            s.ID.Hex(),
			CreatedAt:     s.CreatedAt,
			UpdatedAt:     s.UpdatedAt,
			SessionID:     s.SessionID,
			Active:        s.Active,
			Client:        client,
			ClientChannel: channel,
			Participants:  s.Participants,
			Handover:      false, // Handover detection not implemented in this version
		}
	}
	return resp, nil
}
