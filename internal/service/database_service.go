package service

import (
	"context"
	"fmt"
	"time"

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

// ChatMessage represents a chat message document
type ChatMessage struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	MessageID string             `bson:"message_id"`
	SessionID string             `bson:"session_id"`
	ClientID  string             `bson:"client_id"`
	Content   string             `bson:"content"`
	Role      string             `bson:"role"` // user, assistant, system
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	Metadata  map[string]interface{} `bson:"metadata,omitempty"`
}

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
	
	var message ChatMessage
	err := collection.FindOne(ctx, bson.M{"message_id": messageID}).Decode(&message)
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

// SaveChatMessage saves or updates a chat message
func (db *DatabaseService) SaveChatMessage(ctx context.Context, message *ChatMessage) error {
	collection := db.database.Collection("chat_messages")
	
	message.UpdatedAt = time.Now()
	if message.CreatedAt.IsZero() {
		message.CreatedAt = message.UpdatedAt
	}
	
	filter := bson.M{"message_id": message.MessageID}
	update := bson.M{"$set": message}
	
	_, err := collection.UpdateOne(ctx, filter, update, nil)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
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

// HealthCheck performs a basic health check on the database connection
func (db *DatabaseService) HealthCheck(ctx context.Context) error {
	return db.mongoClient.Ping(ctx, nil)
}