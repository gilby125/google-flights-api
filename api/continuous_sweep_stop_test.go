package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gilby125/google-flights-api/db"
	"github.com/gilby125/google-flights-api/test/mocks"
)

func TestStopContinuousSweep_UsesControlFlagsUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(mocks.MockPostgresDB)
	mockQueue := new(mocks.MockQueue)
	workerManager := newWorkerManagerForTests(mockQueue, mockDB)

	mockDB.
		On(
			"SetContinuousSweepControlFlags",
			mock.Anything,
			mock.MatchedBy(func(p *bool) bool { return p != nil && *p == false }),
			mock.MatchedBy(func(p *bool) bool { return p != nil && *p == false }),
		).
		Return(nil).
		Once()

	mockDB.
		On("GetContinuousSweepProgress", mock.Anything).
		Return(&db.ContinuousSweepProgress{ID: 1, IsRunning: false, IsPaused: false}, nil).
		Once()

	mockQueue.On("CancelProcessing", mock.Anything, "continuous_price_graph").Return(int64(2), nil).Once()
	mockQueue.On("ClearQueue", mock.Anything, "continuous_price_graph").Return(int64(0), nil).Once()

	router := gin.New()
	router.POST("/admin/continuous-sweep/stop", stopContinuousSweep(workerManager, mockDB))

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/admin/continuous-sweep/stop", nil)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	mockDB.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
	mockDB.AssertNotCalled(t, "SaveContinuousSweepProgress", mock.Anything, mock.Anything)
}
