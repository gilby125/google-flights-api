package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MockScheduler is a mock implementation for worker.Scheduler
// Note: This mocks the concrete type, which is less ideal than using interfaces,
// but necessary given the current Manager implementation.
type MockScheduler struct {
	mock.Mock
}

// Start mocks the Start method
func (m *MockScheduler) Start() error {
	args := m.Called()
	return args.Error(0)
}

// Stop mocks the Stop method
func (m *MockScheduler) Stop() {
	m.Called()
}

// AddJob mocks the AddJob method
// Even though the manager doesn't directly call AddJob in the current code,
// we include it for completeness if the interaction changes.
func (m *MockScheduler) AddJob(payload []byte) error {
	args := m.Called(payload)
	return args.Error(0)
}

// Note: We don't need to implement the underlying Cronner methods here,
// as the Manager interacts with the Scheduler's own Start/Stop methods.
