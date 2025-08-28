package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatMessageSuggestion represents AI-generated message suggestions
type ChatMessageSuggestion struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ChatSessionID primitive.ObjectID `bson:"chat_session" json:"chat_session_id" validate:"required"`
	ClientID      primitive.ObjectID `bson:"client" json:"client_id" validate:"required"`
	SuggestionText string            `bson:"suggestion_text" json:"suggestion_text" validate:"required"`
	ConfidenceScore float64          `bson:"confidence_score,omitempty" json:"confidence_score,omitempty"`
	SuggestionType  string           `bson:"suggestion_type,omitempty" json:"suggestion_type,omitempty"`
	Metadata       map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
	IsUsed         bool             `bson:"is_used" json:"is_used"`
	UsedAt         *time.Time       `bson:"used_at,omitempty" json:"used_at,omitempty"`
	CreatedAt      time.Time        `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time        `bson:"updated_at" json:"updated_at"`
}

// TableName returns the collection name for ChatMessageSuggestion
func (ChatMessageSuggestion) TableName() string {
	return "chat_message_suggestions"
}

// BeforeCreate sets the timestamps before creating
func (cms *ChatMessageSuggestion) BeforeCreate() {
	now := time.Now().UTC()
	cms.CreatedAt = now
	cms.UpdatedAt = now
	if cms.ID.IsZero() {
		cms.ID = primitive.NewObjectID()
	}
	if cms.Metadata == nil {
		cms.Metadata = make(map[string]interface{})
	}
}

// BeforeUpdate sets the updated timestamp before updating
func (cms *ChatMessageSuggestion) BeforeUpdate() {
	cms.UpdatedAt = time.Now().UTC()
}

// MarkAsUsed marks the suggestion as used
func (cms *ChatMessageSuggestion) MarkAsUsed() {
	now := time.Now().UTC()
	cms.IsUsed = true
	cms.UsedAt = &now
	cms.BeforeUpdate()
}