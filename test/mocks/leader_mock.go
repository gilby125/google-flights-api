package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MockLeaderElector is a mock implementation of leader election for testing.
type MockLeaderElector struct {
	mock.Mock
}

// Start starts the mock leader elector.
func (m *MockLeaderElector) Start() {
	m.Called()
}

// Stop stops the mock leader elector.
func (m *MockLeaderElector) Stop() {
	m.Called()
}

// IsLeader returns whether this instance is the leader.
func (m *MockLeaderElector) IsLeader() bool {
	args := m.Called()
	return args.Bool(0)
}

// InstanceID returns the instance identifier.
func (m *MockLeaderElector) InstanceID() string {
	args := m.Called()
	return args.String(0)
}
