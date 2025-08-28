// Package service provides business logic for client user types.
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/fraiday-org/api-service/internal/api/dto"
	"github.com/fraiday-org/api-service/internal/models"
	"github.com/fraiday-org/api-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
)

// ClientUserTypeService encapsulates business logic for client user types.
type ClientUserTypeService struct {
	Repo       *repository.ClientUserTypeRepository
	ClientRepo *repository.ClientRepository
}

// NewClientUserTypeService creates a new ClientUserTypeService.
func NewClientUserTypeService(repo *repository.ClientUserTypeRepository, clientRepo *repository.ClientRepository) *ClientUserTypeService {
	return &ClientUserTypeService{
		Repo:       repo,
		ClientRepo: clientRepo,
	}
}

// CreateUserType creates a new user type for a client.
func (s *ClientUserTypeService) CreateUserType(ctx context.Context, clientID string, req *dto.ClientUserTypeCreateRequest) (*dto.ClientUserTypeResponse, error) {
	// Validate client exists
	client, err := s.ClientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("client with ID %s not found", clientID)
	}

	// Check if user type with same type_id already exists for this client
	existing, err := s.Repo.GetByClientAndTypeID(ctx, client.ID, req.TypeID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("user type with ID %s already exists for this client", req.TypeID)
	}

	// Create the new user type
	userType := &models.ClientUserType{
		ClientID:    client.ID,
		TypeID:      req.TypeID,
		Name:        req.Name,
		Description: req.Description,
		Metadata:    req.Metadata,
		IsActive:    true,
	}

	if err := s.Repo.Create(ctx, userType); err != nil {
		return nil, err
	}

	return &dto.ClientUserTypeResponse{
		ID:          userType.ID.Hex(),
		ClientID:    userType.ClientID.Hex(),
		TypeID:      userType.TypeID,
		Name:        userType.Name,
		Description: userType.Description,
		Metadata:    userType.Metadata,
		IsActive:    userType.IsActive,
	}, nil
}

// GetUserType gets a specific user type by client_id and type_id.
func (s *ClientUserTypeService) GetUserType(ctx context.Context, clientID, typeID string) (*dto.ClientUserTypeResponse, error) {
	// Validate client exists
	client, err := s.ClientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("client with ID %s not found", clientID)
	}

	userType, err := s.Repo.GetByClientAndTypeID(ctx, client.ID, typeID)
	if err != nil {
		return nil, err
	}
	if userType == nil {
		return nil, errors.New("user type not found")
	}

	return &dto.ClientUserTypeResponse{
		ID:          userType.ID.Hex(),
		ClientID:    userType.ClientID.Hex(),
		TypeID:      userType.TypeID,
		Name:        userType.Name,
		Description: userType.Description,
		Metadata:    userType.Metadata,
		IsActive:    userType.IsActive,
	}, nil
}

// GetUserTypes gets all user types for a client.
func (s *ClientUserTypeService) GetUserTypes(ctx context.Context, clientID string, includeInactive bool) ([]dto.ClientUserTypeResponse, error) {
	// Validate client exists
	client, err := s.ClientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("client with ID %s not found", clientID)
	}

	filter := bson.M{"client": client.ID}
	if !includeInactive {
		filter["is_active"] = true
	}

	userTypes, err := s.Repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	resp := make([]dto.ClientUserTypeResponse, len(userTypes))
	for i, ut := range userTypes {
		resp[i] = dto.ClientUserTypeResponse{
			ID:          ut.ID.Hex(),
			ClientID:    ut.ClientID.Hex(),
			TypeID:      ut.TypeID,
			Name:        ut.Name,
			Description: ut.Description,
			Metadata:    ut.Metadata,
			IsActive:    ut.IsActive,
		}
	}

	return resp, nil
}

// UpdateUserType updates an existing user type.
func (s *ClientUserTypeService) UpdateUserType(ctx context.Context, clientID, typeID string, req *dto.ClientUserTypeUpdateRequest) (*dto.ClientUserTypeResponse, error) {
	// Validate client exists
	client, err := s.ClientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("client with ID %s not found", clientID)
	}

	// Get existing user type
	userType, err := s.Repo.GetByClientAndTypeID(ctx, client.ID, typeID)
	if err != nil || userType == nil {
		return nil, errors.New("user type not found")
	}

	update := bson.M{}
	if req.Name != nil {
		update["name"] = *req.Name
	}
	if req.Description != nil {
		update["description"] = *req.Description
	}
	if req.Metadata != nil {
		update["metadata"] = req.Metadata
	}
	if req.IsActive != nil {
		update["is_active"] = *req.IsActive
	}

	updated, err := s.Repo.Update(ctx, userType.ID, update)
	if err != nil {
		return nil, err
	}

	return &dto.ClientUserTypeResponse{
		ID:          updated.ID.Hex(),
		ClientID:    updated.ClientID.Hex(),
		TypeID:      updated.TypeID,
		Name:        updated.Name,
		Description: updated.Description,
		Metadata:    updated.Metadata,
		IsActive:    updated.IsActive,
	}, nil
}

// DeleteUserType soft deletes a user type by setting is_active to false.
func (s *ClientUserTypeService) DeleteUserType(ctx context.Context, clientID, typeID string) error {
	// Validate client exists
	client, err := s.ClientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return fmt.Errorf("client with ID %s not found", clientID)
	}

	// Get existing user type
	userType, err := s.Repo.GetByClientAndTypeID(ctx, client.ID, typeID)
	if err != nil || userType == nil {
		return errors.New("user type not found")
	}

	update := bson.M{"is_active": false}
	_, err = s.Repo.Update(ctx, userType.ID, update)
	return err
}

// GetSenderTypeID generates the full sender_type ID for a client user type.
// This is used in the sender_type field of chat messages.
func (s *ClientUserTypeService) GetSenderTypeID(clientID, typeID string) string {
	return fmt.Sprintf("client:%s:%s", clientID, typeID)
}

// ParseSenderType parses a sender_type string to extract client_id and type_id.
// Returns (client_id, type_id) if it's a valid client user type, or ("", "") if not.
func (s *ClientUserTypeService) ParseSenderType(senderType string) (string, string) {
	if senderType == "" || !strings.HasPrefix(senderType, "client:") {
		return "", ""
	}

	parts := strings.Split(senderType, ":")
	if len(parts) != 3 {
		return "", ""
	}

	return parts[1], parts[2] // client_id, type_id
}