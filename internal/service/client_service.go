// Package service provides business logic for clients.
package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"github.com/fraiday-org/api-service/internal/api/dto"
	"github.com/fraiday-org/api-service/internal/models"
	"github.com/fraiday-org/api-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ClientService struct {
	Repo *repository.ClientRepository
}

func NewClientService(repo *repository.ClientRepository) *ClientService {
	return &ClientService{Repo: repo}
}

func generateClientSecret(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
}

func (s *ClientService) CreateClient(ctx context.Context, req *dto.ClientCreateOrUpdateRequest) (*dto.ClientResponse, error) {
	clientID := ""
	if req.ClientID != nil && *req.ClientID != "" {
		clientID = *req.ClientID
	} else {
		clientID = primitive.NewObjectID().Hex()
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	client := &models.Client{
		Name:      req.Name,
		Email:     req.Email,
		ClientID:  clientID,
		ClientKey: generateClientSecret(32),
		IsActive:  isActive,
	}
	if err := s.Repo.Create(ctx, client); err != nil {
		return nil, err
	}
	return &dto.ClientResponse{
		ID:       client.ID.Hex(),
		Name:     client.Name,
		Email:    client.Email,
		ClientID: client.ClientID,
		IsActive: client.IsActive,
	}, nil
}

func (s *ClientService) ListClients(ctx context.Context) ([]dto.ClientResponse, error) {
	clients, err := s.Repo.List(ctx)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.ClientResponse, len(clients))
	for i, c := range clients {
		resp[i] = dto.ClientResponse{
			ID:       c.ID.Hex(),
			Name:     c.Name,
			Email:    c.Email,
			ClientID: c.ClientID,
			IsActive: c.IsActive,
		}
	}
	return resp, nil
}

// GetClient retrieves a client by client_id
func (s *ClientService) GetClient(ctx context.Context, clientID string) (*models.Client, error) {
	return s.Repo.GetByClientID(ctx, clientID)
}

func (s *ClientService) UpdateClient(ctx context.Context, clientID string, req *dto.ClientCreateOrUpdateRequest) (*dto.ClientResponse, error) {
	update := bson.M{}
	if req.Name != "" {
		update["name"] = req.Name
	}
	if req.Email != nil {
		update["email"] = req.Email
	}
	if req.IsActive != nil {
		update["is_active"] = *req.IsActive
	}
	updated, err := s.Repo.Update(ctx, clientID, update)
	if err != nil {
		return nil, err
	}
	return &dto.ClientResponse{
		ID:       updated.ID.Hex(),
		Name:     updated.Name,
		Email:    updated.Email,
		ClientID: updated.ClientID,
		IsActive: updated.IsActive,
	}, nil
}
