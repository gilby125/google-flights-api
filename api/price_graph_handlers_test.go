package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gilby125/google-flights-api/config"
	"github.com/gilby125/google-flights-api/db"
	"github.com/gilby125/google-flights-api/test/mocks"
	"github.com/gilby125/google-flights-api/worker"
)

func newWorkerManagerForTests(queue *mocks.MockQueue, postgres *mocks.MockPostgresDB) *worker.Manager {
	cfg := config.WorkerConfig{Concurrency: 1}
	return worker.NewManager(queue, nil, postgres, nil, cfg, config.FlightConfig{})
}

func TestEnqueuePriceGraphSweepHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockPostgresDB)
	mockQueue := new(mocks.MockQueue)
	workerManager := newWorkerManagerForTests(mockQueue, mockDB)

	router := gin.New()
	router.POST("/admin/price-graph-sweeps", enqueuePriceGraphSweep(mockDB, workerManager))

	reqBody := PriceGraphSweepRequest{
		Origins:           []string{"JFK"},
		Destinations:      []string{"LAX"},
		DepartureDateFrom: DateOnly{Time: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)},
		DepartureDateTo:   DateOnly{Time: time.Date(2025, 2, 10, 0, 0, 0, 0, time.UTC)},
		TripLengths:       []int{5, 7},
		TripType:          "round_trip",
		Class:             "business",
		Stops:             "one_stop",
		Adults:            2,
		Children:          1,
		Currency:          "eur",
	}

	expectedSweepID := 101

	mockDB.On("CreatePriceGraphSweep", mock.Anything, sql.NullInt32{}, len(reqBody.Origins), len(reqBody.Destinations),
		sql.NullInt32{Int32: 5, Valid: true}, sql.NullInt32{Int32: 7, Valid: true}, "EUR").
		Return(expectedSweepID, nil).Once()

	mockQueue.On("Enqueue", mock.Anything, "price_graph_sweep", mock.MatchedBy(func(payload interface{}) bool {
		p, ok := payload.(worker.PriceGraphSweepPayload)
		if !ok {
			return false
		}
		return p.SweepID == expectedSweepID && p.Currency == "EUR" && p.RateLimitMillis == 750
	})).Return("job-id", nil).Once()

	summary := &db.PriceGraphSweep{
		ID:               expectedSweepID,
		Status:           "queued",
		OriginCount:      len(reqBody.Origins),
		DestinationCount: len(reqBody.Destinations),
		Currency:         "EUR",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	mockDB.On("GetPriceGraphSweepByID", mock.Anything, expectedSweepID).Return(summary, nil).Once()

	body, _ := json.Marshal(reqBody)
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/admin/price-graph-sweeps", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusAccepted, rec.Code)

	var resp struct {
		Message string                 `json:"message"`
		SweepID int                    `json:"sweep_id"`
		Sweep   map[string]interface{} `json:"sweep"`
	}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Price graph sweep enqueued", resp.Message)
	assert.Equal(t, expectedSweepID, resp.SweepID)
	assert.Equal(t, "queued", resp.Sweep["status"])

	mockDB.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}

func TestEnqueuePriceGraphSweepHandler_InvalidDates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockPostgresDB)
	mockQueue := new(mocks.MockQueue)
	workerManager := newWorkerManagerForTests(mockQueue, mockDB)

	router := gin.New()
	router.POST("/admin/price-graph-sweeps", enqueuePriceGraphSweep(mockDB, workerManager))

	reqBody := PriceGraphSweepRequest{
		Origins:           []string{"JFK"},
		Destinations:      []string{"LAX"},
		DepartureDateFrom: DateOnly{Time: time.Date(2025, 2, 10, 0, 0, 0, 0, time.UTC)},
		DepartureDateTo:   DateOnly{Time: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)},
		TripType:          "round_trip",
		Class:             "economy",
		Stops:             "nonstop",
		Adults:            1,
		Currency:          "usd",
	}

	body, _ := json.Marshal(reqBody)
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/admin/price-graph-sweeps", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "departure_date_from must be before")

	mockDB.AssertNotCalled(t, "CreatePriceGraphSweep", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mockQueue.AssertNotCalled(t, "Enqueue", mock.Anything, mock.Anything, mock.Anything)
}

func TestEnqueuePriceGraphSweepHandler_SchedulerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockPostgresDB)
	mockQueue := new(mocks.MockQueue)
	workerManager := newWorkerManagerForTests(mockQueue, mockDB)

	router := gin.New()
	router.POST("/admin/price-graph-sweeps", enqueuePriceGraphSweep(mockDB, workerManager))

	reqBody := PriceGraphSweepRequest{
		Origins:           []string{"JFK"},
		Destinations:      []string{"LAX"},
		DepartureDateFrom: DateOnly{Time: time.Now()},
		DepartureDateTo:   DateOnly{Time: time.Now().AddDate(0, 0, 3)},
		TripType:          "round_trip",
		Class:             "economy",
		Stops:             "nonstop",
		Adults:            1,
		Currency:          "usd",
	}

	expectedSweepID := 7
	mockDB.On("CreatePriceGraphSweep", mock.Anything, sql.NullInt32{}, 1, 1,
		sql.NullInt32{Int32: 0, Valid: true}, sql.NullInt32{Int32: 0, Valid: true}, "USD").
		Return(expectedSweepID, nil).Once()

	queueErr := errors.New("queue unavailable")
	mockQueue.On("Enqueue", mock.Anything, "price_graph_sweep", mock.Anything).Return("", queueErr).Once()
	mockDB.On("UpdatePriceGraphSweepStatus", mock.Anything, expectedSweepID, "failed", sql.NullTime{}, sql.NullTime{}, 1).
		Return(nil).Once()

	body, _ := json.Marshal(reqBody)
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/admin/price-graph-sweeps", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.Background())

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), queueErr.Error())

	mockDB.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}
