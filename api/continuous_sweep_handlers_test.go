package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/gilby125/google-flights-api/db"
	"github.com/gilby125/google-flights-api/test/mocks"
	"github.com/gilby125/google-flights-api/worker"
)

func TestUpdateContinuousSweepConfig_TripLengths_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalAirports := db.Top100Airports
	db.Top100Airports = []db.TopAirport{
		{Code: "AAA", Country: "US"},
		{Code: "BBB", Country: "FR"},
		{Code: "CCC", Country: "US"},
	}
	t.Cleanup(func() { db.Top100Airports = originalAirports })

	mockDB := new(mocks.MockPostgresDB)
	mockQueue := new(mocks.MockQueue)
	workerManager := newWorkerManagerForTests(mockQueue, mockDB)

	runner := worker.NewContinuousSweepRunner(mockDB, mockQueue, nil, worker.DefaultContinuousSweepConfig())
	workerManager.SetSweepRunner(runner)

	router := gin.New()
	router.PUT("/admin/continuous-sweep/config", updateContinuousSweepConfig(workerManager))

	reqBody := map[string]any{
		"trip_lengths": []int{7, 3, 3, 5},
	}
	body, _ := json.Marshal(reqBody)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/admin/continuous-sweep/config", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	cfg := runner.GetConfig()
	assert.Equal(t, []int{3, 5, 7}, cfg.TripLengths)

	var resp struct {
		Message string                 `json:"message"`
		Status  db.SweepStatusResponse `json:"status"`
	}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Sweep configuration updated", resp.Message)
	assert.Equal(t, []int{3, 5, 7}, resp.Status.TripLengths)
}

func TestUpdateContinuousSweepConfig_TripLengths_Invalid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalAirports := db.Top100Airports
	db.Top100Airports = []db.TopAirport{
		{Code: "AAA", Country: "US"},
		{Code: "BBB", Country: "FR"},
	}
	t.Cleanup(func() { db.Top100Airports = originalAirports })

	mockDB := new(mocks.MockPostgresDB)
	mockQueue := new(mocks.MockQueue)
	workerManager := newWorkerManagerForTests(mockQueue, mockDB)

	runner := worker.NewContinuousSweepRunner(mockDB, mockQueue, nil, worker.DefaultContinuousSweepConfig())
	workerManager.SetSweepRunner(runner)

	router := gin.New()
	router.PUT("/admin/continuous-sweep/config", updateContinuousSweepConfig(workerManager))

	reqBody := map[string]any{
		"trip_lengths": []int{0, 7},
	}
	body, _ := json.Marshal(reqBody)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/admin/continuous-sweep/config", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	cfg := runner.GetConfig()
	assert.Equal(t, []int{7, 14}, cfg.TripLengths)
}
