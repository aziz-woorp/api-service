package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EventDelivery tracks the delivery of an event to a specific processor
type EventDelivery struct {
	ID                     primitive.ObjectID     `bson:"_id,omitempty" json:"id,omitempty"`
	EventID                primitive.ObjectID     `bson:"event" json:"event_id" validate:"required"`
	EventProcessorConfigID primitive.ObjectID     `bson:"event_processor_config" json:"event_processor_config_id" validate:"required"`
	Status                 DeliveryStatus        `bson:"status" json:"status"`
	MaxAttempts            int                   `bson:"max_attempts" json:"max_attempts"`
	CurrentAttempts        int                   `bson:"current_attempts" json:"current_attempts"`
	RequestPayload         map[string]interface{} `bson:"request_payload,omitempty" json:"request_payload,omitempty"`
	CreatedAt              time.Time             `bson:"created_at" json:"created_at"`
	UpdatedAt              time.Time             `bson:"updated_at" json:"updated_at"`
}

// TableName returns the collection name for EventDelivery
func (EventDelivery) TableName() string {
	return "event_deliveries"
}

// BeforeCreate sets the timestamps before creating
func (ed *EventDelivery) BeforeCreate() {
	now := time.Now().UTC()
	ed.CreatedAt = now
	ed.UpdatedAt = now
	if ed.ID.IsZero() {
		ed.ID = primitive.NewObjectID()
	}
	if ed.Status == "" {
		ed.Status = DeliveryStatusPending
	}
	if ed.MaxAttempts == 0 {
		ed.MaxAttempts = 3
	}
	if ed.RequestPayload == nil {
		ed.RequestPayload = make(map[string]interface{})
	}
}

// BeforeUpdate sets the updated timestamp before updating
func (ed *EventDelivery) BeforeUpdate() {
	ed.UpdatedAt = time.Now().UTC()
}

// CanRetry checks if the delivery can be retried
func (ed *EventDelivery) CanRetry() bool {
	return ed.CurrentAttempts < ed.MaxAttempts && ed.Status == DeliveryStatusFailed
}

// IncrementAttempts increments the current attempts counter
func (ed *EventDelivery) IncrementAttempts() {
	ed.CurrentAttempts++
	ed.BeforeUpdate()
}