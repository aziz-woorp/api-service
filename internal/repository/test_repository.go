package repository

import "context"

// TestRepository defines the interface for test operations
type TestRepository interface {
	GetTestData(ctx context.Context, id string) (*TestData, error)
	CreateTestData(ctx context.Context, data *TestData) error
}

// TestData represents test data structure
type TestData struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Message string `json:"message"`
}