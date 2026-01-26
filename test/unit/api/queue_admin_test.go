package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gilby125/google-flights-api/api"
	"github.com/gilby125/google-flights-api/config"
	"github.com/gilby125/google-flights-api/queue"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newTestRedisQueue(t *testing.T) (*miniredis.Miniredis, *queue.RedisQueue) {
	t.Helper()

	mr := miniredis.RunT(t)
	host, port, ok := strings.Cut(mr.Addr(), ":")
	require.True(t, ok)

	q, err := queue.NewRedisQueue(config.RedisConfig{
		Host:                   host,
		Port:                   port,
		Password:               "",
		DB:                     0,
		QueueGroup:             "test_group",
		QueueStreamPrefix:      "test_stream",
		QueueBlockTimeout:      50 * time.Millisecond,
		QueueVisibilityTimeout: 50 * time.Millisecond,
	})
	require.NoError(t, err)

	return mr, q
}

func TestAdminQueueEndpoints_MetricsBacklogAndList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	_, q := newTestRedisQueue(t)

	ctx := queue.WithEnqueueMeta(context.Background(), queue.EnqueueMeta{
		Actor:     "http",
		RequestID: "req-123",
		Method:    "POST",
		Path:      "/api/v1/search",
		RemoteIP:  "1.2.3.4",
		UserAgent: "unit-test",
	})
	_, err := q.Enqueue(ctx, "flight_search", map[string]any{"hello": "world"})
	require.NoError(t, err)

	router := gin.New()
	router.GET("/api/v1/admin/queue/:name/enqueues", api.GetQueueEnqueueMetrics(q))
	router.GET("/api/v1/admin/queue/:name/backlog", api.GetQueueBacklog(q))
	router.GET("/api/v1/admin/queue/:name/jobs", api.ListQueueJobs(q))

	t.Run("enqueues", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/queue/flight_search/enqueues?minutes=5", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var body struct {
			Queue   string           `json:"queue"`
			Minutes int              `json:"minutes"`
			Sources map[string]int64 `json:"sources"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.Equal(t, "flight_search", body.Queue)
		require.GreaterOrEqual(t, body.Minutes, 1)
		require.NotEmpty(t, body.Sources)
	})

	t.Run("backlog", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/queue/flight_search/backlog?limit=10", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var body struct {
			Queue string      `json:"queue"`
			Limit int         `json:"limit"`
			Jobs  []queue.Job `json:"jobs"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.Equal(t, "flight_search", body.Queue)
		require.NotEmpty(t, body.Jobs)
		require.NotNil(t, body.Jobs[0].EnqueueMeta)
		require.Equal(t, "req-123", body.Jobs[0].EnqueueMeta.RequestID)
	})

	t.Run("list pending", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/queue/flight_search/jobs?state=pending&limit=10&offset=0", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var body struct {
			Queue string      `json:"queue"`
			State string      `json:"state"`
			Count int         `json:"count"`
			Jobs  []queue.Job `json:"jobs"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.Equal(t, "flight_search", body.Queue)
		require.Equal(t, "pending", body.State)
		require.Equal(t, 1, body.Count)
		require.NotEmpty(t, body.Jobs)
	})
}

func TestAdminQueueEndpoints_InvalidQueueName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, q := newTestRedisQueue(t)

	router := gin.New()
	router.GET("/api/v1/admin/queue/:name/enqueues", api.GetQueueEnqueueMetrics(q))

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/queue/not_a_queue/enqueues?minutes=5", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}
