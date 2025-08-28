// Package service provides business logic for event delivery tracking.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/fraiday-org/api-service/internal/models"
	"github.com/fraiday-org/api-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EventDeliveryTrackingService encapsulates business logic for event delivery tracking.
type EventDeliveryTrackingService struct {
	DeliveryRepo *repository.EventDeliveryRepository
	AttemptRepo  *repository.EventDeliveryAttemptRepository
}

// NewEventDeliveryTrackingService creates a new EventDeliveryTrackingService.
func NewEventDeliveryTrackingService(
	deliveryRepo *repository.EventDeliveryRepository,
	attemptRepo *repository.EventDeliveryAttemptRepository,
) *EventDeliveryTrackingService {
	return &EventDeliveryTrackingService{
		DeliveryRepo: deliveryRepo,
		AttemptRepo:  attemptRepo,
	}
}

// CreateDeliveryRecord creates a new event delivery record.
func (s *EventDeliveryTrackingService) CreateDeliveryRecord(
	ctx context.Context,
	eventID primitive.ObjectID,
	processorConfigID primitive.ObjectID,
	requestPayload map[string]interface{},
	maxAttempts int,
) (*models.EventDelivery, error) {
	delivery := &models.EventDelivery{
		EventID:                eventID,
		EventProcessorConfigID: processorConfigID,
		Status:                 models.DeliveryStatusPending,
		MaxAttempts:            maxAttempts,
		CurrentAttempts:        0,
		RequestPayload:         requestPayload,
	}

	if err := s.DeliveryRepo.Create(ctx, delivery); err != nil {
		return nil, fmt.Errorf("failed to create delivery record: %w", err)
	}

	return delivery, nil
}

// RecordDeliveryAttempt records a delivery attempt with its result.
func (s *EventDeliveryTrackingService) RecordDeliveryAttempt(
	ctx context.Context,
	deliveryID primitive.ObjectID,
	attemptNumber int,
	status models.DeliveryStatus,
	statusCode *int,
	errorMessage string,
	requestPayload map[string]interface{},
	responsePayload map[string]interface{},
) (*models.EventDeliveryAttempt, error) {
	// First increment the attempt count
	if err := s.DeliveryRepo.IncrementAttempts(ctx, deliveryID); err != nil {
		return nil, fmt.Errorf("failed to increment attempts: %w", err)
	}

	// Create the attempt record
	attempt := &models.EventDeliveryAttempt{
		EventDeliveryID: deliveryID,
		AttemptNumber:   attemptNumber,
		Status:          status,
		RequestPayload:  requestPayload,
		ResponsePayload: responsePayload,
		ErrorMessage:    errorMessage,
		StartedAt:       time.Now(),
	}

	if statusCode != nil {
		attempt.StatusCode = *statusCode
	}

	if err := s.AttemptRepo.Create(ctx, attempt); err != nil {
		return nil, fmt.Errorf("failed to create attempt record: %w", err)
	}

	// Update delivery status based on attempt result
	if err := s.updateDeliveryStatusFromAttempt(ctx, deliveryID, status); err != nil {
		return nil, fmt.Errorf("failed to update delivery status: %w", err)
	}

	return attempt, nil
}

// updateDeliveryStatusFromAttempt updates the delivery status based on the latest attempt.
func (s *EventDeliveryTrackingService) updateDeliveryStatusFromAttempt(
	ctx context.Context,
	deliveryID primitive.ObjectID,
	attemptStatus models.DeliveryStatus,
) error {
	// Get the current delivery to check attempts
	delivery, err := s.DeliveryRepo.GetByID(ctx, deliveryID)
	if err != nil {
		return fmt.Errorf("failed to get delivery: %w", err)
	}

	var newStatus models.DeliveryStatus
	switch attemptStatus {
	case models.DeliveryStatusCompleted:
		newStatus = models.DeliveryStatusCompleted
	case models.DeliveryStatusFailed:
		if delivery.CurrentAttempts >= delivery.MaxAttempts {
			newStatus = models.DeliveryStatusFailed
		} else {
			newStatus = models.DeliveryStatusPending // Will retry
		}
	default:
		// For other statuses like InProgress, keep current status
		return nil
	}

	return s.DeliveryRepo.UpdateStatus(ctx, deliveryID, newStatus)
}

// GetDeliveryByID retrieves a delivery record by its ID.
func (s *EventDeliveryTrackingService) GetDeliveryByID(ctx context.Context, deliveryID string) (*models.EventDelivery, error) {
	id, err := primitive.ObjectIDFromHex(deliveryID)
	if err != nil {
		return nil, fmt.Errorf("invalid delivery ID: %w", err)
	}

	delivery, err := s.DeliveryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery: %w", err)
	}

	return delivery, nil
}

// GetDeliveriesForEvent retrieves all delivery records for a specific event.
func (s *EventDeliveryTrackingService) GetDeliveriesForEvent(ctx context.Context, eventID string) ([]models.EventDelivery, error) {
	id, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		return nil, fmt.Errorf("invalid event ID: %w", err)
	}

	deliveries, err := s.DeliveryRepo.GetByEventID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get deliveries for event: %w", err)
	}

	return deliveries, nil
}

// GetAttemptsForDelivery retrieves all attempts for a specific delivery.
func (s *EventDeliveryTrackingService) GetAttemptsForDelivery(ctx context.Context, deliveryID string) ([]models.EventDeliveryAttempt, error) {
	id, err := primitive.ObjectIDFromHex(deliveryID)
	if err != nil {
		return nil, fmt.Errorf("invalid delivery ID: %w", err)
	}

	attempts, err := s.AttemptRepo.GetByDeliveryID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get attempts for delivery: %w", err)
	}

	return attempts, nil
}

// GetPendingDeliveries retrieves deliveries that need to be processed or retried.
func (s *EventDeliveryTrackingService) GetPendingDeliveries(ctx context.Context) ([]models.EventDelivery, error) {
	deliveries, err := s.DeliveryRepo.GetPendingDeliveries(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending deliveries: %w", err)
	}

	return deliveries, nil
}

// GetDeliveryStats returns statistics about deliveries.
func (s *EventDeliveryTrackingService) GetDeliveryStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count by status
	statuses := []models.DeliveryStatus{
		models.DeliveryStatusPending,
		models.DeliveryStatusInProgress,
		models.DeliveryStatusCompleted,
		models.DeliveryStatusFailed,
	}

	for _, status := range statuses {
		filter := map[string]interface{}{"status": status}
		count, err := s.DeliveryRepo.Count(ctx, filter)
		if err != nil {
			return nil, fmt.Errorf("failed to count deliveries with status %s: %w", status, err)
		}
		stats[string(status)] = count
	}

	// Total count
	totalCount, err := s.DeliveryRepo.Count(ctx, map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to count total deliveries: %w", err)
	}
	stats["total"] = totalCount

	return stats, nil
}

// RetryFailedDelivery marks a failed delivery for retry by resetting its status.
func (s *EventDeliveryTrackingService) RetryFailedDelivery(ctx context.Context, deliveryID string) error {
	id, err := primitive.ObjectIDFromHex(deliveryID)
	if err != nil {
		return fmt.Errorf("invalid delivery ID: %w", err)
	}

	// Get current delivery to check if it can be retried
	delivery, err := s.DeliveryRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get delivery: %w", err)
	}

	if delivery.Status != models.DeliveryStatusFailed {
		return fmt.Errorf("delivery is not in failed status, cannot retry")
	}

	if delivery.CurrentAttempts >= delivery.MaxAttempts {
		return fmt.Errorf("delivery has exceeded maximum attempts")
	}

	// Reset status to pending for retry
	return s.DeliveryRepo.UpdateStatus(ctx, id, models.DeliveryStatusPending)
}