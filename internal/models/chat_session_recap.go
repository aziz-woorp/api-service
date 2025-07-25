// Package models defines the MongoDB model for chat session recaps.
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatSessionRecap represents a recap for a chat session.
type ChatSessionRecap struct {
	ID        primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	SessionID primitive.ObjectID     `bson:"session_id" json:"session_id"`
	RecapData map[string]interface{} `bson:"recap_data" json:"recap_data"`
	CreatedAt time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time              `bson:"updated_at" json:"updated_at"`
}
