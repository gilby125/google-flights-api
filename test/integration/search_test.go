package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest" // Need httptest for recorder
	"testing"
	"time"

	"github.com/gilby125/google-flights-api/api"
	"github.com/gilby125/google-flights-api/test/mocks" // Need mocks

	// Removed unused comment
	"github.com/gin-gonic/gin" // Need gin
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock" // Need mock
)

func dateOnly(t time.Time) api.DateOnly {
	return api.DateOnly{Time: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)}
}

func TestCreateSearch(t *testing.T) {
	// Using local router and mock queue for this test suite

	// Test cases
	tests := []struct {
		name         string
		payload      interface{}
		mockSetup    func(mq *mocks.Queue, payload interface{}) // Function to set mock expectations
		expectedCode int
	}{
		{
			name: "valid one-way search",
			payload: api.SearchRequest{
				Origin:        "JFK",
				Destination:   "LAX",
				DepartureDate: dateOnly(time.Now().AddDate(0, 0, 7)),
				Adults:        1,
				TripType:      "one_way",
				Class:         "economy",
				Stops:         "nonstop",
				Currency:      "USD",
			},
			mockSetup: func(mq *mocks.Queue, payload interface{}) {
				mq.On("Enqueue", mock.Anything, "flight_search", mock.AnythingOfType("worker.FlightSearchPayload")).Return("job-id-1", nil).Once()
			},
			expectedCode: http.StatusAccepted,
		},
		{
			name: "valid round-trip search",
			payload: api.SearchRequest{
				Origin:        "JFK",
				Destination:   "LAX",
				DepartureDate: dateOnly(time.Now().AddDate(0, 0, 7)),
				ReturnDate:    dateOnly(time.Now().AddDate(0, 0, 14)),
				Adults:        2,
				TripType:      "round_trip",
				Class:         "premium_economy",
				Stops:         "one_stop",
				Currency:      "USD",
			},
			mockSetup: func(mq *mocks.Queue, payload interface{}) {
				mq.On("Enqueue", mock.Anything, "flight_search", mock.AnythingOfType("worker.FlightSearchPayload")).Return("job-id-2", nil).Once()
			},
			expectedCode: http.StatusAccepted,
		},
		{
			name: "valid stops two_stops_plus", // Fixed validation tag in handler
			payload: api.SearchRequest{
				Origin:        "LHR",
				Destination:   "SYD",
				DepartureDate: dateOnly(time.Now().AddDate(0, 1, 0)),
				Adults:        1,
				TripType:      "one_way",
				Class:         "business",
				Stops:         "two_stops_plus",
				Currency:      "GBP",
			},
			mockSetup: func(mq *mocks.Queue, payload interface{}) {
				mq.On("Enqueue", mock.Anything, "flight_search", mock.AnythingOfType("worker.FlightSearchPayload")).Return("job-id-3", nil).Once()
			},
			expectedCode: http.StatusAccepted,
		},
		{
			name: "missing required fields (origin)",
			payload: map[string]interface{}{
				"destination":    "LAX",
				"departure_date": time.Now().AddDate(0, 0, 7).Format(time.RFC3339),
				"adults":         1,
				"trip_type":      "one_way",
				"class":          "economy",
				"stops":          "any",
				"currency":       "USD",
			},
			mockSetup:    func(mq *mocks.Queue, payload interface{}) { /* No enqueue expected */ },
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "invalid departure date (past)",
			payload: api.SearchRequest{
				Origin:        "JFK",
				Destination:   "LAX",
				DepartureDate: dateOnly(time.Now().AddDate(0, 0, -1)), // Past date should fail validation
				Adults:        1,
				TripType:      "one_way",
				Class:         "economy",
				Stops:         "nonstop",
				Currency:      "USD",
			},
			mockSetup:    func(mq *mocks.Queue, payload interface{}) { /* No enqueue expected */ },
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "invalid return date (before departure)",
			payload: api.SearchRequest{
				Origin:        "JFK",
				Destination:   "LAX",
				DepartureDate: dateOnly(time.Now().AddDate(0, 0, 14)),
				ReturnDate:    dateOnly(time.Now().AddDate(0, 0, 7)), // Before departure should fail validation
				Adults:        1,
				TripType:      "round_trip",
				Class:         "economy",
				Stops:         "nonstop",
				Currency:      "USD",
			},
			mockSetup:    func(mq *mocks.Queue, payload interface{}) { /* No enqueue expected */ },
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "unsupported trip type",
			payload: api.SearchRequest{
				Origin:        "JFK",
				Destination:   "LAX",
				DepartureDate: dateOnly(time.Now().AddDate(0, 0, 7)),
				Adults:        1,
				TripType:      "multi_city", // Invalid
				Class:         "economy",
				Stops:         "nonstop",
				Currency:      "USD",
			},
			mockSetup:    func(mq *mocks.Queue, payload interface{}) { /* No enqueue expected */ },
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "invalid currency",
			payload: api.SearchRequest{
				Origin:        "JFK",
				Destination:   "LAX",
				DepartureDate: dateOnly(time.Now().AddDate(0, 0, 7)),
				Adults:        1,
				TripType:      "one_way",
				Class:         "economy",
				Stops:         "nonstop",
				Currency:      "INVALID", // Invalid format should fail validation
			},
			mockSetup:    func(mq *mocks.Queue, payload interface{}) { /* No enqueue expected */ },
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "invalid cabin class",
			payload: api.SearchRequest{
				Origin:        "JFK",
				Destination:   "LAX",
				DepartureDate: dateOnly(time.Now().AddDate(0, 0, 7)),
				Adults:        1,
				TripType:      "one_way",
				Class:         "invalid_class", // Invalid
				Stops:         "nonstop",
				Currency:      "USD",
			},
			mockSetup:    func(mq *mocks.Queue, payload interface{}) { /* No enqueue expected */ },
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "invalid stops value",
			payload: api.SearchRequest{
				Origin:        "JFK",
				Destination:   "LAX",
				DepartureDate: dateOnly(time.Now().AddDate(0, 0, 7)),
				Adults:        1,
				TripType:      "one_way",
				Class:         "economy",
				Stops:         "five_stops", // Invalid
				Currency:      "USD",
			},
			mockSetup:    func(mq *mocks.Queue, payload interface{}) { /* No enqueue expected */ },
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "invalid origin airport code format",
			payload: api.SearchRequest{
				Origin:        "123", // Invalid format should fail validation
				Destination:   "LAX",
				DepartureDate: dateOnly(time.Now().AddDate(0, 0, 7)),
				Adults:        1,
				TripType:      "one_way",
				Class:         "economy",
				Stops:         "nonstop",
				Currency:      "USD",
			},
			mockSetup:    func(mq *mocks.Queue, payload interface{}) { /* No enqueue expected */ },
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "invalid destination airport code format (too long)",
			payload: api.SearchRequest{
				Origin:        "JFK",
				Destination:   "TOOLONG", // Invalid format should fail validation
				DepartureDate: dateOnly(time.Now().AddDate(0, 0, 7)),
				Adults:        1,
				TripType:      "one_way",
				Class:         "economy",
				Stops:         "nonstop",
				Currency:      "USD",
			},
			mockSetup:    func(mq *mocks.Queue, payload interface{}) { /* No enqueue expected */ },
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "zero adults",
			payload: api.SearchRequest{
				Origin:        "JFK",
				Destination:   "LAX",
				DepartureDate: dateOnly(time.Now().AddDate(0, 0, 7)),
				Adults:        0, // Invalid (min=1)
				TripType:      "one_way",
				Class:         "economy",
				Stops:         "nonstop",
				Currency:      "USD",
			},
			mockSetup:    func(mq *mocks.Queue, payload interface{}) { /* No enqueue expected */ },
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "negative adults",
			payload: api.SearchRequest{
				Origin:        "JFK",
				Destination:   "LAX",
				DepartureDate: dateOnly(time.Now().AddDate(0, 0, 7)),
				Adults:        -1, // Invalid (min=1)
				TripType:      "one_way",
				Class:         "economy",
				Stops:         "nonstop",
				Currency:      "USD",
			},
			mockSetup:    func(mq *mocks.Queue, payload interface{}) { /* No enqueue expected */ },
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock for this specific test case
			mockQueue := new(mocks.Queue)
			if tt.mockSetup != nil {
				tt.mockSetup(mockQueue, tt.payload) // Pass payload argument
			}

			// Marshal payload to JSON
			jsonPayload, err := json.Marshal(tt.payload)
			assert.NoError(t, err)

			// Setup router and handler for this test case
			router := gin.Default()                                    // Use local router with default middleware
			router.POST("/api/v1/search", api.CreateSearch(mockQueue)) // Pass the test-specific mock

			// Create request
			req, err := http.NewRequest("POST", "/api/v1/search", bytes.NewBuffer(jsonPayload))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Record response
			w := httptest.NewRecorder() // Use recorder
			router.ServeHTTP(w, req)    // Use the test-specific router

			// Use recorder's results
			wCode := w.Code
			wBody := w.Body

			// Assert status code
			assert.Equal(t, tt.expectedCode, wCode) // Assert recorder code

			if tt.expectedCode == http.StatusAccepted {
				// Parse response
				var resp map[string]interface{}
				err = json.Unmarshal(wBody.Bytes(), &resp) // Decode from recorder body bytes
				assert.NoError(t, err)

				// Assert response contains job_id and message
				assert.Contains(t, resp, "job_id")
				assert.NotEmpty(t, resp["job_id"])
				assert.Contains(t, resp, "message")
				assert.Equal(t, "Flight search job created successfully", resp["message"])

				// Verify mock expectations for this case
				mockQueue.AssertExpectations(t) // Assert mock expectations
			} else {
				// Optional: Verify error message format for bad requests
				var errResp map[string]string
				err = json.Unmarshal(wBody.Bytes(), &errResp) // Decode from recorder body bytes
				// Allow EOF error for empty body which might happen on some validation errors
				if err != nil && err.Error() != "EOF" {
					assert.NoError(t, err, "Error response should be valid JSON or empty")
				}
				if err == nil {
					assert.Contains(t, errResp, "error", "Error response should contain an error key")
					assert.NotEmpty(t, errResp["error"], "Error message should not be empty")
				}

				// Verify no mocks were called unexpectedly
				mockQueue.AssertExpectations(t) // Assert mock expectations (should be none called)
			}
		})
	}
}
