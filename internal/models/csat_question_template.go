// Package models defines the MongoDB model for CSAT question templates.
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CSATQuestionTemplate represents a CSAT question template
type CSATQuestionTemplate struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Client        primitive.ObjectID `bson:"client" json:"client" validate:"required"`
	ClientChannel primitive.ObjectID `bson:"client_channel" json:"client_channel" validate:"required"`
	QuestionText  string             `bson:"question_text" json:"question_text" validate:"required"`
	Options       []string           `bson:"options" json:"options" validate:"required"`
	Order         int                `bson:"order" json:"order" validate:"required"`
	Active        bool               `bson:"active" json:"active"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

// TableName returns the MongoDB collection name for CSATQuestionTemplate.
func (CSATQuestionTemplate) TableName() string {
	return "csat_question_templates"
}

// BeforeCreate sets the timestamps before creating
func (q *CSATQuestionTemplate) BeforeCreate() {
	now := time.Now().UTC()
	q.CreatedAt = now
	q.UpdatedAt = now
	if q.ID.IsZero() {
		q.ID = primitive.NewObjectID()
	}
	if q.Options == nil {
		q.Options = make([]string, 0)
	}
}

// BeforeUpdate sets the updated timestamp before updating
func (q *CSATQuestionTemplate) BeforeUpdate() {
	q.UpdatedAt = time.Now().UTC()
}
