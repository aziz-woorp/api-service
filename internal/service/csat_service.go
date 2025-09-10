// Package service provides business logic for CSAT surveys.
package service

import (
	"context"
	"fmt"
	"strings"
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
	ChatSessionRepo       *repository.ChatSessionRepository
	ThreadService         *ChatSessionThreadService
	EventPublisherService *EventPublisherService
	PayloadService        *PayloadService
}

// NewCSATService creates a new CSATService.
func NewCSATService(
	configRepo *repository.CSATConfigurationRepository,
	questionRepo *repository.CSATQuestionTemplateRepository,
	sessionRepo *repository.CSATSessionRepository,
	responseRepo *repository.CSATResponseRepository,
	chatMessageRepo *repository.ChatMessageRepository,
	chatSessionRepo *repository.ChatSessionRepository,
	threadService *ChatSessionThreadService,
	eventPublisher *EventPublisherService,
	payloadService *PayloadService,
) *CSATService {
	return &CSATService{
		CSATConfigRepo:        configRepo,
		CSATQuestionRepo:      questionRepo,
		CSATSessionRepo:       sessionRepo,
		CSATResponseRepo:      responseRepo,
		ChatMessageRepo:       chatMessageRepo,
		ChatSessionRepo:       chatSessionRepo,
		ThreadService:         threadService,
		EventPublisherService: eventPublisher,
		PayloadService:        payloadService,
	}
}

// parseSessionID extracts base session ID and potential thread information
// from session IDs that may have thread information appended
// e.g., "session_123_thread_456" -> baseSessionID="session_123", threadFromSessionID="thread_456"
func parseSessionID(sessionID string) (baseSessionID string, threadFromSessionID string) {
	// Look for thread suffix pattern like "_thread_xyz"
	if strings.Contains(sessionID, "_thread_") {
		parts := strings.Split(sessionID, "_thread_")
		if len(parts) == 2 {
			return parts[0], "thread_" + parts[1]
		}
	}
	// No thread information found, return original
	return sessionID, ""
}

// TriggerCSATSurveyBySessionID triggers a CSAT survey using external session_id.
func (s *CSATService) TriggerCSATSurveyBySessionID(ctx context.Context, sessionID string) (*models.CSATSession, error) {
	// 0. Parse session ID to extract potential thread information
	baseSessionID, threadFromSessionID := parseSessionID(sessionID)
	
	// 1. Resolve chat session by base session_id (repository will handle startsWith lookup)
	chatSession, err := s.ChatSessionRepo.GetBySessionID(ctx, baseSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find chat session with session_id %s: %w", baseSessionID, err)
	}
	
	// 2. Validate that chat session has client and channel information
	if chatSession.Client == nil {
		return nil, fmt.Errorf("chat session %s missing client information", baseSessionID)
	}
	if chatSession.ClientChannel == nil {
		return nil, fmt.Errorf("chat session %s missing client channel information", baseSessionID)
	}
	
	clientID := *chatSession.Client
	channelID := *chatSession.ClientChannel
	
	// 3. Determine target session context and thread information
	var targetSessionContext string
	var threadSessionID *string
	var threadContext bool
	
	// If session ID had thread information, use that directly
	if threadFromSessionID != "" {
		targetSessionContext = sessionID // Use the full original session ID
		threadSessionID = &threadFromSessionID
		threadContext = true
	} else if s.ThreadService != nil {
		// Try to get the latest active thread (30 minutes inactivity threshold)
		latestThread, err := s.ThreadService.GetActiveThread(ctx, chatSession.ID.Hex(), 30)
		if err == nil && latestThread != nil {
			// Use thread context
			targetSessionContext = latestThread.ThreadSessionID
			threadSessionID = &latestThread.ThreadSessionID
			threadContext = true
		} else {
			// Fall back to main session
			targetSessionContext = baseSessionID
			threadContext = false
		}
	} else {
		// No threading service available, use base session
		targetSessionContext = baseSessionID
		threadContext = false
	}
	
	// 4. Trigger CSAT with resolved context
	return s.triggerCSATSurvey(ctx, targetSessionContext, clientID, channelID, threadSessionID, threadContext)
}

// triggerCSATSurvey is the internal method that creates the CSAT session.
func (s *CSATService) triggerCSATSurvey(ctx context.Context, chatSessionID string, clientID, channelID primitive.ObjectID, threadSessionID *string, threadContext bool) (*models.CSATSession, error) {
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
		ThreadSessionID:      threadSessionID,
		ThreadContext:        threadContext,
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
		"thread_context":  threadContext,
	}
	if threadSessionID != nil {
		eventData["thread_session_id"] = *threadSessionID
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
	
	// Create chat message structure (but don't save to database)
	chatMessageStructure, err := s.createQuestionMessageStructure(session, &currentQuestion)
	if err != nil {
		return fmt.Errorf("failed to create question message structure: %w", err)
	}
	
	// Update session status and add question to sent list
	session.Status = "in_progress"
	session.QuestionsSent = append(session.QuestionsSent, currentQuestion.ID.Hex())
	if err := s.CSATSessionRepo.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to update CSAT session: %w", err)
	}
	
	// Create event data with chat message structure for upstream processing
	eventData := map[string]interface{}{
		"csat_session_id": session.ID.Hex(),
		"question_id":     currentQuestion.ID.Hex(),
		"chat_session_id": session.ChatSessionID,
		"message_type":    "question",
		"chat_message":    chatMessageStructure,
	}
	
	// Get session ID for parent_id (use chat session context)
	chatSessionIDStr := session.ChatSessionID
	
	// Publish CSAT message sent event with entity_type="csat_question"
	_, err = s.EventPublisherService.PublishEvent(
		ctx,
		models.EventTypeCSATMessageSent,
		models.EntityTypeCSATQuestion,
		currentQuestion.ID.Hex(),
		&chatSessionIDStr,
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

// ProcessResponseBySessionID processes a user response using external chat session ID.
func (s *CSATService) ProcessResponseBySessionID(ctx context.Context, sessionID, questionID, responseValue string) (string, error) {
	// 1. Parse session ID to extract base session and potential thread info
	baseSessionID, _ := parseSessionID(sessionID)
	
	// 2. Find chat session using base session ID (to validate session exists)
	_, err := s.ChatSessionRepo.GetBySessionID(ctx, baseSessionID)
	if err != nil {
		return "", fmt.Errorf("failed to find chat session with session_id %s: %w", baseSessionID, err)
	}
	
	// 3. Find active CSAT session for this chat session
	csatSession, err := s.CSATSessionRepo.GetActiveByChatSessionID(ctx, sessionID)
	if err != nil {
		return "", fmt.Errorf("no active CSAT session found for session_id %s: %w", sessionID, err)
	}
	
	if csatSession.Status != "in_progress" {
		return "", fmt.Errorf("CSAT session is not in progress (status: %s)", csatSession.Status)
	}
	
	// 4. Parse question ID
	questionObjID, err := primitive.ObjectIDFromHex(questionID)
	if err != nil {
		return "", fmt.Errorf("invalid question_id format: %w", err)
	}
	
	// 5. Validate question exists in current survey
	questions, err := s.CSATQuestionRepo.GetByClientAndChannel(ctx, csatSession.Client, csatSession.ClientChannel)
	if err != nil {
		return "", fmt.Errorf("failed to get CSAT questions: %w", err)
	}
	
	// Find the question index to validate it exists
	currentQuestionIndex := -1
	for i, q := range questions {
		if q.ID == questionObjID {
			currentQuestionIndex = i
			break
		}
	}
	
	if currentQuestionIndex == -1 {
		return "", fmt.Errorf("question not found in current survey")
	}
	
	// 6. Check if response already exists for this question
	var responseID string
	existingResponse, err := s.CSATResponseRepo.GetBySessionAndQuestion(ctx, csatSession.ID, questionObjID)
	if err == nil && existingResponse != nil {
		// EXISTING RESPONSE: Update only, do NOT send next question
		existingResponse.ResponseValue = responseValue
		if err := s.CSATResponseRepo.Update(ctx, existingResponse); err != nil {
			return "", fmt.Errorf("failed to update CSAT response: %w", err)
		}
		responseID = existingResponse.ID.Hex()
		
		// Return immediately - no question advancement for updates
		return responseID, nil
	} else {
		// NEW RESPONSE: Create response and advance to next question
		response := &models.CSATResponse{
			CSATSession:      csatSession.ID,
			QuestionTemplate: questionObjID,
			ResponseValue:    responseValue,
		}
		
		if err := s.CSATResponseRepo.Create(ctx, response); err != nil {
			return "", fmt.Errorf("failed to create CSAT response: %w", err)
		}
		responseID = response.ID.Hex()
		
		// 7. Advance survey flow only for NEW responses
		// If this is the current question (or beyond), advance to next
		if currentQuestionIndex >= csatSession.CurrentQuestionIndex {
			csatSession.CurrentQuestionIndex = currentQuestionIndex + 1
			if err := s.CSATSessionRepo.Update(ctx, csatSession); err != nil {
				return "", fmt.Errorf("failed to update CSAT session: %w", err)
			}
			
			// Send next question or complete survey
			err = s.SendNextQuestion(ctx, csatSession.ID)
			if err != nil {
				return "", err
			}
		}
		
		return responseID, nil
	}
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
	
	// Create thank you message structure (but don't save to database)
	tempID := primitive.NewObjectID()
	thankYouMessageStructure := map[string]interface{}{
		"id":          tempID.Hex(),
		"sender":      "system",
		"sender_name": "CSAT Survey",
		"sender_type": string(models.SenderTypeSystem),
		"session_id":  session.ChatSessionID,
		"text":        "Thank you for your feedback!",
		"category":    string(models.MessageCategoryInfo),
		"data": map[string]interface{}{
			"csat_message":    true,
			"csat_session_id": session.ID.Hex(),
			"message_type":    "completion",
		},
		"created_at": now,
		"updated_at": now,
	}
	
	// Publish CSAT completed event with thank you message structure
	chatSessionIDStr := session.ChatSessionID
	eventData := map[string]interface{}{
		"csat_session_id": session.ID.Hex(),
		"chat_session_id": session.ChatSessionID,
		"completed_at":    session.CompletedAt,
		"message_type":    "completion",
		"chat_message":    thankYouMessageStructure,
	}
	
	_, err = s.EventPublisherService.PublishEvent(
		ctx,
		models.EventTypeCSATCompleted,
		models.EntityTypeCSATSession,
		session.ID.Hex(),
		&chatSessionIDStr,
		eventData,
	)
	if err != nil {
		return fmt.Errorf("failed to publish CSAT completed event: %w", err)
	}
	
	return nil
}

// createQuestionMessageStructure creates a chat message structure for CSAT questions without database persistence.
func (s *CSATService) createQuestionMessageStructure(session *models.CSATSession, question *models.CSATQuestionTemplate) (map[string]interface{}, error) {
	// Create postback buttons with CSAT payload format
	buttons := make([]map[string]interface{}, 0)
	for _, option := range question.Options {
		button := map[string]interface{}{
			"type":    "postback",
			"text":    option,
			"payload": fmt.Sprintf("csat:%s:%s", question.ID.Hex(), option),
		}
		buttons = append(buttons, button)
	}
	
	// Create buttons attachment (not carousel)
	attachment := map[string]interface{}{
		"type":    "buttons",
		"buttons": buttons,
	}
	
	// Generate a temporary ID for the message structure
	tempID := primitive.NewObjectID()
	
	// Create chat message structure (not a database model)
	chatMessageStructure := map[string]interface{}{
		"id":          tempID.Hex(),
		"sender":      "system",
		"sender_name": "CSAT Survey",
		"sender_type": string(models.SenderTypeSystem),
		"session_id":  session.ChatSessionID, // Use actual chat session ID
		"text":        question.QuestionText,
		"attachments": []map[string]interface{}{attachment},
		"category":    string(models.MessageCategoryInfo),
		"data": map[string]interface{}{
			"csat_message":    true,
			"csat_session_id": session.ID.Hex(),
			"question_id":     question.ID.Hex(),
			"options":         question.Options,
		},
		"created_at": time.Now().UTC(),
		"updated_at": time.Now().UTC(),
	}
	
	return chatMessageStructure, nil
}

// createQuestionMessage creates a chat message for a CSAT question.
func (s *CSATService) createQuestionMessage(session *models.CSATSession, question *models.CSATQuestionTemplate) (*models.ChatMessage, error) {
	// Create postback buttons with CSAT payload format
	var attachments []models.Attachment
	
	// Create postback buttons for options
	buttons := make([]map[string]interface{}, 0)
	for _, option := range question.Options {
		button := map[string]interface{}{
			"type":    "postback",
			"text":    option,
			"payload": fmt.Sprintf("csat:%s:%s", question.ID.Hex(), option),
		}
		buttons = append(buttons, button)
	}
	
	// Create buttons attachment (not carousel)
	attachment := models.Attachment{
		Type:    "buttons",
		Buttons: buttons,
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
