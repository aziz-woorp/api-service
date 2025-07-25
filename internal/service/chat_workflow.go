// Package service provides background workflow triggers for chat messages.
package service

import (
	"context"
	"log"
)

// TriggerChatWorkflow simulates triggering an AI chat workflow in the background.
func TriggerChatWorkflow(ctx context.Context, messageID string, sessionID string) {
	go func() {
		// TODO: Replace with actual AI chat workflow logic.
		log.Printf("[Background] Triggering chat workflow for message_id=%s, session_id=%s", messageID, sessionID)
		// Simulate processing...
	}()
}

// TriggerSuggestionWorkflow simulates triggering a suggestion workflow in the background.
func TriggerSuggestionWorkflow(ctx context.Context, messageID string, sessionID string) {
	go func() {
		// TODO: Replace with actual suggestion workflow logic.
		log.Printf("[Background] Triggering suggestion workflow for message_id=%s, session_id=%s", messageID, sessionID)
		// Simulate processing...
	}()
}
