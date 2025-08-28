package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EventDeliveryAttempt tracks individual delivery attempts for event deliveries
type EventDeliveryAttempt struct {
	ID              primitive.ObjectID     `bson:"_id,omitempty" json:"id,omitempty"`
	EventDeliveryID primitive.ObjectID     `bson:"event_delivery" json:"event_delivery_id" validate:"required"`
	AttemptNumber   int                   `bson:"attempt_number" json:"attempt_number"`
	Status          DeliveryStatus        `bson:"status" json:"status"`
	RequestPayload  map[string]interface{} `bson:"request_payload,omitempty" json:"request_payload,omitempty"`
	ResponsePayload map[string]interface{} `bson:"response_payload,omitempty" json:"response_payload,omitempty"`
	StatusCode      int                   `bson:"status_code,omitempty" json:"status_code,omitempty"`
	ErrorMessage    string                `bson:"error_message,omitempty" json:"error_message,omitempty"`
	StartedAt       time.Time             `bson:"started_at" json:"started_at"`
	CompletedAt     *time.Time            `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
	CreatedAt       time.Time             `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time             `bson:"updated_at" json:"updated_at"`
}

// TableName returns the collection name for EventDeliveryAttempt
func (EventDeliveryAttempt) TableName() string {
	return "event_delivery_attempts"
}

// BeforeCreate sets the timestamps before creating
func (eda *EventDeliveryAttempt) BeforeCreate() {
	now := time.Now().UTC()
	eda.CreatedAt = now
	eda.UpdatedAt = now
	eda.StartedAt = now
	if eda.ID.IsZero() {
		eda.ID = primitive.NewObjectID()
	}
	if eda.Status == "" {
		eda.Status = DeliveryStatusInProgress
	}
	if eda.RequestPayload == nil {
		eda.RequestPayload = make(map[string]interface{})
	}
	if eda.ResponsePayload == nil {
		eda.ResponsePayload = make(map[string]interface{})
	}
}

// BeforeUpdate sets the updated timestamp before updating
func (eda *EventDeliveryAttempt) BeforeUpdate() {
	eda.UpdatedAt = time.Now().UTC()
}

// MarkCompleted marks the attempt as completed
func (eda *EventDeliveryAttempt) MarkCompleted(status DeliveryStatus) {
	now := time.Now().UTC()
	eda.Status = status
	eda.CompletedAt = &now
	eda.BeforeUpdate()
}

// MarkFailed marks the attempt as failed with an error message
func (eda *EventDeliveryAttempt) MarkFailed(errorMessage string) {
	eda.ErrorMessage = errorMessage
	eda.MarkCompleted(DeliveryStatusFailed)
}