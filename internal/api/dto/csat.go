// Package dto defines request and response structures for CSAT API endpoints.
package dto

import (
	"time"

)

// CSATTriggerRequest represents a request to trigger a CSAT survey.
type CSATTriggerRequest struct {
	SessionID string `json:"session_id" validate:"required"`
	Type      string `json:"type" validate:"required,min=1"`
}

// CSATTriggerResponse represents a response after triggering a CSAT survey.
type CSATTriggerResponse struct {
	CSATSessionID string    `json:"csat_session_id"`
	Status        string    `json:"status"`
	TriggeredAt   time.Time `json:"triggered_at"`
	Message       string    `json:"message"`
}

// CSATResponseRequest represents a request to respond to a CSAT question.
type CSATResponseRequest struct {
	SessionID        string `json:"session_id" validate:"required"`
	CSATQuestionID   string `json:"csat_question_id" validate:"required"`
	ResponseValue    string `json:"response_value" validate:"required"`
}

// CSATResponseResponse represents a response after submitting a CSAT response.
type CSATResponseResponse struct {
	ResponseID string `json:"response_id"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}

// CSATConfigurationRequest represents a request to create/update CSAT configuration.
type CSATConfigurationRequest struct {
	Type              string                 `json:"type" validate:"required,min=1"`
	Enabled           bool                   `json:"enabled"`
	TriggerConditions map[string]interface{} `json:"trigger_conditions,omitempty"`
}

// CSATConfigurationResponse represents a CSAT configuration response.
type CSATConfigurationResponse struct {
	ID                string                 `json:"id"`
	ClientID          string                 `json:"client_id"`
	ChannelID         string                 `json:"channel_id"`
	Type              string                 `json:"type"`
	Enabled           bool                   `json:"enabled"`
	TriggerConditions map[string]interface{} `json:"trigger_conditions,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// CSATQuestionRequest represents a request to create/update CSAT questions.
type CSATQuestionRequest struct {
	QuestionText string   `json:"question_text" validate:"required"`
	Options      []string `json:"options" validate:"required"`
	Order        int      `json:"order" validate:"required"`
	Active       bool     `json:"active"`
}

// CSATQuestionsRequest represents a request to update multiple CSAT questions.
type CSATQuestionsRequest struct {
	Questions []CSATQuestionRequest `json:"questions" validate:"required"`
}

// CSATQuestionResponse represents a CSAT question response.
type CSATQuestionResponse struct {
	ID                   string    `json:"id"`
	CSATConfigurationID  string    `json:"csat_configuration_id"`
	QuestionText         string    `json:"question_text"`
	Options              []string  `json:"options"`
	Order                int       `json:"order"`
	Active               bool      `json:"active"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// CSATSessionResponse represents a CSAT session response.
type CSATSessionResponse struct {
	ID                   string     `json:"id"`
	ChatSessionID        string     `json:"chat_session_id"`
	CSATConfigurationID  string     `json:"csat_configuration_id"`
	ClientID             string     `json:"client_id"`
	ChannelID            string     `json:"channel_id"`
	ThreadSessionID      *string    `json:"thread_session_id,omitempty"`
	ThreadContext        bool       `json:"thread_context"`
	Status               string     `json:"status"`
	TriggeredAt          time.Time  `json:"triggered_at"`
	CompletedAt          *time.Time `json:"completed_at,omitempty"`
	CurrentQuestionIndex int        `json:"current_question_index"`
	QuestionsSent        []string   `json:"questions_sent"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// CSATAnalyticsResponse represents CSAT analytics data.
type CSATAnalyticsResponse struct {
	TotalSurveys    int                    `json:"total_surveys"`
	CompletedSurveys int                   `json:"completed_surveys"`
	CompletionRate  float64                `json:"completion_rate"`
	AverageRating   float64                `json:"average_rating,omitempty"`
	ResponseBreakdown map[string]int       `json:"response_breakdown,omitempty"`
	TimeRange       map[string]interface{} `json:"time_range"`
}
