// Package models defines the MongoDB model for chat message feedback.
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatMessageFeedback represents feedback for a chat message.
type ChatMessageFeedback struct {
	ID            primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	ChatMessageID primitive.ObjectID     `bson:"chat_message_id" json:"chat_message_id"`
	Rating        int                    `bson:"rating" json:"rating"`
	Comment       *string                `bson:"comment,omitempty" json:"comment,omitempty"`
	Metadata      map[string]interface{} `bson:"metadata" json:"metadata"`
	CreatedAt     time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time              `bson:"updated_at" json:"updated_at"`
}
