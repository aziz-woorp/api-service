// Package models defines the MongoDB model for CSAT responses.
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CSATResponse represents a CSAT response storage
type CSATResponse struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CSATSession      primitive.ObjectID `bson:"csat_session" json:"csat_session" validate:"required"`
	QuestionTemplate primitive.ObjectID `bson:"question_template" json:"question_template" validate:"required"`
	ResponseValue    string             `bson:"response_value" json:"response_value" validate:"required"`
	RespondedAt      time.Time          `bson:"responded_at" json:"responded_at"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}

// TableName returns the MongoDB collection name for CSATResponse.
func (CSATResponse) TableName() string {
	return "csat_responses"
}

// BeforeCreate sets the timestamps before creating
func (r *CSATResponse) BeforeCreate() {
	now := time.Now().UTC()
	r.CreatedAt = now
	r.UpdatedAt = now
	r.RespondedAt = now
	if r.ID.IsZero() {
		r.ID = primitive.NewObjectID()
	}
}

// BeforeUpdate sets the updated timestamp before updating
func (r *CSATResponse) BeforeUpdate() {
	r.UpdatedAt = time.Now().UTC()
}
