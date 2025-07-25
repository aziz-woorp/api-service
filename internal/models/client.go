// Package models defines the MongoDB model for clients.
package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Client represents a client entity.
type Client struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Email     *string            `bson:"email,omitempty" json:"email,omitempty"`
	ClientID  string             `bson:"client_id" json:"client_id"`
	ClientKey string             `bson:"client_key" json:"client_key"`
	IsActive  bool               `bson:"is_active" json:"is_active"`
}
