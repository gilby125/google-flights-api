package worker_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/gilby125/google-flights-api/queue"
	"github.com/gilby125/google-flights-api/test/mocks" // Import mocks
	"github.com/gilby125/google-flights-api/worker"
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var runSchedulerTests = os.Getenv("ENABLE_SCHEDULER_TESTS") == "1"

func skipUnlessWorkerSchedulerTests(t *testing.T) {
	if !runSchedulerTests {
		t.Skip("set ENABLE_SCHEDULER_TESTS=1 to run worker scheduler tests")
	}
}

// Test NewScheduler creation
func TestNewScheduler(t *testing.T) {
	skipUnlessWorkerSchedulerTests(t)
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
	skipUnlessWorkerSchedulerTests(t)
	mockQueue := new(mocks.MockQueue)
	mockDB := new(mocks.MockPostgresDB)
	mockCron := new(mocks.MockCronner)

	mockRows := new(mocks.MockRows)
	mockDB.On("ListJobs", mock.Anything).Return(mockRows, nil).Once()
	mockRows.On("Next").Return(true).Once()
	mockRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args[0].(*int) = 1
		*args[1].(*string) = "daily price graph"
		*args[2].(*string) = "@every 1m"
		*args[3].(*bool) = true
		*args[4].(**time.Time) = nil
		*args[5].(*time.Time) = time.Now()
		*args[6].(*time.Time) = time.Now()
	}).Once()
	mockRows.On("Next").Return(false).Once()
	mockRows.On("Close").Return(nil).Once()
	mockRows.On("Err").Return(nil).Once()

	// Expect scheduler cron to start
	mockCron.On("Start").Return().Once()
	mockCron.On("AddFunc", mock.Anything, mock.Anything).Return(cron.EntryID(1), nil).Maybe()

	scheduler := worker.NewScheduler(mockQueue, mockDB, mockCron)
	err := scheduler.Start()

	assert.NoError(t, err)
	mockCron.AssertExpectations(t)
}

// Test Scheduler Start method with AddFunc error

// Test Scheduler Stop method
func TestScheduler_Stop(t *testing.T) {
	skipUnlessWorkerSchedulerTests(t)
	mockQueue := new(mocks.MockQueue)
	mockDB := new(mocks.MockPostgresDB)
	mockCron := new(mocks.MockCronner)

	// Mock Stop to return a context that can be waited on
	stopCtx, cancel := context.WithCancel(context.Background())
	mockCron.On("Stop").Return(stopCtx).Once()
	cancel() // Immediately cancel the context for the test

	scheduler := worker.NewScheduler(mockQueue, mockDB, mockCron)
	// Need to start it first so Stop doesn't panic (though mock doesn't care)
	mockRows := new(mocks.MockRows)
	mockDB.On("ListJobs", mock.Anything).Return(mockRows, nil).Once()
	mockRows.On("Next").Return(false).Once()
	mockRows.On("Close").Return(nil).Once()
	mockRows.On("Err").Return(nil).Once()
	mockCron.On("Start").Return()
	mockCron.On("AddFunc", mock.Anything, mock.Anything).Return(cron.EntryID(1), nil).Maybe()
	_ = scheduler.Start()

	scheduler.Stop() // Call the method under test

	mockCron.AssertCalled(t, "Stop")
	mockCron.AssertExpectations(t)
}

// Test Scheduler AddJob method - Success
func TestScheduler_AddJob_Success(t *testing.T) {
	skipUnlessWorkerSchedulerTests(t)
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
	skipUnlessWorkerSchedulerTests(t)
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

func TestScheduler_EnqueuePriceGraphSweep_Success(t *testing.T) {
	skipUnlessWorkerSchedulerTests(t)

	mockQueue := new(mocks.MockQueue)
	mockDB := new(mocks.MockPostgresDB)
	mockCron := new(mocks.MockCronner)

	scheduler := worker.NewScheduler(mockQueue, mockDB, mockCron)

	payload := worker.PriceGraphSweepPayload{
		Origins:           []string{"JFK"},
		Destinations:      []string{"LAX"},
		DepartureDateFrom: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		DepartureDateTo:   time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		TripLengths:       []int{3, 5},
		TripType:          "round_trip",
		Class:             "business",
		Stops:             "nonstop",
		Adults:            2,
		Children:          1,
		InfantsLap:        0,
		InfantsSeat:       0,
		Currency:          "eur",
		RateLimitMillis:   0, // Should default to 750
	}

	ctxMatcher := mock.Anything
	expectedSweepID := 42

	mockDB.On("CreatePriceGraphSweep", ctxMatcher, sql.NullInt32{}, len(payload.Origins), len(payload.Destinations),
		sql.NullInt32{Int32: 3, Valid: true}, sql.NullInt32{Int32: 5, Valid: true}, "EUR").
		Return(expectedSweepID, nil).Once()

	mockQueue.On("Enqueue", ctxMatcher, "price_graph_sweep", mock.MatchedBy(func(arg interface{}) bool {
		p, ok := arg.(worker.PriceGraphSweepPayload)
		if !ok {
			return false
		}
		if p.SweepID != expectedSweepID {
			return false
		}
		if p.Currency != "EUR" {
			return false
		}
		if p.RateLimitMillis != 750 {
			return false
		}
		return reflect.DeepEqual(p.TripLengths, []int{3, 5})
	})).Return("queue-id", nil).Once()

	sweepID, err := scheduler.EnqueuePriceGraphSweep(context.Background(), payload)

	assert.NoError(t, err)
	assert.Equal(t, expectedSweepID, sweepID)
	mockDB.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}

func TestScheduler_EnqueuePriceGraphSweep_Validation(t *testing.T) {
	skipUnlessWorkerSchedulerTests(t)

	mockQueue := new(mocks.MockQueue)
	mockDB := new(mocks.MockPostgresDB)
	mockCron := new(mocks.MockCronner)
	scheduler := worker.NewScheduler(mockQueue, mockDB, mockCron)

	_, err := scheduler.EnqueuePriceGraphSweep(context.Background(), worker.PriceGraphSweepPayload{
		Origins:      []string{},
		Destinations: []string{"LAX"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one origin")

	_, err = scheduler.EnqueuePriceGraphSweep(context.Background(), worker.PriceGraphSweepPayload{
		Origins:      []string{"JFK"},
		Destinations: []string{},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one destination")

	mockDB.AssertNotCalled(t, "CreatePriceGraphSweep", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mockQueue.AssertNotCalled(t, "Enqueue", mock.Anything, mock.Anything, mock.Anything)
}

func TestScheduler_EnqueuePriceGraphSweep_QueueFailure(t *testing.T) {
	skipUnlessWorkerSchedulerTests(t)

	mockQueue := new(mocks.MockQueue)
	mockDB := new(mocks.MockPostgresDB)
	mockCron := new(mocks.MockCronner)
	scheduler := worker.NewScheduler(mockQueue, mockDB, mockCron)

	payload := worker.PriceGraphSweepPayload{
		Origins:           []string{"JFK"},
		Destinations:      []string{"LAX"},
		DepartureDateFrom: time.Now(),
		DepartureDateTo:   time.Now().AddDate(0, 0, 7),
		Currency:          "usd",
	}

	expectedSweepID := 7
	mockDB.On("CreatePriceGraphSweep", mock.Anything, sql.NullInt32{}, 1, 1, sql.NullInt32{Int32: 0, Valid: true}, sql.NullInt32{Int32: 0, Valid: true}, "USD").
		Return(expectedSweepID, nil).Once()

	queueErr := errors.New("queue unavailable")
	mockQueue.On("Enqueue", mock.Anything, "price_graph_sweep", mock.Anything).Return("", queueErr).Once()

	mockDB.On("UpdatePriceGraphSweepStatus", mock.Anything, expectedSweepID, "failed", sql.NullTime{}, sql.NullTime{}, 1).
		Return(nil).Once()

	sweepID, err := scheduler.EnqueuePriceGraphSweep(context.Background(), payload)

	assert.Error(t, err)
	assert.Equal(t, 0, sweepID)
	assert.Contains(t, err.Error(), queueErr.Error())
	mockDB.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}
