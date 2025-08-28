package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ClientChannel represents a client's communication channel configuration
type ClientChannel struct {
	ID            primitive.ObjectID     `bson:"_id,omitempty" json:"id,omitempty"`
	ChannelType   ChannelType           `bson:"channel_type" json:"channel_type" validate:"required"`
	ChannelConfig map[string]interface{} `bson:"channel_config" json:"channel_config" validate:"required"`
	ClientID      primitive.ObjectID     `bson:"client" json:"client_id" validate:"required"`
	IsActive      bool                  `bson:"is_active" json:"is_active"`
	CreatedAt     time.Time             `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time             `bson:"updated_at" json:"updated_at"`
}

// TableName returns the collection name for ClientChannel
func (ClientChannel) TableName() string {
	return "client_channels"
}

// BeforeCreate sets the timestamps before creating
func (cc *ClientChannel) BeforeCreate() {
	now := time.Now().UTC()
	cc.CreatedAt = now
	cc.UpdatedAt = now
	if cc.ID.IsZero() {
		cc.ID = primitive.NewObjectID()
	}
	if !cc.IsActive {
		cc.IsActive = true // Default to active
	}
}

// BeforeUpdate sets the updated timestamp before updating
func (cc *ClientChannel) BeforeUpdate() {
	cc.UpdatedAt = time.Now().UTC()
}