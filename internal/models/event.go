package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Event represents a system event with parent-child relationships
type Event struct {
	ID         primitive.ObjectID     `bson:"_id,omitempty" json:"id,omitempty"`
	EventType  EventType             `bson:"event_type" json:"event_type" validate:"required"`
	EntityType EntityType            `bson:"entity_type" json:"entity_type" validate:"required"`
	EntityID   string                `bson:"entity_id" json:"entity_id" validate:"required"`
	ParentID   string                `bson:"parent_id,omitempty" json:"parent_id,omitempty"`
	Data       map[string]interface{} `bson:"data" json:"data"`
	CreatedAt  time.Time             `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time             `bson:"updated_at" json:"updated_at"`
}

// TableName returns the collection name for Event
func (Event) TableName() string {
	return "events"
}

// BeforeCreate sets the timestamps before creating
func (e *Event) BeforeCreate() {
	now := time.Now().UTC()
	e.CreatedAt = now
	e.UpdatedAt = now
	if e.ID.IsZero() {
		e.ID = primitive.NewObjectID()
	}
	if e.Data == nil {
		e.Data = make(map[string]interface{})
	}
}

// BeforeUpdate sets the updated timestamp before updating
func (e *Event) BeforeUpdate() {
	e.UpdatedAt = time.Now().UTC()
}