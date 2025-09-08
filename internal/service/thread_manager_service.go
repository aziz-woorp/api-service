package service

import (
	"context"
	"fmt"
	"log"
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
		log.Printf("[ThreadManager] Client is nil, threading disabled")
		return false
	}

	log.Printf("[ThreadManager] Checking threading for client %s (client_id: %s)", client.ID.Hex(), client.ClientID)
	log.Printf("[ThreadManager] Client ThreadConfig: %+v", client.ThreadConfig)

	// Check if thread_config exists at root level and is enabled
	if client.ThreadConfig != nil {
		log.Printf("[ThreadManager] Found ThreadConfig: %+v", client.ThreadConfig)
		if enabled, ok := client.ThreadConfig["enabled"].(bool); ok {
			log.Printf("[ThreadManager] Threading enabled: %v", enabled)
			return enabled
		} else {
			log.Printf("[ThreadManager] 'enabled' field not found or not boolean in ThreadConfig")
		}
	} else {
		log.Printf("[ThreadManager] No ThreadConfig found at root level")
		
		// Fallback: check if thread_config exists in Config field (for backward compatibility)
		if threadConfig, exists := client.Config["thread_config"]; exists {
			log.Printf("[ThreadManager] Found thread_config in Config: %+v", threadConfig)
			if config, ok := threadConfig.(map[string]interface{}); ok {
				log.Printf("[ThreadManager] thread_config is map: %+v", config)
				if enabled, ok := config["enabled"].(bool); ok {
					log.Printf("[ThreadManager] Threading enabled (from Config): %v", enabled)
					return enabled
				}
			}
		}
		log.Printf("[ThreadManager] No thread_config found in either ThreadConfig or Config")
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
	filter := bson.M{"parent_session_id": parentSessionID}
	opts := options.FindOne().SetSort(bson.D{{Key: "last_activity", Value: -1}})

	var thread models.ChatSessionThread
	err := tm.chatSessionThreadCollection.FindOne(ctx, filter, opts).Decode(&thread)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find latest thread: %w", err)
	}

	return &thread, nil
}

// getExistingThreadedSessions checks if any threaded sessions exist for a base session ID
func (tm *ThreadManagerService) getExistingThreadedSessions(ctx context.Context, baseSessionID string) ([]*models.ChatSession, error) {
	// Look for sessions that start with baseSessionID# (threaded sessions)
	filter := bson.M{"session_id": bson.M{"$regex": "^" + baseSessionID + "#"}}
	cursor, err := tm.chatSessionCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find existing threaded sessions: %w", err)
	}
	defer cursor.Close(ctx)

	var sessions []*models.ChatSession
	if err = cursor.All(ctx, &sessions); err != nil {
		return nil, fmt.Errorf("failed to decode threaded sessions: %w", err)
	}

	return sessions, nil
}

// createFirstThread creates the first thread for a new session (matching Python behavior)
func (tm *ThreadManagerService) createFirstThread(ctx context.Context, baseSessionID string, client *models.Client, clientChannel *models.ClientChannel) (*models.ChatSession, error) {
	// Generate a thread ID (8 characters like Python)
	threadID := primitive.NewObjectID().Hex()[:8]
	threadedSessionID := tm.FormatThreadSessionID(baseSessionID, threadID)

	log.Printf("[ThreadManager] Creating first thread %s for new session %s", threadID, baseSessionID)

	// Create the session with thread ID
	now := time.Now().UTC()
	session := &models.ChatSession{
		SessionID:     threadedSessionID,
		Active:        true,
		Client:        &client.ID,
		ClientChannel: &clientChannel.ID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Insert the threaded session
	result, err := tm.chatSessionCollection.InsertOne(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create threaded session: %w", err)
	}

	// Set the generated ObjectID back to the session
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		session.ID = oid
	}

	// Create thread tracking record
	thread := &models.ChatSessionThread{
		ThreadID:         threadID,
		ThreadSessionID:  threadedSessionID,
		ParentSessionID:  baseSessionID,
		ChatSessionID:    session.ID,
		Active:           true,
		LastActivity:     now,
	}

	_, err = tm.chatSessionThreadCollection.InsertOne(ctx, thread)
	if err != nil {
		return nil, fmt.Errorf("failed to create thread tracking record: %w", err)
	}

	log.Printf("[ThreadManager] Created first thread %s with session_id %s", threadID, threadedSessionID)
	return session, nil
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

// CloseActiveThreads closes all active threads for a parent session and returns count
func (tm *ThreadManagerService) CloseActiveThreads(ctx context.Context, parentSessionID string) (int, error) {
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

	result, err := tm.chatSessionThreadCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("failed to close active threads: %w", err)
	}

	return int(result.ModifiedCount), nil
}

// CreateNewThread creates a new thread for an existing session context
// This creates both a ChatSessionThread record and a new ChatSession with threaded session_id
// Note: In threading, there's no separate "parent session" - we create threads directly
func (tm *ThreadManagerService) CreateNewThread(ctx context.Context, baseSessionID string, client *models.Client, clientChannel *models.ClientChannel) (*models.ChatSession, error) {
	log.Printf("[ThreadManager] CreateNewThread called for baseSessionID: %s", baseSessionID)
	
	// We don't need a "parent session" - threads are the actual sessions
	// The baseSessionID is just used as an identifier for grouping threads

	// Deactivate any existing active threads for this base session
	log.Printf("[ThreadManager] Deactivating existing active threads for session %s", baseSessionID)
	deactivatedCount, err := tm.CloseActiveThreads(ctx, baseSessionID)
	if err != nil {
		log.Printf("[ThreadManager] ERROR: Failed to close active threads for %s: %v", baseSessionID, err)
		return nil, fmt.Errorf("failed to close active threads: %w", err)
	}
	if deactivatedCount > 0 {
		log.Printf("[ThreadManager] Deactivated %d existing active threads for session %s", deactivatedCount, baseSessionID)
	}

	// Generate new thread ID (8 characters like Python)
	threadID := primitive.NewObjectID().Hex()[:8]
	threadSessionID := tm.FormatThreadSessionID(baseSessionID, threadID)

	log.Printf("[ThreadManager] Creating new thread: %s for session %s, threaded session_id: %s", threadID, baseSessionID, threadSessionID)

	// Create new ChatSession with threaded session_id (matching Python behavior)
	now := time.Now().UTC()
	newChatSession := &models.ChatSession{
		SessionID:     threadSessionID,
		Active:        true,
		Client:        &client.ID,
		ClientChannel: &clientChannel.ID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	log.Printf("[ThreadManager] Inserting new threaded ChatSession with session_id: %s", threadSessionID)
	// Insert the new chat session
	result, err := tm.chatSessionCollection.InsertOne(ctx, newChatSession)
	if err != nil {
		log.Printf("[ThreadManager] ERROR: Failed to create threaded chat session: %v", err)
		return nil, fmt.Errorf("failed to create threaded chat session: %w", err)
	}

	// Set the generated ObjectID back to the session
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		newChatSession.ID = oid
		log.Printf("[ThreadManager] Assigned ObjectID %s to new ChatSession", oid.Hex())
	}

	// Create thread tracking record
	thread := &models.ChatSessionThread{
		ThreadID:         threadID,
		ThreadSessionID:  threadSessionID,
		ParentSessionID:  baseSessionID,
		ChatSessionID:    newChatSession.ID,
		Active:           true,
		LastActivity:     now,
	}

	log.Printf("[ThreadManager] Creating ChatSessionThread tracking record for thread %s", threadID)
	_, err = tm.chatSessionThreadCollection.InsertOne(ctx, thread)
	if err != nil {
		log.Printf("[ThreadManager] ERROR: Failed to create thread tracking record: %v", err)
		return nil, fmt.Errorf("failed to create thread tracking record: %w", err)
	}

	log.Printf("[ThreadManager] Created new thread: %s for session %s with threaded session_id: %s", threadID, baseSessionID, threadSessionID)
	return newChatSession, nil
}

// GetOrCreateActiveThread gets or creates an active thread for a session
// Core function to handle thread management logic for both new and existing sessions.
// This unified method can:
// 1. Create a new session + thread when none exists yet
// 2. Find or create a new thread for an existing session
// Matches Python ThreadManager.get_or_create_active_thread exactly
func (tm *ThreadManagerService) GetOrCreateActiveThread(ctx context.Context, sessionID string, client *models.Client, clientChannel *models.ClientChannel, forceNew bool, inactivityMinutes int) (*models.ChatSession, error) {
	// Parse session ID to get base ID (removing any thread part)
	baseSessionID, _ := tm.ParseSessionID(sessionID)

	log.Printf("[ThreadManager] Processing session %s, base: %s", sessionID, baseSessionID)

	// First, check if any sessions exist with this base ID
	existingSessions, err := tm.getExistingThreadedSessions(ctx, baseSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing sessions: %w", err)
	}
	sessionExists := len(existingSessions) > 0

	// If client wasn't provided, try to get it from existing sessions
	if client == nil && sessionExists {
		if existingSessions[0].Client != nil {
			// Get client from existing session
			var clientObj models.Client
			err = tm.clientCollection.FindOne(ctx, bson.M{"_id": *existingSessions[0].Client}).Decode(&clientObj)
			if err == nil {
				client = &clientObj
			}
		}
	}

	// If client_channel wasn't provided, try to get it from existing sessions
	if clientChannel == nil && sessionExists {
		if existingSessions[0].ClientChannel != nil {
			// Get client channel from existing session
			var channelObj models.ClientChannel
			err = tm.clientCollection.Database().Collection("client_channels").FindOne(ctx, bson.M{"_id": *existingSessions[0].ClientChannel}).Decode(&channelObj)
			if err == nil {
				clientChannel = &channelObj
			}
		}
	}

	// If we have no sessions and no client was provided, we can't proceed
	if !sessionExists && client == nil {
		return nil, fmt.Errorf("cannot create a thread without either an existing session or a client object")
	}

	// Check if threading is enabled for this client
	threadingEnabled := tm.IsThreadingEnabledForClient(ctx, client)
	var clientInactivityMinutes int

	if threadingEnabled {
		// Get inactivity minutes from client ThreadConfig (root level)
		if client.ThreadConfig != nil {
			if minutes, ok := client.ThreadConfig["inactivity_minutes"].(float64); ok {
				clientInactivityMinutes = int(minutes)
				log.Printf("[ThreadManager] Got inactivity_minutes from ThreadConfig (float64): %d", clientInactivityMinutes)
			} else if minutes, ok := client.ThreadConfig["inactivity_minutes"].(int); ok {
				clientInactivityMinutes = minutes
				log.Printf("[ThreadManager] Got inactivity_minutes from ThreadConfig (int): %d", clientInactivityMinutes)
			} else if minutes, ok := client.ThreadConfig["inactivity_minutes"].(int32); ok {
				clientInactivityMinutes = int(minutes)
				log.Printf("[ThreadManager] Got inactivity_minutes from ThreadConfig (int32): %d", clientInactivityMinutes)
			} else {
				log.Printf("[ThreadManager] Could not parse inactivity_minutes from ThreadConfig: %+v (type: %T)", client.ThreadConfig["inactivity_minutes"], client.ThreadConfig["inactivity_minutes"])
			}
		}
		
		// Fallback: check Config field for backward compatibility
		if clientInactivityMinutes <= 0 {
			if threadConfig, exists := client.Config["thread_config"]; exists {
				if config, ok := threadConfig.(map[string]interface{}); ok {
					if minutes, ok := config["inactivity_minutes"].(float64); ok {
						clientInactivityMinutes = int(minutes)
					}
				}
			}
		}
		
		if clientInactivityMinutes <= 0 {
			clientInactivityMinutes = 1440 // Default 24 hours
		}
		log.Printf("[ThreadManager] Threading enabled for client %s with inactivity_minutes=%d", client.ID.Hex(), clientInactivityMinutes)
	} else {
		log.Printf("[ThreadManager] Threading disabled for client %s", client.ID.Hex())
	}

	// If threading is disabled, handle it directly
	if !threadingEnabled {
		// No threading, check if session exists
		var session models.ChatSession
		err = tm.chatSessionCollection.FindOne(ctx, bson.M{"session_id": baseSessionID}).Decode(&session)
		if err == nil {
			log.Printf("[ThreadManager] Using existing non-threaded session %s", baseSessionID)
			return &session, nil
		}

		// Create a new regular session
		if client != nil {
			log.Printf("[ThreadManager] Creating new non-threaded session %s", baseSessionID)
			now := time.Now().UTC()
			session = models.ChatSession{
				SessionID:     baseSessionID,
				Active:        true,
				Client:        &client.ID,
				ClientChannel: &clientChannel.ID,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			_, err = tm.chatSessionCollection.InsertOne(ctx, &session)
			if err != nil {
				return nil, fmt.Errorf("failed to create session: %w", err)
			}
			return &session, nil
		} else {
			return nil, fmt.Errorf("cannot create session: threading is disabled and no session exists")
		}
	}

	// Threading is enabled - check if we need to use existing thread or create new one

	// Use client-specific inactivity minutes or provided default
	log.Printf("[ThreadManager] Input inactivityMinutes: %d, clientInactivityMinutes: %d", inactivityMinutes, clientInactivityMinutes)
	if inactivityMinutes <= 0 {
		inactivityMinutes = clientInactivityMinutes
	}
	log.Printf("[ThreadManager] Final inactivityMinutes: %d", inactivityMinutes)

	// Get latest thread for this parent session
	latestThread, err := tm.GetLatestThread(ctx, baseSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest thread: %w", err)
	}

	// Check if we have a latest thread but it's inactive
	if latestThread != nil && !tm.IsThreadActive(latestThread, inactivityMinutes) && !forceNew {
		log.Printf("[ThreadManager] Found inactive thread %s for session %s (inactive for more than %d minutes)", 
			latestThread.ThreadID, baseSessionID, inactivityMinutes)
	}

	// Use existing thread if active and not forcing new
	if !forceNew && latestThread != nil && tm.IsThreadActive(latestThread, inactivityMinutes) {
		// Get the threaded ChatSession
		var threadedSession models.ChatSession
		err = tm.chatSessionCollection.FindOne(ctx, bson.M{"session_id": latestThread.ThreadSessionID}).Decode(&threadedSession)
		if err != nil {
			return nil, fmt.Errorf("failed to find threaded session: %w", err)
		}

		// Update activity timestamp
		now := time.Now().UTC()
		_, err = tm.chatSessionThreadCollection.UpdateOne(ctx,
			bson.M{"_id": latestThread.ID},
			bson.M{"$set": bson.M{"last_activity": now}})
		if err != nil {
			return nil, fmt.Errorf("failed to update thread activity: %w", err)
		}

		log.Printf("[ThreadManager] Using existing active thread %s for session %s", latestThread.ThreadID, baseSessionID)
		return &threadedSession, nil
	}

	// Check if this is a new session or we're creating a new thread for existing session
	if !sessionExists {
		// Creating first thread for a new session
		log.Printf("[ThreadManager] Creating first thread for new session %s", baseSessionID)
		return tm.createFirstThread(ctx, baseSessionID, client, clientChannel)
	} else {
		// Create a new thread for existing session
		log.Printf("[ThreadManager] Creating new thread for existing session %s", baseSessionID)
		return tm.CreateNewThread(ctx, baseSessionID, client, clientChannel)
	}
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