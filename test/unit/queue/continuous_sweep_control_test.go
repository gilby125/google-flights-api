package queue_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gilby125/google-flights-api/config"
	"github.com/gilby125/google-flights-api/queue"
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

func TestRedisQueue_ContinuousSweepControlFlags(t *testing.T) {
	_, q := newTestRedisQueue(t)
	ctx := context.Background()

	ctrl, err := q.GetContinuousSweepControlFlags(ctx)
	require.NoError(t, err)
	require.Nil(t, ctrl)

	running := true
	paused := false
	ctrl, err = q.SetContinuousSweepControlFlags(ctx, &running, &paused)
	require.NoError(t, err)
	require.NotNil(t, ctrl)
	require.True(t, ctrl.IsRunning)
	require.False(t, ctrl.IsPaused)
	require.False(t, ctrl.LastUpdated.IsZero())
	require.Equal(t, "redis", ctrl.Source)

	ctrl, err = q.GetContinuousSweepControlFlags(ctx)
	require.NoError(t, err)
	require.NotNil(t, ctrl)
	require.True(t, ctrl.IsRunning)
	require.False(t, ctrl.IsPaused)

	paused = true
	ctrl, err = q.SetContinuousSweepControlFlags(ctx, nil, &paused)
	require.NoError(t, err)
	require.NotNil(t, ctrl)
	require.True(t, ctrl.IsRunning, "is_running should be sticky when only updating paused")
	require.True(t, ctrl.IsPaused)

	_, err = q.SetContinuousSweepControlFlags(ctx, nil, nil)
	require.Error(t, err)
}
