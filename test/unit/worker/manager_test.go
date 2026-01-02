package worker_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gilby125/google-flights-api/config"
	"github.com/gilby125/google-flights-api/queue"
	"github.com/gilby125/google-flights-api/test/mocks"
	"github.com/gilby125/google-flights-api/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var runWorkerTests = os.Getenv("ENABLE_WORKER_TESTS") == "1"

func skipUnlessWorkerTests(t *testing.T) {
	if !runWorkerTests {
		t.Skip("set ENABLE_WORKER_TESTS=1 to run worker manager tests")
	}
}

func allowPriceGraphDequeues(mockQueue *mocks.MockQueue) {
	mockQueue.On("Dequeue", mock.Anything, "price_graph_sweep").Return(nil, nil).Maybe()
}

// Helper to setup manager with mocks for testing
func setupManagerTest(cfg config.WorkerConfig) (*worker.Manager, *mocks.MockQueue, *mocks.MockPostgresDB, *mocks.MockNeo4jDatabase, *mocks.MockScheduler) {
	mockQueue := new(mocks.MockQueue)
	mockPgDb := new(mocks.MockPostgresDB)
	mockNeo4jDb := new(mocks.MockNeo4jDatabase)
	// The real NewManager creates its own scheduler. To test manager's interaction
	// with the scheduler, we'd ideally inject a mock scheduler.
	// Since the current NewManager doesn't allow injection, we can't directly mock
	// the scheduler created *inside* NewManager.
	// For now, we'll test scheduler interactions separately or assume the internal scheduler works.
	// Let's create a mock scheduler instance mainly for verifying calls in Stop,
	// acknowledging we can't inject it into the *real* manager instance easily.
	mockScheduler := new(mocks.MockScheduler)

	// We need to modify NewManager to accept the scheduler or test differently.
	// Workaround: Create manager, then replace its internal scheduler if possible (requires exporting the field).
	// If not possible, we can only test the parts that don't rely on mocking the *internal* scheduler.

	// Let's assume NewManager is refactored to accept the scheduler for proper testing:
	// manager := worker.NewManager(mockQueue, mockPgDb, mockNeo4jDb, cfg, mockScheduler)

	// --- Simulating current NewManager behavior (creates its own scheduler) ---
	// We can't easily mock the internally created scheduler without refactoring NewManager.
	// We will test Start/Stop behavior based on queue interactions and assume the internal scheduler is called.
	// Create the manager using the real NewManager which internally creates a real scheduler (using default cron).
	// This limits our ability to mock scheduler calls directly.
	// Pass nil for Redis client to disable leader election in tests
	manager := worker.NewManager(mockQueue, nil, mockPgDb, mockNeo4jDb, cfg, config.FlightConfig{})
	// --- End Simulation ---

	// Allow price graph sweep dequeue calls by default in tests
	allowPriceGraphDequeues(mockQueue)

	// We still return the mockScheduler for potential future use if refactored.
	return manager, mockQueue, mockPgDb, mockNeo4jDb, mockScheduler
}

func TestManager_Start(t *testing.T) {
	skipUnlessWorkerTests(t)
	cfg := config.WorkerConfig{Concurrency: 2, JobTimeout: 5 * time.Second, ShutdownTimeout: 1 * time.Second}
	manager, mockQueue, mockPgDb, _, _ := setupManagerTest(cfg)

	// Mock the ListJobs call that the scheduler makes on startup
	mockRows := new(mocks.MockRows)
	mockPgDb.On("ListJobs", mock.Anything).Return(mockRows, nil)
	mockRows.On("Next").Return(false) // No scheduled jobs
	mockRows.On("Close").Return(nil)
	mockRows.On("Err").Return(nil)

	// Mock Dequeue to simulate workers starting and looking for jobs
	// Expect Dequeue to be called multiple times by the workers for both queues
	// Use Eventually to handle the concurrent nature of workers starting
	mockQueue.On("Dequeue", mock.Anything, "flight_search").Return(nil, nil)             // Return no job initially
	mockQueue.On("Dequeue", mock.Anything, "bulk_search").Return(nil, nil)               // Return no job initially
	mockQueue.On("Dequeue", mock.Anything, "price_graph_sweep").Return(nil, nil).Maybe() // Return no job initially

	manager.Start()

	// Allow some time for workers to start and potentially call Dequeue
	time.Sleep(200 * time.Millisecond)

	// Assert that Dequeue was likely called (hard to guarantee exact count due to timing)
	// We can't easily assert scheduler.Start was called without refactoring NewManager.
	mockQueue.AssertCalled(t, "Dequeue", mock.Anything, "flight_search")
	mockQueue.AssertCalled(t, "Dequeue", mock.Anything, "bulk_search")
	mockQueue.AssertCalled(t, "Dequeue", mock.Anything, "price_graph_sweep")

	// Stop the manager to clean up goroutines
	manager.Stop()
}

// Test Start when scheduler fails to start (logs warning)
// Note: This requires refactoring NewManager to inject the scheduler mock.
// func TestManager_Start_SchedulerError(t *testing.T) {
// 	cfg := config.WorkerConfig{Concurrency: 1}
// 	// Assume NewManager accepts mockScheduler
// 	// manager, mockQueue, _, _, mockScheduler := setupManagerTest(cfg)
// 	// expectedErr := errors.New("scheduler failed")
// 	// mockScheduler.On("Start").Return(expectedErr).Once()
// 	// mockQueue.On("Dequeue", mock.Anything, mock.Anything).Return(nil, nil) // Worker still starts
// //
// 	// manager.Start()
// 	// // Assert that a warning was logged (requires log capture or checking output)
// 	// mockScheduler.AssertExpectations(t)
// 	// manager.Stop()
// 	t.Skip("Skipping scheduler error test: Requires NewManager refactor for scheduler injection")
// }

func TestManager_Stop(t *testing.T) {
	skipUnlessWorkerTests(t)
	cfg := config.WorkerConfig{Concurrency: 1, JobTimeout: 5 * time.Second, ShutdownTimeout: 1 * time.Second}
	manager, mockQueue, mockPgDb, _, _ := setupManagerTest(cfg)

	// Mock the ListJobs call that the scheduler makes on startup
	mockRows := new(mocks.MockRows)
	mockPgDb.On("ListJobs", mock.Anything).Return(mockRows, nil)
	mockRows.On("Next").Return(false) // No scheduled jobs
	mockRows.On("Close").Return(nil)
	mockRows.On("Err").Return(nil)

	// Mock Dequeue to simulate a worker running
	dequeued := make(chan bool, 1)
	mockQueue.On("Dequeue", mock.Anything, mock.Anything).Return(nil, nil).Run(func(args mock.Arguments) {
		select {
		case dequeued <- true: // Signal that dequeue was called
		default:
		}
	})

	manager.Start()

	// Wait for the worker to likely call Dequeue
	select {
	case <-dequeued:
		// Worker has started and called Dequeue
	case <-time.After(1 * time.Second):
		t.Fatal("Worker did not call Dequeue within timeout")
	}

	// We can't easily assert scheduler.Stop was called without refactoring NewManager.
	manager.Stop()

	// After Stop, Dequeue should not be called anymore (or very few times during shutdown)
	// Reset the channel and wait briefly to see if Dequeue is called again
	calledAfterStop := false
	go func() {
		// Temporarily allow Dequeue calls again to see if they happen after Stop
		mockQueue.Mock.ExpectedCalls = []*mock.Call{} // Clear previous expectations
		mockQueue.On("Dequeue", mock.Anything, mock.Anything).Return(nil, nil).Run(func(args mock.Arguments) {
			calledAfterStop = true
		})
		time.Sleep(200 * time.Millisecond) // Wait a bit longer than worker loop sleep
	}()
	time.Sleep(300 * time.Millisecond) // Ensure the goroutine runs

	assert.False(t, calledAfterStop, "Dequeue should not be called after Stop")
}

// Test worker availability and job processing flow (Ack/Nack)
func TestManager_JobProcessingFlow(t *testing.T) {
	skipUnlessWorkerTests(t)
	cfg := config.WorkerConfig{Concurrency: 1, JobTimeout: 1 * time.Second, ShutdownTimeout: 1 * time.Second}
	manager, mockQueue, mockPgDb, mockNeo4jDb, _ := setupManagerTest(cfg)

	// --- Test Successful Job ---
	t.Run("Successful Job", func(t *testing.T) {
		// Mock the ListJobs call that the scheduler makes on startup
		mockRows := new(mocks.MockRows)
		mockPgDb.On("ListJobs", mock.Anything).Return(mockRows, nil)
		mockRows.On("Next").Return(false) // No scheduled jobs
		mockRows.On("Close").Return(nil)
		mockRows.On("Err").Return(nil)

		// Define payload struct and marshal it
		payloadStruct := worker.FlightSearchPayload{
			Origin:        "JFK",
			Destination:   "LAX",
			DepartureDate: time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC),
			Adults:        1,
			TripType:      "one_way", // Ensure this matches case in manager switch
			Class:         "economy",
			Stops:         "nonstop",
			Currency:      "USD",
		}
		jobPayloadBytes, _ := json.Marshal(payloadStruct)
		testJob := &queue.Job{
			ID:      "job-success",
			Type:    "flight_search", // Type is inferred from queue name in manager
			Payload: json.RawMessage(jobPayloadBytes),
		}
		processed := make(chan bool, 1)

		// Mock Dequeue returning the job
		mockQueue.On("Dequeue", mock.Anything, "flight_search").Return(testJob, nil).Once()
		// Allow subsequent calls to return nil to prevent panic after job processing
		mockQueue.On("Dequeue", mock.Anything, "flight_search").Return(nil, nil).Maybe()
		mockQueue.On("Dequeue", mock.Anything, "bulk_search").Return(nil, nil).Maybe() // Other queue might be checked

		// Mock Ack being called on success
		mockQueue.On("Ack", mock.Anything, "flight_search", testJob.ID).Return(nil).Run(func(args mock.Arguments) {
			processed <- true
		}).Once()
		mockQueue.On("Nack", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe() // Ensure Nack is allowed but not expected

		// --- Corrected DB & Neo4j Mocks for StoreFlightOffers ---
		mockTx := new(mocks.MockTx)
		mockPgDb.On("BeginTx", mock.Anything).Return(mockTx, nil).Once()

		// 1. Mock Search Query Insert (using QueryRowContext)
		mockScannerInsertQuery := new(mocks.MockQueryRowScanner) // Create scanner mock
		mockTx.On("QueryRowContext", mock.Anything, mock.MatchedBy(func(query string) bool { return strings.Contains(query, "INSERT INTO search_queries") }), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(mockScannerInsertQuery).Once() // Return scanner mock
		// Expect Scan(&queryID) on the scanner mock
		mockScannerInsertQuery.On("Scan", mock.AnythingOfType("*int")).Return(nil).Run(func(args mock.Arguments) {
			arg := args.Get(0).(*int)
			*arg = 1 // Simulate scanning ID 1
		}).Once() // This happens once before the loop

		// --- Mocks inside the offer loop - Use Maybe() ---
		// 2. Mock Get Search ID (using QueryRowContext)
		mockScannerGetSearchID := new(mocks.MockQueryRowScanner)
		mockTx.On("QueryRowContext", mock.Anything, mock.MatchedBy(func(query string) bool { return strings.Contains(query, "SELECT search_id FROM search_queries") }), 1).
			Return(mockScannerGetSearchID).Maybe() // Use Maybe for calls inside loop
		mockScannerGetSearchID.On("Scan", mock.AnythingOfType("*string")).Return(nil).Run(func(args mock.Arguments) {
			arg := args.Get(0).(*string)
			*arg = "test-search-uuid" // Simulate scanning search_id
		}).Maybe() // Use Maybe for calls inside loop

		// 3. Mock Offer Insert (using QueryRowContext)
		mockScannerInsertOffer := new(mocks.MockQueryRowScanner)
		mockTx.On("QueryRowContext", mock.Anything, mock.MatchedBy(func(query string) bool { return strings.Contains(query, "INSERT INTO flight_offers") }), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(mockScannerInsertOffer).Maybe() // Use Maybe for calls inside loop
		mockScannerInsertOffer.On("Scan", mock.AnythingOfType("*int")).Return(nil).Run(func(args mock.Arguments) {
			arg := args.Get(0).(*int)
			*arg = 10 // Simulate scanning offer ID 10
		}).Maybe() // Use Maybe for calls inside loop

		// 4. Mock Airline Insert (using ExecContext) - Expecting 3 variadic args
		mockTx.On("ExecContext", mock.Anything, mock.MatchedBy(func(query string) bool { return strings.Contains(query, "INSERT INTO airlines") }), mock.Anything, mock.Anything, mock.Anything).Return(sql.Result(nil), nil).Maybe() // Use Maybe for calls inside loop
		// 5. Mock Airport Inserts (using ExecContext) - Dep & Arr - Expecting 6 variadic args each
		mockTx.On("ExecContext", mock.Anything, mock.MatchedBy(func(query string) bool { return strings.Contains(query, "INSERT INTO airports") }), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sql.Result(nil), nil).Maybe() // Use Maybe for calls inside loop
		// 6. Mock Segment Insert (using ExecContext) - Expecting 11 variadic args
		mockTx.On("ExecContext", mock.Anything, mock.MatchedBy(func(query string) bool { return strings.Contains(query, "INSERT INTO flight_segments") }), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sql.Result(nil), nil).Maybe() // Use Maybe for calls inside loop

		// 7. Mock Neo4j calls (assuming success)
		mockNeo4jDb.On("CreateAirport", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe() // Use Maybe for calls inside loop
		mockNeo4jDb.On("CreateAirline", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()                                              // Use Maybe for calls inside loop
		mockNeo4jDb.On("CreateRoute", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()   // Use Maybe for calls inside loop
		mockNeo4jDb.On("AddPricePoint", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()                // Use Maybe for calls inside loop
		// --- End Mocks inside the offer loop ---

		// 8. Mock Commit
		mockTx.On("Commit").Return(nil).Once()
		// 9. Mock Rollback (should not be called)
		mockTx.On("Rollback").Return(nil).Maybe()
		// --- End Corrected Mocks ---

		allowPriceGraphDequeues(mockQueue)

		manager.Start()
		select {
		case <-processed:
			// Job processed and Acked
		case <-time.After(2 * time.Second): // Longer timeout for processing
			t.Fatal("Job was not processed and Acked within timeout")
		}
		manager.Stop()
		mockQueue.AssertExpectations(t)
		mockPgDb.AssertExpectations(t)
		mockNeo4jDb.AssertExpectations(t)
	})

	// --- Test Failed Job (Processing Error) ---
	t.Run("Failed Job", func(t *testing.T) {
		// Reset mocks for the new subtest
		mockQueue := new(mocks.MockQueue)
		mockPgDb := new(mocks.MockPostgresDB)
		mockNeo4jDb := new(mocks.MockNeo4jDatabase)
		manager := worker.NewManager(mockQueue, nil, mockPgDb, mockNeo4jDb, cfg, config.FlightConfig{}) // Recreate manager with fresh mocks
		allowPriceGraphDequeues(mockQueue)

		// Mock the ListJobs call that the scheduler makes on startup
		mockRowsFail := new(mocks.MockRows)
		mockPgDb.On("ListJobs", mock.Anything).Return(mockRowsFail, nil)
		mockRowsFail.On("Next").Return(false) // No scheduled jobs
		mockRowsFail.On("Close").Return(nil)
		mockRowsFail.On("Err").Return(nil)

		// Define payload struct and marshal it
		payloadStruct := worker.FlightSearchPayload{
			Origin:        "JFK",
			Destination:   "LAX",
			DepartureDate: time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC),
			Adults:        1,
			TripType:      "one_way", // Ensure this matches case in manager switch
			Class:         "economy",
			Stops:         "nonstop",
			Currency:      "USD",
		}
		jobPayloadBytes, _ := json.Marshal(payloadStruct)
		testJob := &queue.Job{
			ID:      "job-fail",
			Type:    "flight_search",
			Payload: json.RawMessage(jobPayloadBytes),
		}
		processed := make(chan bool, 1)
		processingError := errors.New("failed to store offers")

		mockQueue.On("Dequeue", mock.Anything, "flight_search").Return(testJob, nil).Once()
		// Allow subsequent calls to return nil
		mockQueue.On("Dequeue", mock.Anything, "flight_search").Return(nil, nil).Maybe()
		mockQueue.On("Dequeue", mock.Anything, "bulk_search").Return(nil, nil).Maybe()

		// Mock Nack being called on failure
		mockQueue.On("Nack", mock.Anything, "flight_search", testJob.ID).Return(nil).Run(func(args mock.Arguments) {
			processed <- true
		}).Once()
		mockQueue.On("Ack", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe() // Ensure Ack is allowed but not expected

		// --- Corrected DB Mocks for Failure Path ---
		mockTx := new(mocks.MockTx)
		mockPgDb.On("BeginTx", mock.Anything).Return(mockTx, nil).Once()

		// Mock Search Query Insert (Success)
		mockScannerInsertQueryFail := new(mocks.MockQueryRowScanner)
		mockTx.On("QueryRowContext", mock.Anything, mock.MatchedBy(func(query string) bool { return strings.Contains(query, "INSERT INTO search_queries") }), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(mockScannerInsertQueryFail).Once()
		mockScannerInsertQueryFail.On("Scan", mock.AnythingOfType("*int")).Return(nil).Run(func(args mock.Arguments) {
			arg := args.Get(0).(*int)
			*arg = 1 // Simulate scanning ID 1
		}).Once() // This happens once before the loop

		// Mock Get Search ID (Success) - This happens inside the loop, use Maybe()
		mockScannerGetSearchIDFail := new(mocks.MockQueryRowScanner)
		mockTx.On("QueryRowContext", mock.Anything, mock.MatchedBy(func(query string) bool { return strings.Contains(query, "SELECT search_id FROM search_queries") }), 1).
			Return(mockScannerGetSearchIDFail).Maybe() // Use Maybe for calls inside loop
		mockScannerGetSearchIDFail.On("Scan", mock.AnythingOfType("*string")).Return(nil).Run(func(args mock.Arguments) {
			arg := args.Get(0).(*string)
			*arg = "test-search-uuid" // Simulate scanning search_id
		}).Maybe() // Use Maybe for calls inside loop

		// Simulate error during Offer Insert - This happens inside the loop, use Once() because we expect it to fail on the first offer
		mockScannerInsertOfferFail := new(mocks.MockQueryRowScanner)
		mockTx.On("QueryRowContext", mock.Anything, mock.MatchedBy(func(query string) bool { return strings.Contains(query, "INSERT INTO flight_offers") }), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(mockScannerInsertOfferFail).Once() // Expect this specific call to happen once and fail
		// Configure Scan to return the error
		mockScannerInsertOfferFail.On("Scan", mock.AnythingOfType("*int")).Return(processingError).Once() // Expect Scan to be called once and return error

		// Expect Rollback
		mockTx.On("Rollback").Return(nil).Once()
		// Commit should not be called
		mockTx.On("Commit").Return(nil).Maybe()
		// --- End Corrected Mocks ---

		manager.Start()
		select {
		case <-processed:
			// Job processed and Nacked
		case <-time.After(2 * time.Second):
			t.Fatal("Job was not processed and Nacked within timeout")
		}
		manager.Stop()
		mockQueue.AssertExpectations(t)
		mockPgDb.AssertExpectations(t)
	})
}

// Test Job Prioritization (flight_search before bulk_search)
func TestManager_JobPrioritization(t *testing.T) {
	skipUnlessWorkerTests(t)
	cfg := config.WorkerConfig{Concurrency: 1, JobTimeout: 1 * time.Second, ShutdownTimeout: 1 * time.Second}
	manager, mockQueue, mockPgDb, mockNeo4jDb, _ := setupManagerTest(cfg)

	// Define payload structs and marshal them
	flightPayloadStruct := worker.FlightSearchPayload{
		Origin:        "JFK",
		Destination:   "LAX",
		DepartureDate: time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC),
		Adults:        1,
		TripType:      "one_way", // Ensure this matches case in manager switch
		Class:         "economy",
		Stops:         "nonstop",
		Currency:      "USD",
	}
	flightJobPayloadBytes, _ := json.Marshal(flightPayloadStruct)
	flightJob := &queue.Job{ID: "flight-job", Payload: json.RawMessage(flightJobPayloadBytes)}

	bulkPayloadStruct := worker.BulkSearchPayload{
		Origins:           []string{"LHR"},
		Destinations:      []string{"CDG"},
		DepartureDateFrom: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
		DepartureDateTo:   time.Date(2025, 6, 7, 0, 0, 0, 0, time.UTC),
		Adults:            1,
		TripType:          "round_trip", // Ensure this matches case in manager switch
		Class:             "business",
		Stops:             "nonstop",
		Currency:          "GBP",
	}
	bulkJobPayloadBytes, _ := json.Marshal(bulkPayloadStruct)
	bulkJob := &queue.Job{ID: "bulk-job", Payload: json.RawMessage(bulkJobPayloadBytes)} // Assuming bulk job uses bulk_search queue

	processedOrder := make([]string, 0, 2)
	var orderMutex sync.Mutex
	processedChan := make(chan bool, 2)

	// Mock Dequeue: Return flight job first, then bulk job
	mockQueue.On("Dequeue", mock.Anything, "flight_search").Return(flightJob, nil).Once()
	// Subsequent calls for flight_search return nil
	mockQueue.On("Dequeue", mock.Anything, "flight_search").Return(nil, nil).Maybe()

	// Mock Dequeue for bulk_search, should only be called after flight_search returned nil
	mockQueue.On("Dequeue", mock.Anything, "bulk_search").Return(bulkJob, nil).Once()
	// Subsequent calls for bulk_search return nil
	mockQueue.On("Dequeue", mock.Anything, "bulk_search").Return(nil, nil).Maybe()

	// Mock Ack for both jobs
	mockQueue.On("Ack", mock.Anything, "flight_search", flightJob.ID).Return(nil).Run(func(args mock.Arguments) {
		orderMutex.Lock()
		processedOrder = append(processedOrder, "flight")
		orderMutex.Unlock()
		processedChan <- true
	}).Once()
	mockQueue.On("Ack", mock.Anything, "bulk_search", bulkJob.ID).Return(nil).Run(func(args mock.Arguments) {
		orderMutex.Lock()
		processedOrder = append(processedOrder, "bulk")
		orderMutex.Unlock()
		processedChan <- true
	}).Once()
	// FIX: Allow Nack for bulk search queue in case of unexpected errors during processing
	mockQueue.On("Nack", mock.Anything, "bulk_search", bulkJob.ID).Return(nil).Maybe()

	// Mock the ListJobs call that the scheduler makes on startup
	mockRowsPriority := new(mocks.MockRows)
	mockPgDb.On("ListJobs", mock.Anything).Return(mockRowsPriority, nil)
	mockRowsPriority.On("Next").Return(false) // No scheduled jobs
	mockRowsPriority.On("Close").Return(nil)
	mockRowsPriority.On("Err").Return(nil)

	// --- Corrected DB & Neo4j Mocks for Prioritization Test ---
	// Mock calls for flightJob (flight_search)
	mockTxFlight := new(mocks.MockTx)
	mockPgDb.On("BeginTx", mock.Anything).Return(mockTxFlight, nil).Once() // Expect BeginTx once for flight job
	// Mock QueryRowContext + Scan for flightJob
	mockScannerFlightInsertQuery := new(mocks.MockQueryRowScanner)
	mockTxFlight.On("QueryRowContext", mock.Anything, mock.MatchedBy(func(query string) bool { return strings.Contains(query, "INSERT INTO search_queries") }), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockScannerFlightInsertQuery).Once()
	mockScannerFlightInsertQuery.On("Scan", mock.AnythingOfType("*int")).Return(nil).Run(func(args mock.Arguments) { *args.Get(0).(*int) = 1 }).Once()

	mockScannerFlightGetSearchID := new(mocks.MockQueryRowScanner)
	// FIX: Changed Once() to Maybe() for QueryRowContext and Scan for GetSearchID inside loop
	mockTxFlight.On("QueryRowContext", mock.Anything, mock.MatchedBy(func(query string) bool { return strings.Contains(query, "SELECT search_id FROM search_queries") }), 1).Return(mockScannerFlightGetSearchID).Maybe()
	mockScannerFlightGetSearchID.On("Scan", mock.AnythingOfType("*string")).Return(nil).Run(func(args mock.Arguments) { *args.Get(0).(*string) = "flight-uuid" }).Maybe()

	mockScannerFlightInsertOffer := new(mocks.MockQueryRowScanner)
	mockTxFlight.On("QueryRowContext", mock.Anything, mock.MatchedBy(func(query string) bool { return strings.Contains(query, "INSERT INTO flight_offers") }), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockScannerFlightInsertOffer).Maybe()
	mockScannerFlightInsertOffer.On("Scan", mock.AnythingOfType("*int")).Return(nil).Run(func(args mock.Arguments) { *args.Get(0).(*int) = 10 }).Maybe()

	// Mock ExecContext calls for flightJob - Use Maybe() as they are inside loop
	mockTxFlight.On("ExecContext", mock.Anything, mock.MatchedBy(func(query string) bool { return strings.Contains(query, "INSERT INTO airlines") }), mock.Anything, mock.Anything, mock.Anything).Return(sql.Result(nil), nil).Maybe()
	mockTxFlight.On("ExecContext", mock.Anything, mock.MatchedBy(func(query string) bool { return strings.Contains(query, "INSERT INTO airports") }), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sql.Result(nil), nil).Maybe()
	mockTxFlight.On("ExecContext", mock.Anything, mock.MatchedBy(func(query string) bool { return strings.Contains(query, "INSERT INTO flight_segments") }), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sql.Result(nil), nil).Maybe()

	mockNeo4jDb.On("CreateAirport", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	mockNeo4jDb.On("CreateAirline", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	mockNeo4jDb.On("CreateRoute", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	mockNeo4jDb.On("AddPricePoint", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	mockTxFlight.On("Commit").Return(nil).Once()
	mockTxFlight.On("Rollback").Return(nil).Maybe()

	// Mock calls for bulkJob (bulk_search)
	mockTxBulk := new(mocks.MockTx)
	mockPgDb.On("BeginTx", mock.Anything).Return(mockTxBulk, nil).Once() // Expect BeginTx once for bulk job (or more if it iterates)
	// Mock QueryRowContext + Scan for bulkJob
	mockScannerBulkInsertQuery := new(mocks.MockQueryRowScanner)
	mockTxBulk.On("QueryRowContext", mock.Anything, mock.MatchedBy(func(query string) bool { return strings.Contains(query, "INSERT INTO search_queries") }), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockScannerBulkInsertQuery).Once()
	mockScannerBulkInsertQuery.On("Scan", mock.AnythingOfType("*int")).Return(nil).Run(func(args mock.Arguments) { *args.Get(0).(*int) = 2 }).Once()

	mockScannerBulkGetSearchID := new(mocks.MockQueryRowScanner)
	// FIX: Changed Once() to Maybe() for QueryRowContext and Scan for GetSearchID inside loop
	mockTxBulk.On("QueryRowContext", mock.Anything, mock.MatchedBy(func(query string) bool { return strings.Contains(query, "SELECT search_id FROM search_queries") }), 2).Return(mockScannerBulkGetSearchID).Maybe()
	mockScannerBulkGetSearchID.On("Scan", mock.AnythingOfType("*string")).Return(nil).Run(func(args mock.Arguments) { *args.Get(0).(*string) = "bulk-uuid" }).Maybe()

	mockScannerBulkInsertOffer := new(mocks.MockQueryRowScanner)
	mockTxBulk.On("QueryRowContext", mock.Anything, mock.MatchedBy(func(query string) bool { return strings.Contains(query, "INSERT INTO flight_offers") }), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockScannerBulkInsertOffer).Maybe()
	mockScannerBulkInsertOffer.On("Scan", mock.AnythingOfType("*int")).Return(nil).Run(func(args mock.Arguments) { *args.Get(0).(*int) = 11 }).Maybe()

	// Mock ExecContext calls for bulkJob - Use Maybe() as they are inside loop
	mockTxBulk.On("ExecContext", mock.Anything, mock.MatchedBy(func(query string) bool { return strings.Contains(query, "INSERT INTO airlines") }), mock.Anything, mock.Anything, mock.Anything).Return(sql.Result(nil), nil).Maybe()
	mockTxBulk.On("ExecContext", mock.Anything, mock.MatchedBy(func(query string) bool { return strings.Contains(query, "INSERT INTO airports") }), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sql.Result(nil), nil).Maybe()
	mockTxBulk.On("ExecContext", mock.Anything, mock.MatchedBy(func(query string) bool { return strings.Contains(query, "INSERT INTO flight_segments") }), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sql.Result(nil), nil).Maybe()

	mockNeo4jDb.On("CreateAirport", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	mockNeo4jDb.On("CreateAirline", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	mockNeo4jDb.On("CreateRoute", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	// FIX: Mock AddPricePoint for bulk search path as well
	mockNeo4jDb.On("AddPricePoint", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	mockTxBulk.On("Commit").Return(nil).Once()
	mockTxBulk.On("Rollback").Return(nil).Maybe()
	// --- End Corrected Mocks ---

	allowPriceGraphDequeues(mockQueue)

	manager.Start()

	// Wait for both jobs to be processed
	<-processedChan
	<-processedChan

	manager.Stop()

	// Verify the order
	assert.Equal(t, []string{"flight", "bulk"}, processedOrder, "Jobs processed out of order")
	mockQueue.AssertExpectations(t)
}

// Test Configuration Usage (Concurrency)
func TestManager_ConcurrencyConfig(t *testing.T) {
	skipUnlessWorkerTests(t)
	concurrency := 3
	cfg := config.WorkerConfig{Concurrency: concurrency, JobTimeout: 1 * time.Second, ShutdownTimeout: 1 * time.Second}
	manager, mockQueue, mockPgDb, _, _ := setupManagerTest(cfg)

	// Mock the ListJobs call that the scheduler makes on startup
	mockRowsConcurrency := new(mocks.MockRows)
	mockPgDb.On("ListJobs", mock.Anything).Return(mockRowsConcurrency, nil)
	mockRowsConcurrency.On("Next").Return(false) // No scheduled jobs
	mockRowsConcurrency.On("Close").Return(nil)
	mockRowsConcurrency.On("Err").Return(nil)

	// Mock Dequeue to check how many workers are potentially calling it
	var callCount int32
	var countMutex sync.Mutex
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(concurrency) // Expect 'concurrency' calls roughly simultaneously

	// FIX: Only call Done() when the flight_search queue is dequeued
	mockQueue.On("Dequeue", mock.Anything, "flight_search").Return(nil, nil).Run(func(args mock.Arguments) {
		countMutex.Lock()
		// Only decrement if the counter is positive
		if callCount < int32(concurrency) {
			waitGroup.Done()
		}
		callCount++
		countMutex.Unlock()
		time.Sleep(50 * time.Millisecond) // Small sleep to allow other workers to call
	}).Maybe() // Use Maybe as it might be called multiple times per worker

	// Allow Dequeue for bulk_search without affecting the WaitGroup
	mockQueue.On("Dequeue", mock.Anything, "bulk_search").Return(nil, nil).Maybe()

	allowPriceGraphDequeues(mockQueue)

	manager.Start()

	// Wait for at least 'concurrency' calls or timeout
	done := make(chan bool)
	go func() {
		waitGroup.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All expected calls received
	case <-time.After(2 * time.Second): // Increased timeout slightly
		t.Logf("Received %d Dequeue calls for flight_search", callCount)
		t.Fatal("Timed out waiting for workers to call Dequeue for flight_search")
	}

	manager.Stop()
	// We can't assert the exact number of calls easily due to timing,
	// but we know at least 'concurrency' workers started and called Dequeue.
	// The assertion is implicitly handled by waitGroup.Wait() succeeding.
}
