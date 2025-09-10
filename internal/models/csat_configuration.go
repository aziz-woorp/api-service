// Package models defines the MongoDB model for CSAT configurations.
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CSATConfiguration represents client-specific CSAT configuration with type support
type CSATConfiguration struct {
	ID               primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Client           primitive.ObjectID     `bson:"client" json:"client" validate:"required"`
	ClientChannel    primitive.ObjectID     `bson:"client_channel" json:"client_channel" validate:"required"`
	Type             string                 `bson:"type" json:"type" validate:"required"`
	Enabled          bool                   `bson:"enabled" json:"enabled"`
	TriggerConditions map[string]interface{} `bson:"trigger_conditions,omitempty" json:"trigger_conditions,omitempty"`
	CreatedAt        time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time              `bson:"updated_at" json:"updated_at"`
}

// TableName returns the MongoDB collection name for CSATConfiguration.
func (CSATConfiguration) TableName() string {
	return "csat_configurations"
}

// BeforeCreate sets the timestamps before creating
func (c *CSATConfiguration) BeforeCreate() {
	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now
	if c.ID.IsZero() {
		c.ID = primitive.NewObjectID()
	}
	if c.TriggerConditions == nil {
		c.TriggerConditions = make(map[string]interface{})
	}
}

// BeforeUpdate sets the updated timestamp before updating
func (c *CSATConfiguration) BeforeUpdate() {
	c.UpdatedAt = time.Now().UTC()
}
