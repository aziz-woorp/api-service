package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ClientUserType represents custom user types for clients
// These custom types can be used as sender_type in chat messages
type ClientUserType struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty" json:"id,omitempty"`
	ClientID    primitive.ObjectID     `bson:"client" json:"client_id" validate:"required"`
	TypeID      string                `bson:"type_id" json:"type_id" validate:"required"`        // Unique identifier for this user type within the client
	Name        string                `bson:"name" json:"name" validate:"required"`             // Display name for the user type
	Description string                `bson:"description,omitempty" json:"description,omitempty"` // Optional description
	Metadata    map[string]interface{} `bson:"metadata" json:"metadata"`                        // Additional metadata for the user type
	IsActive    bool                  `bson:"is_active" json:"is_active"`                      // Controls whether this user type is active
	CreatedAt   time.Time             `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time             `bson:"updated_at" json:"updated_at"`
}

// TableName returns the collection name for ClientUserType
func (ClientUserType) TableName() string {
	return "client_user_types"
}

// BeforeCreate sets the timestamps before creating
func (cut *ClientUserType) BeforeCreate() {
	now := time.Now().UTC()
	cut.CreatedAt = now
	cut.UpdatedAt = now
	if cut.ID.IsZero() {
		cut.ID = primitive.NewObjectID()
	}
	if cut.Metadata == nil {
		cut.Metadata = make(map[string]interface{})
	}
	cut.IsActive = true // Default to active
}

// BeforeUpdate sets the updated timestamp before updating
func (cut *ClientUserType) BeforeUpdate() {
	cut.UpdatedAt = time.Now().UTC()
}