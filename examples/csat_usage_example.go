// Package examples demonstrates how to use the CSAT system.
package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/fraiday-org/api-service/internal/api/dto"
	"github.com/fraiday-org/api-service/internal/models"
)

// CSATUsageExample demonstrates the complete CSAT workflow.
func CSATUsageExample() {
	fmt.Println("=== CSAT System Usage Example ===")
	
	// Example client and channel IDs
	clientID := primitive.NewObjectID()
	channelID := primitive.NewObjectID()
	chatSessionID := "chat_session_123"
	
	fmt.Printf("Client ID: %s\n", clientID.Hex())
	fmt.Printf("Channel ID: %s\n", channelID.Hex())
	fmt.Printf("Chat Session ID: %s\n", chatSessionID)
	
	// Step 1: Configure CSAT for the client/channel
	fmt.Println("\n--- Step 1: Configure CSAT ---")
	configRequest := dto.CSATConfigurationRequest{
		Enabled: true,
		TriggerConditions: map[string]interface{}{
			"trigger_after_messages": 10,
			"trigger_on_session_end": true,
		},
	}
	configJSON, _ := json.MarshalIndent(configRequest, "", "  ")
	fmt.Printf("PUT /clients/%s/channels/%s/csat/config\n", clientID.Hex(), channelID.Hex())
	fmt.Printf("Request Body:\n%s\n", configJSON)
	
	// Step 2: Set up CSAT questions
	fmt.Println("\n--- Step 2: Configure CSAT Questions ---")
	questionsRequest := dto.CSATQuestionsRequest{
		Questions: []dto.CSATQuestionRequest{
			{
				QuestionText: "How would you rate your overall experience?",
				Options:      []string{"5", "4", "3", "2", "1"},
				Order:        1,
				Active:       true,
			},
			{
				QuestionText: "What could we improve?",
				Options:      []string{"Great", "Good", "Average", "Poor"},
				Order:        2,
				Active:       true,
			},
			{
				QuestionText: "Would you recommend our service?",
				QuestionType: "multiple_choice",
				Options:      []string{"Yes, definitely", "Maybe", "No"},
				Order:        3,
				Active:       true,
			},
		},
	}
	questionsJSON, _ := json.MarshalIndent(questionsRequest, "", "  ")
	fmt.Printf("PUT /clients/%s/channels/%s/csat/questions\n", clientID.Hex(), channelID.Hex())
	fmt.Printf("Request Body:\n%s\n", questionsJSON)
	
	// Step 3: Trigger CSAT survey
	fmt.Println("\n--- Step 3: Trigger CSAT Survey ---")
	triggerRequest := dto.CSATTriggerRequest{
		ChatSessionID: chatSessionID,
		ClientID:      clientID.Hex(),
		ChannelID:     channelID.Hex(),
	}
	triggerJSON, _ := json.MarshalIndent(triggerRequest, "", "  ")
	fmt.Printf("POST /csat/trigger\n")
	fmt.Printf("Request Body:\n%s\n", triggerJSON)
	
	// Simulated response
	csatSessionID := primitive.NewObjectID()
	triggerResponse := dto.CSATTriggerResponse{
		CSATSessionID: csatSessionID.Hex(),
		Status:        "pending",
		TriggeredAt:   time.Now().UTC(),
		Message:       "CSAT survey triggered successfully",
	}
	triggerResponseJSON, _ := json.MarshalIndent(triggerResponse, "", "  ")
	fmt.Printf("Response:\n%s\n", triggerResponseJSON)
	
	// Step 4: Simulate user responses
	fmt.Println("\n--- Step 4: User Responses ---")
	
	// Response to first question (rating)
	questionID1 := primitive.NewObjectID()
	response1 := dto.CSATResponseRequest{
		CSATSessionID: csatSessionID.Hex(),
		QuestionID:    questionID1.Hex(),
		ResponseValue: "5",
	}
	response1JSON, _ := json.MarshalIndent(response1, "", "  ")
	fmt.Printf("POST /csat/respond (Question 1 - Rating)\n")
	fmt.Printf("Request Body:\n%s\n", response1JSON)
	
	// Response to second question (text)
	questionID2 := primitive.NewObjectID()
	response2 := dto.CSATResponseRequest{
		CSATSessionID: csatSessionID.Hex(),
		QuestionID:    questionID2.Hex(),
		ResponseValue: "The response time could be faster",
	}
	response2JSON, _ := json.MarshalIndent(response2, "", "  ")
	fmt.Printf("\nPOST /csat/respond (Question 2 - Text)\n")
	fmt.Printf("Request Body:\n%s\n", response2JSON)
	
	// Response to third question (multiple choice)
	questionID3 := primitive.NewObjectID()
	response3 := dto.CSATResponseRequest{
		CSATSessionID: csatSessionID.Hex(),
		QuestionID:    questionID3.Hex(),
		ResponseValue: "Yes, definitely",
	}
	response3JSON, _ := json.MarshalIndent(response3, "", "  ")
	fmt.Printf("\nPOST /csat/respond (Question 3 - Multiple Choice)\n")
	fmt.Printf("Request Body:\n%s\n", response3JSON)
	
	// Step 5: Check session status
	fmt.Println("\n--- Step 5: Check Session Status ---")
	fmt.Printf("GET /csat/sessions/%s\n", csatSessionID.Hex())
	
	// Simulated session response
	sessionResponse := dto.CSATSessionResponse{
		ID:                   csatSessionID.Hex(),
		ChatSessionID:        chatSessionID,
		ClientID:             clientID.Hex(),
		ChannelID:            channelID.Hex(),
		Status:               "completed",
		TriggeredAt:          time.Now().UTC().Add(-5 * time.Minute),
		CompletedAt:          &[]time.Time{time.Now().UTC()}[0],
		CurrentQuestionIndex: 3,
		QuestionsSent:        []string{questionID1.Hex(), questionID2.Hex(), questionID3.Hex()},
		CreatedAt:            time.Now().UTC().Add(-5 * time.Minute),
		UpdatedAt:            time.Now().UTC(),
	}
	sessionResponseJSON, _ := json.MarshalIndent(sessionResponse, "", "  ")
	fmt.Printf("Response:\n%s\n", sessionResponseJSON)
	
	// Step 6: Show event flow
	fmt.Println("\n--- Step 6: Event Flow ---")
	fmt.Println("Events published during CSAT workflow:")
	fmt.Printf("1. %s - CSAT survey triggered\n", models.EventTypeCSATTriggered)
	fmt.Printf("2. %s - Question 1 sent\n", models.EventTypeCSATMessageSent)
	fmt.Printf("3. %s - Question 2 sent\n", models.EventTypeCSATMessageSent)
	fmt.Printf("4. %s - Question 3 sent\n", models.EventTypeCSATMessageSent)
	fmt.Printf("5. %s - Survey completed\n", models.EventTypeCSATCompleted)
	
	fmt.Println("\n--- Chat Messages Generated ---")
	fmt.Println("1. Question: 'How would you rate your overall experience?' (with rating buttons)")
	fmt.Println("2. Question: 'What could we improve?' (text input)")
	fmt.Println("3. Question: 'Would you recommend our service?' (with choice buttons)")
	fmt.Println("4. Thank you message: 'Thank you for your feedback!'")
	
	fmt.Println("\n=== CSAT Workflow Complete ===")
}

// ExampleChatMessage shows how CSAT questions appear as chat messages.
func ExampleChatMessage() {
	fmt.Println("\n=== Example CSAT Chat Message ===")
	
	// Example of a rating question message
	ratingMessage := models.ChatMessage{
		Sender:     "system",
		SenderName: "CSAT Survey",
		SenderType: string(models.SenderTypeSystem),
		Text:       "How would you rate your overall experience?",
		Attachments: []models.Attachment{
			{
				Type: "carousel",
				Carousel: map[string]interface{}{
					"type": "buttons",
					"buttons": []map[string]interface{}{
						{"text": "5", "value": "5", "type": "button"},
						{"text": "4", "value": "4", "type": "button"},
						{"text": "3", "value": "3", "type": "button"},
						{"text": "2", "value": "2", "type": "button"},
						{"text": "1", "value": "1", "type": "button"},
					},
				},
			},
		},
		Category: models.MessageCategoryInfo,
		Data: map[string]interface{}{
			"csat_message":    true,
			"csat_session_id": primitive.NewObjectID().Hex(),
			"question_id":     primitive.NewObjectID().Hex(),
			"question_type":   "rating",
			"options":         []string{"5", "4", "3", "2", "1"},
		},
	}
	
	messageJSON, _ := json.MarshalIndent(ratingMessage, "", "  ")
	fmt.Printf("Rating Question Message:\n%s\n", messageJSON)
}
