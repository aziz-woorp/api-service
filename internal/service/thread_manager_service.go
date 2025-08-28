package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/fraiday-org/api-service/internal/models"
)

// ThreadManagerService manages session threads without modifying existing models
// Provides functionality for creating threads, checking activity, and
// retrieving the appropriate thread for a given session ID
type ThreadManagerService struct {
	chatSessionCollection       *mongo.Collection
	chatSessionThreadCollection *mongo.Collection
	clientCollection           *mongo.Collection
}

// NewThreadManagerService creates a new ThreadManagerService
func NewThreadManagerService(db *mongo.Database) *ThreadManagerService {
	return &ThreadManagerService{
		chatSessionCollection:       db.Collection("chat_sessions"),
		chatSessionThreadCollection: db.Collection("chat_session_threads"),
		clientCollection:           db.Collection("clients"),
	}
}

// FormatThreadSessionID formats the composite session_id
func (tm *ThreadManagerService) FormatThreadSessionID(parentID, threadID string) string {
	return fmt.Sprintf("%s#%s", parentID, threadID)
}

// ParseSessionID parses composite session_id into parent and thread components
func (tm *ThreadManagerService) ParseSessionID(sessionID string) (string, string) {
	if sessionID != "" && strings.Contains(sessionID, "#") {
		parts := strings.SplitN(sessionID, "#", 2)
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	}
	return sessionID, ""
}

// GetBaseSessionIDForEvent gets the base session ID for event payloads, stripping any thread information
// This ensures that external systems always receive consistent session IDs
// regardless of threading being enabled
func (tm *ThreadManagerService) GetBaseSessionIDForEvent(sessionID string) string {
	baseID, _ := tm.ParseSessionID(sessionID)
	return baseID
}

// IsThreadingEnabledForClient checks if threading is enabled for a client
func (tm *ThreadManagerService) IsThreadingEnabledForClient(ctx context.Context, client *models.Client) bool {
	if client == nil {
		return false
	}

	// Check if thread_config exists and is enabled
	if threadConfig, exists := client.Config["thread_config"]; exists {
		if config, ok := threadConfig.(map[string]interface{}); ok {
			if enabled, ok := config["enabled"].(bool); ok {
				return enabled
			}
		}
	}
	return false
}

// IsThreadingEnabledForClientID checks if threading is enabled for a client by ID
func (tm *ThreadManagerService) IsThreadingEnabledForClientID(ctx context.Context, clientID string) (bool, error) {
	clientObjID, err := primitive.ObjectIDFromHex(clientID)
	if err != nil {
		return false, fmt.Errorf("invalid client ID: %w", err)
	}

	var client models.Client
	err = tm.clientCollection.FindOne(ctx, bson.M{"_id": clientObjID}).Decode(&client)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, fmt.Errorf("failed to find client: %w", err)
	}

	return tm.IsThreadingEnabledForClient(ctx, &client), nil
}

// IsThreadingEnabledForSession checks if threading is enabled for a session
func (tm *ThreadManagerService) IsThreadingEnabledForSession(ctx context.Context, sessionID string) (bool, error) {
	baseSessionID, _ := tm.ParseSessionID(sessionID)

	sessionObjID, err := primitive.ObjectIDFromHex(baseSessionID)
	if err != nil {
		return false, fmt.Errorf("invalid session ID: %w", err)
	}

	var session models.ChatSession
	err = tm.chatSessionCollection.FindOne(ctx, bson.M{"_id": sessionObjID}).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, fmt.Errorf("failed to find session: %w", err)
	}

	if session.Client == nil {
		return false, nil
	}
	return tm.IsThreadingEnabledForClientID(ctx, session.Client.Hex())
}

// GetLatestThread gets the latest thread for a parent session
func (tm *ThreadManagerService) GetLatestThread(ctx context.Context, parentSessionID string) (*models.ChatSessionThread, error) {
	parentObjID, err := primitive.ObjectIDFromHex(parentSessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid parent session ID: %w", err)
	}

	filter := bson.M{"parent_session_id": parentObjID}
	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})

	var thread models.ChatSessionThread
	err = tm.chatSessionThreadCollection.FindOne(ctx, filter, opts).Decode(&thread)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find latest thread: %w", err)
	}

	return &thread, nil
}

// IsThreadActive checks if a thread is active based on inactivity minutes
func (tm *ThreadManagerService) IsThreadActive(thread *models.ChatSessionThread, inactivityMinutes int) bool {
	if thread == nil || !thread.Active {
		return false
	}

	if inactivityMinutes <= 0 {
		inactivityMinutes = 1440 // Default 24 hours
	}

	inactivityDuration := time.Duration(inactivityMinutes) * time.Minute
	return time.Since(thread.LastActivity) <= inactivityDuration
}

// GetChatSession retrieves a chat session by ID
func (tm *ThreadManagerService) GetChatSession(ctx context.Context, parentSessionID string) (*models.ChatSession, error) {
	sessionObjID, err := primitive.ObjectIDFromHex(parentSessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	var session models.ChatSession
	err = tm.chatSessionCollection.FindOne(ctx, bson.M{"_id": sessionObjID}).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find session: %w", err)
	}

	return &session, nil
}

// DeactivateThread deactivates a thread
func (tm *ThreadManagerService) DeactivateThread(ctx context.Context, thread *models.ChatSessionThread) error {
	if thread == nil {
		return fmt.Errorf("thread cannot be nil")
	}

	now := time.Now().UTC()
	update := bson.M{
		"$set": bson.M{
			"active":        false,
			"last_activity": now,
		},
	}

	_, err := tm.chatSessionThreadCollection.UpdateOne(ctx, bson.M{"_id": thread.ID}, update)
	if err != nil {
		return fmt.Errorf("failed to deactivate thread: %w", err)
	}

	return nil
}

// CloseActiveThreads closes all active threads for a parent session
func (tm *ThreadManagerService) CloseActiveThreads(ctx context.Context, parentSessionID string) error {
	now := time.Now().UTC()
	filter := bson.M{
		"parent_session_id": parentSessionID,
		"active":            true,
	}
	update := bson.M{
		"$set": bson.M{
			"active":        false,
			"last_activity": now,
		},
	}

	_, err := tm.chatSessionThreadCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to close active threads: %w", err)
	}

	return nil
}

// CreateNewThread creates a new thread for a parent session
func (tm *ThreadManagerService) CreateNewThread(ctx context.Context, parentSessionID string) (*models.ChatSessionThread, error) {
	// Get parent session to inherit properties
	parentSession, err := tm.GetChatSession(ctx, parentSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent session: %w", err)
	}
	if parentSession == nil {
		return nil, fmt.Errorf("parent session not found")
	}

	// Close any existing active threads
	err = tm.CloseActiveThreads(ctx, parentSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to close active threads: %w", err)
	}

	// Generate thread ID and session ID
	threadID := primitive.NewObjectID().Hex()
	threadSessionID := tm.FormatThreadSessionID(parentSessionID, threadID)

	// Create new thread
	now := time.Now().UTC()
	thread := &models.ChatSessionThread{
		ThreadID:         threadID,
		ThreadSessionID:  threadSessionID,
		ParentSessionID:  parentSessionID,
		ChatSessionID:    parentSession.ID,
		Active:           true,
		LastActivity:     now,
	}

	_, err = tm.chatSessionThreadCollection.InsertOne(ctx, thread)
	if err != nil {
		return nil, fmt.Errorf("failed to create thread: %w", err)
	}

	return thread, nil
}

// GetOrCreateActiveThread gets or creates an active thread for a session
func (tm *ThreadManagerService) GetOrCreateActiveThread(ctx context.Context, sessionID string, forceNew bool, inactivityMinutes int) (*models.ChatSessionThread, error) {
	baseSessionID, threadID := tm.ParseSessionID(sessionID)

	// Check if threading is enabled for this session
	enabled, err := tm.IsThreadingEnabledForSession(ctx, baseSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to check threading status: %w", err)
	}
	if !enabled {
		return nil, nil // Threading not enabled
	}

	// If specific thread ID provided, try to get it
	if threadID != "" {
		threadObjID, err := primitive.ObjectIDFromHex(threadID)
		if err != nil {
			return nil, fmt.Errorf("invalid thread ID: %w", err)
		}

		var thread models.ChatSessionThread
		err = tm.chatSessionThreadCollection.FindOne(ctx, bson.M{"_id": threadObjID}).Decode(&thread)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, fmt.Errorf("thread not found")
			}
			return nil, fmt.Errorf("failed to find thread: %w", err)
		}

		return &thread, nil
	}

	// Get latest thread
	latestThread, err := tm.GetLatestThread(ctx, baseSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest thread: %w", err)
	}

	// If force new or no active thread exists, create new one
	if forceNew || latestThread == nil || !tm.IsThreadActive(latestThread, inactivityMinutes) {
		return tm.CreateNewThread(ctx, baseSessionID)
	}

	return latestThread, nil
}

// ListThreads lists all threads for a parent session
func (tm *ThreadManagerService) ListThreads(ctx context.Context, parentSessionID string, includeInactive bool) ([]*models.ChatSessionThread, error) {
	parentObjID, err := primitive.ObjectIDFromHex(parentSessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid parent session ID: %w", err)
	}

	filter := bson.M{"parent_session_id": parentObjID}
	if !includeInactive {
		filter["is_active"] = true
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := tm.chatSessionThreadCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find threads: %w", err)
	}
	defer cursor.Close(ctx)

	var threads []*models.ChatSessionThread
	if err = cursor.All(ctx, &threads); err != nil {
		return nil, fmt.Errorf("failed to decode threads: %w", err)
	}

	return threads, nil
}