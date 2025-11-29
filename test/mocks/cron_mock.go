package mocks

import (
	"context"

	"github.com/gilby125/google-flights-api/worker" // Adjust import path if necessary
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/mock"
)

// MockCronner is a mock implementation of the worker.Cronner interface
type MockCronner struct {
	mock.Mock
}

// Start mocks the Start method
func (m *MockCronner) Start() {
	m.Called()
}

// Stop mocks the Stop method
func (m *MockCronner) Stop() context.Context {
	args := m.Called()
	// Return a context, handle nil case if necessary
	if args.Get(0) == nil {
		// Return a default context if no specific one is provided by the mock setup
		return context.Background()
	}
	return args.Get(0).(context.Context)
}

// AddFunc mocks the AddFunc method
func (m *MockCronner) AddFunc(spec string, cmd func()) (cron.EntryID, error) {
	args := m.Called(spec, cmd) // Note: Mocking functions passed as arguments can be tricky.
	// Often, it's better to assert that *a* function was passed rather than *which* function.
	// Or, the mock setup can provide a specific function to expect.
	// For simplicity here, we just register the call.

	// Return type is (cron.EntryID, error)
	// Assuming EntryID is an int or similar simple type for mocking
	// Return the configured values directly. Assert the type for EntryID.
	entryID := args.Get(0)
	return entryID.(cron.EntryID), args.Error(1)
}

// Ensure MockCronner implements the interface
var _ worker.Cronner = (*MockCronner)(nil)
