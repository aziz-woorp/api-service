// Package models defines the MongoDB model for chat session threads.
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatSessionThread represents a thread within a chat session.
type ChatSessionThread struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ThreadID         string             `bson:"thread_id" json:"thread_id"`
	ThreadSessionID  string             `bson:"thread_session_id" json:"thread_session_id"`
	ParentSessionID  string             `bson:"parent_session_id" json:"parent_session_id"`
	ChatSessionID    primitive.ObjectID `bson:"chat_session_id" json:"chat_session_id"`
	Active           bool               `bson:"active" json:"active"`
	LastActivity     time.Time          `bson:"last_activity" json:"last_activity"`
}
