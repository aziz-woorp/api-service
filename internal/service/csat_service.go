// Package service provides business logic for CSAT surveys.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/fraiday-org/api-service/internal/models"
	"github.com/fraiday-org/api-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CSATService encapsulates business logic for CSAT surveys.
type CSATService struct {
	CSATConfigRepo        *repository.CSATConfigurationRepository
	CSATQuestionRepo      *repository.CSATQuestionTemplateRepository
	CSATSessionRepo       *repository.CSATSessionRepository
	CSATResponseRepo      *repository.CSATResponseRepository
	ChatMessageRepo       *repository.ChatMessageRepository
	EventPublisherService *EventPublisherService
}

// NewCSATService creates a new CSATService.
func NewCSATService(
	configRepo *repository.CSATConfigurationRepository,
	questionRepo *repository.CSATQuestionTemplateRepository,
	sessionRepo *repository.CSATSessionRepository,
	responseRepo *repository.CSATResponseRepository,
	chatMessageRepo *repository.ChatMessageRepository,
	eventPublisher *EventPublisherService,
) *CSATService {
	return &CSATService{
		CSATConfigRepo:        configRepo,
		CSATQuestionRepo:      questionRepo,
		CSATSessionRepo:       sessionRepo,
		CSATResponseRepo:      responseRepo,
		ChatMessageRepo:       chatMessageRepo,
		EventPublisherService: eventPublisher,
	}
}

// TriggerCSATSurvey triggers a CSAT survey for a chat session.
func (s *CSATService) TriggerCSATSurvey(ctx context.Context, chatSessionID string, clientID, channelID primitive.ObjectID) (*models.CSATSession, error) {
	// Check if CSAT is enabled for this client and channel
	config, err := s.CSATConfigRepo.GetByClientAndChannel(ctx, clientID, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get CSAT configuration: %w", err)
	}
	
	if !config.Enabled {
		return nil, fmt.Errorf("CSAT is not enabled for this client and channel")
	}
	
	// Check if there's already an active CSAT session for this chat session
	existingSession, err := s.CSATSessionRepo.GetActiveByChatSessionID(ctx, chatSessionID)
	if err == nil && existingSession != nil {
		return nil, fmt.Errorf("CSAT session already active for this chat session")
	}
	
	// Create new CSAT session
	csatSession := &models.CSATSession{
		ChatSessionID:        chatSessionID,
		Client:               clientID,
		ClientChannel:        channelID,
		Status:               "pending",
		CurrentQuestionIndex: 0,
		QuestionsSent:        make([]string, 0),
	}
	
	if err := s.CSATSessionRepo.Create(ctx, csatSession); err != nil {
		return nil, fmt.Errorf("failed to create CSAT session: %w", err)
	}
	
	// Publish CSAT triggered event
	eventData := map[string]interface{}{
		"csat_session_id": csatSession.ID.Hex(),
		"chat_session_id": chatSessionID,
		"client_id":       clientID.Hex(),
		"channel_id":      channelID.Hex(),
	}
	
	_, err = s.EventPublisherService.PublishEvent(
		ctx,
		models.EventTypeCSATTriggered,
		models.EntityTypeCSATSession,
		csatSession.ID.Hex(),
		nil,
		eventData,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to publish CSAT triggered event: %w", err)
	}
	
	// Send first question
	if err := s.SendNextQuestion(ctx, csatSession.ID); err != nil {
		return nil, fmt.Errorf("failed to send first question: %w", err)
	}
	
	return csatSession, nil
}

// SendNextQuestion sends the next question in the CSAT survey.
func (s *CSATService) SendNextQuestion(ctx context.Context, sessionID primitive.ObjectID) error {
	// Get the CSAT session
	session, err := s.CSATSessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get CSAT session: %w", err)
	}
	
	// Get all questions for this client and channel
	questions, err := s.CSATQuestionRepo.GetByClientAndChannel(ctx, session.Client, session.ClientChannel)
	if err != nil {
		return fmt.Errorf("failed to get CSAT questions: %w", err)
	}
	
	if len(questions) == 0 {
		return fmt.Errorf("no CSAT questions configured for this client and channel")
	}
	
	// Check if we've sent all questions
	if session.CurrentQuestionIndex >= len(questions) {
		return s.CompleteCSATSurvey(ctx, sessionID)
	}
	
	// Get the current question
	currentQuestion := questions[session.CurrentQuestionIndex]
	
	// Create chat message with question
	chatMessage, err := s.createQuestionMessage(session, &currentQuestion)
	if err != nil {
		return fmt.Errorf("failed to create question message: %w", err)
	}
	
	// Save the chat message
	if err := s.ChatMessageRepo.Create(ctx, chatMessage); err != nil {
		return fmt.Errorf("failed to save question message: %w", err)
	}
	
	// Update session status and add question to sent list
	session.Status = "in_progress"
	session.QuestionsSent = append(session.QuestionsSent, currentQuestion.ID.Hex())
	if err := s.CSATSessionRepo.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to update CSAT session: %w", err)
	}
	
	// Publish CSAT message sent event
	eventData := map[string]interface{}{
		"csat_session_id": session.ID.Hex(),
		"question_id":     currentQuestion.ID.Hex(),
		"message_id":      chatMessage.ID.Hex(),
		"chat_session_id": session.ChatSessionID,
	}
	
	sessionIDHex := session.ID.Hex()
	_, err = s.EventPublisherService.PublishEvent(
		ctx,
		models.EventTypeCSATMessageSent,
		models.EntityTypeCSATQuestion,
		currentQuestion.ID.Hex(),
		&sessionIDHex,
		eventData,
	)
	if err != nil {
		return fmt.Errorf("failed to publish CSAT message sent event: %w", err)
	}
	
	return nil
}

// ProcessResponse processes a user response to a CSAT question.
func (s *CSATService) ProcessResponse(ctx context.Context, sessionID primitive.ObjectID, questionID primitive.ObjectID, responseValue string) error {
	// Get the CSAT session
	session, err := s.CSATSessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get CSAT session: %w", err)
	}
	
	if session.Status != "in_progress" {
		return fmt.Errorf("CSAT session is not in progress")
	}
	
	// Save the response
	response := &models.CSATResponse{
		CSATSession:      sessionID,
		QuestionTemplate: questionID,
		ResponseValue:    responseValue,
	}
	
	if err := s.CSATResponseRepo.Create(ctx, response); err != nil {
		return fmt.Errorf("failed to save CSAT response: %w", err)
	}
	
	// Move to next question
	session.CurrentQuestionIndex++
	if err := s.CSATSessionRepo.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to update CSAT session: %w", err)
	}
	
	// Send next question or complete survey
	return s.SendNextQuestion(ctx, sessionID)
}

// CompleteCSATSurvey completes the CSAT survey.
func (s *CSATService) CompleteCSATSurvey(ctx context.Context, sessionID primitive.ObjectID) error {
	// Get the CSAT session
	session, err := s.CSATSessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get CSAT session: %w", err)
	}
	
	// Update session status
	now := time.Now().UTC()
	session.Status = "completed"
	session.CompletedAt = &now
	if err := s.CSATSessionRepo.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to update CSAT session: %w", err)
	}
	
	// Create thank you message
	thankYouMessage := &models.ChatMessage{
		Sender:     "system",
		SenderName: "CSAT Survey",
		SenderType: string(models.SenderTypeSystem),
		SessionID:  primitive.NilObjectID, // Will be set based on chat session
		Text:       "Thank you for your feedback!",
		Category:   models.MessageCategoryInfo,
		Data: map[string]interface{}{
			"csat_message":    true,
			"csat_session_id": session.ID.Hex(),
			"message_type":    "completion",
		},
	}
	
	// Save the thank you message
	if err := s.ChatMessageRepo.Create(ctx, thankYouMessage); err != nil {
		return fmt.Errorf("failed to save thank you message: %w", err)
	}
	
	// Publish CSAT completed event
	eventData := map[string]interface{}{
		"csat_session_id": session.ID.Hex(),
		"chat_session_id": session.ChatSessionID,
		"completed_at":    session.CompletedAt,
	}
	
	_, err = s.EventPublisherService.PublishEvent(
		ctx,
		models.EventTypeCSATCompleted,
		models.EntityTypeCSATSession,
		session.ID.Hex(),
		nil,
		eventData,
	)
	if err != nil {
		return fmt.Errorf("failed to publish CSAT completed event: %w", err)
	}
	
	return nil
}

// createQuestionMessage creates a chat message for a CSAT question.
func (s *CSATService) createQuestionMessage(session *models.CSATSession, question *models.CSATQuestionTemplate) (*models.ChatMessage, error) {
	// Create button attachments for all questions
	var attachments []models.Attachment
	
	// Create button attachments for options
	buttons := make([]map[string]interface{}, 0)
	for _, option := range question.Options {
		button := map[string]interface{}{
			"text":  option,
			"value": option,
			"type":  "button",
		}
		buttons = append(buttons, button)
	}
	
	attachment := models.Attachment{
		Type: "carousel",
		Carousel: map[string]interface{}{
			"type":    "buttons",
			"buttons": buttons,
		},
	}
	attachments = append(attachments, attachment)
	
	return &models.ChatMessage{
		Sender:     "system",
		SenderName: "CSAT Survey",
		SenderType: string(models.SenderTypeSystem),
		SessionID:  primitive.NilObjectID, // Will be set based on chat session
		Text:       question.QuestionText,
		Attachments: attachments,
		Category:   models.MessageCategoryInfo,
		Data: map[string]interface{}{
			"csat_message":    true,
			"csat_session_id": session.ID.Hex(),
			"question_id":     question.ID.Hex(),
			"options":         question.Options,
		},
	}, nil
}
