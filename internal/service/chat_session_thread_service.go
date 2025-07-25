// Package service provides business logic for chat session threads.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/fraiday-org/api-service/internal/api/dto"
	"github.com/fraiday-org/api-service/internal/models"
	"github.com/fraiday-org/api-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatSessionThreadService struct {
	Repo *repository.ChatSessionThreadRepository
}

func NewChatSessionThreadService(repo *repository.ChatSessionThreadRepository) *ChatSessionThreadService {
	return &ChatSessionThreadService{Repo: repo}
}

func (s *ChatSessionThreadService) CreateThread(ctx context.Context, sessionID string) (*dto.ThreadResponse, error) {
	sid, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return nil, errors.New("invalid session_id")
	}
	thread := &models.ChatSessionThread{
		ThreadID:        primitive.NewObjectID().Hex(),
		ThreadSessionID: primitive.NewObjectID().Hex(),
		ParentSessionID: sessionID,
		ChatSessionID:   sid,
		Active:          true,
		LastActivity:    time.Now(),
	}
	if err := s.Repo.Create(ctx, thread); err != nil {
		return nil, err
	}
	return &dto.ThreadResponse{
		ThreadID:        thread.ThreadID,
		ThreadSessionID: thread.ThreadSessionID,
		ParentSessionID: thread.ParentSessionID,
		ChatSessionID:   thread.ChatSessionID.Hex(),
		Active:          thread.Active,
		LastActivity:    thread.LastActivity,
	}, nil
}

func (s *ChatSessionThreadService) ListThreads(ctx context.Context, sessionID string, includeInactive bool) (*dto.ThreadListResponse, error) {
	sid, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return nil, errors.New("invalid session_id")
	}
	threads, err := s.Repo.ListBySessionID(ctx, sid, includeInactive)
	if err != nil {
		return nil, err
	}
	resp := &dto.ThreadListResponse{
		Threads: make([]dto.ThreadResponse, len(threads)),
		Total:   len(threads),
	}
	for i, t := range threads {
		resp.Threads[i] = dto.ThreadResponse{
			ThreadID:        t.ThreadID,
			ThreadSessionID: t.ThreadSessionID,
			ParentSessionID: t.ParentSessionID,
			ChatSessionID:   t.ChatSessionID.Hex(),
			Active:          t.Active,
			LastActivity:    t.LastActivity,
		}
	}
	return resp, nil
}

func (s *ChatSessionThreadService) GetActiveThread(ctx context.Context, sessionID string, inactivityMinutes int) (*dto.ThreadResponse, error) {
	sid, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return nil, errors.New("invalid session_id")
	}
	thread, err := s.Repo.GetActiveThread(ctx, sid, inactivityMinutes)
	if err != nil {
		return nil, err
	}
	return &dto.ThreadResponse{
		ThreadID:        thread.ThreadID,
		ThreadSessionID: thread.ThreadSessionID,
		ParentSessionID: thread.ParentSessionID,
		ChatSessionID:   thread.ChatSessionID.Hex(),
		Active:          thread.Active,
		LastActivity:    thread.LastActivity,
	}, nil
}

func (s *ChatSessionThreadService) CloseThread(ctx context.Context, sessionID string, threadID *string) (bool, error) {
	sid, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return false, errors.New("invalid session_id")
	}
	return s.Repo.CloseThread(ctx, sid, threadID)
}
