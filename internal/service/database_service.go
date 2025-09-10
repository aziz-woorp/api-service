package service

import (
	"context"
	"fmt"
	"time"

	"github.com/fraiday-org/api-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// DatabaseService handles database operations for task workers
type DatabaseService struct {
	logger     *zap.Logger
	mongoClient *mongo.Client
	database   *mongo.Database
}

// NewDatabaseService creates a new database service
func NewDatabaseService(logger *zap.Logger, mongoClient *mongo.Client, dbName string) *DatabaseService {
	return &DatabaseService{
		logger:      logger,
		mongoClient: mongoClient,
		database:    mongoClient.Database(dbName),
	}
}

// ChatMessage alias for models.ChatMessage for backwards compatibility in worker
type ChatMessage = models.ChatMessage

// ChatSession represents a chat session document
type ChatSession struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	SessionID string             `bson:"session_id"`
	ClientID  string             `bson:"client_id"`
	Title     string             `bson:"title,omitempty"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	Metadata  map[string]interface{} `bson:"metadata,omitempty"`
}

// EventDeliveryRecord represents an event delivery record
type EventDeliveryRecord struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	EventID      string             `bson:"event_id"`
	ProcessorID  string             `bson:"processor_id"`
	ClientID     string             `bson:"client_id"`
	EventType    string             `bson:"event_type"`
	EntityType   string             `bson:"entity_type"`
	Status       string             `bson:"status"` // pending, delivered, failed
	Attempts     int                `bson:"attempts"`
	LastAttempt  *time.Time         `bson:"last_attempt,omitempty"`
	DeliveredAt  *time.Time         `bson:"delivered_at,omitempty"`
	CreatedAt    time.Time          `bson:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at"`
	ErrorMessage string             `bson:"error_message,omitempty"`
	Payload      map[string]interface{} `bson:"payload,omitempty"`
}

// GetChatMessage retrieves a chat message by message ID
func (db *DatabaseService) GetChatMessage(ctx context.Context, messageID string) (*ChatMessage, error) {
	collection := db.database.Collection("chat_messages")
	
	// Convert messageID string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return nil, fmt.Errorf("invalid message ID format: %s", messageID)
	}
	
	var message ChatMessage
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&message)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("message not found: %s", messageID)
		}
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	
	return &message, nil
}

// GetChatSession retrieves a chat session by session ID
func (db *DatabaseService) GetChatSession(ctx context.Context, sessionID string) (*ChatSession, error) {
	collection := db.database.Collection("chat_sessions")
	
	var session ChatSession
	err := collection.FindOne(ctx, bson.M{"session_id": sessionID}).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("session not found: %s", sessionID)
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	return &session, nil
}

// GetChatSessionByID retrieves a chat session by its MongoDB _id
func (db *DatabaseService) GetChatSessionByID(ctx context.Context, sessionID string) (*models.ChatSession, error) {
	collection := db.database.Collection("chat_sessions")
	
	// Convert sessionID string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID format: %s", sessionID)
	}
	
	var session models.ChatSession
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("session not found: %s", sessionID)
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	return &session, nil
}

// SaveChatMessage saves or updates a chat message
func (db *DatabaseService) SaveChatMessage(ctx context.Context, message *ChatMessage) error {
	collection := db.database.Collection("chat_messages")
	
	message.UpdatedAt = time.Now()
	if message.CreatedAt.IsZero() {
		message.CreatedAt = message.UpdatedAt
	}
	
	// If message has no ID, create new; otherwise update existing
	if message.ID == primitive.NilObjectID {
		// Create new message
		result, err := collection.InsertOne(ctx, message)
		if err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}
		
		// Set the generated ObjectID back to the message
		if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
			message.ID = oid
		}
	} else {
		// Update existing message
		filter := bson.M{"_id": message.ID}
		update := bson.M{"$set": message}
		
		_, err := collection.UpdateOne(ctx, filter, update)
		if err != nil {
			return fmt.Errorf("failed to update message: %w", err)
		}
	}
	
	return nil
}

// CreateEventDeliveryRecord creates a new event delivery record
func (db *DatabaseService) CreateEventDeliveryRecord(ctx context.Context, record *EventDeliveryRecord) error {
	collection := db.database.Collection("event_delivery_records")
	
	record.CreatedAt = time.Now()
	record.UpdatedAt = record.CreatedAt
	record.Status = "pending"
	record.Attempts = 0
	
	_, err := collection.InsertOne(ctx, record)
	if err != nil {
		return fmt.Errorf("failed to create delivery record: %w", err)
	}
	
	return nil
}

// UpdateEventDeliveryRecord updates an event delivery record
func (db *DatabaseService) UpdateEventDeliveryRecord(ctx context.Context, recordID primitive.ObjectID, updates bson.M) error {
	collection := db.database.Collection("event_delivery_records")
	
	updates["updated_at"] = time.Now()
	
	_, err := collection.UpdateOne(ctx, bson.M{"_id": recordID}, bson.M{"$set": updates})
	if err != nil {
		return fmt.Errorf("failed to update delivery record: %w", err)
	}
	
	return nil
}

// GetSessionContext retrieves context information for a chat session
func (db *DatabaseService) GetSessionContext(ctx context.Context, sessionID string) (map[string]interface{}, error) {
	// Get recent messages from the session for context
	collection := db.database.Collection("chat_messages")
	
	cursor, err := collection.Find(ctx, bson.M{
		"session_id": sessionID,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get session messages: %w", err)
	}
	defer cursor.Close(ctx)
	
	var messages []ChatMessage
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, fmt.Errorf("failed to decode messages: %w", err)
	}
	
	// Build context from messages
	context := map[string]interface{}{
		"session_id":     sessionID,
		"message_count":  len(messages),
		"recent_messages": messages,
	}
	
	return context, nil
}

// GetCSATSession retrieves a CSAT session by ID
func (db *DatabaseService) GetCSATSession(ctx context.Context, sessionID string) (*models.CSATSession, error) {
	collection := db.database.Collection("csat_sessions")
	
	// Convert sessionID string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid CSAT session ID format: %s", sessionID)
	}
	
	var session models.CSATSession
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("CSAT session not found: %s", sessionID)
		}
		return nil, fmt.Errorf("failed to get CSAT session: %w", err)
	}
	
	return &session, nil
}

// GetCSATQuestion retrieves a CSAT question template by ID
func (db *DatabaseService) GetCSATQuestion(ctx context.Context, questionID string) (*models.CSATQuestionTemplate, error) {
	collection := db.database.Collection("csat_question_templates")
	
	// Convert questionID string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(questionID)
	if err != nil {
		return nil, fmt.Errorf("invalid CSAT question ID format: %s", questionID)
	}
	
	var question models.CSATQuestionTemplate
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&question)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("CSAT question not found: %s", questionID)
		}
		return nil, fmt.Errorf("failed to get CSAT question: %w", err)
	}
	
	return &question, nil
}

// GetCSATResponse retrieves a CSAT response by ID
func (db *DatabaseService) GetCSATResponse(ctx context.Context, responseID string) (*models.CSATResponse, error) {
	collection := db.database.Collection("csat_responses")
	
	// Convert responseID string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(responseID)
	if err != nil {
		return nil, fmt.Errorf("invalid CSAT response ID format: %s", responseID)
	}
	
	var response models.CSATResponse
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&response)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("CSAT response not found: %s", responseID)
		}
		return nil, fmt.Errorf("failed to get CSAT response: %w", err)
	}
	
	return &response, nil
}

// HealthCheck performs a basic health check on the database connection
func (db *DatabaseService) HealthCheck(ctx context.Context) error {
	return db.mongoClient.Ping(ctx, nil)
}