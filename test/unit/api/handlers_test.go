package api_test

import (
	"bytes"
	// "context" // Removed unused import
	"database/sql" // Import sql package for sql.Result
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"

	// "strconv" // Removed unused import
	"testing"
	"time"

	"github.com/gilby125/google-flights-api/api"
	"github.com/gilby125/google-flights-api/config"
	"github.com/gilby125/google-flights-api/db" // Import db package
	"github.com/gilby125/google-flights-api/flights"
	"github.com/gilby125/google-flights-api/test/mocks" // Assuming mocks are here
	"github.com/gilby125/google-flights-api/worker"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func dateOnly(t time.Time) api.DateOnly {
	return api.DateOnly{Time: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)}
}

// --- Test Setup ---

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	return router
}

type workerStatusProviderMock struct {
	statuses []worker.WorkerStatus
}

func (m *workerStatusProviderMock) WorkerStatuses() []worker.WorkerStatus {
	return m.statuses
}

func matchFlightSearchPayload(expected worker.FlightSearchPayload) func(worker.FlightSearchPayload) bool {
	return func(actual worker.FlightSearchPayload) bool {
		if actual.Origin != expected.Origin ||
			actual.Destination != expected.Destination ||
			actual.Adults != expected.Adults ||
			actual.Children != expected.Children ||
			actual.InfantsLap != expected.InfantsLap ||
			actual.InfantsSeat != expected.InfantsSeat ||
			actual.TripType != expected.TripType ||
			actual.Class != expected.Class ||
			actual.Stops != expected.Stops ||
			actual.Currency != expected.Currency {
			return false
		}

		if !actual.DepartureDate.Equal(expected.DepartureDate) {
			return false
		}

		switch {
		case expected.ReturnDate.IsZero() && actual.ReturnDate.IsZero():
			return true
		case expected.ReturnDate.IsZero() != actual.ReturnDate.IsZero():
			return false
		default:
			return actual.ReturnDate.Equal(expected.ReturnDate)
		}
	}
}

func matchBulkSearchPayload(expected worker.BulkSearchPayload) func(worker.BulkSearchPayload) bool {
	return func(actual worker.BulkSearchPayload) bool {
		if !reflect.DeepEqual(actual.Origins, expected.Origins) ||
			!reflect.DeepEqual(actual.Destinations, expected.Destinations) ||
			actual.TripLength != expected.TripLength ||
			actual.Adults != expected.Adults ||
			actual.Children != expected.Children ||
			actual.InfantsLap != expected.InfantsLap ||
			actual.InfantsSeat != expected.InfantsSeat ||
			actual.TripType != expected.TripType ||
			actual.Class != expected.Class ||
			actual.Stops != expected.Stops ||
			actual.Currency != expected.Currency {
			return false
		}

		if expected.BulkSearchID > 0 {
			if actual.BulkSearchID != expected.BulkSearchID {
				return false
			}
		} else if actual.BulkSearchID == 0 {
			// Ensure we always set a bulk search record
			return false
		}

		if actual.JobID != expected.JobID {
			return false
		}

		if !actual.DepartureDateFrom.Equal(expected.DepartureDateFrom) ||
			!actual.DepartureDateTo.Equal(expected.DepartureDateTo) {
			return false
		}

		switch {
		case expected.ReturnDateFrom.IsZero() && !actual.ReturnDateFrom.IsZero():
			return false
		case !expected.ReturnDateFrom.IsZero() && !actual.ReturnDateFrom.Equal(expected.ReturnDateFrom):
			return false
		}

		switch {
		case expected.ReturnDateTo.IsZero() && !actual.ReturnDateTo.IsZero():
			return false
		case !expected.ReturnDateTo.IsZero() && !actual.ReturnDateTo.Equal(expected.ReturnDateTo):
			return false
		}

		return true
	}
}

// --- Helper Function Tests ---

func TestParseClass(t *testing.T) {
	assert.Equal(t, flights.Economy, api.ParseClass("economy"))
	assert.Equal(t, flights.PremiumEconomy, api.ParseClass("premium_economy"))
	assert.Equal(t, flights.Business, api.ParseClass("business"))
	assert.Equal(t, flights.First, api.ParseClass("first"))
	assert.Equal(t, flights.Economy, api.ParseClass("unknown")) // Default case
	assert.Equal(t, flights.Economy, api.ParseClass(""))        // Default case
}

func TestParseStops(t *testing.T) {
	assert.Equal(t, flights.Nonstop, api.ParseStops("nonstop"))
	assert.Equal(t, flights.Stop1, api.ParseStops("one_stop"))
	assert.Equal(t, flights.Stop2, api.ParseStops("two_stops"))
	assert.Equal(t, flights.AnyStops, api.ParseStops("any"))
	assert.Equal(t, flights.AnyStops, api.ParseStops("unknown")) // Default case
	assert.Equal(t, flights.AnyStops, api.ParseStops(""))        // Default case
}

// --- Handler Tests ---

func TestCreateSearch_Success(t *testing.T) {
	// Arrange
	mockQueue := new(mocks.Queue)
	router := setupRouter()
	router.POST("/search", api.CreateSearch(mockQueue)) // Use the actual handler constructor

	searchReq := api.SearchRequest{
		Origin:        "LHR",
		Destination:   "JFK",
		DepartureDate: dateOnly(time.Now().AddDate(0, 1, 0)),
		Adults:        1,
		TripType:      "one_way",
		Class:         "economy",
		Stops:         "any",
		Currency:      "USD",
	}
	expectedJobID := "job-123"
	expectedPayload := worker.FlightSearchPayload{
		Origin:        searchReq.Origin,
		Destination:   searchReq.Destination,
		DepartureDate: searchReq.DepartureDate.Time,
		ReturnDate:    searchReq.ReturnDate.Time, // Should be zero time.Time
		Adults:        searchReq.Adults,
		Children:      searchReq.Children,
		InfantsLap:    searchReq.InfantsLap,
		InfantsSeat:   searchReq.InfantsSeat,
		TripType:      searchReq.TripType,
		Class:         searchReq.Class, // Use original string as per handler logic
		Stops:         searchReq.Stops, // Use original string as per handler logic
		Currency:      searchReq.Currency,
	}

	// Configure mock
	mockQueue.On("Enqueue", mock.Anything, "flight_search",
		mock.MatchedBy(matchFlightSearchPayload(expectedPayload)),
	).Return(expectedJobID, nil) // Corrected expected job type

	// Act
	body, _ := json.Marshal(searchReq)
	req, _ := http.NewRequest(http.MethodPost, "/search", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusAccepted, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedJobID, response["job_id"])
	assert.Equal(t, "Flight search job created successfully", response["message"])
	mockQueue.AssertExpectations(t)
}

func TestCreateSearch_BindError(t *testing.T) {
	// Arrange
	mockQueue := new(mocks.Queue) // Not used, but handler needs it
	router := setupRouter()
	router.POST("/search", api.CreateSearch(mockQueue))

	// Act: Send invalid JSON
	req, _ := http.NewRequest(http.MethodPost, "/search", bytes.NewBufferString(`{"origin": "LHR"`)) // Malformed JSON
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	// Optionally check error message structure
}

func TestCreateSearch_ValidationError(t *testing.T) {
	// Arrange
	mockQueue := new(mocks.Queue)
	router := setupRouter()
	router.POST("/search", api.CreateSearch(mockQueue))

	searchReq := api.SearchRequest{
		// Missing required fields like Origin, Destination, etc.
		Adults: 0, // Invalid value
	}

	// Act
	body, _ := json.Marshal(searchReq)
	req, _ := http.NewRequest(http.MethodPost, "/search", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	// Check for specific validation error messages if needed
}

func TestCreateSearch_EnqueueError(t *testing.T) {
	// Arrange
	mockQueue := new(mocks.Queue)
	router := setupRouter()
	router.POST("/search", api.CreateSearch(mockQueue))

	searchReq := api.SearchRequest{
		Origin:        "LHR",
		Destination:   "JFK",
		DepartureDate: dateOnly(time.Now().AddDate(0, 1, 0)),
		Adults:        1,
		TripType:      "one_way",
		Class:         "economy",
		Stops:         "any",
		Currency:      "USD",
	}
	expectedPayload := worker.FlightSearchPayload{
		Origin:        searchReq.Origin,
		Destination:   searchReq.Destination,
		DepartureDate: searchReq.DepartureDate.Time,
		ReturnDate:    searchReq.ReturnDate.Time,
		Adults:        searchReq.Adults,
		Children:      searchReq.Children,
		InfantsLap:    searchReq.InfantsLap,
		InfantsSeat:   searchReq.InfantsSeat,
		TripType:      searchReq.TripType,
		Class:         searchReq.Class, // Use original string as per handler logic
		Stops:         searchReq.Stops, // Use original string as per handler logic
		Currency:      searchReq.Currency,
	}

	// Configure mock to return an error
	mockQueue.On("Enqueue", mock.Anything, "flight_search",
		mock.MatchedBy(matchFlightSearchPayload(expectedPayload)),
	).Return("", assert.AnError) // Corrected expected job type

	// Act
	body, _ := json.Marshal(searchReq)
	req, _ := http.NewRequest(http.MethodPost, "/search", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockQueue.AssertExpectations(t)
}

func TestCreateBulkSearch_Success(t *testing.T) {
	// Arrange
	mockQueue := new(mocks.Queue)
	mockPostgres := new(mocks.MockPostgresDB)
	router := setupRouter()
	router.POST("/bulk-search", api.CreateBulkSearch(mockQueue, mockPostgres))

	bulkReq := api.BulkSearchRequest{
		Origins:           []string{"LHR", "LGW"},
		Destinations:      []string{"JFK", "EWR"},
		DepartureDateFrom: dateOnly(time.Now().AddDate(0, 1, 0)),
		DepartureDateTo:   dateOnly(time.Now().AddDate(0, 1, 7)),
		Adults:            2,
		TripType:          "round_trip",
		Class:             "business",
		Stops:             "nonstop",
		Currency:          "GBP",
	}
	expectedJobID := "bulk-job-456"
	createdBulkSearchID := 123
	totalRoutes := len(bulkReq.Origins) * len(bulkReq.Destinations)
	expectedPayload := worker.BulkSearchPayload{
		Origins:           bulkReq.Origins,
		Destinations:      bulkReq.Destinations,
		DepartureDateFrom: bulkReq.DepartureDateFrom.Time,
		DepartureDateTo:   bulkReq.DepartureDateTo.Time,
		ReturnDateFrom:    bulkReq.ReturnDateFrom.Time, // zero
		ReturnDateTo:      bulkReq.ReturnDateTo.Time,   // zero
		TripLength:        bulkReq.TripLength,          // zero
		Adults:            bulkReq.Adults,
		Children:          bulkReq.Children,
		InfantsLap:        bulkReq.InfantsLap,
		InfantsSeat:       bulkReq.InfantsSeat,
		TripType:          bulkReq.TripType,
		Class:             bulkReq.Class, // Note: Bulk search payload uses string class/stops directly
		Stops:             bulkReq.Stops,
		Currency:          strings.ToUpper(bulkReq.Currency),
		BulkSearchID:      createdBulkSearchID,
	}

	mockPostgres.On("CreateBulkSearchRecord", mock.Anything, mock.Anything, totalRoutes, strings.ToUpper(bulkReq.Currency), "queued").Return(createdBulkSearchID, nil)
	// Configure mock
	mockQueue.On("Enqueue", mock.Anything, "bulk_search",
		mock.MatchedBy(matchBulkSearchPayload(expectedPayload)),
	).Return(expectedJobID, nil)

	// Act
	body, _ := json.Marshal(bulkReq)
	req, _ := http.NewRequest(http.MethodPost, "/bulk-search", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusAccepted, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedJobID, response["job_id"])
	assert.Equal(t, float64(createdBulkSearchID), response["bulk_search_id"])
	assert.Equal(t, "Bulk flight search job created successfully", response["message"])
	mockQueue.AssertExpectations(t)
	mockPostgres.AssertExpectations(t)
}

func TestCreateBulkSearch_BindError(t *testing.T) {
	// Arrange
	mockQueue := new(mocks.Queue)
	mockPostgres := new(mocks.MockPostgresDB)
	router := setupRouter()
	router.POST("/bulk-search", api.CreateBulkSearch(mockQueue, mockPostgres))

	// Act: Send invalid JSON
	req, _ := http.NewRequest(http.MethodPost, "/bulk-search", bytes.NewBufferString(`{"origins": ["LHR"]`)) // Malformed
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateBulkSearch_ValidationError(t *testing.T) {
	// Arrange
	mockQueue := new(mocks.Queue)
	mockPostgres := new(mocks.MockPostgresDB)
	router := setupRouter()
	router.POST("/bulk-search", api.CreateBulkSearch(mockQueue, mockPostgres))

	bulkReq := api.BulkSearchRequest{
		// Missing required fields
	}

	// Act
	body, _ := json.Marshal(bulkReq)
	req, _ := http.NewRequest(http.MethodPost, "/bulk-search", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateBulkSearch_EnqueueError(t *testing.T) {
	// Arrange
	mockQueue := new(mocks.Queue)
	mockPostgres := new(mocks.MockPostgresDB)
	router := setupRouter()
	router.POST("/bulk-search", api.CreateBulkSearch(mockQueue, mockPostgres))

	bulkReq := api.BulkSearchRequest{
		Origins:           []string{"LHR"},
		Destinations:      []string{"JFK"},
		DepartureDateFrom: dateOnly(time.Now().AddDate(0, 1, 0)),
		DepartureDateTo:   dateOnly(time.Now().AddDate(0, 1, 7)),
		Adults:            1,
		TripType:          "one_way",
		Class:             "economy",
		Stops:             "any",
		Currency:          "USD",
	}
	createdBulkSearchID := 789
	totalRoutes := len(bulkReq.Origins) * len(bulkReq.Destinations)
	expectedPayload := worker.BulkSearchPayload{
		Origins:           bulkReq.Origins,
		Destinations:      bulkReq.Destinations,
		DepartureDateFrom: bulkReq.DepartureDateFrom.Time,
		DepartureDateTo:   bulkReq.DepartureDateTo.Time,
		TripLength:        bulkReq.TripLength,
		Adults:            bulkReq.Adults,
		Children:          bulkReq.Children,
		InfantsLap:        bulkReq.InfantsLap,
		InfantsSeat:       bulkReq.InfantsSeat,
		TripType:          bulkReq.TripType,
		Class:             bulkReq.Class,
		Stops:             bulkReq.Stops,
		Currency:          strings.ToUpper(bulkReq.Currency),
		BulkSearchID:      createdBulkSearchID,
	}

	mockPostgres.On("CreateBulkSearchRecord", mock.Anything, mock.Anything, totalRoutes, strings.ToUpper(bulkReq.Currency), "queued").Return(createdBulkSearchID, nil)
	mockQueue.On("Enqueue", mock.Anything, "bulk_search",
		mock.MatchedBy(matchBulkSearchPayload(expectedPayload)),
	).Return("", assert.AnError)
	mockPostgres.On("UpdateBulkSearchStatus", mock.Anything, createdBulkSearchID, "failed").Return(nil)

	// Act
	body, _ := json.Marshal(bulkReq)
	req, _ := http.NewRequest(http.MethodPost, "/bulk-search", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockQueue.AssertExpectations(t)
	mockPostgres.AssertExpectations(t)
}

func TestGetQueueStatus_Success(t *testing.T) {
	// Arrange
	mockQueue := new(mocks.Queue)
	router := setupRouter()
	router.GET("/queue/status", api.GetQueueStatus(mockQueue)) // Use the actual handler constructor

	expectedStatsFS := map[string]int64{"pending": 10, "active": 2}
	expectedStatsBS := map[string]int64{"pending": 5, "active": 1}
	expectedStatsPGS := map[string]int64{"pending": 0, "active": 0}
	expectedStatsCPG := map[string]int64{"pending": 0, "active": 0}

	// Configure mock
	mockQueue.On("GetQueueStats", mock.Anything, "flight_search").Return(expectedStatsFS, nil)
	mockQueue.On("GetQueueStats", mock.Anything, "bulk_search").Return(expectedStatsBS, nil)
	mockQueue.On("GetQueueStats", mock.Anything, "price_graph_sweep").Return(expectedStatsPGS, nil)
	mockQueue.On("GetQueueStats", mock.Anything, "continuous_price_graph").Return(expectedStatsCPG, nil)

	// Act
	req, _ := http.NewRequest(http.MethodGet, "/queue/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]map[string]int64
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedStatsFS, response["flight_search"])
	assert.Equal(t, expectedStatsBS, response["bulk_search"])
	assert.Equal(t, expectedStatsPGS, response["price_graph_sweep"])
	assert.Equal(t, expectedStatsCPG, response["continuous_price_graph"])
	mockQueue.AssertExpectations(t)
}

func TestGetQueueStatus_Error(t *testing.T) {
	// Arrange
	mockQueue := new(mocks.Queue)
	router := setupRouter()
	router.GET("/queue/status", api.GetQueueStatus(mockQueue))

	// Configure mock to return an error for one of the calls
	mockQueue.On("GetQueueStats", mock.Anything, "flight_search").Return(map[string]int64{}, assert.AnError)
	// No need to mock the second call if the first fails

	// Act
	req, _ := http.NewRequest(http.MethodGet, "/queue/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockQueue.AssertExpectations(t) // Ensure the failing call was made
}

func TestGetWorkerStatus(t *testing.T) {
	// Arrange
	router := setupRouter()
	mockProvider := &workerStatusProviderMock{
		statuses: []worker.WorkerStatus{
			{
				ID:            1,
				Status:        "active",
				CurrentJob:    "",
				ProcessedJobs: 3,
				Uptime:        42,
			},
		},
	}
	router.GET("/worker/status", api.GetWorkerStatus(mockProvider, nil, config.WorkerConfig{}))

	// Act
	req, _ := http.NewRequest(http.MethodGet, "/worker/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var response []map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, float64(1), response[0]["id"])
	assert.Equal(t, "active", response[0]["status"])
	assert.Equal(t, float64(3), response[0]["processed_jobs"])
	assert.Equal(t, float64(42), response[0]["uptime"])
	assert.Equal(t, "local", response[0]["source"])
}

func TestGetWorkerStatus_NilManager(t *testing.T) {
	// Arrange
	router := setupRouter()
	router.GET("/worker/status", api.GetWorkerStatus(nil, nil, config.WorkerConfig{}))

	// Act
	req, _ := http.NewRequest(http.MethodGet, "/worker/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var response []map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Empty(t, response)
}

// --- DB Dependent Handler Tests ---

func TestGetAirports_Success(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockPostgresDB)
	mockRows := new(mocks.MockRows)
	router := setupRouter()
	// Register the handler with the router
	router.GET("/airports", api.GetAirports(mockDB))

	expectedAirports := []db.Airport{
		{Code: "LHR", Name: "Heathrow", City: "London", Country: "UK", Latitude: sql.NullFloat64{Float64: 51.47, Valid: true}, Longitude: sql.NullFloat64{Float64: -0.45, Valid: true}},
		{Code: "JFK", Name: "John F Kennedy", City: "New York", Country: "USA", Latitude: sql.NullFloat64{Float64: 40.64, Valid: true}, Longitude: sql.NullFloat64{Float64: -73.77, Valid: true}},
		{Code: "NUL", Name: "Null Island", City: "Nowhere", Country: "NA", Latitude: sql.NullFloat64{Valid: false}, Longitude: sql.NullFloat64{Valid: false}},
	}

	// Configure mock rows
	mockRows.On("Next").Return(true).Times(len(expectedAirports))
	mockRows.On("Next").Return(false) // End of rows

	// Mock Scan calls sequentially
	scanCallCount := 0
	mockRows.On("Scan", mock.AnythingOfType("*string"), mock.AnythingOfType("*string"), mock.AnythingOfType("*string"), mock.AnythingOfType("*string"), mock.AnythingOfType("*sql.NullFloat64"), mock.AnythingOfType("*sql.NullFloat64")).Return(nil).Run(func(args mock.Arguments) {
		if scanCallCount < len(expectedAirports) {
			airport := expectedAirports[scanCallCount]
			*(args.Get(0).(*string)) = airport.Code
			*(args.Get(1).(*string)) = airport.Name
			*(args.Get(2).(*string)) = airport.City
			*(args.Get(3).(*string)) = airport.Country
			*(args.Get(4).(*sql.NullFloat64)) = airport.Latitude
			*(args.Get(5).(*sql.NullFloat64)) = airport.Longitude
			scanCallCount++
		}
	}).Times(len(expectedAirports))

	mockRows.On("Close").Return(nil)
	mockRows.On("Err").Return(nil)

	// Configure mock DB
	mockDB.On("QueryAirports", mock.Anything).Return(mockRows, nil)

	// Act
	req, _ := http.NewRequest(http.MethodGet, "/airports", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req) // Use router to serve the request

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var response []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 3)
	// Check specific fields, handling potential nulls
	assert.Equal(t, "LHR", response[0]["code"])
	assert.Equal(t, 51.47, response[0]["latitude"])
	assert.Equal(t, "NUL", response[2]["code"])
	_, latExists := response[2]["latitude"]
	assert.False(t, latExists) // Latitude should not exist for NUL

	mockDB.AssertExpectations(t)
	mockRows.AssertExpectations(t)
}

func TestGetAirports_DBError(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockPostgresDB)
	router := setupRouter()
	// Register the handler with the router
	router.GET("/airports", api.GetAirports(mockDB))

	// Configure mock DB to return an error
	mockDB.On("QueryAirports", mock.Anything).Return(nil, assert.AnError)

	// Act
	req, _ := http.NewRequest(http.MethodGet, "/airports", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req) // Use router to serve the request

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}

func TestGetAirlines_Success(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockPostgresDB)
	mockRows := new(mocks.MockRows)
	router := setupRouter()
	// Register the handler with the router
	router.GET("/airlines", api.GetAirlines(mockDB))

	expectedAirlines := []db.Airline{
		{Code: "BA", Name: "British Airways", Country: "UK"},
		{Code: "AA", Name: "American Airlines", Country: "USA"},
	}

	// Configure mock rows
	mockRows.On("Next").Return(true).Times(len(expectedAirlines))
	mockRows.On("Next").Return(false)
	mockRows.On("Scan", mock.AnythingOfType("*string"), mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).Return(nil).Run(func(args mock.Arguments) {
		*(args.Get(0).(*string)) = expectedAirlines[0].Code
		*(args.Get(1).(*string)) = expectedAirlines[0].Name
		*(args.Get(2).(*string)) = expectedAirlines[0].Country
	}).Once()
	mockRows.On("Scan", mock.AnythingOfType("*string"), mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).Return(nil).Run(func(args mock.Arguments) {
		*(args.Get(0).(*string)) = expectedAirlines[1].Code
		*(args.Get(1).(*string)) = expectedAirlines[1].Name
		*(args.Get(2).(*string)) = expectedAirlines[1].Country
	}).Once()
	mockRows.On("Close").Return(nil)
	mockRows.On("Err").Return(nil)

	// Configure mock DB
	mockDB.On("QueryAirlines", mock.Anything).Return(mockRows, nil)

	// Act
	req, _ := http.NewRequest(http.MethodGet, "/airlines", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req) // Use router to serve the request

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var response []db.Airline
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedAirlines, response)
	mockDB.AssertExpectations(t)
	mockRows.AssertExpectations(t)
}

func TestGetAirlines_DBError(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockPostgresDB)
	router := setupRouter()
	// Register the handler with the router
	router.GET("/airlines", api.GetAirlines(mockDB))

	mockDB.On("QueryAirlines", mock.Anything).Return(nil, assert.AnError)

	// Act
	req, _ := http.NewRequest(http.MethodGet, "/airlines", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req) // Use router to serve the request

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}

func TestGetSearchByID_Success(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockPostgresDB)
	router := setupRouter()
	// Register the handler with the router
	router.GET("/search/:id", api.GetSearchByID(mockDB))

	searchID := 123
	testTime := time.Now()
	offerID := 456

	mockQuery := &db.SearchQuery{
		ID:            searchID,
		Origin:        "LHR",
		Destination:   "JFK",
		DepartureDate: testTime.AddDate(0, 0, 7),
		ReturnDate:    sql.NullTime{Time: testTime.AddDate(0, 0, 14), Valid: true},
		Status:        "COMPLETED",
		CreatedAt:     testTime,
	}

	mockOfferRows := new(mocks.MockRows)
	mockOffers := []db.FlightOffer{
		{
			ID:               offerID,
			Price:            123.45,
			Currency:         "USD",
			AirlineCodes:     sql.NullString{String: "BA,AA", Valid: true},
			OutboundDuration: sql.NullInt64{Int64: 36000, Valid: true},
			OutboundStops:    sql.NullInt64{Int64: 1, Valid: true},
			ReturnDuration:   sql.NullInt64{Int64: 34000, Valid: true},
			ReturnStops:      sql.NullInt64{Int64: 1, Valid: true},
			CreatedAt:        testTime,
		},
	}
	mockOfferRows.On("Next").Return(true).Once()
	mockOfferRows.On("Next").Return(false)
	mockOfferRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*(args.Get(0).(*int)) = mockOffers[0].ID
		*(args.Get(1).(*float64)) = mockOffers[0].Price
		*(args.Get(2).(*string)) = mockOffers[0].Currency
		*(args.Get(3).(*sql.NullString)) = mockOffers[0].AirlineCodes
		*(args.Get(4).(*sql.NullInt64)) = mockOffers[0].OutboundDuration
		*(args.Get(5).(*sql.NullInt64)) = mockOffers[0].OutboundStops
		*(args.Get(6).(*sql.NullInt64)) = mockOffers[0].ReturnDuration
		*(args.Get(7).(*sql.NullInt64)) = mockOffers[0].ReturnStops
		*(args.Get(8).(*time.Time)) = mockOffers[0].CreatedAt
	})
	mockOfferRows.On("Close").Return(nil)
	mockOfferRows.On("Err").Return(nil)

	mockSegmentRows := new(mocks.MockRows)
	mockSegments := []db.FlightSegment{
		{AirlineCode: "BA", FlightNumber: "123", DepartureAirport: "LHR", ArrivalAirport: "JFK", DepartureTime: testTime.AddDate(0, 0, 7), ArrivalTime: testTime.AddDate(0, 0, 7).Add(8 * time.Hour), Duration: 28800, Airplane: "747", Legroom: "32", IsReturn: false},
		{AirlineCode: "BA", FlightNumber: "456", DepartureAirport: "JFK", ArrivalAirport: "LHR", DepartureTime: testTime.AddDate(0, 0, 14), ArrivalTime: testTime.AddDate(0, 0, 14).Add(7 * time.Hour), Duration: 25200, Airplane: "777", Legroom: "31", IsReturn: true},
	}
	mockSegmentRows.On("Next").Return(true).Times(len(mockSegments))
	mockSegmentRows.On("Next").Return(false)
	mockSegmentRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*(args.Get(0).(*string)) = mockSegments[0].AirlineCode
		*(args.Get(1).(*string)) = mockSegments[0].FlightNumber
		*(args.Get(2).(*string)) = mockSegments[0].DepartureAirport
		*(args.Get(3).(*string)) = mockSegments[0].ArrivalAirport
		*(args.Get(4).(*time.Time)) = mockSegments[0].DepartureTime
		*(args.Get(5).(*time.Time)) = mockSegments[0].ArrivalTime
		*(args.Get(6).(*int)) = mockSegments[0].Duration
		*(args.Get(7).(*string)) = mockSegments[0].Airplane
		*(args.Get(8).(*string)) = mockSegments[0].Legroom
		*(args.Get(9).(*bool)) = mockSegments[0].IsReturn
	}).Once()
	mockSegmentRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*(args.Get(0).(*string)) = mockSegments[1].AirlineCode
		*(args.Get(1).(*string)) = mockSegments[1].FlightNumber
		*(args.Get(2).(*string)) = mockSegments[1].DepartureAirport
		*(args.Get(3).(*string)) = mockSegments[1].ArrivalAirport
		*(args.Get(4).(*time.Time)) = mockSegments[1].DepartureTime
		*(args.Get(5).(*time.Time)) = mockSegments[1].ArrivalTime
		*(args.Get(6).(*int)) = mockSegments[1].Duration
		*(args.Get(7).(*string)) = mockSegments[1].Airplane
		*(args.Get(8).(*string)) = mockSegments[1].Legroom
		*(args.Get(9).(*bool)) = mockSegments[1].IsReturn
	}).Once()
	mockSegmentRows.On("Close").Return(nil)
	mockSegmentRows.On("Err").Return(nil)

	// Configure mock DB
	mockDB.On("GetSearchQueryByID", mock.Anything, searchID).Return(mockQuery, nil)
	mockDB.On("GetFlightOffersBySearchID", mock.Anything, searchID).Return(mockOfferRows, nil)
	mockDB.On("GetFlightSegmentsByOfferID", mock.Anything, offerID).Return(mockSegmentRows, nil)

	// Act
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/search/%d", searchID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req) // Use router to serve the request

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	// Add more detailed assertions on the response body if needed
	mockDB.AssertExpectations(t)
	mockOfferRows.AssertExpectations(t)
	mockSegmentRows.AssertExpectations(t)
}

func TestGetSearchByID_InvalidID(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockPostgresDB) // DB not actually called
	router := setupRouter()
	// Register the handler with the router
	router.GET("/search/:id", api.GetSearchByID(mockDB))

	// Act
	req, _ := http.NewRequest(http.MethodGet, "/search/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req) // Use router to serve the request

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	// Check error message if needed
}

func TestGetSearchByID_QueryNotFound(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockPostgresDB)
	router := setupRouter()
	// Register the handler with the router
	router.GET("/search/:id", api.GetSearchByID(mockDB))
	searchID := 404

	// Mock DB call to return not found error
	mockDB.On("GetSearchQueryByID", mock.Anything, searchID).Return(nil, fmt.Errorf("search query with ID %d not found", searchID)) // Simulate not found

	// Act
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/search/%d", searchID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req) // Use router to serve the request

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockDB.AssertExpectations(t)
}

func TestGetSearchByID_OfferQueryError(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockPostgresDB)
	router := setupRouter()
	// Register the handler with the router
	router.GET("/search/:id", api.GetSearchByID(mockDB))
	searchID := 123
	mockQuery := &db.SearchQuery{ID: searchID, Origin: "LHR", Destination: "JFK"} // Minimal query

	// Mock successful query fetch
	mockDB.On("GetSearchQueryByID", mock.Anything, searchID).Return(mockQuery, nil)
	// Mock offer fetch to return error
	mockDB.On("GetFlightOffersBySearchID", mock.Anything, searchID).Return(nil, assert.AnError)

	// Act
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/search/%d", searchID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req) // Use router to serve the request

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}

// TODO: Add test case for error during segment query
// TODO: Add test case for error during offer row scanning
// TODO: Add test case for error during segment query
// TODO: Add test case for error during offer row scanning
// TODO: Add test case for error during segment row scanning

func TestListSearches_Success(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockPostgresDB)
	router := setupRouter()
	// Register the handler with the router
	router.GET("/search", api.ListSearches(mockDB))

	page := 1
	perPage := 2
	offset := (page - 1) * perPage
	totalCount := 5
	testTime := time.Now()

	mockQueries := []db.SearchQuery{
		{ID: 1, Origin: "LHR", Destination: "JFK", DepartureDate: testTime.AddDate(0, 0, 7), Status: "COMPLETED", CreatedAt: testTime},
		{ID: 2, Origin: "CDG", Destination: "SFO", DepartureDate: testTime.AddDate(0, 0, 8), ReturnDate: sql.NullTime{Time: testTime.AddDate(0, 0, 15), Valid: true}, Status: "PENDING", CreatedAt: testTime.Add(-time.Hour)},
	}

	mockRows := new(mocks.MockRows)
	mockRows.On("Next").Return(true).Times(len(mockQueries))
	mockRows.On("Next").Return(false)
	// Mock Scan for the first row
	mockRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*(args.Get(0).(*int)) = mockQueries[0].ID
		*(args.Get(1).(*string)) = mockQueries[0].Origin
		*(args.Get(2).(*string)) = mockQueries[0].Destination
		*(args.Get(3).(*time.Time)) = mockQueries[0].DepartureDate
		*(args.Get(4).(*sql.NullTime)) = mockQueries[0].ReturnDate
		*(args.Get(5).(*string)) = mockQueries[0].Status
		*(args.Get(6).(*time.Time)) = mockQueries[0].CreatedAt
	}).Once()
	// Mock Scan for the second row
	mockRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*(args.Get(0).(*int)) = mockQueries[1].ID
		*(args.Get(1).(*string)) = mockQueries[1].Origin
		*(args.Get(2).(*string)) = mockQueries[1].Destination
		*(args.Get(3).(*time.Time)) = mockQueries[1].DepartureDate
		*(args.Get(4).(*sql.NullTime)) = mockQueries[1].ReturnDate
		*(args.Get(5).(*string)) = mockQueries[1].Status
		*(args.Get(6).(*time.Time)) = mockQueries[1].CreatedAt
	}).Once()
	mockRows.On("Close").Return(nil)
	mockRows.On("Err").Return(nil)

	// Configure mock DB
	mockDB.On("CountSearches", mock.Anything).Return(totalCount, nil)
	mockDB.On("QuerySearchesPaginated", mock.Anything, perPage, offset).Return(mockRows, nil)

	// Act
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/search?page=%d&per_page=%d", page, perPage), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req) // Use router to serve the request

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(totalCount), response["total"]) // JSON numbers are float64
	assert.Equal(t, float64(page), response["page"])
	assert.Equal(t, float64(perPage), response["per_page"])
	assert.Len(t, response["data"], len(mockQueries))
	// Add more detailed checks on the data if needed

	mockDB.AssertExpectations(t)
	mockRows.AssertExpectations(t)
}

func TestListSearches_CountError(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockPostgresDB)
	router := setupRouter()
	// Register the handler with the router
	router.GET("/search", api.ListSearches(mockDB))

	// Configure mock DB to return error on count
	mockDB.On("CountSearches", mock.Anything).Return(0, assert.AnError)

	// Act
	req, _ := http.NewRequest(http.MethodGet, "/search", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req) // Use router to serve the request

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}

func TestListSearches_QueryError(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockPostgresDB)
	router := setupRouter()
	// Register the handler with the router
	router.GET("/search", api.ListSearches(mockDB))

	page := 1
	perPage := 10
	offset := (page - 1) * perPage
	totalCount := 5

	// Configure mock DB
	mockDB.On("CountSearches", mock.Anything).Return(totalCount, nil)
	mockDB.On("QuerySearchesPaginated", mock.Anything, perPage, offset).Return(nil, assert.AnError)

	// Act
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/search?page=%d&per_page=%d", page, perPage), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req) // Use router to serve the request

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}

func TestListSearches_ScanError(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockPostgresDB)
	router := setupRouter()
	// Register the handler with the router
	router.GET("/search", api.ListSearches(mockDB))

	page := 1
	perPage := 10
	offset := (page - 1) * perPage
	totalCount := 1

	mockRows := new(mocks.MockRows)
	mockRows.On("Next").Return(true).Once()                                                                                                             // Simulate one row
	mockRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError) // Error during scan
	mockRows.On("Close").Return(nil)                                                                                                                    // Expect Close to be called even on Scan error due to defer

	// Configure mock DB
	mockDB.On("CountSearches", mock.Anything).Return(totalCount, nil)
	mockDB.On("QuerySearchesPaginated", mock.Anything, perPage, offset).Return(mockRows, nil)

	// Act
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/search?page=%d&per_page=%d", page, perPage), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req) // Use router to serve the request

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
	mockRows.AssertExpectations(t)
}

func TestDeleteJob_Success(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockPostgresDB)
	router := setupRouter()
	mockTx := new(mocks.MockTx)
	// Register the handler with the router
	router.DELETE("/jobs/:id", api.DeleteJob(mockDB, nil)) // Assuming workerManager is not needed

	jobID := 123

	// Configure mock DB and Tx
	mockDB.On("BeginTx", mock.Anything).Return(mockTx, nil)
	mockDB.On("DeleteJobDetailsByJobID", mock.Anything, mockTx, jobID).Return(nil)
	// Correct mock for DeleteScheduledJobByID: Expects ctx, Tx and int, returns int64 and error
	mockDB.On("DeleteScheduledJobByID", mock.Anything, mock.AnythingOfType("*mocks.MockTx"), jobID).Return(int64(1), nil)
	mockTx.On("Commit").Return(nil)
	mockTx.On("Rollback").Return(nil) // Should not be called ideally, but good practice

	// Act
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/jobs/%d", jobID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req) // Use router to serve the request

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Job deleted successfully", response["message"])

	mockDB.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestDeleteJob_InvalidID(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockPostgresDB) // DB not called
	router := setupRouter()
	// Register the handler with the router
	router.DELETE("/jobs/:id", api.DeleteJob(mockDB, nil))

	// Act
	req, _ := http.NewRequest(http.MethodDelete, "/jobs/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req) // Use router to serve the request

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteJob_BeginTxError(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockPostgresDB)
	router := setupRouter()
	// Register the handler with the router
	router.DELETE("/jobs/:id", api.DeleteJob(mockDB, nil))

	jobID := 123

	// Configure mock DB to return error on BeginTx
	mockDB.On("BeginTx", mock.Anything).Return(nil, assert.AnError)

	// Act
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/jobs/%d", jobID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req) // Use router to serve the request

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
}

func TestDeleteJob_DeleteDetailsError(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockPostgresDB)
	router := setupRouter()
	mockTx := new(mocks.MockTx)
	// Register the handler with the router
	router.DELETE("/jobs/:id", api.DeleteJob(mockDB, nil))

	jobID := 123

	// Configure mock DB and Tx
	mockDB.On("BeginTx", mock.Anything).Return(mockTx, nil)
	mockDB.On("DeleteJobDetailsByJobID", mock.Anything, mockTx, jobID).Return(assert.AnError) // Error here
	mockTx.On("Rollback").Return(nil)                                                         // Expect rollback

	// Act
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/jobs/%d", jobID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req) // Use router to serve the request

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestDeleteJob_DeleteJobError(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockPostgresDB)
	router := setupRouter()
	mockTx := new(mocks.MockTx)
	// Register the handler with the router
	router.DELETE("/jobs/:id", api.DeleteJob(mockDB, nil))

	jobID := 123

	// Configure mock DB and Tx
	mockDB.On("BeginTx", mock.Anything).Return(mockTx, nil)
	mockDB.On("DeleteJobDetailsByJobID", mock.Anything, mockTx, jobID).Return(nil)
	// Correct mock for DeleteScheduledJobByID
	mockDB.On("DeleteScheduledJobByID", mock.Anything, mock.AnythingOfType("*mocks.MockTx"), jobID).Return(int64(0), assert.AnError)
	mockTx.On("Rollback").Return(nil) // Expect rollback

	// Act
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/jobs/%d", jobID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req) // Use router to serve the request

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestDeleteJob_NotFound(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockPostgresDB)
	router := setupRouter()
	mockTx := new(mocks.MockTx)
	// Register the handler with the router
	router.DELETE("/jobs/:id", api.DeleteJob(mockDB, nil))

	jobID := 404 // Use a different ID to simulate not found

	// Configure mock DB and Tx
	mockDB.On("BeginTx", mock.Anything).Return(mockTx, nil)
	mockDB.On("DeleteJobDetailsByJobID", mock.Anything, mockTx, jobID).Return(nil) // Assume details might exist or delete does nothing if not found
	// Correct mock for DeleteScheduledJobByID
	mockDB.On("DeleteScheduledJobByID", mock.Anything, mock.AnythingOfType("*mocks.MockTx"), jobID).Return(int64(0), nil)
	mockTx.On("Rollback").Return(nil) // Expect rollback because job not found

	// Act
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/jobs/%d", jobID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req) // Use router to serve the request

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockDB.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestDeleteJob_CommitError(t *testing.T) {
	// Arrange
	mockDB := new(mocks.MockPostgresDB)
	router := setupRouter()
	mockTx := new(mocks.MockTx)
	// Register the handler with the router
	router.DELETE("/jobs/:id", api.DeleteJob(mockDB, nil))

	jobID := 123

	// Configure mock DB and Tx
	mockDB.On("BeginTx", mock.Anything).Return(mockTx, nil)
	mockDB.On("DeleteJobDetailsByJobID", mock.Anything, mockTx, jobID).Return(nil)
	// Correct mock for DeleteScheduledJobByID
	mockDB.On("DeleteScheduledJobByID", mock.Anything, mock.AnythingOfType("*mocks.MockTx"), jobID).Return(int64(1), nil)
	mockTx.On("Commit").Return(assert.AnError) // Error on commit
	mockTx.On("Rollback").Return(nil)          // Should not be called if commit fails, but mock anyway

	// Act
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/jobs/%d", jobID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req) // Use router to serve the request

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockDB.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}
