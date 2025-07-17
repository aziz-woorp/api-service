package repoimpl

import (
	"context"
	"fmt"

	"github.com/example/api-service/internal/repository"
)

// testRepositoryImpl implements the TestRepository interface
type testRepositoryImpl struct {
	// In a real implementation, this would have database connection, etc.
	data map[string]*repository.TestData
}

// NewTestRepository creates a new test repository instance
func NewTestRepository() repository.TestRepository {
	return &testRepositoryImpl{
		data: make(map[string]*repository.TestData),
	}
}

// GetTestData retrieves test data by ID
func (r *testRepositoryImpl) GetTestData(ctx context.Context, id string) (*repository.TestData, error) {
	if data, exists := r.data[id]; exists {
		return data, nil
	}
	return nil, fmt.Errorf("test data with ID %s not found", id)
}

// CreateTestData creates new test data
func (r *testRepositoryImpl) CreateTestData(ctx context.Context, data *repository.TestData) error {
	if data.ID == "" {
		return fmt.Errorf("test data ID cannot be empty")
	}
	r.data[data.ID] = data
	return nil
}