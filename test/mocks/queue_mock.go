package mocks

import (
	"context"

	"github.com/gilby125/google-flights-api/queue" // Adjust import path if necessary
	"github.com/stretchr/testify/mock"
)

// MockQueue is a mock implementation of the queue.Queue interface
type MockQueue struct {
	mock.Mock
}

// Enqueue mocks the Enqueue method
func (m *MockQueue) Enqueue(ctx context.Context, jobType string, payload interface{}) (string, error) {
	args := m.Called(ctx, jobType, payload)
	return args.String(0), args.Error(1)
}

// Dequeue mocks the Dequeue method
func (m *MockQueue) Dequeue(ctx context.Context, queueName string) (*queue.Job, error) {
	args := m.Called(ctx, queueName)
	// Need to handle the case where the first return value is nil
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*queue.Job), args.Error(1)
}

// Ack mocks the Ack method
func (m *MockQueue) Ack(ctx context.Context, queueName, jobID string) error {
	args := m.Called(ctx, queueName, jobID)
	return args.Error(0)
}

// Nack mocks the Nack method
func (m *MockQueue) Nack(ctx context.Context, queueName, jobID string) error {
	args := m.Called(ctx, queueName, jobID)
	return args.Error(0)
}

// GetJobStatus mocks the GetJobStatus method
func (m *MockQueue) GetJobStatus(ctx context.Context, jobID string) (string, error) {
	args := m.Called(ctx, jobID)
	return args.String(0), args.Error(1)
}

// GetQueueStats mocks the GetQueueStats method
func (m *MockQueue) GetQueueStats(ctx context.Context, queueName string) (map[string]int64, error) {
	args := m.Called(ctx, queueName)
	// Need to handle the case where the first return value is nil
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

// Ensure MockQueue implements the interface
var _ queue.Queue = (*MockQueue)(nil)
