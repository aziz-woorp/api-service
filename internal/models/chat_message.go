// Package models contains MongoDB document models for the API service.
package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// MessageCategory represents the type of message.
type MessageCategory string

const (
	MessageCategoryMessage MessageCategory = "message"
	MessageCategoryError   MessageCategory = "error"
	MessageCategoryInfo    MessageCategory = "info"
	MessageCategoryWarning MessageCategory = "warning"
)

// SenderType represents the sender of the message.
type SenderType string

const (
	SenderTypeUser      SenderType = "user"
	SenderTypeAssistant SenderType = "assistant"
	SenderTypeSystem    SenderType = "system"
	// Custom client types are prefixed with "client:"
)

// Attachment represents a file/image attached to a chat message.
type Attachment struct {
	FileName string                   `bson:"file_name,omitempty" json:"file_name,omitempty"`
	FileType string                   `bson:"file_type,omitempty" json:"file_type,omitempty"`
	FileSize int64                    `bson:"file_size,omitempty" json:"file_size,omitempty"`
	FileURL  string                   `bson:"file_url,omitempty" json:"file_url,omitempty"`
	Type     string                   `bson:"type,omitempty" json:"type,omitempty"` // "file", "image", "carousel", "buttons"
	Carousel map[string]interface{}   `bson:"carousel,omitempty" json:"carousel,omitempty"`
	Buttons  []map[string]interface{} `bson:"buttons,omitempty" json:"buttons,omitempty"` // For postback/reply buttons
}

// ChatMessage represents a chat message document in MongoDB.
type ChatMessage struct {
	ID             primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	ExternalID     string                 `bson:"external_id,omitempty" json:"external_id,omitempty"`
	Sender         string                 `bson:"sender" json:"sender"`
	SenderName     string                 `bson:"sender_name,omitempty" json:"sender_name,omitempty"`
	SenderType     string                 `bson:"sender_type" json:"sender_type"`
	SessionID      primitive.ObjectID     `bson:"session,omitempty" json:"session"` // Reference to ChatSession
	Text           string                 `bson:"text" json:"text"`
	Attachments    []Attachment           `bson:"attachments,omitempty" json:"attachments,omitempty"`
	Data           map[string]interface{} `bson:"data,omitempty" json:"data,omitempty"`
	Category       MessageCategory        `bson:"category" json:"category"`
	Config         map[string]interface{} `bson:"config,omitempty" json:"config,omitempty"`
	Confidence     float64                `bson:"confidence_score,omitempty" json:"confidence_score,omitempty"`
	Edit           bool                   `bson:"edit,omitempty" json:"edit,omitempty"`
	CreatedAt      time.Time              `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt      time.Time              `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}

// TableName returns the MongoDB collection name for ChatMessage.
func (ChatMessage) TableName() string {
	return "chat_messages"
}
