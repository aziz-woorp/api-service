// Package service provides business logic for client channels.
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

// ClientChannelService encapsulates business logic for client channels.
type ClientChannelService struct {
	Repo       *repository.ClientChannelRepository
	ClientRepo *repository.ClientRepository
}

// NewClientChannelService creates a new ClientChannelService.
func NewClientChannelService(repo *repository.ClientChannelRepository, clientRepo *repository.ClientRepository) *ClientChannelService {
	return &ClientChannelService{
		Repo:       repo,
		ClientRepo: clientRepo,
	}
}

// CreateChannel creates a new client channel.
func (s *ClientChannelService) CreateChannel(ctx context.Context, clientID string, req *dto.ClientChannelCreateOrUpdateRequest) (*dto.ClientChannelResponse, error) {
	// Validate client exists
	client, err := s.ClientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, errors.New("client not found")
	}

	// Create channel
	channel := &models.ClientChannel{
		ChannelType:   req.ChannelType,
		ChannelConfig: req.ChannelConfig,
		ClientID:      client.ID,
		IsActive:      true,
	}

	if req.IsActive != nil {
		channel.IsActive = *req.IsActive
	}

	if err := s.Repo.Create(ctx, channel); err != nil {
		return nil, err
	}

	return &dto.ClientChannelResponse{
		ID:            channel.ID.Hex(),
		ChannelType:   channel.ChannelType,
		ChannelConfig: channel.ChannelConfig,
		IsActive:      channel.IsActive,
	}, nil
}

// ListChannels retrieves all active channels for a client.
func (s *ClientChannelService) ListChannels(ctx context.Context, clientID string) ([]dto.ClientChannelResponse, error) {
	// Validate client exists
	client, err := s.ClientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, errors.New("client not found")
	}

	// Get channels
	filter := bson.M{
		"client":    client.ID,
		"is_active": true,
	}

	channels, err := s.Repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	resp := make([]dto.ClientChannelResponse, len(channels))
	for i, c := range channels {
		resp[i] = dto.ClientChannelResponse{
			ID:            c.ID.Hex(),
			ChannelType:   c.ChannelType,
			ChannelConfig: c.ChannelConfig,
			IsActive:      c.IsActive,
		}
	}

	return resp, nil
}

// GetChannelByType retrieves a specific channel for a client by its type.
func (s *ClientChannelService) GetChannelByType(ctx context.Context, clientID, channelType string) (*models.ClientChannel, error) {
	// Validate client exists
	client, err := s.ClientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, errors.New("client not found")
	}

	// Get channel by type
	filter := bson.M{
		"client":       client.ID,
		"channel_type": channelType,
	}

	return s.Repo.GetByFilter(ctx, filter)
}

// UpdateChannel updates an existing client channel.
func (s *ClientChannelService) UpdateChannel(ctx context.Context, channelID string, req *dto.ClientChannelCreateOrUpdateRequest) (*dto.ClientChannelResponse, error) {
	channelObjID, err := primitive.ObjectIDFromHex(channelID)
	if err != nil {
		return nil, errors.New("invalid channel ID")
	}

	update := bson.M{}
	if req.ChannelType != "" {
		update["channel_type"] = req.ChannelType
	}
	if req.ChannelConfig != nil {
		update["channel_config"] = req.ChannelConfig
	}
	if req.IsActive != nil {
		update["is_active"] = *req.IsActive
	}

	updated, err := s.Repo.Update(ctx, channelObjID, update)
	if err != nil {
		return nil, err
	}

	return &dto.ClientChannelResponse{
		ID:            updated.ID.Hex(),
		ChannelType:   updated.ChannelType,
		ChannelConfig: updated.ChannelConfig,
		IsActive:      updated.IsActive,
	}, nil
}

// DeleteChannel soft deletes a client channel by setting is_active to false.
func (s *ClientChannelService) DeleteChannel(ctx context.Context, channelID string) error {
	channelObjID, err := primitive.ObjectIDFromHex(channelID)
	if err != nil {
		return errors.New("invalid channel ID")
	}

	update := bson.M{"is_active": false}
	_, err = s.Repo.Update(ctx, channelObjID, update)
	return err
}