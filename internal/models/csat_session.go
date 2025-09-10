// Package models defines the MongoDB model for CSAT sessions.
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CSATSession represents a CSAT session to track progress
type CSATSession struct {
	ID                   primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	ChatSessionID        string               `bson:"chat_session_id" json:"chat_session_id" validate:"required"`
	Client               primitive.ObjectID   `bson:"client" json:"client" validate:"required"`
	ClientChannel        primitive.ObjectID   `bson:"client_channel" json:"client_channel" validate:"required"`
	ThreadSessionID      *string              `bson:"thread_session_id,omitempty" json:"thread_session_id,omitempty"`
	ThreadContext        bool                 `bson:"thread_context" json:"thread_context"`
	Status               string               `bson:"status" json:"status"` // "pending", "in_progress", "completed", "abandoned"
	TriggeredAt          time.Time            `bson:"triggered_at" json:"triggered_at"`
	CompletedAt          *time.Time           `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
	CurrentQuestionIndex int                  `bson:"current_question_index" json:"current_question_index"`
	QuestionsSent        []string             `bson:"questions_sent" json:"questions_sent"`
	CreatedAt            time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt            time.Time            `bson:"updated_at" json:"updated_at"`
}

// TableName returns the MongoDB collection name for CSATSession.
func (CSATSession) TableName() string {
	return "csat_sessions"
}

// BeforeCreate sets the timestamps before creating
func (s *CSATSession) BeforeCreate() {
	now := time.Now().UTC()
	s.CreatedAt = now
	s.UpdatedAt = now
	s.TriggeredAt = now
	if s.ID.IsZero() {
		s.ID = primitive.NewObjectID()
	}
	if s.Status == "" {
		s.Status = "pending"
	}
	if s.QuestionsSent == nil {
		s.QuestionsSent = make([]string, 0)
	}
}

// BeforeUpdate sets the updated timestamp before updating
func (s *CSATSession) BeforeUpdate() {
	s.UpdatedAt = time.Now().UTC()
}
