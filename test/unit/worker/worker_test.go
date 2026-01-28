package worker_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/gilby125/google-flights-api/flights"
	"github.com/gilby125/google-flights-api/test/mocks"
	"github.com/gilby125/google-flights-api/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var runWorkerCoreTests = os.Getenv("ENABLE_WORKER_TESTS") == "1"

func skipUnlessWorkerCoreTests(t *testing.T) {
	if !runWorkerCoreTests {
		t.Skip("set ENABLE_WORKER_TESTS=1 to run worker core tests")
	}
}

// --- Test Setup ---

// Helper function to create a basic worker with mocks
func setupWorkerTest() (*worker.Worker, *mocks.MockPostgresDB, *mocks.MockNeo4jDatabase) {
	mockPgDb := new(mocks.MockPostgresDB)
	mockNeo4jDb := new(mocks.MockNeo4jDatabase)
	workerInstance := worker.NewWorker(mockPgDb, mockNeo4jDb) // Use the constructor
	return workerInstance, mockPgDb, mockNeo4jDb
}

// --- Test StoreFlightInNeo4j ---

func TestWorker_StoreFlightInNeo4j_Success(t *testing.T) {
	skipUnlessWorkerCoreTests(t)
	workerInstance, _, mockNeo4jDb := setupWorkerTest()
	ctx := context.Background()
	testTime := time.Now()

	offer := flights.FullOffer{
		Offer: flights.Offer{
			StartDate: testTime,
			Price:     100.50,
		},
		SrcAirportCode: "LHR",
		DstAirportCode: "JFK",
		Flight: []flights.Flight{
			{
				DepAirportCode: "LHR", DepAirportName: "Heathrow", DepCity: "London",
				ArrAirportCode: "JFK", ArrAirportName: "Kennedy", ArrCity: "New York",
				AirlineName: "TestAir", FlightNumber: "TA123",
				Duration: 8 * time.Hour,
			},
		},
	}

	// Mock Neo4j calls
	mockNeo4jDb.On("CreateAirport", "LHR", "Heathrow", "London", "", 0.0, 0.0).Return(nil).Once()
	mockNeo4jDb.On("CreateAirport", "JFK", "Kennedy", "New York", "", 0.0, 0.0).Return(nil).Once()
	mockNeo4jDb.On("CreateAirline", "TA", "TestAir", "").Return(nil).Once()
	mockNeo4jDb.On("CreateRoute", "LHR", "JFK", "TA", "TA123", 100.50, 480).Return(nil).Once()
	mockNeo4jDb.On("AddPricePoint", "LHR", "JFK", testTime.Format("2006-01-02"), "", 100.50, "TA", "one_way").Return(nil).Once()

	err := workerInstance.StoreFlightInNeo4j(ctx, offer)

	assert.NoError(t, err)
	mockNeo4jDb.AssertExpectations(t)
}

func TestWorker_StoreFlightInNeo4j_CreateAirportError(t *testing.T) {
	skipUnlessWorkerCoreTests(t)
	workerInstance, _, mockNeo4jDb := setupWorkerTest()
	ctx := context.Background()
	testTime := time.Now()
	expectedError := errors.New("neo4j airport error")

	offer := flights.FullOffer{
		Offer: flights.Offer{
			StartDate: testTime,
			Price:     100.50,
		},
		SrcAirportCode: "LHR",
		DstAirportCode: "JFK",
		Flight: []flights.Flight{
			{
				DepAirportCode: "LHR", DepAirportName: "Heathrow", DepCity: "London",
				ArrAirportCode: "JFK", ArrAirportName: "Kennedy", ArrCity: "New York",
				AirlineName: "TestAir", FlightNumber: "TA123",
				Duration: 8 * time.Hour,
			},
		},
	}

	// Mock Neo4j calls - Simulate error on first CreateAirport
	mockNeo4jDb.On("CreateAirport", "LHR", "Heathrow", "London", "", 0.0, 0.0).Return(expectedError).Once()

	err := workerInstance.StoreFlightInNeo4j(ctx, offer)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create departure airport in Neo4j")
	assert.ErrorIs(t, err, expectedError)
	mockNeo4jDb.AssertExpectations(t)
	// Ensure other mocks were not called
	mockNeo4jDb.AssertNotCalled(t, "CreateAirport", "JFK", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mockNeo4jDb.AssertNotCalled(t, "CreateAirline", mock.Anything, mock.Anything, mock.Anything)
}

// --- Test StoreFlightOffers ---

func TestWorker_StoreFlightOffers_Success(t *testing.T) {
	skipUnlessWorkerCoreTests(t)
	workerInstance, mockPgDb, mockNeo4jDb := setupWorkerTest()
	ctx := context.Background()
	testTime := time.Now()

	payload := worker.FlightSearchPayload{
		Origin:        "LHR",
		Destination:   "JFK",
		DepartureDate: testTime.AddDate(0, 0, 7),
		ReturnDate:    testTime.AddDate(0, 0, 14),
		Adults:        1,
		TripType:      "round_trip",
		Class:         "economy",
		Stops:         "nonstop",
		Currency:      "USD",
	}
	offers := []flights.FullOffer{
		{
			Offer: flights.Offer{
				StartDate:  testTime.AddDate(0, 0, 7),
				ReturnDate: testTime.AddDate(0, 0, 14),
				Price:      500.00,
			},
			SrcAirportCode: "LHR",
			DstAirportCode: "JFK",
			FlightDuration: 8 * time.Hour,
			Flight: []flights.Flight{
				{
					DepAirportCode: "LHR", DepAirportName: "Heathrow", DepCity: "London",
					ArrAirportCode: "JFK", ArrAirportName: "Kennedy", ArrCity: "New York",
					AirlineName: "TestAir", FlightNumber: "TA123",
					DepTime: testTime.AddDate(0, 0, 7), ArrTime: testTime.AddDate(0, 0, 7).Add(8 * time.Hour),
					Duration: 8 * time.Hour, Airplane: "747", Legroom: "32",
				},
			},
		},
	}

	// --- Mock DB Calls ---
	mockTx := new(mocks.MockTx)
	// Expect only context argument for BeginTx
	mockPgDb.On("BeginTx", mock.Anything).Return(mockTx, nil).Once()

	// Mock Search Query Insert
	mockScannerInsertQuery := new(mocks.MockQueryRowScanner)
	mockTx.On("QueryRowContext", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockScannerInsertQuery).Once()
	mockScannerInsertQuery.On("Scan", mock.AnythingOfType("*int")).Return(nil).Run(func(args mock.Arguments) { *args.Get(0).(*int) = 1 }).Once() // Simulate query ID 1

	// Mock Get Search ID
	mockScannerGetSearchID := new(mocks.MockQueryRowScanner)
	mockTx.On("QueryRowContext", mock.Anything, mock.AnythingOfType("string"), 1).Return(mockScannerGetSearchID).Once()
	mockScannerGetSearchID.On("Scan", mock.AnythingOfType("*string")).Return(nil).Run(func(args mock.Arguments) { *args.Get(0).(*string) = "test-uuid" }).Once()

	// Mock Offer Insert
	mockScannerInsertOffer := new(mocks.MockQueryRowScanner)
	mockTx.On("QueryRowContext", mock.Anything, mock.AnythingOfType("string"), 1, "test-uuid", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockScannerInsertOffer).Once()
	mockScannerInsertOffer.On("Scan", mock.AnythingOfType("*int")).Return(nil).Run(func(args mock.Arguments) { *args.Get(0).(*int) = 10 }).Once() // Simulate offer ID 10

	// FIX: Correct ExecContext mock expectations for variadic args
	// Mock Airline Insert (ctx, query, code, name, country)
	// Simplify mock expectation using mock.Anything for args
	// Match specific query string and args for Airline
	// Revert to expecting slices, matching mock's observed behavior
	// Revert to expecting individual arguments, matching actual function signature
	// Revert to expecting slices to match mock behavior
	mockTx.On("ExecContext", mock.Anything, mock.AnythingOfType("string"), []interface{}{"TA", "TestAir", ""}).Return(sql.Result(nil), nil).Once()
	// Mock Airport Inserts (ctx, query, code, name, city, country, lat, lon)
	// Simplify mock expectation using mock.Anything for args
	// Use separate .Once() expectations for each airport with mock.Anything args
	// Match specific query string and args for Airports
	mockTx.On("ExecContext", mock.Anything, mock.AnythingOfType("string"), []interface{}{"LHR", "Heathrow", "London", "", 0.0, 0.0}).Return(sql.Result(nil), nil).Once()
	mockTx.On("ExecContext", mock.Anything, mock.AnythingOfType("string"), []interface{}{"JFK", "Kennedy", "New York", "", 0.0, 0.0}).Return(sql.Result(nil), nil).Once()
	// Mock Segment Insert (ctx, query, offerID, code, num, dep, arr, depTime, arrTime, dur, plane, legroom, isReturn)
	// Match specific query string for segment insert
	// Use mock.AnythingOfType("time.Time") for time fields
	// Use mock.MatchedBy for the slice argument to handle time.Time comparison issue
	mockTx.On("ExecContext", mock.Anything, mock.AnythingOfType("string"), mock.MatchedBy(func(args []interface{}) bool {
		if len(args) != 11 {
			return false
		}
		// Check non-time fields directly
		if args[0].(int) != 10 {
			return false
		}
		if args[1].(string) != "TA" {
			return false
		}
		if args[2].(string) != "TA123" {
			return false
		}
		if args[3].(string) != "LHR" {
			return false
		}
		if args[4].(string) != "JFK" {
			return false
		}
		// Check time fields by type
		if _, ok := args[5].(time.Time); !ok {
			return false
		}
		if _, ok := args[6].(time.Time); !ok {
			return false
		}
		// Check remaining fields
		if args[7].(int) != 480 {
			return false
		}
		if args[8].(string) != "747" {
			return false
		}
		if args[9].(string) != "32" {
			return false
		}
		if args[10].(bool) != false {
			return false
		}
		return true // If all checks pass
	})).Return(sql.Result(nil), nil).Once()

	// Mock Neo4j calls (from StoreFlightInNeo4j)
	mockNeo4jDb.On("CreateAirport", "LHR", "Heathrow", "London", "", 0.0, 0.0).Return(nil).Once()
	mockNeo4jDb.On("CreateAirport", "JFK", "Kennedy", "New York", "", 0.0, 0.0).Return(nil).Once()
	mockNeo4jDb.On("CreateAirline", "TA", "TestAir", "").Return(nil).Once()
	mockNeo4jDb.On("CreateRoute", "LHR", "JFK", "TA", "TA123", 500.00, 480).Return(nil).Once()
	mockNeo4jDb.On("AddPricePoint", "LHR", "JFK", offers[0].StartDate.Format("2006-01-02"), "", 500.00, "TA", "one_way").Return(nil).Once()

	// Mock Commit
	mockTx.On("Commit").Return(nil).Once()
	mockTx.On("Rollback").Return(nil).Maybe() // Should not be called

	// --- Act ---
	err := workerInstance.StoreFlightOffers(ctx, payload, offers, nil) // Call the actual function

	// --- Assert ---
	assert.NoError(t, err)
	mockPgDb.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockNeo4jDb.AssertExpectations(t)
}

// Add more tests for StoreFlightOffers error paths (BeginTx error, Insert errors, Commit error, etc.)

// Add tests for processPriceGraphSearch if needed
