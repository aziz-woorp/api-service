// Package models defines the MongoDB model for chat sessions.
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatSession represents a chat session document.
type ChatSession struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	SessionID     string               `bson:"session_id" json:"session_id"`
	CreatedAt     time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time            `bson:"updated_at" json:"updated_at"`
	Active        bool                 `bson:"active" json:"active"`
	Client        *primitive.ObjectID  `bson:"client,omitempty" json:"client,omitempty"`
	ClientChannel *primitive.ObjectID  `bson:"client_channel,omitempty" json:"client_channel,omitempty"`
	Participants  []string             `bson:"participants,omitempty" json:"participants,omitempty"`
}
