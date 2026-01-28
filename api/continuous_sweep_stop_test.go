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

	router := gin.New()
	router.POST("/admin/continuous-sweep/stop", stopContinuousSweep(nil, mockDB))

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/admin/continuous-sweep/stop", nil)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	mockDB.AssertExpectations(t)
	mockDB.AssertNotCalled(t, "SaveContinuousSweepProgress", mock.Anything, mock.Anything)
}
