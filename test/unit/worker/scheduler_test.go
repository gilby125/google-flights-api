package worker_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	// "time" // Removed unused import

	// "github.com/gilby125/google-flights-api/db" // Removed unused import
	"github.com/gilby125/google-flights-api/queue"
	"github.com/gilby125/google-flights-api/worker"
	"github.com/gilby125/google-flights-api/test/mocks" // Import mocks
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test NewScheduler creation
func TestNewScheduler(t *testing.T) {
	mockQueue := new(mocks.MockQueue)
	mockDB := new(mocks.MockPostgresDB) // Scheduler depends on DB too

	// Test with nil cronner (should use default)
	schedulerDefault := worker.NewScheduler(mockQueue, mockDB, nil)
	assert.NotNil(t, schedulerDefault, "Scheduler should be created with default cronner")

	// Test with provided mock cronner
	mockCron := new(mocks.MockCronner)
	schedulerMocked := worker.NewScheduler(mockQueue, mockDB, mockCron)
	assert.NotNil(t, schedulerMocked, "Scheduler should be created with mock cronner")
	// We could potentially assert that schedulerMocked.cron == mockCron if the field were exported,
	// but testing behavior via methods is generally preferred.
}

// Test Scheduler Start method
func TestScheduler_Start(t *testing.T) {
	mockQueue := new(mocks.MockQueue)
	mockDB := new(mocks.MockPostgresDB)
	mockCron := new(mocks.MockCronner)

	// Expect Start and AddFunc to be called
	mockCron.On("Start").Return().Once()
	// Expect the example "@every 1m" job to be added
	mockCron.On("AddFunc", "@every 1m", mock.AnythingOfType("func()")).Return(cron.EntryID(1), nil).Once()

	scheduler := worker.NewScheduler(mockQueue, mockDB, mockCron)
	err := scheduler.Start()

	assert.NoError(t, err)
	mockCron.AssertExpectations(t)
}

// Test Scheduler Start method with AddFunc error
func TestScheduler_Start_AddFuncError(t *testing.T) {
	mockQueue := new(mocks.MockQueue)
	mockDB := new(mocks.MockPostgresDB)
	mockCron := new(mocks.MockCronner)
	expectedError := errors.New("failed to add cron func")

	// Expect Start to be called
	mockCron.On("Start").Return().Once()
	// Expect AddFunc to be called and return an error
	mockCron.On("AddFunc", "@every 1m", mock.AnythingOfType("func()")).Return(cron.EntryID(0), expectedError).Once()

	scheduler := worker.NewScheduler(mockQueue, mockDB, mockCron)
	err := scheduler.Start()

	assert.Error(t, err)
	assert.ErrorContains(t, err, expectedError.Error())
	mockCron.AssertExpectations(t)
}

// Test Scheduler Stop method
func TestScheduler_Stop(t *testing.T) {
	mockQueue := new(mocks.MockQueue)
	mockDB := new(mocks.MockPostgresDB)
	mockCron := new(mocks.MockCronner)

	// Mock Stop to return a context that can be waited on
	stopCtx, cancel := context.WithCancel(context.Background())
	mockCron.On("Stop").Return(stopCtx).Once()
	cancel() // Immediately cancel the context for the test

	scheduler := worker.NewScheduler(mockQueue, mockDB, mockCron)
	// Need to start it first so Stop doesn't panic (though mock doesn't care)
	mockCron.On("Start").Return()
	mockCron.On("AddFunc", mock.Anything, mock.Anything).Return(cron.EntryID(1), nil)
	_ = scheduler.Start()

	scheduler.Stop() // Call the method under test

	mockCron.AssertCalled(t, "Stop")
	mockCron.AssertExpectations(t)
}

// Test Scheduler AddJob method - Success
func TestScheduler_AddJob_Success(t *testing.T) {
	mockQueue := new(mocks.MockQueue)
	mockDB := new(mocks.MockPostgresDB)
	mockCron := new(mocks.MockCronner) // Cronner is not used by AddJob, but needed for NewScheduler

	payload := []byte(`{"data":"test"}`)
	expectedQueueName := "scheduled_jobs"
	expectedJobID := "mockJobID" // Mock Enqueue will return this

	// Expect Enqueue to be called with the correct queue name and marshaled job data
	mockQueue.On("Enqueue", mock.Anything, expectedQueueName, mock.MatchedBy(func(jobBytes []byte) bool {
		var job queue.Job
		err := json.Unmarshal(jobBytes, &job)
		assert.NoError(t, err)
		assert.Equal(t, "scheduled_job", job.Type)
		assert.Equal(t, json.RawMessage(payload), job.Payload)
		assert.NotEmpty(t, job.ID) // ID is generated internally
		return true
	})).Return(expectedJobID, nil).Once()

	scheduler := worker.NewScheduler(mockQueue, mockDB, mockCron)
	err := scheduler.AddJob(payload)

	assert.NoError(t, err)
	mockQueue.AssertExpectations(t)
}

// Test Scheduler AddJob method - Enqueue Error (e.g., queue full)
func TestScheduler_AddJob_EnqueueError(t *testing.T) {
	mockQueue := new(mocks.MockQueue)
	mockDB := new(mocks.MockPostgresDB)
	mockCron := new(mocks.MockCronner)
	expectedError := errors.New("queue is full")

	payload := []byte(`{"data":"test"}`)
	expectedQueueName := "scheduled_jobs"

	// Expect Enqueue to be called and return an error
	mockQueue.On("Enqueue", mock.Anything, expectedQueueName, mock.AnythingOfType("[]uint8")).Return("", expectedError).Once()

	scheduler := worker.NewScheduler(mockQueue, mockDB, mockCron)
	err := scheduler.AddJob(payload)

	assert.Error(t, err)
	assert.ErrorContains(t, err, expectedError.Error())
	assert.ErrorContains(t, err, "failed to enqueue job") // Check error wrapping
	mockQueue.AssertExpectations(t)
}

// Test Scheduler AddJob method - Marshal Error (unlikely with []byte, but good practice)
func TestScheduler_AddJob_MarshalError(t *testing.T) {
	// This test is more theoretical as marshaling queue.Job with a []byte payload is unlikely to fail.
	// However, if the Job struct or AddJob logic changes, this might become relevant.
	// We can simulate it by passing an unmarshalable payload type if AddJob took interface{},
	// but since it takes []byte, we'll skip the direct simulation.

	// If AddJob were modified to take an interface{} payload:
	// mockQueue := new(mocks.MockQueue)
	// mockDB := new(mocks.MockPostgresDB)
	// mockCron := new(mocks.MockCronner)
	//
	// // Payload that cannot be marshaled (e.g., a channel)
	// invalidPayload := make(chan int)
	//
	// // We don't expect Enqueue to be called if marshaling fails
	//
	// scheduler := worker.NewScheduler(mockQueue, mockDB, mockCron)
	// // Assuming AddJob is modified to accept interface{}
	// // err := scheduler.AddJob(invalidPayload)
	// // assert.Error(t, err)
	// // assert.ErrorContains(t, err, "failed to marshal job")
	// // mockQueue.AssertNotCalled(t, "Enqueue", mock.Anything, mock.Anything, mock.Anything)

	t.Skip("Skipping marshal error test as current AddJob signature makes it hard to trigger")
}
